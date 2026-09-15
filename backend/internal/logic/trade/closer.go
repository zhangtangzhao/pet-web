package trade

import (
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"pet/backend/internal/logic/notify"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
)

// closeOrder 事务内关单 + 释放商品 + 支付流水置失败
func closeOrder(sc *svc.ServiceContext, o *model.Order, target int, reason string, failPayment bool) error {
	now := time.Now()
	return sc.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Order{}).
			Where("id = ? AND status = ?", o.ID, model.OrderPending).
			Updates(map[string]any{
				"status":        target,
				"canceled_at":   &now,
				"cancel_reason": reason,
				"updated_at":    now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil // 状态已流转（并发/幂等）
		}
		// 回滚锁定的优惠券：未过期回到可用，已过期置为过期
		if o.CouponID > 0 {
			if err := tx.Exec(
				`UPDATE member_coupon SET status = CASE
					WHEN EXISTS (SELECT 1 FROM coupon_template t WHERE t.id = member_coupon.template_id
					             AND t.valid_end IS NOT NULL AND t.valid_end < now())
					THEN ?::smallint ELSE ?::smallint END,
				 order_id = NULL
				 WHERE id = ? AND status = ?`,
				model.CouponExpired, model.CouponUsable, o.CouponID, model.CouponLocked).Error; err != nil {
				return err
			}
		}
		var items []model.OrderItem
		if err := tx.Where("order_id = ?", o.ID).Find(&items).Error; err != nil {
			return err
		}
		for _, it := range items {
			if err := tx.Exec(
				"UPDATE pet_product SET status = ?, updated_at = now() WHERE id = ? AND status = ?",
				model.ProductOnSale, it.ProductID, model.ProductLocked).Error; err != nil {
				return err
			}
		}
		if failPayment {
			return tx.Model(&model.Payment{}).
				Where("order_no = ? AND pay_type = ? AND status = ?",
					o.OrderNo, model.PayTypePurchase, model.PayStatusPending).
				Update("status", model.PayStatusFail).Error
		}
		return nil
	})
}

// CloseExpiredOrders 扫描关闭超时未支付订单
func CloseExpiredOrders(sc *svc.ServiceContext) {
	var orders []model.Order
	if err := sc.DB.Where("status = ? AND expire_at < now()", model.OrderPending).
		Limit(200).Find(&orders).Error; err != nil {
		logx.Errorf("扫描超时订单失败: %v", err)
		return
	}
	for i := range orders {
		if err := closeOrder(sc, &orders[i], model.OrderClosed, "超时未支付，自动关闭", true); err != nil {
			logx.Errorf("关闭订单 %s 失败: %v", orders[i].OrderNo, err)
		} else {
			logx.Infof("订单 %s 超时关闭", orders[i].OrderNo)
			if err := notify.Enqueue(sc, orders[i].MemberID, model.NotifySceneOrder,
				"close:"+orders[i].OrderNo, "订单已关闭", "订单超时未支付已自动关闭", orders[i].OrderNo); err != nil {
				logx.Errorf("关单通知入队失败 orderNo=%s: %v", orders[i].OrderNo, err)
			}
		}
	}
}

// AutoConfirmOrders 扫描超时未确认的已支付订单，自动确认完成（镜像用户侧 ConfirmOrder 的 CAS，无 member 条件）
func AutoConfirmOrders(sc *svc.ServiceContext) {
	days := sc.Config.Trade.AutoConfirmDays
	if days <= 0 {
		return // ≤0 关闭
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	var orders []model.Order
	if err := sc.DB.Where("status = ? AND paid_at < ?", model.OrderPaid, cutoff).
		Limit(200).Find(&orders).Error; err != nil {
		logx.Errorf("扫描待自动确认订单失败: %v", err)
		return
	}
	for i := range orders {
		now := time.Now()
		res := sc.DB.Model(&model.Order{}).
			Where("id = ? AND status = ?", orders[i].ID, model.OrderPaid).
			Updates(map[string]any{
				"status":       model.OrderCompleted,
				"completed_at": &now,
				"updated_at":   now,
			})
		if res.Error != nil {
			logx.Errorf("自动确认订单 %s 失败: %v", orders[i].OrderNo, res.Error)
			continue
		}
		if res.RowsAffected == 0 {
			continue // 状态已流转（并发/幂等）
		}
		logx.Infof("订单 %s 超过 %d 天未确认，自动完成", orders[i].OrderNo, days)
		if err := notify.Enqueue(sc, orders[i].MemberID, model.NotifySceneOrder,
			"confirm:"+orders[i].OrderNo, "订单已确认完成", "订单已自动确认完成", orders[i].OrderNo); err != nil {
			logx.Errorf("自动确认通知入队失败 orderNo=%s: %v", orders[i].OrderNo, err)
		}
	}
}

// StartOrderCloser 启动交易定时任务（每分钟）：启动即跑一轮，之后超时关单 + 自动确认收货
func StartOrderCloser(sc *svc.ServiceContext) {
	CloseExpiredOrders(sc)
	AutoConfirmOrders(sc)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		CloseExpiredOrders(sc)
		AutoConfirmOrders(sc)
	}
}

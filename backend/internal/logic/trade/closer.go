package trade

import (
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

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
		}
	}
}

// StartOrderCloser 启动超时关单定时任务（每分钟）
func StartOrderCloser(sc *svc.ServiceContext) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		CloseExpiredOrders(sc)
	}
}

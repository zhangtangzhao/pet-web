// Package marketing 营销自动化：购物车放弃提醒 + 沉睡用户召回券（notify 唯一键幂等）。
package marketing

import (
	"context"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/notify"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
)

// StartAutomation 启动营销自动化任务：启动即跑一轮，之后每 6 小时一轮
func StartAutomation(ctx context.Context, sc *svc.ServiceContext) {
	runAutomation(sc)
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runAutomation(sc)
		}
	}
}

func runAutomation(sc *svc.ServiceContext) {
	remindAbandonedCarts(sc)
	recallDormantMembers(sc)
}

// remindAbandonedCarts 购物车加入 N 小时未结算 → 每月至多提醒一次
func remindAbandonedCarts(sc *svc.ServiceContext) {
	hours := sc.Config.Automation.CartRemindHours
	if hours <= 0 {
		return
	}
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	var memberIDs []int64
	if err := sc.DB.Model(&model.Cart{}).
		Distinct("member_id").
		Where("updated_at < ?", cutoff).
		Limit(300).Pluck("member_id", &memberIDs).Error; err != nil {
		logx.Errorf("automation: 扫描购物车失败: %v", err)
		return
	}
	month := time.Now().Format("2006-01")
	for _, memberID := range memberIDs {
		if err := notify.Enqueue(sc, memberID, model.NotifySceneCare,
			"cartremind:"+strconv.FormatInt(memberID, 10)+":"+month,
			"购物车提醒", "购物车的宝贝还在等你，尽快结算哦", ""); err != nil {
			logx.Errorf("automation: 购物车提醒入队失败 member=%d: %v", memberID, err)
		}
	}
}

// recallDormantMembers 距最近支付订单超过 N 天（或从未下单）→ 发召回券 + 通知，每月至多一次
func recallDormantMembers(sc *svc.ServiceContext) {
	days := sc.Config.Automation.DormantDays
	if days <= 0 {
		return
	}
	month := time.Now().Format("2006-01")
	var memberIDs []int64
	if err := sc.DB.Raw(`
		SELECT m.id FROM member m
		WHERE m.status = 1 AND m.blacklist = 0 AND m.created_at < now() - INTERVAL '7 days'
		  AND NOT EXISTS (SELECT 1 FROM orders o WHERE o.member_id = m.id AND o.status IN (20, 30)
		                  AND o.paid_at > now() - MAKE_INTERVAL(days => ?))
		  AND NOT EXISTS (SELECT 1 FROM notification n WHERE n.member_id = m.id AND n.biz_key = ?)
		LIMIT 200`,
		days, "dormant:"+month).Scan(&memberIDs).Error; err != nil {
		logx.Errorf("automation: 扫描沉睡用户失败: %v", err)
		return
	}
	for _, memberID := range memberIDs {
		if granted := grantDormantCoupon(sc, memberID); granted {
			_ = notify.Enqueue(sc, memberID, model.NotifySceneCoupon,
				"dormant:"+strconv.FormatInt(memberID, 10)+":"+month,
				"好久不见", "送你一张专属优惠券，回到首页看看吧", "")
		}
	}
}

// grantDormantCoupon 发放沉睡召回券（模板余量校验；券本身不幂等，幂等由通知键+调用频率控制）
func grantDormantCoupon(sc *svc.ServiceContext, memberID int64) bool {
	templateID := sc.Config.Automation.DormantCouponID
	if templateID <= 0 {
		return false
	}
	var tpl model.CouponTemplate
	if err := sc.DB.Where("id = ? AND status = 1 AND issued_count < total_count", templateID).
		First(&tpl).Error; err != nil {
		return false
	}
	err := sc.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Exec(
			"UPDATE coupon_template SET issued_count = issued_count + 1, updated_at = now() WHERE id = ? AND issued_count < total_count",
			templateID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return common.ErrCouponSoldOut
		}
		return tx.Create(&model.MemberCoupon{
			ID:         common.NewID(),
			MemberID:   memberID,
			TemplateID: templateID,
			Status:     model.CouponUsable,
			ReceivedAt: time.Now(),
		}).Error
	})
	return err == nil
}

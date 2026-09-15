// remind.go 优惠券到期提醒：惰性过期补扫（可用→已过期）+ 3 天内到期券入通知队列。
// biz_key=couponexp:<memberCouponID> 幂等，重复扫描只提醒一次。
package marketing

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"pet/backend/internal/logic/notify"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
)

const couponRemindAhead = 3 * 24 * time.Hour // 提前提醒窗口

// StartCouponReminders 启动券提醒定时任务（启动即跑一轮，之后每 30 分钟）
func StartCouponReminders(ctx context.Context, sc *svc.ServiceContext) {
	runCouponReminderRound(sc)
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runCouponReminderRound(sc)
		}
	}
}

func runCouponReminderRound(sc *svc.ServiceContext) {
	// ① 惰性过期补扫：可用但模板 valid_end 已过的批量置为已过期
	if err := sc.DB.Exec(
		`UPDATE member_coupon SET status = ? WHERE status = ?
		 AND EXISTS (SELECT 1 FROM coupon_template t WHERE t.id = member_coupon.template_id
		             AND t.valid_end IS NOT NULL AND t.valid_end < now())`,
		model.CouponExpired, model.CouponUsable).Error; err != nil {
		logx.Errorf("券过期补扫失败: %v", err)
	}

	// ② 即将到期提醒：3 天内到期的可用券逐条入通知队列（biz_key 幂等恰好一次）
	now := time.Now()
	var rows []struct {
		ID       int64
		MemberID int64
		Name     string
	}
	if err := sc.DB.Table("member_coupon AS mc").
		Joins("JOIN coupon_template t ON t.id = mc.template_id").
		Where("mc.status = ? AND t.valid_end >= ? AND t.valid_end <= ?",
			model.CouponUsable, now, now.Add(couponRemindAhead)).
		Order("mc.id").Select("mc.id, mc.member_id, t.name").Scan(&rows).Error; err != nil {
		logx.Errorf("扫描即将到期券失败: %v", err)
		return
	}
	for _, r := range rows {
		if err := notify.Enqueue(sc, r.MemberID, model.NotifySceneCoupon,
			"couponexp:"+formatID(r.ID), "优惠券即将到期", r.Name+"即将到期，尽快使用", ""); err != nil {
			logx.Errorf("券到期提醒入队失败 memberCoupon=%d: %v", r.ID, err)
		}
	}
}

// Package growth 用户增长：积分（签到/消费/评价/邀请奖励、兑换优惠券扣减）
// 与邀请归因。积分变动一律走 Credit/Deduct 原子 SQL 并落 points_log 流水。
package growth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func norm(sc *svc.ServiceContext, n int) int {
	if n <= 0 {
		return 0
	}
	return n
}

// Credit 积分入账（原子 UPDATE RETURNING + 流水）
func Credit(sc *svc.ServiceContext, memberID, change int64, reason, ref string) (int64, error) {
	return CreditTx(sc.DB, memberID, change, reason, ref)
}

// CreditTx 事务内积分入账
func CreditTx(db *gorm.DB, memberID, change int64, reason, ref string) (int64, error) {
	if change <= 0 {
		return 0, nil
	}
	var balance int64
	err := db.Raw(
		`UPDATE member SET points = points + ?, updated_at = now() WHERE id = ? RETURNING points`,
		change, memberID).Scan(&balance).Error
	if err != nil {
		return 0, err
	}
	return balance, db.Create(&model.PointsLog{
		ID: common.NewID(), MemberID: memberID,
		Change: change, BalanceAfter: balance, Reason: reason, Ref: ref,
	}).Error
}

// Deduct 积分扣减（余额不足返回 ErrPointsNotEnough，调用方自行回滚事务）
func Deduct(sc *svc.ServiceContext, memberID, change int64, reason, ref string) error {
	return DeductTx(sc.DB, memberID, change, reason, ref)
}

// DeductTx 事务内积分扣减
func DeductTx(db *gorm.DB, memberID, change int64, reason, ref string) error {
	if change <= 0 {
		return nil
	}
	res := db.Exec(
		`UPDATE member SET points = points - ?, updated_at = now() WHERE id = ? AND points >= ?`,
		change, memberID, change)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrPointsNotEnough
	}
	var balance int64
	if err := db.Raw(`SELECT points FROM member WHERE id = ?`, memberID).Scan(&balance).Error; err != nil {
		return err
	}
	return db.Create(&model.PointsLog{
		ID: common.NewID(), MemberID: memberID,
		Change: -change, BalanceAfter: balance, Reason: reason, Ref: ref,
	}).Error
}

// SignIn 每日签到：Redis SetNX 当日唯一，成功加 Growth.SignPoints + 连签奖励
func SignIn(sc *svc.ServiceContext, memberID int64) (int64, bool, error) {
	pts := norm(sc, sc.Config.Growth.SignPoints)
	if pts == 0 {
		return 0, false, common.NewErr(400, 41806, "签到活动未开启")
	}
	key := "points:signin:" + strconv.FormatInt(memberID, 10) + ":" + time.Now().Format("20060102")
	ok, err := sc.Rdb.SetNX(context.Background(), key, 1, 25*time.Hour).Result()
	if err != nil {
		return 0, false, err
	}
	if !ok {
		return 0, false, nil // 今日已签到（幂等，不报错）
	}
	balance, err := Credit(sc, memberID, int64(pts), model.PointsReasonSign, time.Now().Format("2006-01-02"))
	if err != nil {
		return 0, true, err
	}
	// 连签奖励（3 天 / 7 天）：写入 points_log，reason=sign_bonus
	if streak := currentStreak(sc, memberID); streak > 0 {
		bonus := 0
		if streak == 3 {
			bonus = sc.Config.Growth.SignBonus3
		} else if streak == 7 {
			bonus = sc.Config.Growth.SignBonus7
		}
		if bonus > 0 {
			_, _ = Credit(sc, memberID, int64(bonus), "sign_bonus",
				"streak:"+time.Now().Format("2006-01-02"))
		}
	}
	return balance, true, nil
}

// signDates 查询某月已签日期（含补签）
func signDates(sc *svc.ServiceContext, memberID int64, month string) ([]string, error) {
	var refs []string
	err := sc.DB.Model(&model.PointsLog{}).
		Where("member_id = ? AND reason IN ? AND ref LIKE ?", memberID,
			[]string{model.PointsReasonSign, "sign_makeup"}, month+"-%").
		Order("ref ASC").Distinct("ref").Pluck("ref", &refs).Error
	return refs, err
}

// currentStreak 截至今日（含今日）的连续签到天数
func currentStreak(sc *svc.ServiceContext, memberID int64) int {
	day := time.Now()
	streak := 0
	for i := 0; i < 400; i++ {
		ref := day.Format("2006-01-02")
		var cnt int64
		sc.DB.Model(&model.PointsLog{}).
			Where("member_id = ? AND reason IN ? AND ref = ?", memberID,
				[]string{model.PointsReasonSign, "sign_makeup"}, ref).
			Count(&cnt)
		if cnt > 0 {
			streak++
		} else if i > 0 || day.Format("2006-01-02") != time.Now().Format("2006-01-02") {
			break
		} else if i == 0 {
			// 今日未签不影响昨日连签链的延续判断
		}
		day = day.AddDate(0, 0, -1)
	}
	return streak
}

// SignCalendar 签到日历
func SignCalendar(sc *svc.ServiceContext, memberID int64, month string) (*types.SignCalendarResp, error) {
	if len(month) != 7 {
		month = time.Now().Format("2006-01")
	}
	dates, err := signDates(sc, memberID, month)
	if err != nil {
		return nil, err
	}
	today := time.Now().Format("2006-01-02")
	signedToday := false
	for _, d := range dates {
		if d == today {
			signedToday = true
			break
		}
	}
	return &types.SignCalendarResp{
		Month:       month,
		SignedDates: dates,
		Streak:      currentStreak(sc, memberID),
		SignedToday: signedToday,
	}, nil
}

// SignMakeup 补签：本月内未签的过去日期，消耗积分
func SignMakeup(sc *svc.ServiceContext, memberID int64, date string) error {
	cost := sc.Config.Growth.SignMakeupCost
	if cost <= 0 {
		return common.ErrSignMakeupInvalid
	}
	day, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return common.ErrSignMakeupInvalid
	}
	today := time.Now()
	if day.After(today) || day.Month() != today.Month() || day.Format("2006-01-02") == today.Format("2006-01-02") {
		return common.ErrSignMakeupInvalid // 仅限本月过去的日期
	}
	var cnt int64
	if err := sc.DB.Model(&model.PointsLog{}).
		Where("member_id = ? AND reason IN ? AND ref = ?", memberID,
			[]string{model.PointsReasonSign, "sign_makeup"}, date).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return common.ErrSignMakeupInvalid
	}
	if err := Deduct(sc, memberID, int64(cost), "sign_makeup", date); err != nil {
		return err
	}
	// 补签落签到流水（日历可见），积分奖励不补发
	var balance int64
	if err := sc.DB.Model(&model.Member{}).Select("points").
		Where("id = ?", memberID).Scan(&balance).Error; err != nil {
		return err
	}
	return sc.DB.Create(&model.PointsLog{
		ID: common.NewID(), MemberID: memberID,
		Change: 0, BalanceAfter: balance, Reason: model.PointsReasonSign, Ref: date,
	}).Error
}

// MyPoints 积分余额 + 流水分页
func MyPoints(sc *svc.ServiceContext, memberID int64, page, size int) (*types.PointsResp, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 20
	}
	var m model.Member
	if err := sc.DB.Select("points").First(&m, memberID).Error; err != nil {
		return nil, err
	}
	var total int64
	if err := sc.DB.Model(&model.PointsLog{}).Where("member_id = ?", memberID).Count(&total).Error; err != nil {
		return nil, err
	}
	var logs []model.PointsLog
	if err := sc.DB.Where("member_id = ?", memberID).
		Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&logs).Error; err != nil {
		return nil, err
	}
	list := make([]types.PointsLogView, 0, len(logs))
	for _, l := range logs {
		list = append(list, types.PointsLogView{
			ID:        strconv.FormatInt(l.ID, 10),
			Change:    l.Change,
			Balance:   l.BalanceAfter,
			Reason:    pointsReasonText(l.Reason),
			Ref:       l.Ref,
			CreatedAt: l.CreatedAt.Format(time.RFC3339),
		})
	}
	return &types.PointsResp{
		Points: m.Points,
		Total:  total,
		List:   list,
	}, nil
}

func pointsReasonText(r string) string {
	switch r {
	case model.PointsReasonSign:
		return "每日签到"
	case model.PointsReasonOrder:
		return "消费奖励"
	case model.PointsReasonReview:
		return "评价奖励"
	case model.PointsReasonInvite:
		return "邀请好友"
	case model.PointsReasonInvited:
		return "好友邀请奖励"
	case model.PointsReasonExchange:
		return "兑换优惠券"
	}
	return r
}

// RewardOrder 支付成功返积分（向下取整每元 N 分）；失败仅日志不阻断支付落账
func RewardOrder(sc *svc.ServiceContext, memberID int64, payAmount decimal.Decimal, orderNo string) {
	per := norm(sc, sc.Config.Growth.OrderPointsPerYuan)
	if per == 0 {
		return
	}
	change := int64(payAmount.IntPart()) * int64(per)
	if change <= 0 {
		return
	}
	if _, err := Credit(sc, memberID, change, model.PointsReasonOrder, orderNo); err != nil {
		logx.Errorf("消费积分发放失败 member=%d order=%s: %v", memberID, orderNo, err)
	}
}

// RewardOrderTx 事务内支付返积分（随支付落账同事务，天然幂等一次）；失败仅日志不阻断
func RewardOrderTx(tx *gorm.DB, sc *svc.ServiceContext, memberID int64, payAmount decimal.Decimal, orderNo string) {
	per := norm(sc, sc.Config.Growth.OrderPointsPerYuan)
	if per == 0 {
		return
	}
	change := int64(payAmount.IntPart()) * int64(per)
	if change <= 0 {
		return
	}
	if _, err := CreditTx(tx, memberID, change, model.PointsReasonOrder, orderNo); err != nil {
		logx.Errorf("消费积分发放失败 member=%d order=%s: %v", memberID, orderNo, err)
	}
}

// RewardReview 首次评价奖励
func RewardReview(sc *svc.ServiceContext, memberID int64, orderNo string) {
	pts := norm(sc, sc.Config.Growth.ReviewPoints)
	if pts == 0 {
		return
	}
	if _, err := Credit(sc, memberID, int64(pts), model.PointsReasonReview, orderNo); err != nil {
		logx.Errorf("评价积分发放失败 member=%d order=%s: %v", memberID, orderNo, err)
	}
}

// ─────────────────────────── 邀请 ───────────────────────────

// GenInviteCode 由雪花 ID 派生 8 位 base36 邀请码（唯一性由主键保证）
func GenInviteCode(id int64) string {
	return strings.ToUpper(fmt.Sprintf("%08s", strconv.FormatInt(id, 36)))
}

// AttachInvite 新注册会员邀请归因：双方各得 Growth.InviteRewardPoints。
// 邀请码无效时静默忽略（不影响注册主流程）；异常仅日志。
func AttachInvite(sc *svc.ServiceContext, newMemberID int64, inviteCode string) {
	code := strings.TrimSpace(inviteCode)
	if code == "" {
		return
	}
	var inviter model.Member
	if err := sc.DB.Where("invite_code = ? AND id <> ?", code, newMemberID).First(&inviter).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logx.Errorf("邀请码查询失败 member=%d code=%s: %v", newMemberID, code, err)
		}
		return
	}
	if err := sc.DB.Model(&model.Member{}).Where("id = ?", newMemberID).
		Update("invited_by", inviter.ID).Error; err != nil {
		logx.Errorf("邀请归因失败 member=%d inviter=%d: %v", newMemberID, inviter.ID, err)
		return
	}
	pts := int64(norm(sc, sc.Config.Growth.InviteRewardPoints))
	if pts > 0 {
		_, _ = Credit(sc, inviter.ID, pts, model.PointsReasonInvite, strconv.FormatInt(newMemberID, 10))
		_, _ = Credit(sc, newMemberID, pts, model.PointsReasonInvited, strconv.FormatInt(inviter.ID, 10))
	}
}

// InviteSummary 我的邀请码 + 已邀请人数
func InviteSummary(sc *svc.ServiceContext, memberID int64) (*types.InviteResp, error) {
	var m model.Member
	if err := sc.DB.Select("invite_code").First(&m, memberID).Error; err != nil {
		return nil, err
	}
	if m.InviteCode == "" {
		if err := sc.DB.Model(&model.Member{}).Where("id = ?", memberID).
			Update("invite_code", GenInviteCode(memberID)).Error; err != nil {
			return nil, err
		}
		m.InviteCode = GenInviteCode(memberID)
	}
	var invited int64
	if err := sc.DB.Model(&model.Member{}).Where("invited_by = ?", memberID).Count(&invited).Error; err != nil {
		return nil, err
	}
	return &types.InviteResp{
		InviteCode: m.InviteCode,
		Invited:    invited,
		RewardEach: int64(norm(sc, sc.Config.Growth.InviteRewardPoints)),
	}, nil
}

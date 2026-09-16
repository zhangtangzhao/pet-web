// Package risk 风控：下单频控、评价联系方式拦截、会员黑名单，命中即落 risk_log 留痕。
package risk

import (
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
)

var phonePattern = regexp.MustCompile(`1[3-9]\d{9}`)

const orderWindow = 10 * time.Minute
const orderMaxInWindow = 5

func isBlacklisted(sc *svc.ServiceContext, memberID int64) bool {
	var m model.Member
	if err := sc.DB.Select("blacklist").First(&m, memberID).Error; err != nil {
		return false
	}
	return m.Blacklist == 1
}

func logRisk(sc *svc.ServiceContext, memberID int64, rule, detail string) {
	_ = sc.DB.Create(&model.RiskLog{
		ID:       common.NewID(),
		MemberID: memberID,
		Rule:     rule,
		Detail:   truncate(detail, 250),
	}).Error
}

func truncate(s string, n int) string {
	rs := []rune(s)
	if len(rs) > n {
		return string(rs[:n])
	}
	return s
}

// CheckOrderAllowed 下单前置风控：黑名单 + 频控（近 10 分钟 ≥5 单拦截）
func CheckOrderAllowed(sc *svc.ServiceContext, memberID int64) error {
	if isBlacklisted(sc, memberID) {
		logRisk(sc, memberID, model.RiskRuleBlacklist, "黑名单用户尝试下单")
		return common.ErrMemberBlacklisted
	}
	var cnt int64
	if err := sc.DB.Model(&model.Order{}).
		Where("member_id = ? AND created_at > ?", memberID, time.Now().Add(-orderWindow)).
		Count(&cnt).Error; err != nil {
		return nil // 风控查询失败不阻断主流程
	}
	if cnt >= orderMaxInWindow {
		logRisk(sc, memberID, model.RiskRuleOrderFreq,
			strconv.FormatInt(cnt, 10)+" 单/"+strconv.Itoa(int(orderWindow.Minutes()))+" 分钟")
		return common.ErrRiskOrderFreq
	}
	return nil
}

// CheckReviewContent 评价内容风控：黑名单 + 手机号等联系方式拦截
func CheckReviewContent(sc *svc.ServiceContext, memberID int64, content string) error {
	if isBlacklisted(sc, memberID) {
		logRisk(sc, memberID, model.RiskRuleBlacklist, "黑名单用户尝试评价")
		return common.ErrMemberBlacklisted
	}
	if content == "" || utf8.RuneCountInString(content) > 500 {
		return nil
	}
	if phonePattern.MatchString(content) {
		logRisk(sc, memberID, model.RiskRuleReviewContact, truncate(content, 120))
		return common.ErrReviewContact
	}
	return nil
}

// CheckTradeAction 通用交易动作校验（领券等）：黑名单
func CheckTradeAction(sc *svc.ServiceContext, memberID int64, action string) error {
	if isBlacklisted(sc, memberID) {
		logRisk(sc, memberID, model.RiskRuleBlacklist, "黑名单用户尝试"+action)
		return common.ErrMemberBlacklisted
	}
	return nil
}

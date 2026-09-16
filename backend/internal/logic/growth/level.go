// level.go 会员等级：成长值 = 累计实付（元取整），只增不减。
// V1 ≥1000（98折）V2 ≥5000（95折）V3 ≥20000（92折）；
// 首达等级自动发放升级礼包券（模板见迁移 017），level_reached 保证恰发一次。
package growth

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

var levelThresholds = []struct {
	Level     int
	Name      string
	Threshold int64
	Rate      float64 // 商品金额（券后）折扣率
	CouponID  int64   // 升级礼包券模板，0=无
}{
	{1, "V1", 1000, 0.98, 5111},
	{2, "V2", 5000, 0.95, 5112},
	{3, "V3", 20000, 0.92, 5113},
}

// LevelOf 按成长值推导等级
func LevelOf(growth int64) int {
	level := 0
	for _, l := range levelThresholds {
		if growth >= l.Threshold {
			level = l.Level
		}
	}
	return level
}

// LevelRate 等级折扣率（越级越高折扣；非 levels 内为 1）
func LevelRate(level int) float64 {
	for _, l := range levelThresholds {
		if l.Level == level {
			return l.Rate
		}
	}
	return 1
}

// LevelName 等级名
func LevelName(level int) string {
	for _, l := range levelThresholds {
		if l.Level == level {
			return l.Name
		}
	}
	return "V0"
}

// LevelView 我的等级信息
func LevelView(sc *svc.ServiceContext, memberID int64) (*types.LevelView, error) {
	var m model.Member
	if err := sc.DB.Select("growth_value", "level_reached").First(&m, memberID).Error; err != nil {
		return nil, err
	}
	level := LevelOf(m.GrowthValue)
	v := &types.LevelView{
		GrowthValue: m.GrowthValue,
		Level:       level,
		LevelName:   LevelName(level),
		Discount:    strconv.FormatFloat(LevelRate(level), 'f', 2, 64),
	}
	for _, l := range levelThresholds {
		if l.Level > level {
			v.NextThreshold = l.Threshold
			v.NextDiscount = strconv.FormatFloat(l.Rate, 'f', 2, 64)
			break
		}
	}
	return v, nil
}

// OnPaidTx 支付事务内累计成长值；升级礼包随事务发放（level_reached 恰好一次）。
// 返回是否升级及新等级（未升级时 newLevel 无意义）。
func OnPaidTx(tx *gorm.DB, memberID int64, payYuan int64) (bool, int, error) {
	if payYuan <= 0 {
		return false, 0, nil
	}
	if err := tx.Model(&model.Member{}).Where("id = ?", memberID).
		Update("growth_value", gorm.Expr("growth_value + ?", payYuan)).Error; err != nil {
		return false, 0, err
	}
	var m model.Member
	if err := tx.Select("growth_value", "level_reached").First(&m, memberID).Error; err != nil {
		return false, 0, err
	}
	newLevel := LevelOf(m.GrowthValue)
	if newLevel <= m.LevelReached {
		return false, newLevel, nil
	}
	couponID := int64(0)
	for _, l := range levelThresholds {
		if l.Level == newLevel {
			couponID = l.CouponID
		}
	}
	if couponID > 0 {
		if err := tx.Create(&model.MemberCoupon{
			ID:         common.NewID(),
			MemberID:   memberID,
			TemplateID: couponID,
			Status:     model.CouponUsable,
			ReceivedAt: time.Now(),
		}).Error; err != nil {
			return false, newLevel, err
		}
	}
	if err := tx.Model(&model.Member{}).Where("id = ?", memberID).
		Update("level_reached", newLevel).Error; err != nil {
		return false, newLevel, err
	}
	return true, newLevel, nil
}

// LevelCouponTemplateID 等级对应的升级券模板
func LevelCouponTemplateID(level int) int64 {
	for _, l := range levelThresholds {
		if l.Level == level {
			return l.CouponID
		}
	}
	return 0
}

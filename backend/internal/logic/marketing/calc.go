package marketing

import (
	"time"

	"github.com/shopspring/decimal"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
)

var decimal100 = decimal.NewFromInt(100)

// CalcDiscount 按模板类型计算抵扣金额：
// 满减/立减 = discount_amount；折扣 = min(base*(100-pct)/100, max_discount_amount)
// 结果 clamp 到 [0, base]。
func CalcDiscount(t *model.CouponTemplate, base decimal.Decimal) decimal.Decimal {
	var d decimal.Decimal
	switch t.Type {
	case model.CouponTypeThreshold, model.CouponTypeCash:
		d = t.DiscountAmount
	case model.CouponTypeDiscount:
		if t.DiscountPercent <= 0 || t.DiscountPercent >= 100 {
			return decimal.Zero
		}
		d = base.Mul(decimal100.Sub(decimal.NewFromInt(int64(t.DiscountPercent)))).Div(decimal100)
		if t.MaxDiscountAmount.GreaterThan(decimal.Zero) && d.GreaterThan(t.MaxDiscountAmount) {
			d = t.MaxDiscountAmount
		}
	default:
		return decimal.Zero
	}
	if d.LessThan(decimal.Zero) {
		d = decimal.Zero
	}
	if d.GreaterThan(base) {
		d = base
	}
	return d
}

// ValidateUsable 校验用户券在 base 消费下是否可用；可用返回模板
func ValidateUsable(t *model.CouponTemplate, base decimal.Decimal) error {
	now := time.Now()
	if t.Status != 1 {
		return common.ErrCouponUnusable
	}
	if t.ValidStart != nil && now.Before(*t.ValidStart) {
		return common.ErrCouponUnusable
	}
	if t.ValidEnd != nil && now.After(*t.ValidEnd) {
		return common.ErrCouponUnusable
	}
	if base.LessThan(t.ThresholdAmount) {
		return common.ErrCouponThreshold
	}
	return nil
}

// CouponTypeText 模板类型文案
func CouponTypeText(t int) string {
	switch t {
	case model.CouponTypeThreshold:
		return "满减券"
	case model.CouponTypeDiscount:
		return "折扣券"
	case model.CouponTypeCash:
		return "立减券"
	}
	return "优惠券"
}

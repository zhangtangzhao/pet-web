// growthops.go RFM 会员分群 + 智能定价建议（纯 SQL 统计）。
package manage

import (
	"strconv"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// MemberSegments RFM 分群统计
func MemberSegments(sc *svc.ServiceContext) ([]types.SegmentRow, error) {
	type row struct {
		Segment string
		Cnt     int64
	}
	var rows []row
	if err := sc.DB.Raw(`
		WITH stat AS (
			SELECT m.id, MAX(o.paid_at) AS last_paid,
			       COUNT(o.id) FILTER (WHERE o.status >= 20) AS pay_cnt,
			       COALESCE(SUM(o.pay_amount) FILTER (WHERE o.status >= 20), 0) AS total
			FROM member m LEFT JOIN orders o ON o.member_id = m.id AND o.status >= 20
			WHERE m.status = 1 GROUP BY m.id
		)
		SELECT CASE
			WHEN pay_cnt = 0 THEN 'potential'
			WHEN last_paid > now() - INTERVAL '30 days' AND total >= 5000 THEN 'vip'
			WHEN last_paid > now() - INTERVAL '30 days' THEN 'active'
			ELSE 'dormant' END AS segment, COUNT(*) AS cnt
		FROM stat GROUP BY segment`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	name := map[string]string{
		"vip": "高价值会员", "active": "活跃会员", "dormant": "沉睡会员", "potential": "潜在新客",
	}
	out := make([]types.SegmentRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, types.SegmentRow{Segment: r.Segment, Name: name[r.Segment], Count: r.Cnt})
	}
	return out, nil
}

// SegmentMembers 某分群会员列表（Top 50）
func SegmentMembers(sc *svc.ServiceContext, segment string) ([]types.SegmentMemberRow, error) {
	base := `
		SELECT m.id, m.nickname, m.phone,
		       COALESCE(SUM(o.pay_amount) FILTER (WHERE o.status >= 20), 0) AS total,
		       COUNT(o.id) FILTER (WHERE o.status >= 20) AS cnt,
		       MAX(o.paid_at) AS last_paid
		FROM member m LEFT JOIN orders o ON o.member_id = m.id AND o.status >= 20
		WHERE m.status = 1 GROUP BY m.id HAVING true`
	switch segment {
	case "vip":
		base += ` AND MAX(o.paid_at) > now() - INTERVAL '30 days' AND COALESCE(SUM(o.pay_amount) FILTER (WHERE o.status >= 20), 0) >= 5000`
	case "active":
		base += ` AND MAX(o.paid_at) > now() - INTERVAL '30 days' AND COALESCE(SUM(o.pay_amount) FILTER (WHERE o.status >= 20), 0) < 5000`
	case "dormant":
		base += ` AND (SELECT COUNT(o.id) FROM orders o WHERE o.member_id = m.id AND o.status >= 20) > 0 AND (MAX(o.paid_at) IS NULL OR MAX(o.paid_at) <= now() - INTERVAL '30 days')`
	case "potential":
		base += ` AND (SELECT COUNT(o.id) FROM orders o WHERE o.member_id = m.id AND o.status >= 20) = 0`
	default:
		return nil, common.ErrParam
	}
	base += ` ORDER BY total DESC LIMIT 50`
	var rows []struct {
		ID       int64
		Nickname string
		Phone    string
		Total    decimal.Decimal
		Cnt      int64
	}
	if err := sc.DB.Raw(base).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]types.SegmentMemberRow, 0, len(rows))
	for _, r := range rows {
		last := "-"
		out = append(out, types.SegmentMemberRow{
			MemberID: strconv.FormatInt(r.ID, 10), Nickname: r.Nickname, Phone: r.Phone,
			Total: r.Total.StringFixed(2), OrderCount: r.Cnt, LastPaid: last,
		})
	}
	return out, nil
}

// SegmentIssue 分群一键发券
func SegmentIssue(sc *svc.ServiceContext, segment string, templateID int64) (int64, error) {
	members, err := SegmentMembers(sc, segment)
	if err != nil {
		return 0, err
	}
	var granted int64
	for _, m := range members {
		id, _ := strconv.ParseInt(m.MemberID, 10, 64)
		err := sc.DB.Transaction(func(tx *gorm.DB) error {
			res := tx.Exec("UPDATE coupon_template SET issued_count = issued_count + 1 WHERE id = ? AND status = 1 AND issued_count < total_count", templateID)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return common.ErrCouponSoldOut
			}
			return tx.Exec(`INSERT INTO member_coupon (id, member_id, template_id, status, received_at) VALUES (?, ?, ?, 1, now())`, common.NewID(), id, templateID).Error
		})
		if err == nil {
			granted++
		}
	}
	return granted, nil
}

// PriceSuggest 智能定价建议
func PriceSuggest(sc *svc.ServiceContext, productID int64) (*types.PriceSuggestResp, error) {
	var p model.PetProduct
	if err := sc.DB.Select("id", "breed_id", "price").First(&p, productID).Error; err != nil {
		return nil, common.ErrNotFound
	}
	var sold struct {
		Avg decimal.Decimal
		Min decimal.Decimal
		Max decimal.Decimal
		Cnt int64
	}
	_ = sc.DB.Raw(`
		SELECT COALESCE(AVG(oi.price),0) AS avg, COALESCE(MIN(oi.price),0) AS min,
		       COALESCE(MAX(oi.price),0) AS max, COUNT(*) AS cnt
		FROM order_item oi JOIN orders o ON o.id = oi.order_id
		JOIN pet_product pp ON pp.id = oi.product_id
		WHERE pp.breed_id = ? AND o.status IN (20, 30)`, p.BreedID).Scan(&sold).Error
	var onSale int64
	sc.DB.Model(&model.PetProduct{}).Where("breed_id = ? AND status = ?", p.BreedID, model.ProductOnSale).Count(&onSale)
	resp := &types.PriceSuggestResp{
		BreedID: strconv.FormatInt(p.BreedID, 10), SoldCount: sold.Cnt, OnSaleCount: onSale,
		AvgPrice: sold.Avg.StringFixed(2), MinPrice: sold.Min.StringFixed(2), MaxPrice: sold.Max.StringFixed(2),
	}
	if sold.Cnt > 0 {
		resp.SuggestLow = sold.Avg.Mul(decimal.NewFromFloat(0.9)).StringFixed(2)
		resp.SuggestHigh = sold.Avg.Mul(decimal.NewFromFloat(1.1)).StringFixed(2)
	}
	return resp, nil
}

// report.go 报表深化：经营总览（今日/本月 GMV）+ 近 N 天成交趋势 +
// 商品销量 Top10 + 品类分布 + 待处理售后数。成交口径 = 已支付及之后状态（20/30/60）按实付计。
package manage

import (
	"fmt"

	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func AdminReport(sc *svc.ServiceContext, days int) (*types.AdminReportResp, error) {
	if days <= 0 || days > 90 {
		days = 30
	}
	resp := &types.AdminReportResp{
		Daily:       []types.AdminDailyRow{},
		TopProducts: []types.AdminRankRow{},
		Breeds:      []types.AdminRankRow{},
	}

	var todayGmv, monthGmv *float64
	if err := sc.DB.Model(&model.Order{}).
		Select("COALESCE(SUM(pay_amount),0)").
		Where("status IN ? AND paid_at >= CURRENT_DATE", []int{model.OrderPaid, model.OrderCompleted, model.OrderRefunded}).
		Scan(&todayGmv).Error; err != nil {
		return nil, err
	}
	if err := sc.DB.Model(&model.Order{}).
		Select("COUNT(*)").
		Where("status IN ? AND paid_at >= CURRENT_DATE", []int{model.OrderPaid, model.OrderCompleted, model.OrderRefunded}).
		Scan(&resp.Summary.TodayOrders).Error; err != nil {
		return nil, err
	}
	if err := sc.DB.Model(&model.Order{}).
		Select("COALESCE(SUM(pay_amount),0)").
		Where("status IN ? AND paid_at >= date_trunc('month', now())", []int{model.OrderPaid, model.OrderCompleted, model.OrderRefunded}).
		Scan(&monthGmv).Error; err != nil {
		return nil, err
	}
	if err := sc.DB.Model(&model.Order{}).
		Select("COUNT(*)").
		Where("status IN ? AND paid_at >= date_trunc('month', now())", []int{model.OrderPaid, model.OrderCompleted, model.OrderRefunded}).
		Scan(&resp.Summary.MonthOrders).Error; err != nil {
		return nil, err
	}
	resp.Summary.TodayGMV = money2(todayGmv)
	resp.Summary.MonthGMV = money2(monthGmv)
	if err := sc.DB.Model(&model.AfterSale{}).
		Select("COUNT(*)").
		Where("status = ?", model.AfterSalePending).
		Scan(&resp.Summary.PendingAfters).Error; err != nil {
		return nil, err
	}

	// 近 N 天成交趋势
	var daily []struct {
		Day    string
		Cnt    int64
		Amount *float64
	}
	if err := sc.DB.Model(&model.Order{}).
		Select(`TO_CHAR(date_trunc('day', paid_at), 'MM-DD') AS day, COUNT(*) AS cnt,
		        COALESCE(SUM(pay_amount),0) AS amount`).
		Where("status IN ? AND paid_at >= CURRENT_DATE - ?::int", []int{model.OrderPaid, model.OrderCompleted, model.OrderRefunded}, days).
		Group("day").Order("day").Scan(&daily).Error; err != nil {
		return nil, err
	}
	for _, d := range daily {
		resp.Daily = append(resp.Daily, types.AdminDailyRow{Date: d.Day, Orders: d.Cnt, GMV: money2(d.Amount)})
	}

	// 商品销量 Top10
	var top []struct {
		Title  string
		Cnt    int64
		Amount *float64
	}
	if err := sc.DB.Table("order_item i").
		Select("i.product_title AS title, COUNT(*) AS cnt, COALESCE(SUM(i.price * i.quantity),0) AS amount").
		Joins("JOIN orders o ON o.id = i.order_id AND o.status IN ?", []int{model.OrderPaid, model.OrderCompleted, model.OrderRefunded}).
		Group("i.product_title").Order("cnt DESC").Limit(10).
		Scan(&top).Error; err != nil {
		return nil, err
	}
	for _, t := range top {
		resp.TopProducts = append(resp.TopProducts, types.AdminRankRow{Name: t.Title, Count: t.Cnt, Amount: money2(t.Amount)})
	}

	// 品类分布
	var breeds []struct {
		Breed  string
		Cnt    int64
		Amount *float64
	}
	if err := sc.DB.Table("order_item i").
		Select("COALESCE(NULLIF(i.breed_name,''),'其他') AS breed, COUNT(*) AS cnt, COALESCE(SUM(i.price * i.quantity),0) AS amount").
		Joins("JOIN orders o ON o.id = i.order_id AND o.status IN ?", []int{model.OrderPaid, model.OrderCompleted, model.OrderRefunded}).
		Group("breed").Order("cnt DESC").Limit(10).
		Scan(&breeds).Error; err != nil {
		return nil, err
	}
	for _, b := range breeds {
		resp.Breeds = append(resp.Breeds, types.AdminRankRow{Name: b.Breed, Count: b.Cnt, Amount: money2(b.Amount)})
	}
	return resp, nil
}

func money2(p *float64) string {
	if p == nil {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", *p)
}

// insight.go 活动日历 + 数据大屏（运营视图聚合）
package manage

import (
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

var provincePrefixes = []string{
	"北京", "天津", "上海", "重庆", "河北", "山西", "辽宁", "吉林", "黑龙江",
	"江苏", "浙江", "安徽", "福建", "江西", "山东", "河南", "湖北", "湖南",
	"广东", "广西", "海南", "四川", "贵州", "云南", "西藏", "陕西", "甘肃",
	"青海", "宁夏", "新疆", "内蒙古", "香港", "澳门", "台湾",
}

// ActivityCalendar 活动日历：秒杀 / 拼团 / 优惠券统一排期 + 同商品时间窗冲突检测
func ActivityCalendar(sc *svc.ServiceContext, days int) (*types.ActivityCalendarResp, error) {
	if days < 1 || days > 90 {
		days = 30
	}
	from := time.Now().AddDate(0, 0, -1)
	to := time.Now().AddDate(0, 0, days)
	items := make([]types.ActivityItem, 0, 32)

	var flashes []model.FlashSale
	if err := sc.DB.Where("end_at > ? AND start_at < ?", from, to).Limit(200).Find(&flashes).Error; err == nil {
		for _, f := range flashes {
			items = append(items, types.ActivityItem{
				Type: "flash", TypeText: "秒杀", ID: strconv.FormatInt(f.ID, 10),
				Name:      "秒杀 · " + strconv.FormatInt(f.SalePrice.IntPart(), 10) + "元",
				ProductID: strconv.FormatInt(f.ProductID, 10),
				StartAt:   f.StartAt.Format("2006-01-02 15:04"), EndAt: f.EndAt.Format("2006-01-02 15:04"),
				Status: f.Status,
			})
		}
	}
	var groups []model.GroupBuy
	if err := sc.DB.Where("status = 1").Limit(200).Find(&groups).Error; err == nil {
		for _, g := range groups {
			items = append(items, types.ActivityItem{
				Type: "group", TypeText: "拼团", ID: strconv.FormatInt(g.ID, 10),
				Name:      strconv.Itoa(g.Size) + "人团 · " + g.Price.StringFixed(0) + "元",
				ProductID: strconv.FormatInt(g.ProductID, 10),
				StartAt:   g.CreatedAt.Format("2006-01-02 15:04"), EndAt: "-",
				Status: g.Status,
			})
		}
	}
	var coupons []model.CouponTemplate
	if err := sc.DB.Where("status = 1").Limit(200).Find(&coupons).Error; err == nil {
		for _, c := range coupons {
			start, end := "-", "-"
			if c.ValidStart != nil {
				start = c.ValidStart.Format("2006-01-02")
			}
			if c.ValidEnd != nil {
				end = c.ValidEnd.Format("2006-01-02")
			}
			items = append(items, types.ActivityItem{
				Type: "coupon", TypeText: "优惠券", ID: strconv.FormatInt(c.ID, 10),
				Name: c.Name, StartAt: start, EndAt: end, Status: c.Status,
			})
		}
	}
	// 冲突检测：同商品同时段启用中的秒杀与拼团
	conflicts := []string{}
	for _, f := range flashes {
		if f.Status != 1 {
			continue
		}
		for _, g := range groups {
			if g.Status != 1 || g.ProductID != f.ProductID {
				continue
			}
			if g.CreatedAt.Before(f.EndAt) {
				conflicts = append(conflicts,
					"商品 "+strconv.FormatInt(f.ProductID, 10)+" 的秒杀与拼团时间窗重叠")
			}
		}
	}
	if conflicts == nil {
		conflicts = []string{}
	}
	return &types.ActivityCalendarResp{Items: items, Conflicts: conflicts}, nil
}

// RealtimeDashboard 数据大屏
func RealtimeDashboard(sc *svc.ServiceContext) (*types.RealtimeDashboardResp, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	resp := &types.RealtimeDashboardResp{
		Hourly:      make([]types.HourlyPoint, 0, 24),
		Funnel:      map[string]int64{},
		Regions:     []types.RegionRow{},
		GeneratedAt: now.Format("2006-01-02 15:04:05"),
	}
	// 今日 GMV / 订单
	type payAgg struct {
		Amount decimal.Decimal
		Count  int64
	}
	var agg payAgg
	_ = sc.DB.Model(&model.Payment{}).
		Select("COALESCE(SUM(amount), 0) AS amount, COUNT(*) AS count").
		Where("status = ? AND pay_type IN ? AND created_at >= ?",
			model.PayStatusSuccess, []int{model.PayTypePurchase, model.PayTypeTail}, todayStart).
		Scan(&agg).Error
	resp.TodayGmv = agg.Amount.StringFixed(2)
	resp.TodayOrders = agg.Count
	// 24 小时分布
	type hourRow struct {
		Hour   int
		Amount decimal.Decimal
		Count  int64
	}
	var hours []hourRow
	_ = sc.DB.Model(&model.Payment{}).
		Select("EXTRACT(HOUR FROM created_at)::int AS hour, COALESCE(SUM(amount),0) AS amount, COUNT(*) AS count").
		Where("status = ? AND pay_type IN ? AND created_at >= ?",
			model.PayStatusSuccess, []int{model.PayTypePurchase, model.PayTypeTail}, todayStart).
		Group("hour").Scan(&hours).Error
	hourMap := map[int]hourRow{}
	for _, h := range hours {
		hourMap[h.Hour] = h
	}
	for i := 0; i < 24; i++ {
		h := hourMap[i]
		resp.Hourly = append(resp.Hourly, types.HourlyPoint{
			Hour: padHour(i), Gmv: h.Amount.StringFixed(2), Orders: h.Count,
		})
	}
	// 漏斗（近 30 天）
	funnelFrom := now.AddDate(0, 0, -30)
	for _, item := range []struct {
		label  string
		status int
	}{
		{"created", -1}, {"paid", model.OrderPaid}, {"completed", model.OrderCompleted}, {"refunded", model.OrderRefunded},
	} {
		var cnt int64
		if item.label == "created" {
			sc.DB.Model(&model.Order{}).Where("created_at >= ?", funnelFrom).Count(&cnt)
		} else if item.label == "paid" {
			sc.DB.Model(&model.Order{}).Where("created_at >= ? AND status >= ?", funnelFrom, item.status).Count(&cnt)
		} else {
			sc.DB.Model(&model.Order{}).Where("created_at >= ? AND status = ?", funnelFrom, item.status).Count(&cnt)
		}
		resp.Funnel[item.label] = cnt
	}
	// 地域分布（收货地址省份前缀，近 30 天已支付订单）
	var orders []model.Order
	if err := sc.DB.Select("ship_address").
		Where("created_at >= ? AND ship_address <> '' AND status >= ?", funnelFrom, model.OrderPaid).
		Limit(5000).Find(&orders).Error; err == nil {
		regionMap := map[string]int64{}
		for _, o := range orders {
			for _, p := range provincePrefixes {
				if len(o.ShipAddress) >= len(p) && o.ShipAddress[:len(p)] == p {
					regionMap[p]++
					break
				}
			}
		}
		for province, cnt := range regionMap {
			resp.Regions = append(resp.Regions, types.RegionRow{Province: province, Orders: cnt})
		}
		// 按订单数降序取前 10
		for i := 0; i < len(resp.Regions); i++ {
			for j := i + 1; j < len(resp.Regions); j++ {
				if resp.Regions[j].Orders > resp.Regions[i].Orders {
					resp.Regions[i], resp.Regions[j] = resp.Regions[j], resp.Regions[i]
				}
			}
		}
		if len(resp.Regions) > 10 {
			resp.Regions = resp.Regions[:10]
		}
	}
	return resp, nil
}

func padHour(h int) string {
	if h < 10 {
		return "0" + strconv.Itoa(h)
	}
	return strconv.Itoa(h)
}

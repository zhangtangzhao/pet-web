package manage

import (
	"time"

	"github.com/shopspring/decimal"

	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// Overview 运营看板（在售宠物数 / 今日订单 / 今日 GMV / 会员总数）
func Overview(sc *svc.ServiceContext) (*types.OverviewResp, error) {
	resp := &types.OverviewResp{}
	today := time.Now().Format("2006-01-02")

	if err := sc.DB.Model(&model.PetProduct{}).
		Where("status = ?", model.ProductOnSale).Count(&resp.OnSaleCount).Error; err != nil {
		return nil, err
	}
	if err := sc.DB.Model(&model.Order{}).
		Where("created_at >= ? AND status >= ?", today, model.OrderPaid).Count(&resp.TodayOrders).Error; err != nil {
		return nil, err
	}
	var gmv decimal.Decimal
	if err := sc.DB.Model(&model.Order{}).
		Where("created_at >= ? AND status >= ?", today, model.OrderPaid).
		Select("COALESCE(SUM(pay_amount), 0)").Scan(&gmv).Error; err != nil {
		return nil, err
	}
	resp.TodayGMV = gmv.StringFixed(2)
	if err := sc.DB.Model(&model.Member{}).Count(&resp.MemberCount).Error; err != nil {
		return nil, err
	}
	return resp, nil
}

// Package manage 库存预警 + 风控记录（平台端视图）
package manage

import (
	"strconv"

	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// AdminStockAlerts 在售商品中库存 ≤ 阈值的清单（低到高排序）
func AdminStockAlerts(sc *svc.ServiceContext) (*types.StockAlertResp, error) {
	var products []model.PetProduct
	if err := sc.DB.
		Where("status = ? AND stock <= stock_warn_threshold", model.ProductOnSale).
		Order("stock ASC, id DESC").Limit(100).
		Find(&products).Error; err != nil {
		return nil, err
	}
	list := make([]types.StockAlertRow, 0, len(products))
	for _, p := range products {
		list = append(list, types.StockAlertRow{
			ProductID: strconv.FormatInt(p.ID, 10),
			Title:     p.Title,
			Stock:     p.Stock,
			Threshold: p.StockWarnThreshold,
			Sales:     p.Sales,
		})
	}
	return &types.StockAlertResp{List: list}, nil
}

// AdminRiskLogs 风控拦截记录（分页，可按规则筛选）
func AdminRiskLogs(sc *svc.ServiceContext, req *types.RiskLogListReq) (*types.RiskLogResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	query := sc.DB.Model(&model.RiskLog{})
	if req.Rule != "" {
		query = query.Where("rule = ?", req.Rule)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var logs []model.RiskLog
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).
		Find(&logs).Error; err != nil {
		return nil, err
	}
	list := make([]types.RiskLogRow, 0, len(logs))
	for _, l := range logs {
		list = append(list, types.RiskLogRow{
			ID:        strconv.FormatInt(l.ID, 10),
			MemberID:  strconv.FormatInt(l.MemberID, 10),
			Rule:      l.Rule,
			Detail:    l.Detail,
			CreatedAt: l.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &types.RiskLogResp{Total: total, List: list}, nil
}

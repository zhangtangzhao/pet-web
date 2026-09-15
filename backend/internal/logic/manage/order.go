package manage

import (
	"github.com/shopspring/decimal"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/trade"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// AdminOrderList 订单列表（平台端，全量 + 筛选）
func AdminOrderList(sc *svc.ServiceContext, req *types.AdminOrderListReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	query := sc.DB.Model(&model.Order{})
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.OrderNo != "" {
		query = query.Where("order_no = ?", req.OrderNo)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where("contact_name ILIKE ? OR contact_phone ILIKE ?", kw, kw)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var orders []model.Order
	if err := query.Order("created_at DESC, id DESC").Offset((page - 1) * size).Limit(size).Find(&orders).Error; err != nil {
		return nil, err
	}
	list, err := trade.BuildOrderViews(sc, orders)
	if err != nil {
		return nil, err
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// AdminRefund 平台端退款（受理即落账，简化实现）
func AdminRefund(sc *svc.ServiceContext, req *types.RefundReq) error {
	if req.Reason == "" {
		return common.NewErr(400, 40001, "请填写退款原因")
	}
	var amount *decimal.Decimal
	if req.Amount != "" {
		d, err := decimal.NewFromString(req.Amount)
		if err != nil {
			return common.ErrParam
		}
		amount = &d
	}
	return trade.RefundOrder(sc, req.OrderNo, req.Reason, amount, common.NewBizNo("RF"))
}

// AdminOrderDetail 订单详情（平台端）
func AdminOrderDetail(sc *svc.ServiceContext, orderNo string) (*types.OrderView, error) {
	var o model.Order
	if err := sc.DB.Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		return nil, common.ErrNotFound
	}
	views, err := trade.BuildOrderViews(sc, []model.Order{o})
	if err != nil {
		return nil, err
	}
	return views[0], nil
}

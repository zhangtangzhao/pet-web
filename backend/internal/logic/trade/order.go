package trade

import (
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

const orderPayTTL = 30 * time.Minute

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, common.ErrParam
	}
	return id, nil
}

func releaseProduct(db *gorm.DB, productID int64) {
	db.Exec("UPDATE pet_product SET status = ?, updated_at = now() WHERE id = ? AND status = ?",
		model.ProductOnSale, productID, model.ProductLocked)
}

// CreateOrder 创建订单（活体防超卖：原子占位），并尝试预支付
func CreateOrder(sc *svc.ServiceContext, memberID int64, req *types.CreateOrderReq) (*types.CreateOrderResp, error) {
	productID, err := parseID(req.ProductID)
	if err != nil {
		return nil, err
	}
	if req.ContactName == "" || req.ContactPhone == "" {
		return nil, common.NewErr(400, 40001, "请填写联系人信息")
	}

	// 原子占位：仅"在售"可被锁定（防并发多人下单同一只）
	res := sc.DB.Exec(
		"UPDATE pet_product SET status = ?, updated_at = now() WHERE id = ? AND status = ?",
		model.ProductLocked, productID, model.ProductOnSale)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		var p model.PetProduct
		if err := sc.DB.First(&p, productID).Error; err != nil {
			return nil, common.ErrNotFound
		}
		if p.Status == model.ProductLocked {
			return nil, common.ErrGoodsLocked
		}
		return nil, common.ErrGoodsOffline
	}

	var p model.PetProduct
	if err := sc.DB.First(&p, productID).Error; err != nil {
		releaseProduct(sc.DB, productID)
		return nil, common.ErrNotFound
	}
	var breed model.Breed
	_ = sc.DB.First(&breed, p.BreedID).Error

	now := time.Now()
	order := model.Order{
		ID:           common.NewID(),
		OrderNo:      common.NewBizNo("P"),
		MemberID:     memberID,
		TotalAmount:  p.Price,
		PayAmount:    p.Price,
		Status:       model.OrderPending,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		Remark:       req.Remark,
		ExpireAt:     now.Add(orderPayTTL),
	}
	item := model.OrderItem{
		ID:           common.NewID(),
		ProductID:    p.ID,
		ProductTitle: p.Title,
		ProductImage: p.MainImage,
		BreedName:    breed.Name,
		Price:        p.Price,
		Quantity:     1,
	}
	payment := model.Payment{
		ID:        common.NewID(),
		PaymentNo: common.NewBizNo("PAY"),
		OrderID:   order.ID,
		OrderNo:   order.OrderNo,
		MemberID:  memberID,
		Amount:    p.Price,
		Channel:   model.PayChannelMini,
		PayType:   model.PayTypePurchase,
		Status:    model.PayStatusPending,
	}

	err = sc.DB.Transaction(func(tx *gorm.DB) error {
		order.ID = common.NewID()
		item.OrderID = order.ID
		payment.OrderID = order.ID
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		return tx.Create(&payment).Error
	})
	if err != nil {
		releaseProduct(sc.DB, productID)
		return nil, err
	}

	resp := &types.CreateOrderResp{
		OrderNo:   order.OrderNo,
		PayAmount: order.PayAmount.StringFixed(2),
		ExpireAt:  order.ExpireAt.Format(time.RFC3339),
	}
	// 尝试预支付（未配置支付/未绑定微信时返回空参数，前端可通过 prepay 接口重试）
	payParams, _ := prepayOrder(sc, memberID, &order)
	resp.PayParams = payParams
	return resp, nil
}

// OrderList 我的订单列表
func OrderList(sc *svc.ServiceContext, memberID int64, req *types.OrderListReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}
	query := sc.DB.Model(&model.Order{}).Where("member_id = ?", memberID)
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var orders []model.Order
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&orders).Error; err != nil {
		return nil, err
	}
	list, err := BuildOrderViews(sc, orders)
	if err != nil {
		return nil, err
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// OrderDetail 订单详情（归属校验）
func OrderDetail(sc *svc.ServiceContext, memberID int64, orderNo string) (*types.OrderView, error) {
	var o model.Order
	if err := sc.DB.Where("order_no = ? AND member_id = ?", orderNo, memberID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	views, err := BuildOrderViews(sc, []model.Order{o})
	if err != nil {
		return nil, err
	}
	return views[0], nil
}

// CancelOrder 取消订单（仅待支付），事务内关单 + 释放商品
func CancelOrder(sc *svc.ServiceContext, memberID int64, orderNo, reason string) error {
	var o model.Order
	if err := sc.DB.Where("order_no = ? AND member_id = ?", orderNo, memberID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNotFound
		}
		return err
	}
	return closeOrder(sc, &o, model.OrderCanceled, reason, true)
}

// ConfirmOrder 确认完成（已支付 → 已完成）
func ConfirmOrder(sc *svc.ServiceContext, memberID int64, orderNo string) error {
	now := time.Now()
	res := sc.DB.Model(&model.Order{}).
		Where("order_no = ? AND member_id = ? AND status = ?", orderNo, memberID, model.OrderPaid).
		Updates(map[string]any{"status": model.OrderCompleted, "completed_at": &now, "updated_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrOrderState
	}
	return nil
}

// BuildOrderViews 批量组装订单视图
func BuildOrderViews(sc *svc.ServiceContext, orders []model.Order) ([]*types.OrderView, error) {
	if len(orders) == 0 {
		return []*types.OrderView{}, nil
	}
	orderIDs := make([]int64, 0, len(orders))
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
	}
	var items []model.OrderItem
	if err := sc.DB.Where("order_id IN ?", orderIDs).Find(&items).Error; err != nil {
		return nil, err
	}
	itemMap := map[int64][]types.OrderItemView{}
	for _, it := range items {
		itemMap[it.OrderID] = append(itemMap[it.OrderID], types.OrderItemView{
			ProductID:    strconv.FormatInt(it.ProductID, 10),
			ProductTitle: it.ProductTitle,
			ProductImage: it.ProductImage,
			BreedName:    it.BreedName,
			Price:        it.Price.StringFixed(2),
			Quantity:     it.Quantity,
		})
	}
	views := make([]*types.OrderView, 0, len(orders))
	for _, o := range orders {
		v := &types.OrderView{
			OrderNo:      o.OrderNo,
			Status:       o.Status,
			StatusText:   model.OrderStatusText(o.Status),
			TotalAmount:  o.TotalAmount.StringFixed(2),
			PayAmount:    o.PayAmount.StringFixed(2),
			ContactName:  o.ContactName,
			ContactPhone: o.ContactPhone,
			Remark:       o.Remark,
			ExpireAt:     o.ExpireAt.Format(time.RFC3339),
			CreatedAt:    o.CreatedAt.Format(time.RFC3339),
			Items:        itemMap[o.ID],
		}
		if v.Items == nil {
			v.Items = []types.OrderItemView{}
		}
		if o.PaidAt != nil {
			v.PaidAt = o.PaidAt.Format(time.RFC3339)
		}
		views = append(views, v)
	}
	return views, nil
}

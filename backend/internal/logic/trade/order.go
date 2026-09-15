package trade

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/marketing"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func payTTL(sc *svc.ServiceContext) time.Duration {
	minutes := sc.Config.Trade.PayTimeoutMinutes
	if minutes <= 0 {
		minutes = 15
	}
	return time.Duration(minutes) * time.Minute
}

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, common.ErrParam
	}
	return id, nil
}

func strconvSingle(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	return parseID(s)
}

func strconvSlice(ss []string) ([]int64, error) {
	if len(ss) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0, len(ss))
	for _, s := range ss {
		id, err := parseID(s)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
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

	// 增值服务：必须全部启用，汇总服务费
	serviceIDs, err := strconvSlice(req.ServiceIDs)
	if err != nil {
		releaseProduct(sc.DB, productID)
		return nil, err
	}
	serviceFee, svcItems, err := marketing.LoadOrderServices(sc, serviceIDs)
	if err != nil {
		releaseProduct(sc.DB, productID)
		return nil, err
	}

	// 金额计算：total = 商品价 + 服务费；pay = total - 券抵扣（保底 0.01）+ 运费（券不抵运费）
	couponID, err := strconvSingle(req.CouponID)
	if err != nil {
		releaseProduct(sc.DB, productID)
		return nil, err
	}
	base := p.Price.Add(serviceFee)
	discount := decimal.Zero
	var couponName string
	if couponID > 0 {
		var mc model.MemberCoupon
		if err := sc.DB.Where("id = ? AND member_id = ?", couponID, memberID).First(&mc).Error; err != nil {
			releaseProduct(sc.DB, productID)
			return nil, common.ErrCouponUnusable
		}
		var tpl model.CouponTemplate
		if err := sc.DB.First(&tpl, mc.TemplateID).Error; err != nil {
			releaseProduct(sc.DB, productID)
			return nil, common.ErrCouponUnusable
		}
		if mc.Status != model.CouponUsable {
			releaseProduct(sc.DB, productID)
			return nil, common.ErrCouponUnusable
		}
		if err := marketing.ValidateUsable(&tpl, base); err != nil {
			releaseProduct(sc.DB, productID)
			return nil, err
		}
		discount = marketing.CalcDiscount(&tpl, base)
		couponName = tpl.Name
	}

	// 配送方式：必须启用；托运配送类必填收货地址
	shipMethodID, err := parseID(req.ShipMethodID)
	if err != nil {
		releaseProduct(sc.DB, productID)
		return nil, err
	}
	var sm model.ShipMethod
	if err := sc.DB.Where("id = ? AND status = ?", shipMethodID, model.ShipMethodOn).First(&sm).Error; err != nil {
		releaseProduct(sc.DB, productID)
		return nil, common.NewErr(400, 40002, "配送方式不可用")
	}
	shipAddress := strings.TrimSpace(req.ShipAddress)
	if sm.Kind == model.KindShip {
		if shipAddress == "" {
			releaseProduct(sc.DB, productID)
			return nil, common.NewErr(400, 40003, "请填写收货地址")
		}
	}
	if rs := []rune(shipAddress); len(rs) > 255 {
		shipAddress = string(rs[:255])
	}

	payAmount := base.Sub(discount)
	if payAmount.LessThan(decimal.NewFromFloat(0.01)) {
		payAmount = decimal.NewFromFloat(0.01)
	}
	payAmount = payAmount.Add(sm.Fee)

	now := time.Now()
	order := model.Order{
		ID:             common.NewID(),
		OrderNo:        common.NewBizNo("P"),
		MemberID:       memberID,
		TotalAmount:    base,
		DiscountAmount: discount,
		ServiceFee:     serviceFee,
		PayAmount:      payAmount,
		CouponID:       couponID,
		CouponInfo:     couponName,
		ServiceItems:   marketing.ServiceSnapshotJSON(svcItems),
		Status:         model.OrderPending,
		ContactName:    req.ContactName,
		ContactPhone:   req.ContactPhone,
		Remark:         req.Remark,
		ShipMethodID:   sm.ID,
		ShipMethodName: sm.Name,
		ShipFee:        sm.Fee,
		ShipAddress:    shipAddress,
		ExpireAt:       now.Add(payTTL(sc)),
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
		Amount:    payAmount,
		Channel:   model.PayChannelMini,
		PayType:   model.PayTypePurchase,
		Status:    model.PayStatusPending,
	}

	err = sc.DB.Transaction(func(tx *gorm.DB) error {
		// 用券：事务内原子锁定（并发用同一张券只有一单成功）
		if couponID > 0 {
			res := tx.Exec(
				`UPDATE member_coupon SET status = ?, order_id = ?
				 WHERE id = ? AND member_id = ? AND status = ?
				   AND NOT EXISTS (SELECT 1 FROM coupon_template t WHERE t.id = member_coupon.template_id
				                   AND t.valid_end IS NOT NULL AND t.valid_end < now())`,
				model.CouponLocked, order.ID, couponID, memberID, model.CouponUsable)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return common.ErrCouponUnusable
			}
		}
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
	orderNos := make([]string, 0, len(orders))
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
		orderNos = append(orderNos, o.OrderNo)
	}
	// 入口态：已评价 / 最新售后状态
	reviewedSet := map[string]struct{}{}
	var reviewed []model.OrderReview
	if err := sc.DB.Select("order_no").Where("order_no IN ?", orderNos).Find(&reviewed).Error; err != nil {
		return nil, err
	}
	for _, r := range reviewed {
		reviewedSet[r.OrderNo] = struct{}{}
	}
	aftersaleMap := map[string]int{}
	var afters []model.AfterSale
	if err := sc.DB.Select("order_no", "status").Where("order_no IN ?", orderNos).
		Order("id DESC").Find(&afters).Error; err != nil {
		return nil, err
	}
	for _, a := range afters {
		if _, ok := aftersaleMap[a.OrderNo]; !ok { // id DESC，首次出现即最新
			aftersaleMap[a.OrderNo] = a.Status
		}
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
		_, isReviewed := reviewedSet[o.OrderNo]
		v := &types.OrderView{
			OrderNo:         o.OrderNo,
			Status:          o.Status,
			StatusText:      model.OrderStatusText(o.Status),
			TotalAmount:     o.TotalAmount.StringFixed(2),
			DiscountAmount:  o.DiscountAmount.StringFixed(2),
			ServiceFee:      o.ServiceFee.StringFixed(2),
			ShipFee:         o.ShipFee.StringFixed(2),
			PayAmount:       o.PayAmount.StringFixed(2),
			CouponInfo:      o.CouponInfo,
			ServiceItems:    o.ServiceItems,
			ContactName:     o.ContactName,
			ContactPhone:    o.ContactPhone,
			Remark:          o.Remark,
			ShipMethod:      o.ShipMethodName,
			ShipAddress:     o.ShipAddress,
			ShipStatus:      o.ShipStatus,
			ShipNo:          o.ShipNo,
			ExpireAt:        o.ExpireAt.Format(time.RFC3339),
			CreatedAt:       o.CreatedAt.Format(time.RFC3339),
			Items:           itemMap[o.ID],
			Reviewed:        isReviewed,
			AftersaleStatus: aftersaleMap[o.OrderNo],
		}
		if v.Items == nil {
			v.Items = []types.OrderItemView{}
		}
		if o.PaidAt != nil {
			v.PaidAt = o.PaidAt.Format(time.RFC3339)
		}
		if o.ShippedAt != nil {
			v.ShippedAt = o.ShippedAt.Format(time.RFC3339)
		}
		if o.DeliveredAt != nil {
			v.DeliveredAt = o.DeliveredAt.Format(time.RFC3339)
		}
		if o.CompletedAt != nil {
			v.CompletedAt = o.CompletedAt.Format(time.RFC3339)
		}
		views = append(views, v)
	}
	return views, nil
}

// pointsshop.go 积分商城：券 / 实物 / 免运费卡三类兑换，实物走 points_order 履约。
package marketing

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/growth"
	"pet/backend/internal/logic/risk"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func pointsTypeText(t int) string {
	switch t {
	case model.PointsProductCoupon:
		return "优惠券"
	case model.PointsProductGoods:
		return "实物"
	case model.PointsProductFreeShip:
		return "免运费卡"
	}
	return "权益"
}

func pointsProductView(p model.PointsProduct, couponName string) types.PointsProductView {
	return types.PointsProductView{
		ID:               strconv.FormatInt(p.ID, 10),
		Name:             p.Name,
		Image:            p.Image,
		PointsCost:       p.PointsCost,
		Stock:            p.Stock,
		Type:             p.Type,
		TypeText:         pointsTypeText(p.Type),
		CouponTemplateID: strconv.FormatInt(p.CouponTemplateID, 10),
		CouponName:       couponName,
		Description:      p.Description,
	}
}

// PointsShop 商城列表（公开）
func PointsShop(sc *svc.ServiceContext) (*types.PointsShopResp, error) {
	var products []model.PointsProduct
	if err := sc.DB.Where("status = ?", 1).Order("sort ASC, id ASC").Limit(50).Find(&products).Error; err != nil {
		return nil, err
	}
	tplNames := map[int64]string{}
	tplIDs := make([]int64, 0, len(products))
	for _, p := range products {
		if p.Type == model.PointsProductCoupon && p.CouponTemplateID > 0 {
			tplIDs = append(tplIDs, p.CouponTemplateID)
		}
	}
	if len(tplIDs) > 0 {
		var tpls []model.CouponTemplate
		if err := sc.DB.Select("id", "name").Where("id IN ?", tplIDs).Find(&tpls).Error; err == nil {
			for _, t := range tpls {
				tplNames[t.ID] = t.Name
			}
		}
	}
	list := make([]types.PointsProductView, 0, len(products))
	for _, p := range products {
		list = append(list, pointsProductView(p, tplNames[p.CouponTemplateID]))
	}
	return &types.PointsShopResp{List: list}, nil
}

// PointsExchange 积分兑换（券/实物/免运费卡），积分扣减与库存 CAS 同事务
func PointsExchange(sc *svc.ServiceContext, memberID int64, productID int64, req *types.PointsExchangeReq) error {
	if err := risk.CheckTradeAction(sc, memberID, "积分兑换"); err != nil {
		return err
	}
	var p model.PointsProduct
	if err := sc.DB.Where("id = ? AND status = 1 AND stock > 0", productID).First(&p).Error; err != nil {
		return common.ErrPointsProduct
	}
	err := sc.DB.Transaction(func(tx *gorm.DB) error {
		// 扣库存（CAS）
		res := tx.Exec(
			"UPDATE points_product SET stock = stock - 1, updated_at = now() WHERE id = ? AND stock > 0",
			productID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return common.ErrPointsProduct
		}
		// 扣积分
		if err := growth.DeductTx(tx, memberID, int64(p.PointsCost),
			"points_shop", strconv.FormatInt(productID, 10)); err != nil {
			return err
		}
		switch p.Type {
		case model.PointsProductCoupon:
			// 发券（模板校验在券模板行上重做一次）
			var tpl model.CouponTemplate
			if err := tx.First(&tpl, p.CouponTemplateID).Error; err != nil {
				return common.ErrPointsProduct
			}
			return tx.Create(&model.MemberCoupon{
				ID: common.NewID(), MemberID: memberID, TemplateID: tpl.ID,
				Status: model.CouponUsable, ReceivedAt: time.Now(),
			}).Error
		case model.PointsProductFreeShip:
			return tx.Exec(
				"UPDATE member SET free_ship_cards = free_ship_cards + 1, updated_at = now() WHERE id = ?",
				memberID).Error
		case model.PointsProductGoods:
			contact := strings.TrimSpace(req.Contact)
			phone := strings.TrimSpace(req.Phone)
			address := strings.TrimSpace(req.Address)
			if contact == "" || phone == "" || address == "" {
				return common.NewErr(400, 40001, "实物兑换请填写收货人与地址")
			}
			return tx.Create(&model.PointsOrder{
				ID: common.NewID(), OrderNo: common.NewBizNo("PS"),
				MemberID: memberID, ProductID: p.ID,
				ProductName: p.Name, Image: p.Image, PointsCost: p.PointsCost,
				Status:  model.PointsOrderPending,
				Contact: contact, Phone: phone, Address: address,
			}).Error
		}
		return nil
	})
	if err != nil {
		// 失败回滚库存（事务自动回滚，无需手动）
		return err
	}
	return nil
}

// MyPointsOrders 我的积分订单
func MyPointsOrders(sc *svc.ServiceContext, memberID int64) (*types.PointsOrderListResp, error) {
	var orders []model.PointsOrder
	if err := sc.DB.Where("member_id = ?", memberID).Order("id DESC").Limit(50).Find(&orders).Error; err != nil {
		return nil, err
	}
	list := make([]types.PointsOrderView, 0, len(orders))
	for _, o := range orders {
		list = append(list, types.PointsOrderView{
			ID: strconv.FormatInt(o.ID, 10), OrderNo: o.OrderNo,
			ProductName: o.ProductName, Image: o.Image, PointsCost: o.PointsCost,
			Status: o.Status, StatusText: pointsOrderStatusText(o.Status),
			ShipNo: o.ShipNo, Contact: o.Contact, Phone: o.Phone, Address: o.Address,
			CreatedAt: o.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &types.PointsOrderListResp{List: list}, nil
}

func pointsOrderStatusText(s int) string {
	switch s {
	case model.PointsOrderPending:
		return "待发货"
	case model.PointsOrderShipped:
		return "已发货"
	case model.PointsOrderCompleted:
		return "已完成"
	}
	return "未知"
}

// ─────────────────────────── 平台端 ───────────────────────────

func AdminPointsProducts(sc *svc.ServiceContext) (*types.PointsShopResp, error) {
	var products []model.PointsProduct
	if err := sc.DB.Order("sort ASC, id ASC").Limit(100).Find(&products).Error; err != nil {
		return nil, err
	}
	list := make([]types.PointsProductView, 0, len(products))
	for _, p := range products {
		list = append(list, pointsProductView(p, ""))
	}
	return &types.PointsShopResp{List: list}, nil
}

func PointsProductUpsert(sc *svc.ServiceContext, req *types.PointsProductUpsertReq) error {
	if strings.TrimSpace(req.Name) == "" || req.PointsCost <= 0 || req.Stock < 0 {
		return common.ErrParam
	}
	var tplID int64
	if req.CouponTemplateID != "" {
		id, err := strconv.ParseInt(req.CouponTemplateID, 10, 64)
		if err != nil || id <= 0 {
			return common.ErrParam
		}
		tplID = id
	}
	status := 1
	if req.Status != nil && *req.Status == 0 {
		status = 0
	}
	if req.ID == "" {
		return sc.DB.Create(&model.PointsProduct{
			ID: common.NewID(), Name: req.Name, Image: req.Image,
			PointsCost: req.PointsCost, Stock: req.Stock, Type: req.Type,
			CouponTemplateID: tplID, Description: req.Description,
			Status: status, Sort: req.Sort,
		}).Error
	}
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil || id <= 0 {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.PointsProduct{}).Where("id = ?", id).Updates(map[string]any{
		"name": req.Name, "image": req.Image, "points_cost": req.PointsCost,
		"stock": req.Stock, "type": req.Type, "coupon_template_id": tplID,
		"description": req.Description, "status": status, "sort": req.Sort,
		"updated_at": time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func AdminPointsOrders(sc *svc.ServiceContext, req *types.PageReq) (*types.PointsOrderListResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	var total int64
	if err := sc.DB.Model(&model.PointsOrder{}).Count(&total).Error; err != nil {
		return nil, err
	}
	var orders []model.PointsOrder
	if err := sc.DB.Order("id DESC").Offset((page - 1) * size).Limit(size).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	list := make([]types.PointsOrderView, 0, len(orders))
	for _, o := range orders {
		list = append(list, types.PointsOrderView{
			ID: strconv.FormatInt(o.ID, 10), OrderNo: o.OrderNo,
			ProductName: o.ProductName, Image: o.Image, PointsCost: o.PointsCost,
			Status: o.Status, StatusText: pointsOrderStatusText(o.Status),
			ShipNo: o.ShipNo, Contact: o.Contact, Phone: o.Phone, Address: o.Address,
			CreatedAt: o.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &types.PointsOrderListResp{List: list}, nil
}

func PointsOrderShip(sc *svc.ServiceContext, id int64, shipNo string) error {
	shipNo = strings.TrimSpace(shipNo)
	if shipNo == "" {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.PointsOrder{}).
		Where("id = ? AND status = ?", id, model.PointsOrderPending).
		Updates(map[string]any{"status": model.PointsOrderShipped, "ship_no": shipNo, "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrOrderState
	}
	return nil
}

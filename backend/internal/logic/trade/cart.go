// Package trade 购物车：加购/勾选/删除 + 批量合并结算。
package trade

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/marketing"
	"pet/backend/internal/metrics"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// CartList 我的购物车（含商品实时状态；下架商品置灰不可结算）
func CartList(sc *svc.ServiceContext, memberID int64) (*types.CartListResp, error) {
	var carts []model.Cart
	if err := sc.DB.Where("member_id = ?", memberID).Order("updated_at DESC").Find(&carts).Error; err != nil {
		return nil, err
	}
	resp := &types.CartListResp{List: []types.CartView{}}
	if len(carts) == 0 {
		return resp, nil
	}
	productIDs := make([]int64, 0, len(carts))
	for _, c := range carts {
		productIDs = append(productIDs, c.ProductID)
	}
	var products []model.PetProduct
	if err := sc.DB.Where("id IN ?", productIDs).Find(&products).Error; err != nil {
		return nil, err
	}
	prodMap := map[int64]model.PetProduct{}
	for _, p := range products {
		prodMap[p.ID] = p
	}
	skuMap := map[int64]model.ProductSku{}
	skuIDs := make([]int64, 0, len(carts))
	for _, c := range carts {
		if c.SkuID > 0 {
			skuIDs = append(skuIDs, c.SkuID)
		}
	}
	if len(skuIDs) > 0 {
		var skus []model.ProductSku
		if err := sc.DB.Where("id IN ?", skuIDs).Find(&skus).Error; err != nil {
			return nil, err
		}
		for _, s := range skus {
			skuMap[s.ID] = s
		}
	}
	checked := 0
	for _, c := range carts {
		p := prodMap[c.ProductID]
		price := p.Price
		specs := ""
		if s, ok := skuMap[c.SkuID]; ok && s.Status == 1 {
			price = s.Price
			specs = s.Specs
		}
		onSale := p.Status == model.ProductOnSale
		if c.Checked == 1 && onSale {
			checked++
		}
		resp.List = append(resp.List, types.CartView{
			ID:           strconv.FormatInt(c.ID, 10),
			ProductID:    strconv.FormatInt(c.ProductID, 10),
			ProductTitle: p.Title,
			ProductImage: p.MainImage,
			Price:        price.StringFixed(2),
			OrigPrice:    p.Price.StringFixed(2),
			SkuID:        strconv.FormatInt(c.SkuID, 10),
			SkuSpecs:     specs,
			Checked:      c.Checked,
			OnSale:       onSale,
			CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		})
	}
	resp.CheckedCount = checked
	return resp, nil
}

// CartAdd 加入购物车（同商品同规格幂等：重复加购回到勾选态）
func CartAdd(sc *svc.ServiceContext, memberID int64, req *types.CartAddReq) error {
	productID, err := strconv.ParseInt(req.ProductID, 10, 64)
	if err != nil || productID <= 0 {
		return common.ErrParam
	}
	var skuID int64
	if req.SkuID != "" {
		if skuID, err = strconv.ParseInt(req.SkuID, 10, 64); err != nil || skuID <= 0 {
			return common.ErrParam
		}
		var sku model.ProductSku
		if err := sc.DB.Where("id = ? AND product_id = ? AND status = ?", skuID, productID, 1).
			First(&sku).Error; err != nil {
			return common.ErrSkuInvalid
		}
	}
	var p model.PetProduct
	if err := sc.DB.First(&p, productID).Error; err != nil {
		return common.ErrNotFound
	}
	if p.Status != model.ProductOnSale {
		return common.ErrGoodsOffline
	}
	var exist model.Cart
	err = sc.DB.Where("member_id = ? AND product_id = ?", memberID, productID).First(&exist).Error
	if err == nil {
		return sc.DB.Model(&exist).Updates(map[string]any{"sku_id": skuID, "checked": 1, "updated_at": time.Now()}).Error
	} else if err != gorm.ErrRecordNotFound {
		return err
	}
	return sc.DB.Create(&model.Cart{
		ID:        common.NewID(),
		MemberID:  memberID,
		ProductID: productID,
		SkuID:     skuID,
		Checked:   1,
	}).Error
}

// CartUpdate 勾选/取消勾选
func CartUpdate(sc *svc.ServiceContext, memberID, cartID int64, checked *int) error {
	updates := map[string]any{"updated_at": time.Now()}
	if checked != nil {
		v := 0
		if *checked != 0 {
			v = 1
		}
		updates["checked"] = v
	}
	res := sc.DB.Model(&model.Cart{}).
		Where("id = ? AND member_id = ?", cartID, memberID).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// CartDelete 删除单个；cartID<=0 清空
func CartDelete(sc *svc.ServiceContext, memberID, cartID int64) error {
	q := sc.DB.Where("member_id = ?", memberID)
	if cartID > 0 {
		q = q.Where("id = ?", cartID)
	}
	return q.Delete(&model.Cart{}).Error
}

// createCartOrder 购物车批量结算：多商品一单（无增值服务/定金；券与等级折扣按整单生效）
func createCartOrder(sc *svc.ServiceContext, memberID int64, req *types.CreateOrderReq) (*types.CreateOrderResp, error) {
	if err := checkAgreement(req.Agree); err != nil {
		return nil, err
	}
	if req.ContactName == "" || req.ContactPhone == "" {
		return nil, common.NewErr(400, 40001, "请填写联系人信息")
	}
	cartIDs := make([]int64, 0, len(req.CartIds))
	for _, s := range req.CartIds {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil || id <= 0 {
			return nil, common.ErrParam
		}
		cartIDs = append(cartIDs, id)
	}
	if len(cartIDs) == 0 || len(cartIDs) > 20 {
		return nil, common.ErrCartEmpty
	}
	var carts []model.Cart
	if err := sc.DB.Where("id IN ? AND member_id = ?", cartIDs, memberID).Find(&carts).Error; err != nil {
		return nil, err
	}
	if len(carts) == 0 {
		return nil, common.ErrCartEmpty
	}
	productIDs := make([]int64, 0, len(carts))
	for _, c := range carts {
		productIDs = append(productIDs, c.ProductID)
	}
	// 批量原子占位：任一商品不可锁则全部失败
	res := sc.DB.Exec(
		"UPDATE pet_product SET status = ?, updated_at = now() WHERE id IN ? AND status = ?",
		model.ProductLocked, productIDs, model.ProductOnSale)
	if res.Error != nil {
		return nil, res.Error
	}
	if int(res.RowsAffected) != len(uniqueInt64(productIDs)) {
		sc.DB.Exec("UPDATE pet_product SET status = ?, updated_at = now() WHERE id IN ? AND status = ?",
			model.ProductOnSale, productIDs, model.ProductLocked)
		return nil, common.ErrGoodsLocked
	}
	var products []model.PetProduct
	if err := sc.DB.Where("id IN ?", productIDs).Find(&products).Error; err != nil {
		releaseProducts(sc.DB, productIDs)
		return nil, err
	}
	prodMap := map[int64]model.PetProduct{}
	for _, p := range products {
		prodMap[p.ID] = p
	}
	// 秒杀：逐商品尝试占名额，失败统一回补
	flashByProduct := map[int64]*model.FlashSale{}
	for _, c := range carts {
		if fs := ActiveFlashSaleOf(sc, c.ProductID); fs != nil && ClaimFlashSale(sc.DB, fs.ID) {
			flashByProduct[c.ProductID] = fs
		}
	}
	fail := func(err error) (*types.CreateOrderResp, error) {
		for _, fs := range flashByProduct {
			ReleaseFlashSale(sc.DB, fs.ID)
		}
		releaseProducts(sc.DB, productIDs)
		return nil, err
	}
	// 行项目：规格价/秒杀价优先
	items := make([]model.OrderItem, 0, len(carts))
	base := decimal.Zero
	for _, c := range carts {
		p := prodMap[c.ProductID]
		if p.Status != model.ProductLocked {
			return fail(common.ErrGoodsOffline)
		}
		price := p.Price
		specs := ""
		if c.SkuID > 0 {
			var sku model.ProductSku
			if err := sc.DB.Where("id = ? AND product_id = ? AND status = ?", c.SkuID, c.ProductID, 1).
				First(&sku).Error; err != nil {
				return fail(common.ErrSkuInvalid)
			}
			price = sku.Price
			specs = sku.Specs
		}
		if fs, ok := flashByProduct[c.ProductID]; ok {
			price = fs.SalePrice
		}
		base = base.Add(price)
		var breed model.Breed
		_ = sc.DB.First(&breed, p.BreedID).Error
		items = append(items, model.OrderItem{
			ID:           common.NewID(),
			ProductID:    p.ID,
			ProductTitle: p.Title,
			ProductImage: p.MainImage,
			BreedName:    breed.Name,
			Price:        price,
			Quantity:     1,
			SkuSpecs:     specs,
		})
	}
	// 配送方式 + 地址
	shipMethodID, err := strconv.ParseInt(req.ShipMethodID, 10, 64)
	if err != nil || shipMethodID <= 0 {
		return fail(common.ErrParam)
	}
	var sm model.ShipMethod
	if err := sc.DB.Where("id = ? AND status = ?", shipMethodID, model.ShipMethodOn).First(&sm).Error; err != nil {
		return fail(common.NewErr(400, 40002, "配送方式不可用"))
	}
	shipAddress := strings.TrimSpace(req.ShipAddress)
	if req.AddressID != "" {
		if addrID, perr := strconv.ParseInt(req.AddressID, 10, 64); perr == nil {
			var addr model.MemberAddress
			if err := sc.DB.Where("id = ? AND member_id = ?", addrID, memberID).First(&addr).Error; err == nil {
				shipAddress = addr.Address
			}
		}
	}
	if sm.Kind == model.KindShip && shipAddress == "" {
		return fail(common.NewErr(400, 40003, "请填写收货地址"))
	}
	if rs := []rune(shipAddress); len(rs) > 255 {
		shipAddress = string(rs[:255])
	}
	// 优惠券（整单）
	discount := decimal.Zero
	couponID := int64(0)
	couponName := ""
	if req.CouponID != "" {
		if couponID, err = strconv.ParseInt(req.CouponID, 10, 64); err != nil || couponID <= 0 {
			return fail(common.ErrParam)
		}
		var mc model.MemberCoupon
		if err := sc.DB.Where("id = ? AND member_id = ?", couponID, memberID).First(&mc).Error; err != nil {
			return fail(common.ErrCouponUnusable)
		}
		var tpl model.CouponTemplate
		if err := sc.DB.First(&tpl, mc.TemplateID).Error; err != nil {
			return fail(common.ErrCouponUnusable)
		}
		if mc.Status != model.CouponUsable {
			return fail(common.ErrCouponUnusable)
		}
		if err := marketing.ValidateUsable(&tpl, base); err != nil {
			return fail(err)
		}
		discount = marketing.CalcDiscount(&tpl, base)
		couponName = tpl.Name
	}
	// 等级折扣
	afterCoupon := base.Sub(discount)
	if afterCoupon.LessThan(decimal.NewFromFloat(0.01)) {
		afterCoupon = decimal.NewFromFloat(0.01)
	}
	goodsPay := afterCoupon
	levelDiscount := decimal.Zero
	if rate := memberLevelRate(sc, memberID); rate < 1 {
		goodsPay = afterCoupon.Mul(decimal.NewFromFloat(rate)).Round(2)
		levelDiscount = afterCoupon.Sub(goodsPay)
	}
	payAmount := goodsPay.Add(sm.Fee)

	// 免运费卡
	useFreeShip := freeShipApplies(sc, memberID, req.UseFreeShip, sm.Fee)
	shipFee := sm.Fee
	if useFreeShip {
		shipFee = decimal.Zero
		payAmount = goodsPay
	}

	// 自提门店解析
	var pickupStore *model.Store
	if sm.Kind == model.KindPickup {
		pickupStore, err = resolvePickupStore(sc, req.StoreID)
		if err != nil {
			return fail(err)
		}
		if shipAddress == "" {
			shipAddress = resolveStoreAddress(pickupStore)
		}
	}

	now := time.Now()
	order := model.Order{
		ID:             common.NewID(),
		OrderNo:        common.NewBizNo("P"),
		MemberID:       memberID,
		TotalAmount:    base,
		DiscountAmount: discount,
		LevelDiscount:  levelDiscount,
		PayAmount:      payAmount,
		CouponID:       couponID,
		CouponInfo:     couponName,
		Status:         model.OrderPending,
		ContactName:    req.ContactName,
		ContactPhone:   req.ContactPhone,
		Remark:         req.Remark,
		ShipMethodID:   sm.ID,
		ShipMethodName: sm.Name,
		ShipFee:        shipFee,
		ShipAddress:    shipAddress,
		StoreID:        storeIDOf(pickupStore),
		ExpireAt:       now.Add(payTTL(sc)),
	}
	stampAgreement(&order)
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
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = order.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}
		if useFreeShip {
			if err := consumeFreeShipTx(tx, memberID); err != nil {
				return err
			}
		}
		return tx.Where("id IN ? AND member_id = ?", cartIDs, memberID).Delete(&model.Cart{}).Error
	})
	if err != nil {
		return fail(err)
	}
	metrics.OrdersCreated.Inc()
	resp := &types.CreateOrderResp{
		OrderNo:   order.OrderNo,
		PaymentNo: payment.PaymentNo,
		PayAmount: payAmount.StringFixed(2),
		ExpireAt:  order.ExpireAt.Format(time.RFC3339),
	}
	payParams, _ := prepayOrder(sc, memberID, &order)
	resp.PayParams = payParams
	return resp, nil
}

func releaseProducts(db *gorm.DB, productIDs []int64) {
	db.Exec("UPDATE pet_product SET status = ?, updated_at = now() WHERE id IN ? AND status = ?",
		model.ProductOnSale, productIDs, model.ProductLocked)
}

func uniqueInt64(ids []int64) []int64 {
	seen := map[int64]struct{}{}
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// detailImagesOf 解析商品详情长图 JSON
func detailImagesOf(raw string) []string {
	images := []string{}
	_ = json.Unmarshal([]byte(raw), &images)
	return images
}

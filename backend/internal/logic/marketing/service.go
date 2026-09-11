package marketing

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// ─────────────────────────── 增值服务 ───────────────────────────

func ServiceList(sc *svc.ServiceContext) ([]types.ServiceItemView, error) {
	var items []model.ServiceItem
	if err := sc.DB.Where("status = 1").Order("sort ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return serviceViews(items), nil
}

func AdminServiceList(sc *svc.ServiceContext) ([]types.ServiceItemView, error) {
	var items []model.ServiceItem
	if err := sc.DB.Order("sort ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return serviceViews(items), nil
}

func AdminServiceUpsert(sc *svc.ServiceContext, req *types.ServiceUpsertReq) error {
	if req.Name == "" {
		return common.ErrParam
	}
	price, err := parseDecimal(req.Price)
	if err != nil || price.LessThanOrEqual(decimal.Zero) {
		return common.ErrParam
	}
	orig, err := parseDecimal(req.OriginalPrice)
	if err != nil {
		return common.ErrParam
	}
	if orig.LessThan(price) {
		orig = price
	}
	item := model.ServiceItem{
		Name:          req.Name,
		Description:   req.Description,
		OriginalPrice: orig,
		Price:         price,
		Sort:          req.Sort,
		Status:        normStatus(req.Status),
	}
	if req.ID == "" {
		item.ID = common.NewID()
		return sc.DB.Create(&item).Error
	}
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil || id <= 0 {
		return common.ErrParam
	}
	item.ID = id
	res := sc.DB.Model(&model.ServiceItem{}).Where("id = ?", id).Updates(map[string]any{
		"name": item.Name, "description": item.Description,
		"original_price": item.OriginalPrice, "price": item.Price,
		"sort": item.Sort, "status": item.Status, "updated_at": time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func AdminServiceDelete(sc *svc.ServiceContext, id int64) error {
	res := sc.DB.Delete(&model.ServiceItem{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func serviceViews(items []model.ServiceItem) []types.ServiceItemView {
	list := make([]types.ServiceItemView, 0, len(items))
	for _, it := range items {
		list = append(list, types.ServiceItemView{
			ID:            formatID(it.ID),
			Name:          it.Name,
			Description:   it.Description,
			OriginalPrice: it.OriginalPrice.StringFixed(2),
			Price:         it.Price.StringFixed(2),
		})
	}
	return list
}

// ─────────────────────────── 下单金额计算 ───────────────────────────

// LoadOrderServices 校验并加载下单勾选的服务项（必须全部启用），返回服务费合计
func LoadOrderServices(sc *svc.ServiceContext, serviceIDs []int64) (fee decimal.Decimal, items []model.ServiceItem, err error) {
	fee = decimal.Zero
	if len(serviceIDs) == 0 {
		return
	}
	if err = sc.DB.Where("id IN ? AND status = 1", serviceIDs).Order("sort ASC, id ASC").Find(&items).Error; err != nil {
		return
	}
	if len(items) != len(serviceIDs) {
		err = common.ErrServiceInvalid
		return
	}
	for _, it := range items {
		fee = fee.Add(it.Price)
	}
	return
}

// ServiceSnapshotJSON 订单服务快照 [{id,name,price}]
func ServiceSnapshotJSON(items []model.ServiceItem) string {
	if len(items) == 0 {
		return ""
	}
	type snap struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Price string `json:"price"`
	}
	list := make([]snap, 0, len(items))
	for _, it := range items {
		list = append(list, snap{ID: it.ID, Name: it.Name, Price: it.Price.StringFixed(2)})
	}
	b, err := json.Marshal(list)
	if err != nil {
		return ""
	}
	return string(b)
}

// ─────────────────────────── 发放（管理端/注册赠送） ───────────────────────────

// Issue 单张发放：校验库存与个人限领，不走领券时段限制
func Issue(sc *svc.ServiceContext, memberID, templateID int64) error {
	return sc.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Exec(
			`UPDATE coupon_template SET issued_count = issued_count + 1, updated_at = now()
			 WHERE id = ? AND status = 1 AND (total_count = 0 OR issued_count < total_count)`, templateID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return common.ErrCouponSoldOut
		}
		var cnt int64
		if err := tx.Model(&model.MemberCoupon{}).
			Where("member_id = ? AND template_id = ?", memberID, templateID).
			Count(&cnt).Error; err != nil {
			return err
		}
		var t model.CouponTemplate
		if err := tx.First(&t, templateID).Error; err != nil {
			return err
		}
		if t.PerLimit > 0 && cnt >= int64(t.PerLimit) {
			return common.ErrCouponPerLimit
		}
		return tx.Create(&model.MemberCoupon{
			ID:         common.NewID(),
			MemberID:   memberID,
			TemplateID: templateID,
			Status:     model.CouponUsable,
			ReceivedAt: time.Now(),
		}).Error
	})
}

// ─────────────────────────── 通用工具 ───────────────────────────

func money(f float64) string { return dec(f).StringFixed(2) }

func dec(f float64) decimal.Decimal { return decimal.NewFromFloat(f) }

func parseDecimal(s string) (decimal.Decimal, error) {
	if s == "" {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(s)
}

func parseTimePtr(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return &t, nil
		}
	}
	return nil, common.ErrParam
}

func normStatus(s int) int {
	if s == 0 {
		return 1
	}
	return s
}

package marketing

import (
	"errors"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/growth"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func formatID(id int64) string { return strconv.FormatInt(id, 10) }

// ParseIDs 字符串 ID 列表 → int64
func ParseIDs(ss []string) ([]int64, error) {
	ids := make([]int64, 0, len(ss))
	for _, s := range ss {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil || id <= 0 {
			return nil, common.ErrParam
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ─────────────────────────── 用户端 ───────────────────────────

// CouponCenter 领券中心：在投放时段内、未发完、未达个人限领的启用模板
func CouponCenter(sc *svc.ServiceContext, memberID int64) ([]types.CouponTemplateView, error) {
	now := time.Now()
	var templates []model.CouponTemplate
	if err := sc.DB.
		Where("status = 1 AND (pickup_start IS NULL OR pickup_start <= ?) AND (pickup_end IS NULL OR pickup_end >= ?)", now, now).
		Where("(total_count = 0 OR issued_count < total_count)").
		Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, err
	}

	// 个人限领过滤
	received := map[int64]int64{}
	if len(templates) > 0 {
		tids := make([]int64, 0, len(templates))
		for _, t := range templates {
			tids = append(tids, t.ID)
		}
		var rows []struct {
			TemplateID int64
			Cnt        int64
		}
		if err := sc.DB.Model(&model.MemberCoupon{}).
			Select("template_id, COUNT(*) AS cnt").
			Where("member_id = ? AND template_id IN ?", memberID, tids).
			Group("template_id").Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			received[r.TemplateID] = r.Cnt
		}
	}

	list := make([]types.CouponTemplateView, 0, len(templates))
	for _, t := range templates {
		if t.PerLimit > 0 && received[t.ID] >= int64(t.PerLimit) {
			continue // 已达个人限领，不在领券中心露出
		}
		list = append(list, templateView(t))
	}
	return list, nil
}

// Claim 领取优惠券：积分兑换扣分（如配置）+ CAS 扣减模板库存（并发安全）+ 个人限领校验
func Claim(sc *svc.ServiceContext, memberID, templateID int64) error {
	now := time.Now()
	return sc.DB.Transaction(func(tx *gorm.DB) error {
		var t0 model.CouponTemplate
		if err := tx.First(&t0, templateID).Error; err != nil {
			return common.ErrNotFound
		}
		if t0.PointsCost > 0 {
			// 积分兑换券：先扣分（不足即失败回滚）
			if err := growth.DeductTx(tx, memberID, int64(t0.PointsCost),
				model.PointsReasonExchange, strconv.FormatInt(templateID, 10)); err != nil {
				return err
			}
		}
		res := tx.Exec(
			`UPDATE coupon_template SET issued_count = issued_count + 1, updated_at = ?
			 WHERE id = ? AND status = 1
			   AND (pickup_start IS NULL OR pickup_start <= ?)
			   AND (pickup_end IS NULL OR pickup_end >= ?)
			   AND (total_count = 0 OR issued_count < total_count)`,
			now, templateID, now, now)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			var t model.CouponTemplate
			if err := tx.First(&t, templateID).Error; err != nil {
				return common.ErrNotFound
			}
			if t.Status != 1 || (t.PickupStart != nil && now.Before(*t.PickupStart)) ||
				(t.PickupEnd != nil && now.After(*t.PickupEnd)) {
				return common.ErrCouponNotInTime
			}
			if t.TotalCount > 0 && t.IssuedCount >= t.TotalCount {
				return common.ErrCouponSoldOut
			}
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
			ReceivedAt: now,
		}).Error
	})
}

// MyCoupons 我的券列表；status>0 过滤。惰性过期：可用但已过 valid_end 的置为已过期
func MyCoupons(sc *svc.ServiceContext, memberID int64, status int) ([]types.MyCouponView, error) {
	now := time.Now()
	// 惰性过期
	sc.DB.Exec(
		`UPDATE member_coupon SET status = ? WHERE member_id = ? AND status = ?
		 AND EXISTS (SELECT 1 FROM coupon_template t WHERE t.id = member_coupon.template_id
		             AND t.valid_end IS NOT NULL AND t.valid_end < ?)`,
		model.CouponExpired, memberID, model.CouponUsable, now)

	query := sc.DB.Table("member_coupon AS mc").
		Joins("JOIN coupon_template t ON t.id = mc.template_id").
		Where("mc.member_id = ?", memberID)
	if status > 0 {
		query = query.Where("mc.status = ?", status)
	}
	var rows []struct {
		ID              int64
		TemplateID      int64
		Name            string
		Type            int
		ThresholdAmount float64
		DiscountAmount  float64
		DiscountPercent int
		ValidEnd        *time.Time
		Status          int
		ReceivedAt      time.Time
	}
	if err := query.Select(`
		mc.id, mc.template_id, mc.status, mc.received_at,
		t.name, t.type, t.threshold_amount, t.discount_amount, t.discount_percent, t.valid_end`).
		Order("mc.status ASC, mc.received_at DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	list := make([]types.MyCouponView, 0, len(rows))
	for _, r := range rows {
		v := types.MyCouponView{
			ID:            formatID(r.ID),
			TemplateID:    formatID(r.TemplateID),
			Name:          r.Name,
			Type:          r.Type,
			Threshold:     money(r.ThresholdAmount),
			Discount:      money(r.DiscountAmount),
			DiscountValid: r.Type != model.CouponTypeDiscount,
			Percent:       r.DiscountPercent,
			Status:        r.Status,
			StatusText:    model.CouponStatusText(r.Status),
			ReceivedAt:    r.ReceivedAt.Format(time.RFC3339),
		}
		if r.ValidEnd != nil {
			v.ValidEnd = r.ValidEnd.Format(time.RFC3339)
		}
		list = append(list, v)
	}
	return list, nil
}

// UsableCoupons 某笔消费（商品价+勾选服务费）下可用的券及预估抵扣
func UsableCoupons(sc *svc.ServiceContext, memberID int64, productID int64, serviceIDs []int64) ([]types.UsableCouponView, error) {
	var p model.PetProduct
	if err := sc.DB.First(&p, productID).Error; err != nil {
		return nil, common.ErrNotFound
	}
	_, svcItems, err := LoadOrderServices(sc, serviceIDs)
	if err != nil {
		return nil, err
	}
	base := p.Price
	for _, it := range svcItems {
		base = base.Add(it.Price)
	}
	now := time.Now()
	var rows []struct {
		MCID            int64
		Name            string
		Type            int
		ThresholdAmount float64
		DiscountAmount  float64
		DiscountPercent int
		ValidEnd        *time.Time
	}
	if err := sc.DB.Table("member_coupon AS mc").
		Joins("JOIN coupon_template t ON t.id = mc.template_id").
		Where("mc.member_id = ? AND mc.status = ?", memberID, model.CouponUsable).
		Where("(t.valid_start IS NULL OR t.valid_start <= ?) AND (t.valid_end IS NULL OR t.valid_end >= ?)", now, now).
		Order("mc.received_at DESC").Select(`
		mc.id AS mc_id, t.name, t.type, t.threshold_amount,
		t.discount_amount, t.discount_percent, t.valid_end`).Scan(&rows).Error; err != nil {
		return nil, err
	}

	list := make([]types.UsableCouponView, 0, len(rows))
	for _, r := range rows {
		t := model.CouponTemplate{
			Type: r.Type, ThresholdAmount: dec(r.ThresholdAmount),
			DiscountAmount: dec(r.DiscountAmount), DiscountPercent: r.DiscountPercent,
			ValidEnd: r.ValidEnd, Status: 1,
		}
		if err := ValidateUsable(&t, base); err != nil {
			continue // 不满足门槛/不可用的不露出
		}
		v := types.UsableCouponView{
			ID:            formatID(r.MCID),
			Name:          r.Name,
			Type:          r.Type,
			Threshold:     money(r.ThresholdAmount),
			Discount:      CalcDiscount(&t, base).StringFixed(2),
			DiscountValid: r.Type != model.CouponTypeDiscount,
			Percent:       r.DiscountPercent,
		}
		if r.ValidEnd != nil {
			v.ValidEnd = r.ValidEnd.Format(time.RFC3339)
		}
		list = append(list, v)
	}
	return list, nil
}

// ─────────────────────────── 平台端 ───────────────────────────

func AdminCouponList(sc *svc.ServiceContext, req *types.CouponAdminListReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}
	query := sc.DB.Model(&model.CouponTemplate{})
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		query = query.Where("name ILIKE ?", "%"+req.Keyword+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.CouponTemplate
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, err
	}
	views := make([]types.CouponTemplateView, 0, len(list))
	for _, t := range list {
		views = append(views, templateView(t))
	}
	return &types.PageResp{Total: total, List: views}, nil
}

func AdminCouponUpsert(sc *svc.ServiceContext, req *types.CouponUpsertReq) error {
	if req.Name == "" {
		return common.ErrParam
	}
	t := model.CouponTemplate{
		Name:            req.Name,
		Type:            req.Type,
		DiscountPercent: req.DiscountPercent,
		TotalCount:      req.TotalCount,
		PerLimit:        req.PerLimit,
		NewUserOnly:     req.NewUserOnly,
		PointsCost:      req.PointsCost,
		Status:          normStatus(req.Status),
	}
	if t.Type < 1 || t.Type > 3 {
		return common.ErrParam
	}
	if t.PerLimit < 0 {
		t.PerLimit = 1
	}
	var err error
	if t.ThresholdAmount, err = parseDecimal(req.ThresholdAmount); err != nil {
		return common.ErrParam
	}
	if t.DiscountAmount, err = parseDecimal(req.DiscountAmount); err != nil {
		return common.ErrParam
	}
	if t.MaxDiscountAmount, err = parseDecimal(req.MaxDiscountAmount); err != nil {
		return common.ErrParam
	}
	if t.PickupStart, err = parseTimePtr(req.PickupStart); err != nil {
		return common.ErrParam
	}
	if t.PickupEnd, err = parseTimePtr(req.PickupEnd); err != nil {
		return common.ErrParam
	}
	if t.ValidStart, err = parseTimePtr(req.ValidStart); err != nil {
		return common.ErrParam
	}
	if t.ValidEnd, err = parseTimePtr(req.ValidEnd); err != nil {
		return common.ErrParam
	}

	if req.ID == "" {
		t.ID = common.NewID()
		return sc.DB.Create(&t).Error
	}
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil || id <= 0 {
		return common.ErrParam
	}
	t.ID = id
	// 已发放数量不可编辑回退
	res := sc.DB.Model(&model.CouponTemplate{}).Where("id = ?", id).Omit("IssuedCount").Updates(map[string]any{
		"name": t.Name, "type": t.Type, "threshold_amount": t.ThresholdAmount,
		"discount_amount": t.DiscountAmount, "discount_percent": t.DiscountPercent,
		"max_discount_amount": t.MaxDiscountAmount, "total_count": t.TotalCount,
		"per_limit": t.PerLimit, "new_user_only": t.NewUserOnly, "points_cost": t.PointsCost,
		"pickup_start": t.PickupStart, "pickup_end": t.PickupEnd,
		"valid_start": t.ValidStart, "valid_end": t.ValidEnd,
		"status": t.Status, "updated_at": time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func AdminCouponDelete(sc *svc.ServiceContext, id int64) error {
	var cnt int64
	if err := sc.DB.Model(&model.MemberCoupon{}).Where("template_id = ?", id).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return common.ErrReferenced
	}
	res := sc.DB.Delete(&model.CouponTemplate{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// IssueToMembers 管理端定向发放（逐人按限领发放，部分失败不影响其余）
func IssueToMembers(sc *svc.ServiceContext, templateID int64, memberIDs []int64) (issued int, failed int) {
	for _, mid := range memberIDs {
		if err := Issue(sc, mid, templateID); err != nil {
			failed++
			continue
		}
		issued++
	}
	return issued, failed
}

// GrantNewUserCoupons 新用户注册自动赠送 new_user_only=1 的启用券（异步调用，失败仅日志）
func GrantNewUserCoupons(sc *svc.ServiceContext, memberID int64) {
	var ids []int64
	if err := sc.DB.Model(&model.CouponTemplate{}).
		Where("status = 1 AND new_user_only = 1").Pluck("id", &ids).Error; err != nil {
		logx.Errorf("查询新人券失败 member=%d: %v", memberID, err)
		return
	}
	for _, tid := range ids {
		if err := Issue(sc, memberID, tid); err != nil {
			logx.Errorf("发放新人券失败 member=%d template=%d: %v", memberID, tid, err)
		} else {
			logx.Infof("新人券发放成功 member=%d template=%d", memberID, tid)
		}
	}
}

// ─────────────────────────── 内部工具 ───────────────────────────

func templateView(t model.CouponTemplate) types.CouponTemplateView {
	v := types.CouponTemplateView{
		ID:                formatID(t.ID),
		Name:              t.Name,
		Type:              t.Type,
		TypeText:          CouponTypeText(t.Type),
		ThresholdAmount:   t.ThresholdAmount.StringFixed(2),
		DiscountAmount:    t.DiscountAmount.StringFixed(2),
		DiscountPercent:   t.DiscountPercent,
		MaxDiscountAmount: t.MaxDiscountAmount.StringFixed(2),
		TotalCount:        t.TotalCount,
		IssuedCount:       t.IssuedCount,
		PerLimit:          t.PerLimit,
		NewUserOnly:       t.NewUserOnly,
		PointsCost:        t.PointsCost,
		Status:            t.Status,
		UpdatedAt:         t.UpdatedAt.Format(time.RFC3339),
	}
	if t.PickupStart != nil {
		v.PickupStart = t.PickupStart.Format(time.RFC3339)
	}
	if t.PickupEnd != nil {
		v.PickupEnd = t.PickupEnd.Format(time.RFC3339)
	}
	if t.ValidStart != nil {
		v.ValidStart = t.ValidStart.Format(time.RFC3339)
	}
	if t.ValidEnd != nil {
		v.ValidEnd = t.ValidEnd.Format(time.RFC3339)
	}
	return v
}

func isNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

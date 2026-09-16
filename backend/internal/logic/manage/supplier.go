// supplier.go 供货商管理：活体来源档案（联系/执照），商品可挂供货商与检疫证明。
package manage

import (
	"strings"
	"time"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func supplierView(m *model.Supplier) types.SupplierView {
	return types.SupplierView{
		ID:        strconvI64(m.ID),
		Name:      m.Name,
		Contact:   m.Contact,
		Phone:     m.Phone,
		Address:   m.Address,
		LicenseNo: m.LicenseNo,
		Remark:    m.Remark,
		Status:    m.Status,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
	}
}

// AdminSupplierList 供货商列表（含停用）
func AdminSupplierList(sc *svc.ServiceContext, req *types.SupplierListReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	q := sc.DB.Model(&model.Supplier{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.Supplier
	if err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, err
	}
	// 引用计数（任意状态，删除拦截同口径）
	cntMap := map[int64]int64{}
	if len(rows) > 0 {
		ids := make([]int64, 0, len(rows))
		for i := range rows {
			ids = append(ids, rows[i].ID)
		}
		type cntRow struct {
			SupplierID int64
			Cnt        int64
		}
		var crs []cntRow
		if err := sc.DB.Model(&model.PetProduct{}).
			Select("supplier_id AS supplier_id, COUNT(*) AS cnt").
			Where("supplier_id IN ?", ids).Group("supplier_id").Scan(&crs).Error; err != nil {
			return nil, err
		}
		for _, c := range crs {
			cntMap[c.SupplierID] = c.Cnt
		}
	}
	list := make([]types.SupplierView, 0, len(rows))
	for i := range rows {
		v := supplierView(&rows[i])
		v.ProductCount = int(cntMap[rows[i].ID])
		list = append(list, v)
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// AdminSupplierUpsert 新增 / 编辑（有 ID 为编辑）
func AdminSupplierUpsert(sc *svc.ServiceContext, req *types.SupplierUpsertReq) error {
	name := strings.TrimSpace(req.Name)
	if name == "" || len([]rune(name)) > 64 {
		return common.ErrParam
	}
	contact := strings.TrimSpace(req.Contact)
	phone := strings.TrimSpace(req.Phone)
	address := strings.TrimSpace(req.Address)
	licenseNo := strings.TrimSpace(req.LicenseNo)
	remark := strings.TrimSpace(req.Remark)
	if len([]rune(contact)) > 32 || len(phone) > 20 || len([]rune(address)) > 255 ||
		len(licenseNo) > 64 || len([]rune(remark)) > 255 {
		return common.ErrParam
	}
	if req.ID == "" {
		return sc.DB.Create(&model.Supplier{ID: common.NewID(), Name: name, Contact: contact,
			Phone: phone, Address: address, LicenseNo: licenseNo, Remark: remark, Status: 1}).Error
	}
	id, err := strconvI64Err(req.ID)
	if err != nil {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.Supplier{}).Where("id = ?", id).Updates(map[string]any{
		"name": name, "contact": contact, "phone": phone,
		"address": address, "license_no": licenseNo, "remark": remark,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// AdminSupplierStatus 启用 / 停用
func AdminSupplierStatus(sc *svc.ServiceContext, id int64, status int) error {
	if status != 0 && status != 1 {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.Supplier{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// AdminSupplierDelete 删除（已挂商品保留 supplier_id 快照，商品侧按 ID 兜底显示）
func AdminSupplierDelete(sc *svc.ServiceContext, id int64) error {
	var cnt int64
	if err := sc.DB.Model(&model.PetProduct{}).Where("supplier_id = ?", id).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return common.ErrSupplierReferenced
	}
	res := sc.DB.Where("id = ?", id).Delete(&model.Supplier{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

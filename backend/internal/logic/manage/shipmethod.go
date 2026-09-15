// shipmethod.go 配送/托运方式：管理端 CRUD（用户端下单读取见 logic/trade/ship.go）。
package manage

import (
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func strconvI64(id int64) string { return strconv.FormatInt(id, 10) }

func strconvI64Err(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, common.ErrParam
	}
	return id, nil
}

func view(m *model.ShipMethod) types.ShipMethodView {
	return types.ShipMethodView{
		ID:          strconvI64(m.ID),
		Name:        m.Name,
		Kind:        m.Kind,
		Description: m.Description,
		Fee:         m.Fee.StringFixed(2),
		Sort:        m.Sort,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
	}
}

// AdminShipMethodList 配送方式列表（含停用，sort ASC 分页）
func AdminShipMethodList(sc *svc.ServiceContext, req *types.ShipMethodListReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	q := sc.DB.Model(&model.ShipMethod{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.ShipMethod
	if err := q.Order("sort ASC, id ASC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.ShipMethodView, 0, len(rows))
	for i := range rows {
		list = append(list, view(&rows[i]))
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// AdminShipMethodUpsert 新增 / 编辑（有 ID 为编辑，可覆盖更新）
func AdminShipMethodUpsert(sc *svc.ServiceContext, req *types.ShipMethodUpsertReq) error {
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 32 {
		return common.ErrParam
	}
	if req.Kind != model.KindPickup && req.Kind != model.KindShip {
		return common.ErrParam
	}
	description := strings.TrimSpace(req.Description)
	if len(description) > 128 {
		return common.ErrParam
	}
	fee, err := decimal.NewFromString(strings.TrimSpace(req.Fee))
	if err != nil || fee.IsNegative() || fee.GreaterThan(decimal.NewFromInt(999999)) {
		return common.ErrParam
	}
	if req.Status != model.ShipMethodOff && req.Status != model.ShipMethodOn {
		return common.ErrParam
	}
	updates := map[string]any{
		"name":        name,
		"kind":        req.Kind,
		"description": description,
		"fee":         fee,
		"sort":        req.Sort,
		"status":      req.Status,
	}
	if req.ID == "" {
		return sc.DB.Create(&model.ShipMethod{ID: common.NewID(), Name: name, Kind: req.Kind,
			Description: description, Fee: fee, Sort: req.Sort, Status: req.Status}).Error
	}
	id, err := strconvI64Err(req.ID)
	if err != nil {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.ShipMethod{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// AdminShipMethodStatus 启用 / 停用
func AdminShipMethodStatus(sc *svc.ServiceContext, id int64, status int) error {
	if status != model.ShipMethodOff && status != model.ShipMethodOn {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.ShipMethod{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// AdminShipMethodDelete 删除（历史订单已快照名称与运费，不受影响）
func AdminShipMethodDelete(sc *svc.ServiceContext, id int64) error {
	res := sc.DB.Where("id = ?", id).Delete(&model.ShipMethod{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

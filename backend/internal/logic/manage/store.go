// store.go 自提门店管理
package manage

import (
	"strconv"
	"strings"
	"time"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func storeView(s model.Store) types.StoreView {
	return types.StoreView{
		ID:            strconv.FormatInt(s.ID, 10),
		Name:          s.Name,
		Address:       s.Address,
		Phone:         s.Phone,
		BusinessHours: s.BusinessHours,
		Status:        s.Status,
		Sort:          s.Sort,
		CreatedAt:     s.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func AdminStoreList(sc *svc.ServiceContext) ([]types.StoreView, error) {
	var stores []model.Store
	if err := sc.DB.Order("sort ASC, id ASC").Limit(100).Find(&stores).Error; err != nil {
		return nil, err
	}
	list := make([]types.StoreView, 0, len(stores))
	for _, s := range stores {
		list = append(list, storeView(s))
	}
	return list, nil
}

func AdminStoreUpsert(sc *svc.ServiceContext, req *types.StoreUpsertReq) error {
	name := strings.TrimSpace(req.Name)
	address := strings.TrimSpace(req.Address)
	if name == "" || len([]rune(name)) > 64 || len(address) > 255 {
		return common.ErrParam
	}
	status := 1
	if req.Status != nil && *req.Status == 0 {
		status = 0
	}
	if req.ID == "" {
		return sc.DB.Create(&model.Store{
			ID: common.NewID(), Name: name, Address: address,
			Phone: req.Phone, BusinessHours: req.BusinessHours,
			Status: status, Sort: req.Sort,
		}).Error
	}
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil || id <= 0 {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.Store{}).Where("id = ?", id).Updates(map[string]any{
		"name": name, "address": address, "phone": req.Phone,
		"business_hours": req.BusinessHours, "status": status,
		"sort": req.Sort, "updated_at": time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func AdminStoreStatus(sc *svc.ServiceContext, id int64, status int) error {
	if status != 0 && status != 1 {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.Store{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func AdminStoreDelete(sc *svc.ServiceContext, id int64) error {
	res := sc.DB.Delete(&model.Store{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

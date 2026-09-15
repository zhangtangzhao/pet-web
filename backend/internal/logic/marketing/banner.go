// banner.go 首页运营位：管理端 CRUD（用户端首页读取见 logic/pet/banner.go）。
package marketing

import (
	"strconv"
	"strings"
	"time"

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

// jumpTypes 跳转类型白名单（与 mobile 端 BANNER_NAV 映射一一对应）
var jumpTypes = map[string]bool{
	"me": true, "recommend": true, "coupon": true, "orders": true,
	"notify": true, "favorites": true, "product": true,
	"category": true, "search": true, "custom": true,
}

func view(b *model.Banner) types.BannerView {
	return types.BannerView{
		ID:        strconvI64(b.ID),
		Title:     b.Title,
		SubTitle:  b.SubTitle,
		Icon:      b.Icon,
		JumpType:  b.JumpType,
		Target:    b.Target,
		Sort:      b.Sort,
		Status:    b.Status,
		CreatedAt: b.CreatedAt.Format(time.RFC3339),
	}
}

// AdminBannerList 平台端 banner 列表（含下架，id DESC 分页）
func AdminBannerList(sc *svc.ServiceContext, req *types.BannerListReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	q := sc.DB.Model(&model.Banner{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.Banner
	if err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.BannerView, 0, len(rows))
	for i := range rows {
		list = append(list, view(&rows[i]))
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// AdminBannerUpsert 新增 / 编辑（有 ID 为编辑，可覆盖更新）
func AdminBannerUpsert(sc *svc.ServiceContext, req *types.BannerUpsertReq) error {
	title := strings.TrimSpace(req.Title)
	if title == "" || len(title) > 32 || !jumpTypes[req.JumpType] {
		return common.ErrParam
	}
	if len(req.SubTitle) > 64 || len(req.Icon) > 255 || len(req.Target) > 255 {
		return common.ErrParam
	}
	updates := map[string]any{
		"title":     title,
		"sub_title": req.SubTitle,
		"icon":      req.Icon,
		"jump_type": req.JumpType,
		"target":    req.Target,
		"sort":      req.Sort,
		"status":    req.Status,
	}
	if req.ID == "" {
		return sc.DB.Create(&model.Banner{ID: common.NewID(), Title: title, SubTitle: req.SubTitle,
			Icon: req.Icon, JumpType: req.JumpType, Target: req.Target, Sort: req.Sort, Status: req.Status}).Error
	}
	id, err := strconvI64Err(req.ID)
	if err != nil {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.Banner{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// AdminBannerStatus 上下架
func AdminBannerStatus(sc *svc.ServiceContext, id int64, status int) error {
	if status != model.BannerOff && status != model.BannerOn {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.Banner{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// AdminBannerDelete 删除
func AdminBannerDelete(sc *svc.ServiceContext, id int64) error {
	res := sc.DB.Where("id = ?", id).Delete(&model.Banner{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

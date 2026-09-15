// Package pet banner.go 用户端首页运营位（仅上架，sort ASC）
package pet

import (
	"strconv"
	"time"

	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// ListBanners 首页上架 banner 列表
func ListBanners(sc *svc.ServiceContext) ([]types.BannerView, error) {
	var rows []model.Banner
	if err := sc.DB.Where("status = ?", model.BannerOn).
		Order("sort ASC, id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.BannerView, 0, len(rows))
	for _, b := range rows {
		list = append(list, types.BannerView{
			ID:        strconv.FormatInt(b.ID, 10),
			Title:     b.Title,
			SubTitle:  b.SubTitle,
			Icon:      b.Icon,
			JumpType:  b.JumpType,
			Target:    b.Target,
			Sort:      b.Sort,
			Status:    b.Status,
			CreatedAt: b.CreatedAt.Format(time.RFC3339),
		})
	}
	return list, nil
}

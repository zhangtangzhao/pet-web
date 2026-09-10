package pet

import (
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// Favorite 收藏（幂等），成功则累加计数
func Favorite(sc *svc.ServiceContext, memberID int64, productIDStr string) error {
	productID, err := parseID(productIDStr)
	if err != nil {
		return err
	}
	var cnt int64
	if err := sc.DB.Model(&model.PetProduct{}).Where("id = ?", productID).Count(&cnt).Error; err != nil || cnt == 0 {
		return common.ErrNotFound
	}
	fav := model.MemberFavorite{ID: common.NewID(), MemberID: memberID, ProductID: productID}
	res := sc.DB.Create(&fav)
	if res.Error != nil {
		// 唯一约束冲突 = 重复收藏，幂等返回成功
		return nil
	}
	if res.RowsAffected == 1 {
		return sc.DB.Model(&model.PetProduct{}).Where("id = ?", productID).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count + 1")).Error
	}
	return nil
}

// Unfavorite 取消收藏（幂等），成功则递减计数
func Unfavorite(sc *svc.ServiceContext, memberID int64, productIDStr string) error {
	productID, err := parseID(productIDStr)
	if err != nil {
		return err
	}
	res := sc.DB.Where("member_id = ? AND product_id = ?", memberID, productID).
		Delete(&model.MemberFavorite{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return sc.DB.Model(&model.PetProduct{}).
			Where("id = ? AND favorite_count > 0", productID).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count - 1")).Error
	}
	return nil
}

// FavoriteList 我的收藏列表
func FavoriteList(sc *svc.ServiceContext, memberID int64, req *types.FavoriteListReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}
	var total int64
	if err := sc.DB.Model(&model.MemberFavorite{}).Where("member_id = ?", memberID).Count(&total).Error; err != nil {
		return nil, err
	}
	var favs []model.MemberFavorite
	if err := sc.DB.Where("member_id = ?", memberID).
		Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&favs).Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(favs))
	for _, f := range favs {
		ids = append(ids, f.ProductID)
	}
	list := make([]types.ProductCard, 0, len(ids))
	if len(ids) > 0 {
		var products []model.PetProduct
		if err := sc.DB.Where("id IN ?", ids).Find(&products).Error; err != nil {
			return nil, err
		}
		cards, err := buildCards(sc, products)
		if err != nil {
			return nil, err
		}
		list = cards
	}
	return &types.PageResp{Total: total, List: list}, nil
}

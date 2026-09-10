package manage

import (
	"strconv"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, common.ErrParam
	}
	return id, nil
}

func existsByID(db *gorm.DB, dest any, id int64) bool {
	return db.First(dest, id).Error == nil
}

// UpsertCategory 新建/编辑分类
func UpsertCategory(sc *svc.ServiceContext, req *types.CategoryUpsertReq) error {
	if req.Name == "" {
		return common.ErrParam
	}
	if req.ID == "" {
		return sc.DB.Create(&model.Category{Name: req.Name, Icon: req.Icon, Sort: req.Sort, Status: 1}).Error
	}
	id, err := parseID(req.ID)
	if err != nil {
		return err
	}
	res := sc.DB.Model(&model.Category{}).Where("id = ?", id).
		Updates(map[string]any{"name": req.Name, "icon": req.Icon, "sort": req.Sort})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// DeleteCategory 删除分类（存在品种/商品时拒绝）
func DeleteCategory(sc *svc.ServiceContext, idStr string) error {
	id, err := parseID(idStr)
	if err != nil {
		return err
	}
	var breedCnt, productCnt int64
	if err := sc.DB.Model(&model.Breed{}).Where("category_id = ?", id).Count(&breedCnt).Error; err != nil {
		return err
	}
	if err := sc.DB.Model(&model.PetProduct{}).Where("category_id = ?", id).Count(&productCnt).Error; err != nil {
		return err
	}
	if breedCnt > 0 || productCnt > 0 {
		return common.ErrReferenced
	}
	res := sc.DB.Delete(&model.Category{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// UpsertBreed 新建/编辑品种
func UpsertBreed(sc *svc.ServiceContext, req *types.BreedUpsertReq) error {
	if req.Name == "" || req.CategoryID == "" {
		return common.ErrParam
	}
	catID, err := parseID(req.CategoryID)
	if err != nil {
		return err
	}
	if !existsByID(sc.DB, &model.Category{}, catID) {
		return common.NewErr(400, 41208, "所属分类不存在")
	}
	if req.ID == "" {
		return sc.DB.Create(&model.Breed{
			CategoryID: catID, Name: req.Name, Intro: req.Intro, Cover: req.Cover, Sort: req.Sort, Status: 1,
		}).Error
	}
	id, err := parseID(req.ID)
	if err != nil {
		return err
	}
	res := sc.DB.Model(&model.Breed{}).Where("id = ?", id).
		Updates(map[string]any{"category_id": catID, "name": req.Name, "intro": req.Intro, "cover": req.Cover, "sort": req.Sort})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// DeleteBreed 删除品种（存在商品时拒绝）
func DeleteBreed(sc *svc.ServiceContext, idStr string) error {
	id, err := parseID(idStr)
	if err != nil {
		return err
	}
	var productCnt int64
	if err := sc.DB.Model(&model.PetProduct{}).Where("breed_id = ?", id).Count(&productCnt).Error; err != nil {
		return err
	}
	if productCnt > 0 {
		return common.ErrReferenced
	}
	res := sc.DB.Delete(&model.Breed{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

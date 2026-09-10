package pet

import (
	"strconv"

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

// Categories 分类列表（含启用品种数）
func Categories(sc *svc.ServiceContext) ([]types.CategoryItem, error) {
	var cats []model.Category
	if err := sc.DB.Where("status = 1").Order("sort ASC, id ASC").Find(&cats).Error; err != nil {
		return nil, err
	}
	type cntRow struct {
		CategoryID int64
		Cnt        int64
	}
	var rows []cntRow
	if err := sc.DB.Model(&model.Breed{}).
		Select("category_id, COUNT(*) AS cnt").
		Where("status = 1").
		Group("category_id").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	cntMap := make(map[int64]int64, len(rows))
	for _, r := range rows {
		cntMap[r.CategoryID] = r.Cnt
	}
	list := make([]types.CategoryItem, 0, len(cats))
	for _, c := range cats {
		list = append(list, types.CategoryItem{
			ID:       strconv.FormatInt(c.ID, 10),
			Name:     c.Name,
			Icon:     c.Icon,
			BreedCnt: cntMap[c.ID],
		})
	}
	return list, nil
}

// Breeds 品种列表（可按分类过滤）
func Breeds(sc *svc.ServiceContext, categoryID string) ([]types.BreedItem, error) {
	query := sc.DB.Model(&model.Breed{}).Where("status = 1")
	if categoryID != "" {
		id, err := parseID(categoryID)
		if err != nil {
			return nil, err
		}
		query = query.Where("category_id = ?", id)
	}
	var breeds []model.Breed
	if err := query.Order("sort ASC, id ASC").Find(&breeds).Error; err != nil {
		return nil, err
	}
	list := make([]types.BreedItem, 0, len(breeds))
	for _, b := range breeds {
		list = append(list, types.BreedItem{
			ID:         strconv.FormatInt(b.ID, 10),
			CategoryID: strconv.FormatInt(b.CategoryID, 10),
			Name:       b.Name,
			Cover:      b.Cover,
		})
	}
	return list, nil
}

// Home 首页聚合：分类入口 + 热销推荐（轮播位一期由 Home 硬编码占位，后台配置二期）
func Home(sc *svc.ServiceContext) (*types.HomeResp, error) {
	cats, err := Categories(sc)
	if err != nil {
		return nil, err
	}
	var products []model.PetProduct
	if err := sc.DB.Where("status = ?", model.ProductOnSale).
		Order("sales DESC, id DESC").Limit(10).Find(&products).Error; err != nil {
		return nil, err
	}
	hot, err := buildCards(sc, products)
	if err != nil {
		return nil, err
	}
	return &types.HomeResp{Categories: cats, Hot: hot}, nil
}

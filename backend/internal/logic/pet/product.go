package pet

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func genderText(g int) string {
	if g == 1 {
		return "公"
	}
	if g == 2 {
		return "母"
	}
	return "未知"
}

func ageText(birth *time.Time) string {
	if birth == nil {
		return ""
	}
	months := int(math.Floor(time.Since(*birth).Hours() / 24 / 30.44))
	if months < 1 {
		return "1个月内"
	}
	if months < 12 {
		return fmt.Sprintf("%d个月", months)
	}
	y := months / 12
	m := months % 12
	if m == 0 {
		return fmt.Sprintf("%d岁", y)
	}
	return fmt.Sprintf("%d岁%d个月", y, m)
}

// buildCards 组装商品卡片（批量查品种名）
func buildCards(sc *svc.ServiceContext, products []model.PetProduct) ([]types.ProductCard, error) {
	breedIDs := make([]int64, 0, len(products))
	for _, p := range products {
		breedIDs = append(breedIDs, p.BreedID)
	}
	nameMap := map[int64]string{}
	if len(breedIDs) > 0 {
		var breeds []model.Breed
		if err := sc.DB.Where("id IN ?", breedIDs).Find(&breeds).Error; err != nil {
			return nil, err
		}
		for _, b := range breeds {
			nameMap[b.ID] = b.Name
		}
	}
	list := make([]types.ProductCard, 0, len(products))
	for _, p := range products {
		list = append(list, types.ProductCard{
			ID:            strconv.FormatInt(p.ID, 10),
			Title:         p.Title,
			MainImage:     p.MainImage,
			Price:         p.Price.StringFixed(2),
			OriginalPrice: p.OriginalPrice.StringFixed(2),
			BreedName:     nameMap[p.BreedID],
			PetGender:     p.PetGender,
			FavoriteCount: p.FavoriteCount,
			Sales:         p.Sales,
			Status:        p.Status,
		})
	}
	return list, nil
}

// ListProducts 用户端商品列表（仅含在售）
func ListProducts(sc *svc.ServiceContext, req *types.ProductListReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}
	query := sc.DB.Model(&model.PetProduct{}).Where("status = ?", model.ProductOnSale)
	if req.CategoryID != "" {
		id, err := parseID(req.CategoryID)
		if err != nil {
			return nil, err
		}
		query = query.Where("category_id = ?", id)
	}
	if req.BreedID != "" {
		id, err := parseID(req.BreedID)
		if err != nil {
			return nil, err
		}
		query = query.Where("breed_id = ?", id)
	}
	if req.Gender > 0 {
		query = query.Where("pet_gender = ?", req.Gender)
	}
	if req.PriceMin > 0 {
		query = query.Where("price >= ?", fmt.Sprintf("%.2f", float64(req.PriceMin)/100))
	}
	if req.PriceMax > 0 {
		query = query.Where("price <= ?", fmt.Sprintf("%.2f", float64(req.PriceMax)/100))
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where("title ILIKE ?", kw)
	}
	switch req.Sort {
	case "sales":
		query = query.Order("sales DESC, id DESC")
	case "price_asc":
		query = query.Order("price ASC, id DESC")
	case "price_desc":
		query = query.Order("price DESC, id DESC")
	default:
		query = query.Order("created_at DESC, id DESC")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var products []model.PetProduct
	if err := query.Offset((page - 1) * size).Limit(size).Find(&products).Error; err != nil {
		return nil, err
	}
	list, err := buildCards(sc, products)
	if err != nil {
		return nil, err
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// ProductDetail 商品详情（游客可看在售；已下架/草稿返回 41001）
func ProductDetail(sc *svc.ServiceContext, memberID int64, idStr string) (*types.ProductDetail, error) {
	id, err := parseID(idStr)
	if err != nil {
		return nil, err
	}
	var p model.PetProduct
	if err := sc.DB.First(&p, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	if p.Status == model.ProductDraft || p.Status == model.ProductOffSale {
		return nil, common.ErrGoodsOffline
	}

	var images []model.ProductImage
	if err := sc.DB.Where("product_id = ?", id).Order("sort ASC, id ASC").Find(&images).Error; err != nil {
		return nil, err
	}
	imgs := make([]string, 0, len(images))
	for _, im := range images {
		imgs = append(imgs, im.URL)
	}

	var breed model.Breed
	_ = sc.DB.First(&breed, p.BreedID).Error
	var category model.Category
	_ = sc.DB.First(&category, p.CategoryID).Error

	isFav := false
	if memberID > 0 {
		var cnt int64
		_ = sc.DB.Model(&model.MemberFavorite{}).
			Where("member_id = ? AND product_id = ?", memberID, id).
			Count(&cnt).Error
		isFav = cnt > 0
		// 浏览计数（异步，不阻塞响应）
		go sc.DB.Model(&model.PetProduct{}).Where("id = ?", id).
			UpdateColumn("view_count", gorm.Expr("view_count + 1"))
	}

	var birth string
	if p.BirthDate != nil {
		birth = p.BirthDate.Format("2006-01-02")
	}

	detail := &types.ProductDetail{
		ProductCard: types.ProductCard{
			ID:            strconv.FormatInt(p.ID, 10),
			Title:         p.Title,
			MainImage:     p.MainImage,
			Price:         p.Price.StringFixed(2),
			OriginalPrice: p.OriginalPrice.StringFixed(2),
			BreedName:     breed.Name,
			PetGender:     p.PetGender,
			FavoriteCount: p.FavoriteCount,
			Sales:         p.Sales,
			Status:        p.Status,
		},
		Images:     imgs,
		VideoURL:   p.VideoURL,
		VideoCover: p.VideoCover,
		DetailHTML: p.DetailHTML,
		PetProfile: types.PetProfile{
			Gender:      p.PetGender,
			GenderText:  genderText(p.PetGender),
			BirthDate:   birth,
			AgeText:     ageText(p.BirthDate),
			VaccineDesc: p.VaccineDesc,
			DewormDesc:  p.DewormDesc,
			BodyType:    p.BodyType,
			CoatColor:   p.CoatColor,
			Personality: p.Personality,
			HealthDesc:  p.HealthDesc,
		},
		Breed: types.BreedItem{
			ID:         strconv.FormatInt(breed.ID, 10),
			CategoryID: strconv.FormatInt(breed.CategoryID, 10),
			Name:       breed.Name,
			Cover:      breed.Cover,
		},
		Category:   types.CategoryRef{ID: strconv.FormatInt(category.ID, 10), Name: category.Name},
		IsFavorite: isFav,
	}
	return detail, nil
}

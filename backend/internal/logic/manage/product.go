package manage

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// AdminProductList 商品列表（平台端，含全部状态）
func AdminProductList(sc *svc.ServiceContext, req *types.AdminProductListReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	query := sc.DB.Model(&model.PetProduct{})
	if req.Status >= 0 {
		query = query.Where("status = ?", req.Status)
	}
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
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where("title ILIKE ? OR spu_no ILIKE ?", kw, kw)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var products []model.PetProduct
	if err := query.Order("created_at DESC, id DESC").Offset((page - 1) * size).Limit(size).Find(&products).Error; err != nil {
		return nil, err
	}

	// 批量取品种名
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
	list := make([]types.AdminProductView, 0, len(products))
	for _, p := range products {
		v := types.AdminProductView{
			ID:            strconv.FormatInt(p.ID, 10),
			SpuNo:         p.SpuNo,
			Title:         p.Title,
			MainImage:     p.MainImage,
			Price:         p.Price.StringFixed(2),
			OriginalPrice: p.OriginalPrice.StringFixed(2),
			BreedName:     nameMap[p.BreedID],
			PetGender:     p.PetGender,
			Status:        p.Status,
			Sales:         p.Sales,
			ViewCount:     p.ViewCount,
			FavoriteCount: p.FavoriteCount,
			CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		}
		list = append(list, v)
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// AdminProductDetail 编辑回显（含相册）
func AdminProductDetail(sc *svc.ServiceContext, idStr string) (*types.AdminProductDetailResp, error) {
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
	var images []model.ProductImage
	if err := sc.DB.Where("product_id = ?", id).Order("sort ASC, id ASC").Find(&images).Error; err != nil {
		return nil, err
	}
	imgs := make([]string, 0, len(images))
	for _, im := range images {
		imgs = append(imgs, im.URL)
	}
	var birth string
	if p.BirthDate != nil {
		birth = p.BirthDate.Format("2006-01-02")
	}
	return &types.AdminProductDetailResp{
		ID:            strconv.FormatInt(p.ID, 10),
		Title:         p.Title,
		CategoryID:    strconv.FormatInt(p.CategoryID, 10),
		BreedID:       strconv.FormatInt(p.BreedID, 10),
		Price:         p.Price.StringFixed(2),
		OriginalPrice: p.OriginalPrice.StringFixed(2),
		Status:        p.Status,
		PetProfileUpsert: types.PetProfileUpsert{
			PetGender:   p.PetGender,
			BirthDate:   birth,
			VaccineDesc: p.VaccineDesc,
			DewormDesc:  p.DewormDesc,
			BodyType:    p.BodyType,
			CoatColor:   p.CoatColor,
			Personality: p.Personality,
			HealthDesc:  p.HealthDesc,
		},
		MainImage:  p.MainImage,
		Images:     imgs,
		VideoURL:   p.VideoURL,
		VideoCover: p.VideoCover,
		DetailHTML: p.DetailHTML,
	}, nil
}

// UpsertProduct 新建/编辑商品（活体单件：stock 恒为 1）
func UpsertProduct(sc *svc.ServiceContext, req *types.ProductUpsertReq) error {
	if req.Title == "" || req.MainImage == "" {
		return common.ErrParam
	}
	catID, err := parseID(req.CategoryID)
	if err != nil {
		return err
	}
	breedID, err := parseID(req.BreedID)
	if err != nil {
		return err
	}
	if !existsByID(sc.DB, &model.Category{}, catID) || !existsByID(sc.DB, &model.Breed{}, breedID) {
		return common.NewErr(400, 41208, "分类或品种不存在")
	}
	price, err := decimal.NewFromString(strings.TrimSpace(req.Price))
	if err != nil || price.Sign() <= 0 || price.GreaterThan(decimal.NewFromInt(99999999)) {
		return common.NewErr(400, 40001, "价格不合法")
	}
	orig := price
	if req.OriginalPrice != "" {
		orig, err = decimal.NewFromString(strings.TrimSpace(req.OriginalPrice))
		if err != nil || orig.Sign() <= 0 {
			return common.NewErr(400, 40001, "划线价不合法")
		}
	}
	var birth *time.Time
	if req.BirthDate != "" {
		t, e := time.ParseInLocation("2006-01-02", req.BirthDate, time.Local)
		if e != nil {
			return common.NewErr(400, 40001, "出生日期格式应为 yyyy-MM-dd")
		}
		birth = &t
	}

	err = sc.DB.Transaction(func(tx *gorm.DB) error {
		if req.ID == "" {
			p := model.PetProduct{
				ID:            common.NewID(),
				SpuNo:         newSpuNo(),
				Title:         req.Title,
				CategoryID:    catID,
				BreedID:       breedID,
				Price:         price,
				OriginalPrice: orig,
				Status:        model.ProductDraft,
				PetGender:     req.PetGender,
				BirthDate:     birth,
				VaccineDesc:   req.VaccineDesc,
				DewormDesc:    req.DewormDesc,
				BodyType:      req.BodyType,
				CoatColor:     req.CoatColor,
				Personality:   req.Personality,
				HealthDesc:    req.HealthDesc,
				MainImage:     req.MainImage,
				VideoURL:      req.VideoURL,
				VideoCover:    req.VideoCover,
				DetailHTML:    req.DetailHTML,
				Stock:         1,
			}
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
			return replaceImages(tx, p.ID, req.Images)
		}
		id, e := parseID(req.ID)
		if e != nil {
			return e
		}
		var p model.PetProduct
		if err := tx.First(&p, id).Error; err != nil {
			return err
		}
		// 已售出/锁定中的活体不允许改单
		if p.Status == model.ProductSold || p.Status == model.ProductLocked {
			return common.ErrGoodsState
		}
		if err := tx.Model(&model.PetProduct{}).Where("id = ?", id).Updates(map[string]any{
			"title": req.Title, "category_id": catID, "breed_id": breedID,
			"price": price, "original_price": orig,
			"pet_gender": req.PetGender, "birth_date": birth,
			"vaccine_desc": req.VaccineDesc, "deworm_desc": req.DewormDesc,
			"body_type": req.BodyType, "coat_color": req.CoatColor,
			"personality": req.Personality, "health_desc": req.HealthDesc,
			"main_image": req.MainImage, "video_url": req.VideoURL,
			"video_cover": req.VideoCover, "detail_html": req.DetailHTML,
		}).Error; err != nil {
			return err
		}
		return replaceImages(tx, id, req.Images)
	})
	return err
}

// replaceImages 全量替换相册（简单可靠，量级：个位数图片）
func replaceImages(tx *gorm.DB, productID int64, urls []string) error {
	if err := tx.Where("product_id = ?", productID).Delete(&model.ProductImage{}).Error; err != nil {
		return err
	}
	for i, u := range urls {
		if strings.TrimSpace(u) == "" {
			continue
		}
		if err := tx.Create(&model.ProductImage{
			ID: common.NewID(), ProductID: productID, URL: u, Sort: i,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// UpdateProductStatus 上下架（锁定/已售出状态禁止变更）
func UpdateProductStatus(sc *svc.ServiceContext, req *types.ProductStatusReq) error {
	id, err := parseID(req.ID)
	if err != nil {
		return err
	}
	if req.Status != model.ProductOnSale && req.Status != model.ProductOffSale && req.Status != model.ProductDraft {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.PetProduct{}).
		Where("id = ? AND status NOT IN ?", id, []int{model.ProductLocked, model.ProductSold}).
		Update("status", req.Status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrGoodsState
	}
	return nil
}

// UploadToken COS 预签名直传凭证
func UploadToken(sc *svc.ServiceContext, req *types.UploadTokenReq) (*types.UploadTokenResp, error) {
	dir := req.Dir
	if dir == "" {
		dir = "misc"
	}
	uploadURL, fileURL, err := sc.COSUploadToken(dir, req.ContentType)
	if err != nil {
		return nil, err
	}
	return &types.UploadTokenResp{UploadURL: uploadURL, FileURL: fileURL, Method: "PUT"}, nil
}

func newSpuNo() string {
	return fmt.Sprintf("PET%s", time.Now().Format("20060102150405"))
}

package manage

import (
	"encoding/json"
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
	var nextVaccine, nextDeworm string
	if p.NextVaccineDate != nil {
		nextVaccine = p.NextVaccineDate.Format("2006-01-02")
	}
	if p.NextDewormDate != nil {
		nextDeworm = p.NextDewormDate.Format("2006-01-02")
	}
	supplierID := ""
	if p.SupplierID > 0 {
		supplierID = strconv.FormatInt(p.SupplierID, 10)
	}
	var skus []model.ProductSku
	_ = sc.DB.Where("product_id = ?", p.ID).Order("sort ASC, id ASC").Find(&skus).Error
	skuRows := make([]types.SkuRow, 0, len(skus))
	for _, s := range skus {
		skuRows = append(skuRows, types.SkuRow{
			ID:     strconv.FormatInt(s.ID, 10),
			Specs:  s.Specs,
			Price:  s.Price.StringFixed(2),
			Sort:   s.Sort,
			Status: s.Status,
		})
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
		MainImage:          p.MainImage,
		Images:             imgs,
		VideoURL:           p.VideoURL,
		VideoCover:         p.VideoCover,
		DetailHTML:         p.DetailHTML,
		SupplierID:         supplierID,
		QuarantineCertURL:  p.QuarantineCertURL,
		NextVaccineDate:    nextVaccine,
		NextDewormDate:     nextDeworm,
		DetailImages:       parseDetailImages(p.DetailImages),
		StockWarnThreshold: strconv.Itoa(p.StockWarnThreshold),
		Skus:               skuRows,
		CertType:           p.CertType,
		CertNo:             p.CertNo,
		ChipNo:             p.ChipNo,
		PreSale:            p.PreSale,
		ScheduledOffSaleAt: func() string {
			if p.ScheduledOffSaleAt != nil {
				return p.ScheduledOffSaleAt.Format("2006-01-02 15:04:05")
			}
			return ""
		}(),
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
	// 供货与检疫信息（均可选）
	supplierID := int64(0)
	if req.SupplierID != "" {
		v, e := parseID(req.SupplierID)
		if e != nil {
			return e
		}
		if !existsByID(sc.DB, &model.Supplier{}, v) {
			return common.NewErr(400, 41208, "供货商不存在")
		}
		supplierID = v
	}
	certURL := strings.TrimSpace(req.QuarantineCertURL)
	if len(certURL) > 512 {
		return common.NewErr(400, 40001, "检疫证明地址过长")
	}
	parseDay := func(s, field string) (*time.Time, error) {
		if s == "" {
			return nil, nil
		}
		t, e := time.ParseInLocation("2006-01-02", s, time.Local)
		if e != nil {
			return nil, common.NewErr(400, 40001, field+"格式应为 yyyy-MM-dd")
		}
		return &t, nil
	}
	nextVaccine, err := parseDay(req.NextVaccineDate, "疫苗到期日")
	if err != nil {
		return err
	}
	nextDeworm, err := parseDay(req.NextDewormDate, "驱虫到期日")
	if err != nil {
		return err
	}
	detailImages, err := normDetailImages(req.DetailImages)
	if err != nil {
		return err
	}
	threshold := 1
	if req.StockWarnThreshold != "" {
		if v, e := strconv.Atoi(strings.TrimSpace(req.StockWarnThreshold)); e == nil && v >= 0 && v <= 100 {
			threshold = v
		}
	}
	skus, err := normSkus(req.Skus)
	if err != nil {
		return err
	}
	hasSku := 0
	if len(skus) > 0 {
		hasSku = 1
	}
	// 预售/血统/芯片/定时下架
	var preSalePrice *decimal.Decimal
	var preSaleETA *time.Time
	if req.PreSalePrice != "" {
		v, e := decimal.NewFromString(strings.TrimSpace(req.PreSalePrice))
		if e == nil && v.Sign() > 0 {
			preSalePrice = &v
		}
	}
	if req.PreSaleETA != "" {
		t, e := time.ParseInLocation("2006-01-02", req.PreSaleETA, time.Local)
		if e == nil {
			preSaleETA = &t
		}
	}
	var scheduledOffSale *time.Time
	if req.ScheduledOffSaleAt != "" {
		t, e := time.Parse(time.RFC3339, req.ScheduledOffSaleAt)
		if e == nil {
			scheduledOffSale = &t
		}
	}

	err = sc.DB.Transaction(func(tx *gorm.DB) error {
		if req.ID == "" {
			p := model.PetProduct{
				ID:                 common.NewID(),
				SpuNo:              newSpuNo(),
				Title:              req.Title,
				CategoryID:         catID,
				BreedID:            breedID,
				Price:              price,
				OriginalPrice:      orig,
				Status:             model.ProductDraft,
				PetGender:          req.PetGender,
				BirthDate:          birth,
				VaccineDesc:        req.VaccineDesc,
				DewormDesc:         req.DewormDesc,
				BodyType:           req.BodyType,
				CoatColor:          req.CoatColor,
				Personality:        req.Personality,
				HealthDesc:         req.HealthDesc,
				MainImage:          req.MainImage,
				VideoURL:           req.VideoURL,
				VideoCover:         req.VideoCover,
				DetailHTML:         req.DetailHTML,
				Stock:              1,
				SupplierID:         supplierID,
				QuarantineCertURL:  certURL,
				NextVaccineDate:    nextVaccine,
				NextDewormDate:     nextDeworm,
				HasSKU:             hasSku,
				DetailImages:       detailImages,
				StockWarnThreshold: threshold,
				CertType:           req.CertType,
				CertNo:             req.CertNo,
				ChipNo:             req.ChipNo,
				PreSale:            req.PreSale,
				PreSaleETA:         preSaleETA,
				ScheduledOffSaleAt: scheduledOffSale,
			}
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
			if err := replaceImages(tx, p.ID, req.Images); err != nil {
				return err
			}
			return replaceSkus(tx, p.ID, skus)
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
			"supplier_id": supplierID, "quarantine_cert_url": certURL,
			"next_vaccine_date": nextVaccine, "next_deworm_date": nextDeworm,
			"has_sku": hasSku, "detail_images": detailImages,
			"stock_warn_threshold": threshold,
			"cert_type":            req.CertType, "cert_no": req.CertNo, "chip_no": req.ChipNo,
			"pre_sale": req.PreSale, "pre_sale_price": preSalePrice, "pre_sale_eta": preSaleETA,
			"scheduled_off_sale_at": scheduledOffSale,
		}).Error; err != nil {
			return err
		}
		if err := replaceImages(tx, id, req.Images); err != nil {
			return err
		}
		return replaceSkus(tx, id, skus)
	})
	return err
}

// normSkus 校验并规范化 SKU 列表（全量替换）
func normSkus(items []types.SkuUpsertItem) ([]model.ProductSku, error) {
	if len(items) == 0 {
		return nil, nil
	}
	if len(items) > 20 {
		return nil, common.NewErr(400, 40001, "规格数量过多（≤20）")
	}
	out := make([]model.ProductSku, 0, len(items))
	for i, it := range items {
		specs := strings.TrimSpace(it.Specs)
		if specs == "" || len([]rune(specs)) > 128 {
			return nil, common.NewErr(400, 40001, "规格描述不合法")
		}
		price, e := decimal.NewFromString(strings.TrimSpace(it.Price))
		if e != nil || price.Sign() <= 0 || price.GreaterThan(decimal.NewFromInt(99999999)) {
			return nil, common.NewErr(400, 40001, "规格价格不合法")
		}
		status := 1
		if it.Status != nil && *it.Status == 0 {
			status = 0
		}
		sort := it.Sort
		if sort == 0 {
			sort = i
		}
		out = append(out, model.ProductSku{
			ID: common.NewID(), Specs: specs, Price: price, Sort: sort, Status: status,
		})
	}
	return out, nil
}

// replaceSkus 全量替换商品 SKU（产品侧只存 ID 外键，事务内先删后插）
func replaceSkus(tx *gorm.DB, productID int64, skus []model.ProductSku) error {
	if err := tx.Where("product_id = ?", productID).Delete(&model.ProductSku{}).Error; err != nil {
		return err
	}
	for i := range skus {
		skus[i].ProductID = productID
		if err := tx.Create(&skus[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func normDetailImages(urls []string) (string, error) {
	clean := make([]string, 0, len(urls))
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		if len(u) > 512 {
			return "", common.NewErr(400, 40001, "详情图地址过长")
		}
		clean = append(clean, u)
		if len(clean) >= 20 {
			break
		}
	}
	if len(clean) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(clean)
	if err != nil {
		return "[]", nil
	}
	return string(b), nil
}

func parseDetailImages(raw string) []string {
	out := []string{}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
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

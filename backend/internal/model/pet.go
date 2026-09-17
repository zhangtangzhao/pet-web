package model

import (
	"time"

	"github.com/shopspring/decimal"

)

// 商品状态
const (
	ProductDraft   = 0 // 草稿
	ProductOnSale  = 1 // 在售
	ProductOffSale = 2 // 下架
	ProductLocked  = 3 // 已下单待支付（占位）
	ProductSold    = 4 // 已售出
)

type Category struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	Sort      int       `json:"sort"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Category) TableName() string { return "category" }

type Breed struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	CategoryID int64     `json:"categoryId"`
	Name       string    `json:"name"`
	Intro      string    `json:"intro"`
	Cover      string    `json:"cover"`
	Sort       int       `json:"sort"`
	Status     int       `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (Breed) TableName() string { return "breed" }

type PetProduct struct {
	ID                 int64           `gorm:"primaryKey" json:"id"`
	SpuNo              string          `json:"spuNo"`
	Title              string          `json:"title"`
	CategoryID         int64           `json:"categoryId"`
	BreedID            int64           `json:"breedId"`
	Price              decimal.Decimal `gorm:"type:numeric(10,2)" json:"price"`
	OriginalPrice      decimal.Decimal `gorm:"type:numeric(10,2)" json:"originalPrice"`
	Status             int             `json:"status"`
	PetGender          int             `json:"petGender"`
	BirthDate          *time.Time      `json:"birthDate"`
	VaccineDesc        string          `json:"vaccineDesc"`
	DewormDesc         string          `json:"dewormDesc"`
	BodyType           string          `json:"bodyType"`
	CoatColor          string          `json:"coatColor"`
	Personality        string          `json:"personality"`
	HealthDesc         string          `json:"healthDesc"`
	MainImage          string          `json:"mainImage"`
	VideoURL           string          `json:"videoUrl"`
	VideoCover         string          `json:"videoCover"`
	DetailHTML         string          `json:"detailHtml"`
	Stock              int             `json:"stock"`
	Sales              int             `json:"sales"`
	ViewCount          int             `json:"viewCount"`
	FavoriteCount      int             `json:"favoriteCount"`
	SupplierID         int64           `json:"supplierId"`
	QuarantineCertURL  string          `json:"quarantineCertUrl"`
	NextVaccineDate    *time.Time      `json:"nextVaccineDate"`
	NextDewormDate     *time.Time      `json:"nextDewormDate"`
	HasSKU             int             `json:"hasSku"`       // 1=挂了启用中的规格
	DetailImages       string          `json:"detailImages"` // JSON 数组（详情长图 URL）
	StockWarnThreshold int             `json:"stockWarnThreshold"`
	PreSale            int             `json:"preSale"`    // 1=预售（未到窝）
	PreSaleETA         *time.Time      `json:"preSaleEta"` // 预计到窝日期
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
}

func (PetProduct) TableName() string { return "pet_product" }

type ProductImage struct {
	ID        int64  `gorm:"primaryKey" json:"id"`
	ProductID int64  `json:"productId"`
	URL       string `json:"url"`
	Sort      int    `json:"sort"`
}

func (ProductImage) TableName() string { return "product_image" }

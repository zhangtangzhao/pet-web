package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// 秒杀活动：每商品同时至多一个启用中（uk_flash_sale_active 部分唯一索引）。
// 名额原子扣减（sold+1 WHERE sold<stock），关单/失败路径按 flash_sale_id 回补。
const (
	FlashSaleOff = 0
	FlashSaleOn  = 1
)

type FlashSale struct {
	ID        int64           `gorm:"primaryKey" json:"id"`
	ProductID int64           `json:"productId"`
	SalePrice decimal.Decimal `gorm:"type:numeric(10,2)" json:"salePrice"`
	Stock     int             `json:"stock"`
	Sold      int             `json:"sold"`
	StartAt   time.Time       `json:"startAt"`
	EndAt     time.Time       `json:"endAt"`
	Status    int             `json:"status"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func (FlashSale) TableName() string { return "flash_sale" }

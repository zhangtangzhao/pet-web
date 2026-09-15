package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// 配送方式类别
const (
	KindPickup = 1 // 到店自提
	KindShip   = 2 // 托运配送
)

// 上下架状态
const (
	ShipMethodOff = 0
	ShipMethodOn  = 1
)

type ShipMethod struct {
	ID          int64           `gorm:"primaryKey" json:"id"`
	Name        string          `json:"name"`
	Kind        int             `json:"kind"`
	Description string          `json:"description"`
	Fee         decimal.Decimal `gorm:"type:numeric(10,2)" json:"fee"`
	Sort        int             `json:"sort"`
	Status      int             `json:"status"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

func (ShipMethod) TableName() string { return "ship_method" }

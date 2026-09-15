package model

import (
	"time"
)

// banner 状态
const (
	BannerOff = 0 // 下架
	BannerOn  = 1 // 上架
)

type Banner struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Title     string    `json:"title"`
	SubTitle  string    `json:"subTitle"`
	Icon      string    `json:"icon"` // emoji 或图片 URL
	JumpType  string    `json:"jumpType"`
	Target    string    `json:"target"`
	Sort      int       `json:"sort"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Banner) TableName() string { return "pet_banner" }

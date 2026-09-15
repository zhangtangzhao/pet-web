package model

import (
	"time"
)

// 评价状态
const (
	ReviewShown  = 1 // 显示
	ReviewHidden = 0 // 隐藏（管理端管控）
)

type OrderReview struct {
	ID           int64      `gorm:"primaryKey" json:"id"`
	OrderNo      string     `json:"orderNo"`
	MemberID     int64      `json:"memberId"`
	ProductID    int64      `json:"productId"`
	ProductTitle string     `json:"productTitle"`
	Rating       int        `json:"rating"`
	Content      string     `json:"content"`
	Images       string     `json:"images"` // JSON 数组（COS 图片 URL）
	Status       int        `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	Reply        string     `json:"reply"`    // 官方回复（管理端维护）
	RepliedAt    *time.Time `json:"repliedAt"`
}

func (OrderReview) TableName() string { return "order_review" }

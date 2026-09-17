package model

import (
	"time"
)

// ─────────────────────────── 物流轨迹 ───────────────────────────

type OrderTrace struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	OrderNo    string    `json:"orderNo"`
	HappenedAt time.Time `json:"happenedAt"`
	StatusDesc string    `json:"statusDesc"`
	Detail     string    `json:"detail"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (OrderTrace) TableName() string { return "order_trace" }

// ─────────────────────────── 多门店自提 ───────────────────────────

type Store struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	Phone         string    `json:"phone"`
	BusinessHours string    `json:"businessHours"`
	Status        int       `json:"status"`
	Sort          int       `json:"sort"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (Store) TableName() string { return "store" }

// ─────────────────────────── 积分商城 ───────────────────────────

const (
	PointsProductCoupon   = 1 // 优惠券
	PointsProductGoods    = 2 // 实物
	PointsProductFreeShip = 3 // 免运费卡
	PointsOrderPending    = 0 // 待发货
	PointsOrderShipped    = 1 // 已发货
	PointsOrderCompleted  = 2 // 已完成
)

type PointsProduct struct {
	ID               int64     `gorm:"primaryKey" json:"id"`
	Name             string    `json:"name"`
	Image            string    `json:"image"`
	PointsCost       int       `json:"pointsCost"`
	Stock            int       `json:"stock"`
	Type             int       `json:"type"`
	CouponTemplateID int64     `json:"couponTemplateId"`
	Description      string    `json:"description"`
	Status           int       `json:"status"`
	Sort             int       `json:"sort"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func (PointsProduct) TableName() string { return "points_product" }

type PointsOrder struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	OrderNo     string    `json:"orderNo"`
	MemberID    int64     `json:"memberId"`
	ProductID   int64     `json:"productId"`
	ProductName string    `json:"productName"`
	Image       string    `json:"image"`
	PointsCost  int       `json:"pointsCost"`
	Status      int       `json:"status"`
	ShipNo      string    `json:"shipNo"`
	Contact     string    `json:"contact"`
	Phone       string    `json:"phone"`
	Address     string    `json:"address"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (PointsOrder) TableName() string { return "points_order" }

// ─────────────────────────── 晒单广场 ───────────────────────────

const (
	PostPending = 0 // 待审核
	PostShown   = 1 // 显示
	PostHidden  = 2 // 已隐藏
)

type CommunityPost struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	MemberID  int64     `json:"memberId"`
	Content   string    `json:"content"`
	Images    string    `json:"images"` // JSON 数组
	Status    int       `json:"status"`
	LikeCount int       `json:"likeCount"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (CommunityPost) TableName() string { return "community_post" }

type CommunityPostLike struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	PostID    int64     `json:"postId"`
	MemberID  int64     `json:"memberId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (CommunityPostLike) TableName() string { return "community_post_like" }

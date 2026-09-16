package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// 券状态
const (
	CouponUsable   = 1 // 可用
	CouponLocked   = 2 // 已锁定（待支付订单）
	CouponUsed     = 3 // 已使用
	CouponExpired  = 4 // 已过期
	CouponDisabled = 5 // 已作废
)

// 券模板类型
const (
	CouponTypeThreshold = 1 // 满减
	CouponTypeDiscount  = 2 // 折扣
	CouponTypeCash      = 3 // 无门槛立减
)

func CouponStatusText(s int) string {
	switch s {
	case CouponUsable:
		return "可用"
	case CouponLocked:
		return "已锁定"
	case CouponUsed:
		return "已使用"
	case CouponExpired:
		return "已过期"
	case CouponDisabled:
		return "已作废"
	}
	return "未知"
}

type ServiceItem struct {
	ID            int64           `gorm:"primaryKey" json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	OriginalPrice decimal.Decimal `gorm:"type:numeric(10,2)" json:"originalPrice"`
	Price         decimal.Decimal `gorm:"type:numeric(10,2)" json:"price"`
	GuaranteeDays int             `json:"guaranteeDays"` // >0 为健康保障服务，随单快照延长售后窗口
	Sort          int             `json:"sort"`
	Status        int             `json:"status"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

func (ServiceItem) TableName() string { return "service_item" }

type CouponTemplate struct {
	ID                int64           `gorm:"primaryKey" json:"id"`
	Name              string          `json:"name"`
	Type              int             `json:"type"`
	ThresholdAmount   decimal.Decimal `gorm:"type:numeric(10,2)" json:"thresholdAmount"`
	DiscountAmount    decimal.Decimal `gorm:"type:numeric(10,2)" json:"discountAmount"`
	DiscountPercent   int             `json:"discountPercent"`
	MaxDiscountAmount decimal.Decimal `gorm:"type:numeric(10,2)" json:"maxDiscountAmount"`
	TotalCount        int             `json:"totalCount"`
	IssuedCount       int             `json:"issuedCount"`
	PerLimit          int             `json:"perLimit"`
	NewUserOnly       int             `json:"newUserOnly"`
	PointsCost        int             `json:"pointsCost"` // >0 需用积分兑换领取
	PickupStart       *time.Time      `json:"pickupStart"`
	PickupEnd         *time.Time      `json:"pickupEnd"`
	ValidStart        *time.Time      `json:"validStart"`
	ValidEnd          *time.Time      `json:"validEnd"`
	Status            int             `json:"status"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

func (CouponTemplate) TableName() string { return "coupon_template" }

type MemberCoupon struct {
	ID         int64      `gorm:"primaryKey" json:"id"`
	MemberID   int64      `json:"memberId"`
	TemplateID int64      `json:"templateId"`
	Status     int        `json:"status"`
	OrderID    int64      `json:"orderId"`
	ReceivedAt time.Time  `json:"receivedAt"`
	UsedAt     *time.Time `json:"usedAt"`
}

func (MemberCoupon) TableName() string { return "member_coupon" }

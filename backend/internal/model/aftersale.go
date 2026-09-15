package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// 售后状态
const (
	AfterSalePending = 1 // 待审核
	AfterSaleAgreed  = 2 // 已同意（退款受理）
	AfterSaleRefused = 3 // 已拒绝
	AfterSaleCancel  = 4 // 已撤销
)

func AfterSaleStatusText(s int) string {
	switch s {
	case AfterSalePending:
		return "待审核"
	case AfterSaleAgreed:
		return "已同意"
	case AfterSaleRefused:
		return "已拒绝"
	case AfterSaleCancel:
		return "已撤销"
	}
	return "未知"
}

type AfterSale struct {
	ID              int64           `gorm:"primaryKey" json:"id"`
	AfterSaleNo     string          `json:"afterSaleNo"`
	OrderNo         string          `json:"orderNo"`
	MemberID        int64           `json:"memberId"`
	Reason          string          `json:"reason"`
	RefundAmount    decimal.Decimal `gorm:"type:numeric(10,2)" json:"refundAmount"`
	Status          int             `json:"status"`
	AdminNote       string          `json:"adminNote"`
	RefundPaymentNo string          `json:"refundPaymentNo"`
	AuditAt         *time.Time      `json:"auditAt"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

func (AfterSale) TableName() string { return "after_sale" }

package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// 订单状态
const (
	OrderPending   = 10 // 待支付
	OrderPaid      = 20 // 已支付
	OrderCompleted = 30 // 已完成
	OrderCanceled  = 40 // 已取消
	OrderClosed    = 45 // 已关闭(超时)
	OrderRefunding = 50 // 退款中
	OrderRefunded  = 60 // 已退款
)

func OrderStatusText(s int) string {
	switch s {
	case OrderPending:
		return "待支付"
	case OrderPaid:
		return "已支付"
	case OrderCompleted:
		return "已完成"
	case OrderCanceled:
		return "已取消"
	case OrderClosed:
		return "已关闭"
	case OrderRefunding:
		return "退款中"
	case OrderRefunded:
		return "已退款"
	}
	return "未知"
}

// 支付流水
const (
	PayChannelMini    = 1 // 小程序 JSAPI
	PayChannelH5Jsapi = 2 // 公众号 JSAPI
	PayChannelH5      = 3 // H5 支付
)

const (
	PayStatusPending = 0
	PayStatusSuccess = 1
	PayStatusFail    = 2
	PayStatusRefund  = 3
)

const (
	PayTypePurchase = 1 // 支付
	PayTypeRefund   = 2 // 退款
)

type Order struct {
	ID             int64           `gorm:"primaryKey" json:"id"`
	OrderNo        string          `json:"orderNo"`
	MemberID       int64           `json:"memberId"`
	TotalAmount    decimal.Decimal `gorm:"type:numeric(10,2)" json:"totalAmount"`
	DiscountAmount decimal.Decimal `gorm:"type:numeric(10,2)" json:"discountAmount"`
	ServiceFee     decimal.Decimal `gorm:"type:numeric(10,2)" json:"serviceFee"`
	PayAmount      decimal.Decimal `gorm:"type:numeric(10,2)" json:"payAmount"`
	CouponID       int64           `json:"couponId"`
	CouponInfo     string          `json:"couponInfo"`
	ServiceItems   string          `json:"serviceItems"`
	Status         int             `json:"status"`
	ContactName    string          `json:"contactName"`
	ContactPhone   string          `json:"contactPhone"`
	Remark         string          `json:"remark"`
	PaidAt         *time.Time      `json:"paidAt"`
	CompletedAt    *time.Time      `json:"completedAt"`
	CanceledAt     *time.Time      `json:"canceledAt"`
	CancelReason   string          `json:"cancelReason"`
	ExpireAt       time.Time       `json:"expireAt"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

func (Order) TableName() string { return "orders" }

type OrderItem struct {
	ID           int64           `gorm:"primaryKey" json:"id"`
	OrderID      int64           `json:"orderId"`
	ProductID    int64           `json:"productId"`
	ProductTitle string          `json:"productTitle"`
	ProductImage string          `json:"productImage"`
	BreedName    string          `json:"breedName"`
	Price        decimal.Decimal `gorm:"type:numeric(10,2)" json:"price"`
	Quantity     int             `json:"quantity"`
}

func (OrderItem) TableName() string { return "order_item" }

type Payment struct {
	ID            int64           `gorm:"primaryKey" json:"id"`
	PaymentNo     string          `json:"paymentNo"`
	OrderID       int64           `json:"orderId"`
	OrderNo       string          `json:"orderNo"`
	MemberID      int64           `json:"memberId"`
	Amount        decimal.Decimal `gorm:"type:numeric(10,2)" json:"amount"`
	Channel       int             `json:"channel"`
	PayType       int             `json:"payType"`
	TransactionID string          `json:"transactionId"`
	Status        int             `json:"status"`
	CallbackAt    *time.Time      `json:"callbackAt"`
	RawNotify     string          `json:"rawNotify"`
	CreatedAt     time.Time       `json:"createdAt"`
}

func (Payment) TableName() string { return "payment" }

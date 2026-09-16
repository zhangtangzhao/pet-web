package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// 订单状态
const (
	OrderPending     = 10 // 待支付（全款单待付全款 / 定金单待付定金）
	OrderDepositPaid = 15 // 已付定金，待补尾款
	OrderPaid        = 20 // 已支付（全款 / 尾款已补齐）
	OrderCompleted   = 30 // 已完成
	OrderCanceled    = 40 // 已取消
	OrderClosed      = 45 // 已关闭(超时)
	OrderRefunding   = 50 // 退款中
	OrderRefunded    = 60 // 已退款
)

func OrderStatusText(s int) string {
	switch s {
	case OrderPending:
		return "待支付"
	case OrderDepositPaid:
		return "待补尾款"
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

// 配送子状态（订单状态机不变）
const (
	ShipPending   = 0 // 待配送
	ShipTransit   = 1 // 配送中
	ShipDelivered = 2 // 已送达
)

func ShipStatusText(s int) string {
	switch s {
	case ShipTransit:
		return "配送中"
	case ShipDelivered:
		return "已送达"
	}
	return "待配送"
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
	PayTypePurchase = 1 // 支付（全款 / 定金首笔）
	PayTypeRefund   = 2 // 退款
	PayTypeTail     = 3 // 定金单补尾款
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
	DepositAmount  decimal.Decimal `gorm:"type:numeric(10,2)" json:"depositAmount"` // >0 即定金单
	TailExpireAt   *time.Time      `json:"tailExpireAt"`                            // 尾款补款截止（付定金起 N 天）
	FlashSaleID    int64           `json:"flashSaleId"`                             // 命中的秒杀活动，0=无
	GuaranteeDays  int             `json:"guaranteeDays"`                           // 健康保障天数快照
	LevelDiscount  decimal.Decimal `gorm:"type:numeric(10,2)" json:"levelDiscount"` // 等级折扣优惠金额快照
	ContactName    string          `json:"contactName"`
	ContactPhone   string          `json:"contactPhone"`
	Remark         string          `json:"remark"`
	ShipMethodID   int64           `json:"-"`
	ShipMethodName string          `json:"-"`
	ShipFee        decimal.Decimal `gorm:"type:numeric(10,2)" json:"-"`
	ShipAddress    string          `json:"-"`
	ShipStatus     int             `json:"-"`
	ShipNo         string          `json:"-"`
	ShippedAt      *time.Time      `json:"-"`
	DeliveredAt    *time.Time      `json:"-"`
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

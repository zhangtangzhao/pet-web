package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// ─────────────────────────── 砍价 ───────────────────────────

const (
	BargainRunning = 0 // 进行中
	BargainReady   = 1 // 已到底可购
	BargainBought  = 2 // 已购买
	BargainExpired = 3 // 已过期
)

type BargainActivity struct {
	ID            int64           `gorm:"primaryKey" json:"id"`
	ProductID     int64           `json:"productId"`
	BottomPrice   decimal.Decimal `gorm:"type:numeric(10,2)" json:"bottomPrice"`
	DurationHours int             `json:"durationHours"`
	MaxHelpers    int             `json:"maxHelpers"`
	Status        int             `json:"status"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

func (BargainActivity) TableName() string { return "bargain_activity" }

type BargainLaunch struct {
	ID           int64           `gorm:"primaryKey" json:"id"`
	ActivityID   int64           `json:"activityId"`
	MemberID     int64           `json:"memberId"`
	CurrentPrice decimal.Decimal `gorm:"type:numeric(10,2)" json:"currentPrice"`
	HelperCount  int             `json:"helperCount"`
	Status       int             `json:"status"`
	ExpireAt     time.Time       `json:"expireAt"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

func (BargainLaunch) TableName() string { return "bargain_launch" }

type BargainHelp struct {
	ID             int64           `gorm:"primaryKey" json:"id"`
	LaunchID       int64           `json:"launchId"`
	HelperMemberID int64           `json:"helperMemberId"`
	Amount         decimal.Decimal `gorm:"type:numeric(10,2)" json:"amount"`
	CreatedAt      time.Time       `json:"createdAt"`
}

func (BargainHelp) TableName() string { return "bargain_help" }

// ─────────────────────────── 竞拍 ───────────────────────────

const (
	AuctionNotStarted     = 0
	AuctionRunning        = 1
	AuctionSold           = 2 // 已成交
	AuctionFailed         = 3 // 流拍
	AuctionDepositPending = 0
	AuctionDepositPaid    = 1
	AuctionDepositRefund  = 2
)

type Auction struct {
	ID              int64           `gorm:"primaryKey" json:"id"`
	ProductID       int64           `json:"productId"`
	StartPrice      decimal.Decimal `gorm:"type:numeric(10,2)" json:"startPrice"`
	StepPrice       decimal.Decimal `gorm:"type:numeric(10,2)" json:"stepPrice"`
	DepositAmount   decimal.Decimal `gorm:"type:numeric(10,2)" json:"depositAmount"`
	StartAt         time.Time       `json:"startAt"`
	EndAt           time.Time       `json:"endAt"`
	Status          int             `json:"status"`
	HighestMemberID int64           `json:"highestMemberId"`
	HighestPrice    decimal.Decimal `gorm:"type:numeric(10,2)" json:"highestPrice"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

func (Auction) TableName() string { return "auction" }

type AuctionDeposit struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	AuctionID int64     `json:"auctionId"`
	MemberID  int64     `json:"memberId"`
	PaymentNo string    `json:"paymentNo"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (AuctionDeposit) TableName() string { return "auction_deposit" }

// ─────────────────────────── 任务中心 ───────────────────────────

type MemberTask struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	MemberID     int64     `json:"memberId"`
	TaskKey      string    `json:"taskKey"`
	RewardPoints int       `json:"rewardPoints"`
	TaskDate     time.Time `json:"taskDate"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (MemberTask) TableName() string { return "member_task" }

// ─────────────────────────── 服务预约 / 发票 ───────────────────────────

const (
	BookingPending  = 0 // 待到店
	BookingDone     = 1 // 已完成
	BookingCanceled = 2 // 已取消
	InvoicePending  = 0
	InvoiceIssued   = 1
)

type ServiceBooking struct {
	ID          int64           `gorm:"primaryKey" json:"id"`
	BookingNo   string          `json:"bookingNo"`
	MemberID    int64           `json:"memberId"`
	ServiceID   int64           `json:"serviceId"`
	ServiceName string          `json:"serviceName"`
	StoreID     int64           `json:"storeId"`
	BookingDate time.Time       `json:"bookingDate"`
	TimeSlot    string          `json:"timeSlot"`
	Contact     string          `json:"contact"`
	Phone       string          `json:"phone"`
	PetName     string          `json:"petName"`
	Price       decimal.Decimal `gorm:"type:numeric(10,2)" json:"price"`
	Status      int             `json:"status"`
	VerifyCode  string          `json:"verifyCode"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

func (ServiceBooking) TableName() string { return "service_booking" }

type Invoice struct {
	ID        int64           `gorm:"primaryKey" json:"id"`
	MemberID  int64           `json:"memberId"`
	OrderNo   string          `json:"orderNo"`
	TitleType int             `json:"titleType"`
	Title     string          `json:"title"`
	TaxNo     string          `json:"taxNo"`
	Amount    decimal.Decimal `gorm:"type:numeric(10,2)" json:"amount"`
	Status    int             `json:"status"`
	Link      string          `json:"link"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func (Invoice) TableName() string { return "invoice" }

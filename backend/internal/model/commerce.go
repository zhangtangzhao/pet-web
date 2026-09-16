package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// ─────────────────────────── 购物车 ───────────────────────────

type Cart struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	MemberID  int64     `json:"memberId"`
	ProductID int64     `json:"productId"`
	SkuID     int64     `json:"skuId"`
	Checked   int       `json:"checked"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Cart) TableName() string { return "cart" }

// ─────────────────────────── SKU 规格 ───────────────────────────

type ProductSku struct {
	ID        int64           `gorm:"primaryKey" json:"id"`
	ProductID int64           `json:"productId"`
	Specs     string          `json:"specs"`
	Price     decimal.Decimal `gorm:"type:numeric(10,2)" json:"price"`
	Sort      int             `json:"sort"`
	Status    int             `json:"status"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func (ProductSku) TableName() string { return "product_sku" }

// ─────────────────────────── 拼团 ───────────────────────────

const (
	GroupTeamOpen   = 0 // 进行中
	GroupTeamOK     = 1 // 成团
	GroupTeamFailed = 2 // 超时未成团（已退款）
)

type GroupBuy struct {
	ID        int64           `gorm:"primaryKey" json:"id"`
	ProductID int64           `json:"productId"`
	Price     decimal.Decimal `gorm:"type:numeric(10,2)" json:"price"`
	Size      int             `json:"size"`
	Hours     int             `json:"hours"`
	Status    int             `json:"status"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func (GroupBuy) TableName() string { return "group_buy" }

type GroupTeam struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	GroupBuyID     int64     `json:"groupBuyId"`
	LeaderMemberID int64     `json:"leaderMemberId"`
	MemberCount    int       `json:"memberCount"`
	Status         int       `json:"status"`
	ExpireAt       time.Time `json:"expireAt"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (GroupTeam) TableName() string { return "group_team" }

// ─────────────────────────── 宠物档案 ───────────────────────────

type PetProfile struct {
	ID              int64      `gorm:"primaryKey" json:"id"`
	MemberID        int64      `json:"memberId"`
	Name            string     `json:"name"`
	BreedName       string     `json:"breedName"`
	Gender          int        `json:"gender"`
	Birthday        *time.Time `json:"birthday"`
	Weight          *float64   `gorm:"type:numeric(5,2)" json:"weight"`
	Avatar          string     `json:"avatar"`
	VaccineAt       *time.Time `json:"vaccineAt"`
	NextVaccineDate *time.Time `json:"nextVaccineDate"`
	NextDewormDate  *time.Time `json:"nextDewormDate"`
	Remark          string     `json:"remark"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func (PetProfile) TableName() string { return "pet_profile" }

// ─────────────────────────── 风控 ───────────────────────────

const (
	RiskRuleOrderFreq     = "order_freq"
	RiskRuleReviewContact = "review_contact"
	RiskRuleBlacklist     = "member_blacklist"
)

type RiskLog struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	MemberID  int64     `json:"memberId"`
	Rule      string    `json:"rule"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}

func (RiskLog) TableName() string { return "risk_log" }

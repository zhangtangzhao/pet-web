package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type InsuranceProduct struct {
	ID        int64           `gorm:"primaryKey" json:"id"`
	Name      string          `json:"name"`
	Company   string          `json:"company"`
	CoverDesc string          `json:"coverDesc"`
	Price     decimal.Decimal `gorm:"type:numeric(10,2)" json:"price"`
	Status    int             `json:"status"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func (InsuranceProduct) TableName() string { return "insurance_product" }

type InsuranceApply struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	MemberID    int64     `json:"memberId"`
	ProductID   int64     `json:"productId"`
	ProductName string    `json:"productName"`
	Contact     string    `json:"contact"`
	Phone       string    `json:"phone"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (InsuranceApply) TableName() string { return "insurance_apply" }

type StudService struct {
	ID          int64           `gorm:"primaryKey" json:"id"`
	MemberID    int64           `json:"memberId"`
	BreedName   string          `json:"breedName"`
	PetName     string          `json:"petName"`
	HealthCerts string          `json:"healthCerts"`
	Price       decimal.Decimal `gorm:"type:numeric(10,2)" json:"price"`
	Description string          `json:"description"`
	Status      int             `json:"status"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

func (StudService) TableName() string { return "stud_service" }

type Distributor struct {
	ID              int64           `gorm:"primaryKey" json:"id"`
	MemberID        int64           `json:"memberId"`
	Level           int             `json:"level"`
	CommissionRate  decimal.Decimal `gorm:"type:numeric(5,2)" json:"commissionRate"`
	TotalCommission decimal.Decimal `gorm:"type:numeric(10,2)" json:"totalCommission"`
	Balance         decimal.Decimal `gorm:"type:numeric(10,2)" json:"balance"`
	Status          int             `json:"status"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

func (Distributor) TableName() string { return "distributor" }

type DistributorCommission struct {
	ID            int64           `gorm:"primaryKey" json:"id"`
	DistributorID int64           `json:"distributorId"`
	OrderNo       string          `json:"orderNo"`
	Amount        decimal.Decimal `gorm:"type:numeric(10,2)" json:"amount"`
	Status        int             `json:"status"`
	CreatedAt     time.Time       `json:"createdAt"`
}

func (DistributorCommission) TableName() string { return "distributor_commission" }

type DistributorWithdrawal struct {
	ID            int64           `gorm:"primaryKey" json:"id"`
	DistributorID int64           `json:"distributorId"`
	Amount        decimal.Decimal `gorm:"type:numeric(10,2)" json:"amount"`
	Status        int             `json:"status"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

func (DistributorWithdrawal) TableName() string { return "distributor_withdrawal" }

type HomeConfig struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Config    string    `json:"config"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (HomeConfig) TableName() string { return "home_config" }

package model

import "time"

// 管理端操作审计（adminAuth 中间件自动记录非 GET 请求）
type AdminAuditLog struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	AdminID   int64     `json:"adminId"`
	AdminName string    `json:"adminName"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"createdAt"`
}

func (AdminAuditLog) TableName() string { return "admin_audit_log" }

const (
	SensitiveWordOff = 0
	SensitiveWordOn  = 1
)

type SensitiveWord struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Word      string    `json:"word"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

func (SensitiveWord) TableName() string { return "sensitive_word" }

// ─────────────────────────── 供货商 ───────────────────────────

type Supplier struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Contact   string    `json:"contact"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	LicenseNo string    `json:"licenseNo"`
	Remark    string    `json:"remark"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Supplier) TableName() string { return "supplier" }

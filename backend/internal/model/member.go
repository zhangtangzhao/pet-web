package model

import (
	"time"

	"gorm.io/gorm"
)

// 微信授权端类型
const (
	WxAppMini = 1 // 小程序
	WxAppH5   = 2 // 公众号 H5
	WxAppOpen = 3 // App 开放平台
)

type Member struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	Nickname    string     `json:"nickname"`
	Avatar      string     `json:"avatar"`
	Phone       string     `json:"phone"`
	Gender      int        `json:"gender"`
	Status      int        `json:"status"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func (Member) TableName() string { return "member" }

type WechatAuth struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	MemberID   int64     `json:"memberId"`
	AppType    int       `json:"appType"`
	OpenID     string    `json:"openId"`
	UnionID    string    `json:"unionId"`
	SessionKey string    `json:"-"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (WechatAuth) TableName() string { return "wechat_auth" }

// ─────────────────────────── 收藏 ───────────────────────────

type MemberFavorite struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	MemberID  int64     `json:"memberId"`
	ProductID int64     `json:"productId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (MemberFavorite) TableName() string { return "member_favorite" }

// ─────────────────────────── 平台用户 ───────────────────────────

const (
	AdminRoleSuper    = "super_admin"
	AdminRoleOperator = "operator"
)

type AdminUser struct {
	ID           int64      `gorm:"primaryKey" json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Nickname     string     `json:"nickname"`
	Role         string     `json:"role"`
	Status       int        `json:"status"`
	LastLoginAt  *time.Time `json:"lastLoginAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (AdminUser) TableName() string { return "admin_user" }

// ─────────────────────────── 通用更新 ───────────────────────────

func TouchUpdate(db *gorm.DB, table string, id int64) error {
	return db.Exec("UPDATE "+table+" SET updated_at = now() WHERE id = ?", id).Error
}

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
	ID                  int64      `gorm:"primaryKey" json:"id"`
	Nickname            string     `json:"nickname"`
	Avatar              string     `json:"avatar"`
	Phone               string     `json:"phone"`
	Gender              int        `json:"gender"`
	Status              int        `json:"status"`
	Points              int64      `json:"points"`
	GrowthValue         int64      `json:"growthValue"`  // 成长值 = 累计实付（元），只增不减
	LevelReached        int        `json:"levelReached"` // 已发放升级礼包的最高等级
	InviteCode          string     `json:"inviteCode"`
	InvitedBy           int64      `json:"invitedBy"`
	Blacklist           int        `json:"blacklist"` // 1=黑名单：可登录但禁交易/评价/领券
	Birthday            *time.Time `json:"birthday"`
	FreeShipCards       int        `json:"freeShipCards"`       // 免运费卡（积分商城权益）
	DeleteRequestedAt   *time.Time `json:"deleteRequestedAt"`   // 注销申请时间
	DeleteCooldownUntil *time.Time `json:"deleteCooldownUntil"` // 冷却截止，过期即匿名化
	LastLoginAt         *time.Time `json:"lastLoginAt"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

func (Member) TableName() string { return "member" }

// ─────────────────────────── 收货地址簿 ───────────────────────────

type MemberAddress struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	MemberID  int64     `json:"memberId"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	IsDefault int       `json:"isDefault"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (MemberAddress) TableName() string { return "member_address" }

// ─────────────────────────── 积分流水 ───────────────────────────

// 积分变动原因
const (
	PointsReasonSign     = "sign"          // 每日签到
	PointsReasonOrder    = "order"         // 消费返积分
	PointsReasonReview   = "review"        // 评价奖励
	PointsReasonInvite   = "invite"        // 邀请人奖励
	PointsReasonInvited  = "invite_reward" // 被邀请人奖励
	PointsReasonExchange = "exchange"      // 积分兑换优惠券
)

type PointsLog struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	MemberID     int64     `json:"memberId"`
	Change       int64     `json:"change"`
	BalanceAfter int64     `json:"balanceAfter"`
	Reason       string    `json:"reason"`
	Ref          string    `json:"ref"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (PointsLog) TableName() string { return "points_log" }

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
	AdminRoleSupport  = "support"
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

package model

import (
	"time"
)

// 通知场景
const (
	NotifySceneCsReply = 1 // 客服回复（离线提醒）
	NotifySceneOrder   = 2 // 订单状态（退款到账/超时关单/售后结果/自动确认）
	NotifySceneCoupon  = 3 // 优惠券提醒（到期提醒）
	NotifySceneCare    = 4 // 疫苗/驱虫护理提醒
	NotifySceneContent = 5 // 内容审核结果（晒单通过/隐藏）
)

// 投递状态
const (
	NotifyPending   = 0 // 待投递
	NotifyDelivered = 1 // 已投递
	NotifyFailed    = 2 // 失败（重试耗尽/用户未订阅等终态）
	NotifyDegrade   = 3 // 降级（模板未配置，仅落库留痕）
)

// 投递渠道
const (
	NotifyChannelNone = 0
	NotifyChannelMini = 1 // 小程序订阅消息
	NotifyChannelH5   = 2 // 公众号模板消息
)

type Notification struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	MemberID  int64      `json:"memberId"`
	Scene     int        `json:"scene"`
	BizKey    string     `json:"bizKey"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	OrderNo   string     `json:"orderNo"`
	Status    int        `json:"status"`
	Retry     int        `json:"retry"`
	Channel   int        `json:"channel"`
	ReadAt    *time.Time `json:"readAt"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (Notification) TableName() string { return "notification" }

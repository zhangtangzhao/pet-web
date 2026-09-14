package model

import "time"

// 客服会话状态
const (
	CsSessionOpen   = 1 // 进行中
	CsSessionClosed = 2 // 已结束
)

// 发送方角色
const (
	CsRoleMember = 1
	CsRoleAdmin  = 2
)

// 消息类型
const (
	CsMsgText  = 1
	CsMsgImage = 2
)

func CsStatusText(s int) string {
	if s == CsSessionClosed {
		return "已结束"
	}
	return "进行中"
}

type CsSession struct {
	ID              int64      `gorm:"primaryKey" json:"id"`
	MemberID        int64      `json:"memberId"`
	Status          int        `json:"status"`
	UnreadAdmin     int        `json:"unreadAdmin"`
	UnreadMember    int        `json:"unreadMember"`
	LastMessageText string     `json:"lastMessageText"`
	LastMessageAt   *time.Time `json:"lastMessageAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func (CsSession) TableName() string { return "cs_session" }

type CsMessage struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	SessionID  int64     `json:"sessionId"`
	SenderRole int       `json:"senderRole"`
	SenderID   int64     `json:"senderId"`
	MsgType    int       `json:"msgType"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (CsMessage) TableName() string { return "cs_message" }

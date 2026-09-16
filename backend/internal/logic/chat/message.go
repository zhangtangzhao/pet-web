package chat

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/hub"
	"pet/backend/internal/logic/notify"
	"pet/backend/internal/logic/sensitive"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

const (
	maxTextLen    = 500
	maxContentLen = 2048 // 图片 URL 上限
	previewLen    = 100  // 会话列表预览最大字符
)

// broadcaster 会话广播所需的最小 Hub 能力
type broadcaster interface {
	PushMember(id int64, v any)
	PushAdmins(v any)
}

func hubEvent(typ string, data any) hub.Event { return hub.Event{Type: typ, Data: data} }

// sessionUpdate 推送事件数据（双端共用）
type sessionUpdate struct {
	SessionID       string `json:"sessionId"`
	Status          int    `json:"status"`
	UnreadAdmin     int    `json:"unreadAdmin"`
	UnreadMember    int    `json:"unreadMember"`
	LastMessageText string `json:"lastMessageText"`
	LastMessageAt   string `json:"lastMessageAt"`
}

// Send 发送消息：落库 → 更新会话（预览/对方未读/重开）→ 广播。发送走 REST，WS 仅负责推送。
// 会员传 sessionID=0（自动定位本人会话）；客服传目标会话 ID。
func Send(sc *svc.ServiceContext, senderRole int, actorID, sessionID int64, req *types.CsSendReq) (*types.CsMsgOut, error) {
	msgType := req.MsgType
	if msgType == 0 {
		msgType = model.CsMsgText
	}
	if msgType != model.CsMsgText && msgType != model.CsMsgImage {
		return nil, common.ErrParam
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, common.ErrParam
	}
	if msgType == model.CsMsgText && utf8.RuneCountInString(content) > maxTextLen {
		return nil, common.ErrParam
	}
	if len(content) > maxContentLen {
		return nil, common.ErrParam
	}
	// 敏感词拦截（导流线下交易等）；仅文本校验
	if msgType == model.CsMsgText && sensitive.Contains(sc.DB, content) {
		return nil, common.ErrSensitive
	}

	var s *model.CsSession
	var err error
	if senderRole == model.CsRoleMember {
		s, err = EnsureSession(sc, actorID) // 成员侧 API 不出现 sessionID，天然免越权
	} else {
		s, err = SessionByID(sc, sessionID)
	}
	if err != nil {
		return nil, err
	}

	now := time.Now()
	msg := model.CsMessage{
		ID:         common.NewID(),
		SessionID:  s.ID,
		SenderRole: senderRole,
		SenderID:   actorID,
		MsgType:    msgType,
		Content:    content,
		CreatedAt:  now,
	}
	preview := previewText(msgType, content)

	err = sc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}
		updates := map[string]any{
			"last_message_text": preview,
			"last_message_at":   now,
			"updated_at":        now,
			"status":            model.CsSessionOpen, // 已结束会话自动重开
		}
		if senderRole == model.CsRoleMember {
			updates["unread_admin"] = gorm.Expr("unread_admin + 1")
		} else {
			updates["unread_member"] = gorm.Expr("unread_member + 1")
		}
		return tx.Model(&model.CsSession{}).Where("id = ?", s.ID).Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}

	out := msgOut(msg)
	if fresh, ferr := SessionByID(sc, s.ID); ferr == nil {
		broadcastSession(sc.Hub, fresh)
	}
	sc.Hub.PushMember(s.MemberID, hubEvent("new_message", out))
	sc.Hub.PushAdmins(hubEvent("new_message", out))

	// 客服回复且会员不在线 → 入队微信通知（在线者已实时收到，不打扰）
	if senderRole == model.CsRoleAdmin && !sc.Hub.MemberOnline(s.MemberID) {
		if err := notify.Enqueue(sc, s.MemberID, model.NotifySceneCsReply,
			"cs:"+strconv.FormatInt(msg.ID, 10), "客服回复", preview, ""); err != nil {
			logx.Errorf("客服回复通知入队失败 msgID=%d: %v", msg.ID, err)
		}
	}
	return out, nil
}

// History 历史消息游标分页：before 向上翻页（取更早），after 断线补拉；均按时间正序返回
func History(sc *svc.ServiceContext, sessionID, before, after int64, limit int) (*types.CsHistoryResp, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	q := sc.DB.Model(&model.CsMessage{}).Where("session_id = ?", sessionID)
	order := "id ASC"
	reverse := false
	if before > 0 {
		q = q.Where("id < ?", before)
		order = "id DESC"
		reverse = true
	} else if after > 0 {
		q = q.Where("id > ?", after)
	} else {
		order = "id DESC" // 无游标时取最新一页
		reverse = true
	}
	var msgs []model.CsMessage
	if err := q.Order(order).Limit(limit + 1).Find(&msgs).Error; err != nil {
		return nil, err
	}
	hasMore := len(msgs) > limit
	if hasMore {
		msgs = msgs[:limit]
	}
	if reverse {
		for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 { // 反转回时间正序
			msgs[i], msgs[j] = msgs[j], msgs[i]
		}
	}
	list := make([]types.CsMsgOut, 0, len(msgs))
	for _, m := range msgs {
		list = append(list, *msgOut(m))
	}
	return &types.CsHistoryResp{List: list, HasMore: hasMore}, nil
}

// MarkMemberRead 会员已读：清零本人会话的会员侧未读并广播
func MarkMemberRead(sc *svc.ServiceContext, memberID int64) error {
	s, err := EnsureSession(sc, memberID)
	if err != nil {
		return err
	}
	if err := sc.DB.Model(&model.CsSession{}).
		Where("id = ? AND unread_member > 0", s.ID).Update("unread_member", 0).Error; err != nil {
		return err
	}
	fresh, err := SessionByID(sc, s.ID)
	if err != nil {
		return err
	}
	broadcastSession(sc.Hub, fresh)
	return nil
}

// MarkSessionRead 客服已读：清零指定会话的客服侧未读并广播
func MarkSessionRead(sc *svc.ServiceContext, sessionID int64) error {
	if err := sc.DB.Model(&model.CsSession{}).
		Where("id = ? AND unread_admin > 0", sessionID).Update("unread_admin", 0).Error; err != nil {
		return err
	}
	fresh, err := SessionByID(sc, sessionID)
	if err != nil {
		return err
	}
	broadcastSession(sc.Hub, fresh)
	return nil
}

// ─────────────────────────── 内部 ───────────────────────────

func msgOut(m model.CsMessage) *types.CsMsgOut {
	return &types.CsMsgOut{
		ID:         strconv.FormatInt(m.ID, 10),
		SessionID:  strconv.FormatInt(m.SessionID, 10),
		SenderRole: m.SenderRole,
		SenderID:   strconv.FormatInt(m.SenderID, 10),
		MsgType:    m.MsgType,
		Content:    m.Content,
		CreatedAt:  m.CreatedAt.Format(time.RFC3339),
	}
}

func previewText(msgType int, content string) string {
	if msgType == model.CsMsgImage {
		return "[图片]"
	}
	r := []rune(content)
	if len(r) > previewLen {
		return string(r[:previewLen])
	}
	return content
}

func broadcastSession(h broadcaster, s *model.CsSession) {
	data := sessionUpdate{
		SessionID:       strconv.FormatInt(s.ID, 10),
		Status:          s.Status,
		UnreadAdmin:     s.UnreadAdmin,
		UnreadMember:    s.UnreadMember,
		LastMessageText: s.LastMessageText,
	}
	if s.LastMessageAt != nil {
		data.LastMessageAt = s.LastMessageAt.Format(time.RFC3339)
	}
	h.PushMember(s.MemberID, hubEvent("session_update", data))
	h.PushAdmins(hubEvent("session_update", data))
}

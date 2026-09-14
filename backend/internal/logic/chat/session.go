package chat

import (
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// EnsureSession 取会员会话，不存在则创建（并发下依赖 member_id 唯一约束兜底）
func EnsureSession(sc *svc.ServiceContext, memberID int64) (*model.CsSession, error) {
	var s model.CsSession
	err := sc.DB.Where("member_id = ?", memberID).First(&s).Error
	if err == nil {
		return &s, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	s = model.CsSession{
		ID:       common.NewID(),
		MemberID: memberID,
		Status:   model.CsSessionOpen,
	}
	if err := sc.DB.Create(&s).Error; err != nil {
		// 并发同时首条消息：唯一冲突后回读
		if err2 := sc.DB.Where("member_id = ?", memberID).First(&s).Error; err2 == nil {
			return &s, nil
		}
		return nil, err
	}
	return &s, nil
}

// SessionByID 客服侧按 ID 取会话
func SessionByID(sc *svc.ServiceContext, id int64) (*model.CsSession, error) {
	var s model.CsSession
	if err := sc.DB.First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrCsSession
		}
		return nil, err
	}
	return &s, nil
}

// AdminSessions 会话列表（进行中在前，按最后消息时间倒序），附带会员摘要
func AdminSessions(sc *svc.ServiceContext) ([]types.CsSessionItem, error) {
	var rows []struct {
		ID              int64
		MemberID        int64
		Status          int
		UnreadAdmin     int
		LastMessageText string
		LastMessageAt   *time.Time
		CreatedAt       time.Time
		Nickname        string
		Phone           string
		Avatar          string
	}
	err := sc.DB.Table("cs_session AS s").
		Joins("JOIN member m ON m.id = s.member_id").
		Select(`s.id, s.member_id, s.status, s.unread_admin, s.last_message_text,
		        s.last_message_at, s.created_at, m.nickname, m.phone, m.avatar`).
		Order("s.status ASC, COALESCE(s.last_message_at, s.created_at) DESC").
		Limit(200).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	list := make([]types.CsSessionItem, 0, len(rows))
	for _, r := range rows {
		item := types.CsSessionItem{
			ID:              strconv.FormatInt(r.ID, 10),
			MemberID:        strconv.FormatInt(r.MemberID, 10),
			Nickname:        r.Nickname,
			Phone:           r.Phone,
			Avatar:          r.Avatar,
			Status:          r.Status,
			UnreadAdmin:     r.UnreadAdmin,
			LastMessageText: r.LastMessageText,
			CreatedAt:       r.CreatedAt.Format(time.RFC3339),
		}
		if r.LastMessageAt != nil {
			item.LastMessageAt = r.LastMessageAt.Format(time.RFC3339)
		}
		list = append(list, item)
	}
	return list, nil
}

// CloseSession 客服结束会话（幂等）
func CloseSession(sc *svc.ServiceContext, id int64) error {
	s, err := SessionByID(sc, id)
	if err != nil {
		return err
	}
	res := sc.DB.Model(&model.CsSession{}).
		Where("id = ? AND status = ?", id, model.CsSessionOpen).
		Updates(map[string]any{"status": model.CsSessionClosed, "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		s.Status = model.CsSessionClosed
		broadcastSession(sc.Hub, s)
	}
	return nil
}

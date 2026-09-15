package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

const refreshKeyFmt = "refresh:member:%d:%s"

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// IssueTokens 签发 access + refresh，refresh 存 Redis 可吊销
func IssueTokens(sc *svc.ServiceContext, m *model.Member, isNew bool) (*types.LoginResp, error) {
	cfg := sc.Config.Auth
	access, err := common.GenToken(cfg.MemberAccessSecret, time.Duration(cfg.MemberAccessExpire)*time.Second, m.ID, common.TokenTypeMember, "")
	if err != nil {
		return nil, err
	}
	raw := randomToken()
	refreshTTL := time.Duration(sc.Config.RefreshTokenExpireDays) * 24 * time.Hour
	if err := sc.Rdb.Set(context.Background(), fmt.Sprintf(refreshKeyFmt, m.ID, raw), "1", refreshTTL).Err(); err != nil {
		return nil, err
	}
	return &types.LoginResp{
		AccessToken:  access,
		RefreshToken: fmt.Sprintf("%d.%s", m.ID, raw),
		ExpiresIn:    cfg.MemberAccessExpire,
		Member:       MemberView(sc, m, isNew),
	}, nil
}

// MemberView 会员信息视图
func MemberView(sc *svc.ServiceContext, m *model.Member, isNew bool) types.MemberInfo {
	var wxCnt int64
	_ = sc.DB.Model(&model.WechatAuth{}).Where("member_id = ?", m.ID).Count(&wxCnt).Error
	phone := m.Phone
	if len(phone) == 11 {
		phone = phone[:3] + "****" + phone[7:]
	}
	return types.MemberInfo{
		ID:        fmt.Sprintf("%d", m.ID),
		Nickname:  m.Nickname,
		Avatar:    m.Avatar,
		Phone:     phone,
		Gender:    m.Gender,
		HasWxBind: wxCnt > 0,
		IsNew:     isNew,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
	}
}

func touchLastLogin(db *gorm.DB, memberID int64) {
	now := time.Now()
	_ = db.Model(&model.Member{}).Where("id = ?", memberID).Updates(map[string]any{
		"last_login_at": &now,
	}).Error
}

// Refresh 校验并轮换 refresh_token
func Refresh(sc *svc.ServiceContext, refreshToken string) (*types.LoginResp, error) {
	if refreshToken == "" {
		return nil, common.ErrUnauthorized
	}
	// refresh_token 中不含 uid，遍历无法定位 → 采用 token 自携带 uid 的双段格式
	var memberID int64
	var raw string
	if n, err := fmt.Sscanf(refreshToken, "%d.%s", &memberID, &raw); err != nil || n != 2 {
		return nil, common.ErrUnauthorized
	}
	key := fmt.Sprintf(refreshKeyFmt, memberID, raw)
	val, err := sc.Rdb.GetDel(context.Background(), key).Result()
	if err != nil || val != "1" {
		return nil, common.ErrTokenExpired
	}
	var m model.Member
	if err := sc.DB.First(&m, memberID).Error; err != nil {
		return nil, common.ErrTokenExpired
	}
	if m.Status != 1 {
		return nil, common.ErrForbidden
	}
	touchLastLogin(sc.DB, m.ID)
	return IssueTokens(sc, &m, false)
}

// Logout 吊销 refresh_token
func Logout(sc *svc.ServiceContext, memberID int64, refreshToken string) {
	if refreshToken == "" {
		return
	}
	sc.Rdb.Del(context.Background(), fmt.Sprintf(refreshKeyFmt, memberID, refreshToken))
}

// RefreshTokenFormat 生成 "{uid}.{random}" 格式
func RefreshTokenFormat(uid int64) string {
	return fmt.Sprintf("%d.%s", uid, randomToken())
}

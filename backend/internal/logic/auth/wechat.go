package auth

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/marketing"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// WechatMiniLogin 小程序登录
func WechatMiniLogin(sc *svc.ServiceContext, code string) (*types.LoginResp, error) {
	if code == "" {
		return nil, common.ErrParam
	}
	s, err := sc.MiniCode2Session(context.Background(), code)
	if err != nil {
		return nil, err
	}
	return wechatLogin(sc, model.WxAppMini, s.OpenID, s.UnionID, s.SessionKey)
}

// WechatH5Login 公众号网页授权登录
func WechatH5Login(sc *svc.ServiceContext, code string) (*types.LoginResp, error) {
	if code == "" {
		return nil, common.ErrParam
	}
	t, err := sc.H5Code2OpenID(context.Background(), code)
	if err != nil {
		return nil, err
	}
	return wechatLogin(sc, model.WxAppH5, t.OpenID, t.UnionID, "")
}

func H5OAuthURL(sc *svc.ServiceContext, redirect string) (string, error) {
	return sc.H5OAuthURL(redirect)
}

func createWxAuth(sc *svc.ServiceContext, memberID int64, appType int, openid, unionid, sessionKey string) error {
	return sc.DB.Create(&model.WechatAuth{
		ID:         common.NewID(),
		MemberID:   memberID,
		AppType:    appType,
		OpenID:     openid,
		UnionID:    unionid,
		SessionKey: sessionKey,
	}).Error
}

// wechatLogin 三段式：已有绑定 → unionid 归并 → 新用户注册
func wechatLogin(sc *svc.ServiceContext, appType int, openid, unionid, sessionKey string) (*types.LoginResp, error) {
	var auth model.WechatAuth
	err := sc.DB.Where("app_type = ? AND openid = ?", appType, openid).First(&auth).Error
	if err == nil {
		var m model.Member
		if err := sc.DB.First(&m, auth.MemberID).Error; err != nil {
			return nil, common.ErrInternal
		}
		if m.Status != 1 {
			return nil, common.ErrForbidden
		}
		if unionid != "" && auth.UnionID == "" {
			_ = sc.DB.Model(&model.WechatAuth{}).Where("id = ?", auth.ID).Update("union_id", unionid).Error
		}
		touchLastLogin(sc.DB, m.ID)
		return IssueTokens(sc, &m, false)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// unionid 归并：同一自然人多端同一账号
	if unionid != "" {
		var bind model.WechatAuth
		if e := sc.DB.Where("union_id = ?", unionid).First(&bind).Error; e == nil {
			var m model.Member
			if err := sc.DB.First(&m, bind.MemberID).Error; err == nil {
				if err := createWxAuth(sc, m.ID, appType, openid, unionid, sessionKey); err != nil {
					return nil, err
				}
				touchLastLogin(sc.DB, m.ID)
				return IssueTokens(sc, &m, false)
			}
		}
	}

	// 新用户
	m := model.Member{ID: common.NewID(), Nickname: "微信用户", Status: 1}
	if err := sc.DB.Create(&m).Error; err != nil {
		return nil, err
	}
	go marketing.GrantNewUserCoupons(sc, m.ID)
	if err := createWxAuth(sc, m.ID, appType, openid, unionid, sessionKey); err != nil {
		return nil, err
	}
	touchLastLogin(sc.DB, m.ID)
	return IssueTokens(sc, &m, true)
}

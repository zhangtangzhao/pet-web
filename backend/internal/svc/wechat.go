package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"pet/backend/internal/common"
)

type WxSession struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type WxOAuthToken struct {
	AccessToken string `json:"access_token"`
	OpenID      string `json:"openid"`
	UnionID     string `json:"unionid"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

var wxHTTP = &http.Client{Timeout: 5 * time.Second}

func wxGetJSON(rawURL string, out any) error {
	resp, err := wxHTTP.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return err
	}
	return nil
}

// MiniCode2Session 小程序 code 换 openid/session_key
func (sc *ServiceContext) MiniCode2Session(ctx context.Context, code string) (*WxSession, error) {
	c := sc.Config.WeChat
	if c.MiniAppID == "" {
		return nil, common.NewErr(500, 50001, "小程序登录未配置")
	}
	u := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(c.MiniAppID), url.QueryEscape(c.MiniSecret), url.QueryEscape(code))
	var s WxSession
	if err := wxGetJSON(u, &s); err != nil {
		return nil, common.ErrWxAPI
	}
	if s.ErrCode != 0 {
		return nil, common.NewErr(502, 50001, "微信登录失败:"+s.ErrMsg)
	}
	return &s, nil
}

// H5OAuthURL 构造公众号网页授权跳转地址（snsapi_userinfo）
func (sc *ServiceContext) H5OAuthURL(redirect string) (string, error) {
	c := sc.Config.WeChat
	if c.H5AppID == "" {
		return "", common.NewErr(500, 50001, "H5 微信授权未配置")
	}
	rd, _ := url.QueryUnescape(redirect)
	if rd == "" {
		rd = redirect
	}
	return fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_userinfo&state=pet#wechat_redirect",
		c.H5AppID, url.QueryEscape(rd)), nil
}

// H5Code2OpenID 网页授权 code 换 openid/unionid
func (sc *ServiceContext) H5Code2OpenID(ctx context.Context, code string) (*WxOAuthToken, error) {
	c := sc.Config.WeChat
	if c.H5AppID == "" {
		return nil, common.NewErr(500, 50001, "H5 微信授权未配置")
	}
	u := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		url.QueryEscape(c.H5AppID), url.QueryEscape(c.H5Secret), url.QueryEscape(code))
	var t WxOAuthToken
	if err := wxGetJSON(u, &t); err != nil {
		return nil, common.ErrWxAPI
	}
	if t.ErrCode != 0 {
		return nil, common.NewErr(502, 50001, "微信授权失败:"+t.ErrMsg)
	}
	return &t, nil
}

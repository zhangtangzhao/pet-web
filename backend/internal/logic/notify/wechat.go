package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"pet/backend/internal/model"
	"pet/backend/internal/svc"
)

// wxAPIError 微信业务错误（errcode 非 0）
type wxAPIError struct {
	ErrCode int
	ErrMsg  string
}

func (e *wxAPIError) Error() string { return fmt.Sprintf("wx errcode=%d %s", e.ErrCode, e.ErrMsg) }

var gatMtx sync.Mutex // 全局 access_token 刷新单飞

// globalAccessToken 微信全局 access_token：Redis 缓存 7000s（官方 7200s 提前失效），并发单飞刷新
func globalAccessToken(sc *svc.ServiceContext, appType int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("wx:gat:%d", appType)
	if tok, err := sc.Rdb.Get(ctx, key).Result(); err == nil && tok != "" {
		return tok, nil
	}
	gatMtx.Lock()
	defer gatMtx.Unlock()
	if tok, err := sc.Rdb.Get(ctx, key).Result(); err == nil && tok != "" {
		return tok, nil
	}

	var appID, secret string
	if appType == model.WxAppMini {
		appID, secret = sc.Config.WeChat.MiniAppID, sc.Config.WeChat.MiniSecret
	} else {
		appID, secret = sc.Config.WeChat.H5AppID, sc.Config.WeChat.H5Secret
	}
	if appID == "" {
		return "", &wxAPIError{ErrCode: -1, ErrMsg: "微信应用未配置"}
	}
	u := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(appID), url.QueryEscape(secret))
	var t struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := wxGetJSON(u, &t); err != nil {
		return "", err
	}
	if t.ErrCode != 0 || t.AccessToken == "" {
		return "", &wxAPIError{ErrCode: t.ErrCode, ErrMsg: t.ErrMsg}
	}
	_ = sc.Rdb.Set(ctx, key, t.AccessToken, 7000*time.Second).Err()
	return t.AccessToken, nil
}

// wxGetJSON / wxPostJSON 微信 API 调用（5s 超时），统一解析 errcode
var wxClient = &http.Client{Timeout: 5 * time.Second}

func wxGetJSON(rawURL string, out any) error {
	resp, err := wxClient.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeWx(resp.Body, out)
}

func wxPostJSON(rawURL string, body any, out any) error {
	bs, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := wxClient.Post(rawURL, "application/json", bytes.NewReader(bs))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeWx(resp.Body, out)
}

func decodeWx(r io.Reader, out any) error {
	bs, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, out); err != nil {
		return err
	}
	var ec struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	_ = json.Unmarshal(bs, &ec)
	if ec.ErrCode != 0 {
		return &wxAPIError{ErrCode: ec.ErrCode, ErrMsg: ec.ErrMsg}
	}
	return nil
}

// 模板字段约定：thing1=内容（≤20字）、thing2=单号、time3=时间
type tmplValue struct {
	Value string `json:"value"`
}

func tmplPayload(n *model.Notification, page, tmplID, openID string) map[string]any {
	data := map[string]tmplValue{
		"thing1": {Value: n.Content},
		"time3":  {Value: n.CreatedAt.Format("2006-01-02 15:04")},
	}
	if n.OrderNo != "" {
		data["thing2"] = tmplValue{Value: n.OrderNo}
	}
	return map[string]any{
		"touser":      openID,
		"template_id": tmplID,
		"data":        data,
		"page":        page,
		"url":         "",
	}
}

func notifyPage(scene int) string {
	if scene == model.NotifySceneCsReply {
		return "pages/service-chat/index"
	}
	return "pages/orders/index"
}

// sendMiniSubscribe 小程序订阅消息（cgi-bin/message/subscribe/send）
func sendMiniSubscribe(ctx context.Context, sc *svc.ServiceContext, tmplID, openID string, n *model.Notification) error {
	tok, err := globalAccessToken(sc, model.WxAppMini)
	if err != nil {
		return err
	}
	body := tmplPayload(n, notifyPage(n.Scene), tmplID, openID)
	delete(body, "url") // 订阅消息无 url 字段
	var out map[string]any
	return wxPostJSON("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token="+tok, body, &out)
}

// sendH5Template 公众号模板消息（cgi-bin/message/template/send）
func sendH5Template(ctx context.Context, sc *svc.ServiceContext, tmplID, openID string, n *model.Notification) error {
	tok, err := globalAccessToken(sc, model.WxAppH5)
	if err != nil {
		return err
	}
	body := tmplPayload(n, "", tmplID, openID)
	delete(body, "page") // 模板消息无 page 字段
	var out map[string]any
	return wxPostJSON("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token="+tok, body, &out)
}

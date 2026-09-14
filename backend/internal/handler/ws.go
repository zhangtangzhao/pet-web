package handler

import (
	"net/http"

	"github.com/coder/websocket"

	"pet/backend/internal/common"
	"pet/backend/internal/hub"
	"pet/backend/internal/svc"
)

// CsWS 人工客服 WebSocket 端点：GET /api/ws/cs?token=<accessToken>
// 浏览器 WebSocket 无法携带 Authorization 头，统一以 query 传 token；
// 先按 admin secret 校验再按 member secret，claims.typ 决定连接角色。
func CsWS(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		role, uid, ok := identifyWS(sc, token)
		if !ok {
			common.Err(w, common.ErrUnauthorized)
			return
		}
		ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"}, // 鉴权由 token 负责；开发环境 vite 代理与本站同源之外还有小程序端
		})
		if err != nil {
			return
		}
		sc.Hub.Serve(role, uid, ws)
	}
}

func identifyWS(sc *svc.ServiceContext, token string) (string, int64, bool) {
	if uid, ok := parseQueryAs(token, sc.Config.Auth.AdminAccessSecret, common.TokenTypeAdmin); ok {
		return hub.RoleAdmin, uid, true
	}
	if uid, ok := parseQueryAs(token, sc.Config.Auth.MemberAccessSecret, common.TokenTypeMember); ok {
		return hub.RoleMember, uid, true
	}
	return "", 0, false
}

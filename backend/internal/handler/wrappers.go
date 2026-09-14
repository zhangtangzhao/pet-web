package handler

import (
	"context"
	"net/http"

	"pet/backend/internal/common"
	"pet/backend/internal/svc"
)

type ctxKey int

const (
	ctxMemberID ctxKey = iota
	ctxAdminID
	ctxAdminRole
)

// memberAuth 用户端 JWT 校验（独立 secret，typ=member）
func memberAuth(sc *svc.ServiceContext, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := parseAs(r, sc.Config.Auth.MemberAccessSecret, common.TokenTypeMember)
		if !ok {
			common.Err(w, common.ErrUnauthorized)
			return
		}
		next(w, r.WithContext(withID(r.Context(), ctxMemberID, uid)))
	}
}

// adminAuth 平台端 JWT 校验（独立 secret，typ=admin）
func adminAuth(sc *svc.ServiceContext, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := parseAs(r, sc.Config.Auth.AdminAccessSecret, common.TokenTypeAdmin)
		if !ok {
			common.Err(w, common.ErrUnauthorized)
			return
		}
		ctx := withID(r.Context(), ctxAdminID, uid)
		next(w, r.WithContext(ctx))
	}
}

func withID(ctx context.Context, key ctxKey, id int64) context.Context {
	return context.WithValue(ctx, key, id)
}

func parseAs(r *http.Request, secret, typ string) (int64, bool) {
	return parseQueryAs(common.BearerToken(r), secret, typ)
}

// parseQueryAs 校验裸 token（WebSocket 升级请求无法携带 Authorization 头，走 query 参数）
func parseQueryAs(token, secret, typ string) (int64, bool) {
	if token == "" {
		return 0, false
	}
	mc, err := common.ParseToken(secret, token)
	if err != nil {
		return 0, false
	}
	if t, _ := mc["typ"].(string); t != typ {
		return 0, false
	}
	uid := common.UIDFromClaims(mc)
	if uid <= 0 {
		return 0, false
	}
	return uid, true
}

// memberID 取当前登录会员 ID
func memberID(r *http.Request) int64 {
	v, _ := r.Context().Value(ctxMemberID).(int64)
	return v
}

// adminID 取当前管理员 ID
func adminID(r *http.Request) int64 {
	v, _ := r.Context().Value(ctxAdminID).(int64)
	return v
}

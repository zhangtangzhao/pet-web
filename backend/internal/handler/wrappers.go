package handler

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
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

// adminAuth 平台端 JWT 校验（独立 secret，typ=admin）+ 角色门禁；非 GET 请求异步落操作审计
func adminAuth(sc *svc.ServiceContext, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := parseAs(r, sc.Config.Auth.AdminAccessSecret, common.TokenTypeAdmin)
		if !ok {
			common.Err(w, common.ErrUnauthorized)
			return
		}
		var u model.AdminUser
		if err := sc.DB.Select("role", "status").First(&u, uid).Error; err != nil || u.Status != 1 {
			common.Err(w, common.ErrUnauthorized)
			return
		}
		if !rbacAllow(u.Role, r.Method, r.URL.Path) {
			common.Err(w, common.ErrForbidden)
			return
		}
		ctx := withID(r.Context(), ctxAdminID, uid)
		if r.Method != http.MethodGet {
			go auditAdmin(sc, uid, r.Method, r.URL.Path, clientIP(r))
		}
		next(w, r.WithContext(ctx))
	}
}

// rbacAllow 角色门禁：super_admin 全通过；
// operator（运营）禁退款、会员写、账号管理；support（客服）仅客服会话 + 只读订单/评价/概览 + 评价回复
func rbacAllow(role, method, path string) bool {
	if role == "" || role == model.AdminRoleSuper {
		return true
	}
	isGet := method == http.MethodGet
	switch role {
	case model.AdminRoleOperator:
		if strings.HasSuffix(path, "/refund") {
			return false
		}
		if strings.HasPrefix(path, "/api/admin/members") && !isGet {
			return false
		}
		if strings.HasPrefix(path, "/api/admin/admins") {
			return false
		}
		return true
	case model.AdminRoleSupport:
		if strings.Contains(path, "/cs/") {
			return true
		}
		if isGet && (strings.HasPrefix(path, "/api/admin/overview") ||
			strings.HasPrefix(path, "/api/admin/orders") ||
			strings.HasPrefix(path, "/api/admin/reviews")) {
			return true
		}
		if !isGet && strings.HasSuffix(path, "/reply") {
			return true
		}
		return false
	}
	return false
}

// auditAdmin 异步记录管理端写操作（method+path+ip），失败仅日志不影响请求
func auditAdmin(sc *svc.ServiceContext, adminID int64, method, path, ip string) {
	a := model.AdminAuditLog{
		ID: common.NewID(), AdminID: adminID, Method: method, IP: ip,
	}
	if rs := []rune(path); len(rs) > 128 {
		a.Path = string(rs[:128])
	} else {
		a.Path = path
	}
	var u model.AdminUser
	if err := sc.DB.Select("username").First(&u, adminID).Error; err == nil {
		a.AdminName = u.Username
	}
	if err := sc.DB.Create(&a).Error; err != nil {
		logx.Errorf("审计日志落库失败 admin=%d path=%s: %v", adminID, path, err)
	}
}

func clientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		return strings.TrimSpace(strings.Split(xf, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
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

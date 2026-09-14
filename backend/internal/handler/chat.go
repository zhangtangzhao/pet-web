package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/chat"
	"pet/backend/internal/logic/manage"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// parseCsPath 解析路径中的会话 ID
func parseCsPath(r *http.Request) (int64, error) {
	var path types.IDPathReq
	if err := httpx.ParsePath(r, &path); err != nil {
		return 0, common.ErrParam
	}
	return parseID64(path.ID)
}

// ─────────────────────────── 用户端 ───────────────────────────

// CsMemberSend POST /api/cs/messages
func CsMemberSend(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.CsSendReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		out, err := chat.Send(sc, model.CsRoleMember, memberID(r), 0, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, out)
	})
}

// CsMemberMessages GET /api/cs/messages?before=&after=&limit=
func CsMemberMessages(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.CsHistoryReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		s, err := chat.EnsureSession(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		resp, err := chat.History(sc, s.ID, req.Before, req.After, req.Limit)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// CsMemberRead POST /api/cs/read
func CsMemberRead(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		if err := chat.MarkMemberRead(sc, memberID(r)); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// MemberUploadToken POST /api/upload-token —— 聊天图片直传凭证（目录固定 chat）
func MemberUploadToken(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadTokenReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		req.Dir = "chat"
		resp, err := manage.UploadToken(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// ─────────────────────────── 平台端 ───────────────────────────

// AdminCsSessions GET /api/admin/cs/sessions
func AdminCsSessions(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		list, err := chat.AdminSessions(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	})
}

// AdminCsMessages GET/POST /api/admin/cs/sessions/:id/messages
func AdminCsMessages(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		id, err := parseCsPath(r)
		if err != nil {
			common.Err(w, err)
			return
		}
		if r.Method == http.MethodPost {
			var req types.CsSendReq
			if err := common.ParseBody(r, &req); err != nil {
				common.Err(w, err)
				return
			}
			out, err := chat.Send(sc, model.CsRoleAdmin, adminID(r), id, &req)
			if err != nil {
				common.Err(w, err)
				return
			}
			common.OK(w, out)
			return
		}
		var req types.CsHistoryReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if _, err := chat.SessionByID(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		resp, err := chat.History(sc, id, req.Before, req.After, req.Limit)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// AdminCsRead POST /api/admin/cs/sessions/:id/read
func AdminCsRead(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		id, err := parseCsPath(r)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := chat.MarkSessionRead(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// AdminCsClose POST /api/admin/cs/sessions/:id/close
func AdminCsClose(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		id, err := parseCsPath(r)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := chat.CloseSession(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

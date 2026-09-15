package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/notify"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// NotifyTmpl GET /api/notify/tmpl —— 返回已配置的微信通知模板 ID（小程序 requestSubscribeMessage 用）
func NotifyTmpl(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		c := sc.Config.Notify
		common.OK(w, types.NotifyTmplResp{
			MiniTmplCsReply: c.MiniTmplCsReply,
			MiniTmplOrder:   c.MiniTmplOrder,
			H5TmplCsReply:   c.H5TmplCsReply,
			H5TmplOrder:     c.H5TmplOrder,
		})
	})
}

// NotifyList GET /api/notify/list —— 站内消息中心列表（分页 + 未读数）
func NotifyList(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.PageReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := notify.List(sc, memberID(r), &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// NotifyRead POST /api/notify/read —— 全部已读
func NotifyRead(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		if err := notify.ReadAll(sc, memberID(r)); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// NotifyReadOne POST /api/notify/:id/read —— 单条已读
func NotifyReadOne(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		id, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := notify.ReadOne(sc, memberID(r), id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

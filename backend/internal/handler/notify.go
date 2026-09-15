package handler

import (
	"net/http"

	"pet/backend/internal/common"
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

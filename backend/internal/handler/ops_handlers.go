// ops_handlers.go 平台端新增能力：报表深化、秒杀管理、审计日志、敏感词。
package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/manage"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func AdminReport(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Days int `form:"days,optional"`
		}
		_ = httpx.ParseForm(r, &req)
		resp, err := manage.AdminReport(sc, req.Days)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminFlashSaleList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.PageReq
		_ = httpx.ParseForm(r, &req)
		resp, err := manage.AdminFlashSaleList(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminFlashSaleUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.FlashSaleUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.AdminFlashSaleUpsert(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminFlashSaleStatus(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var req types.BannerStatusReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		id, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.AdminFlashSaleStatus(sc, id, req.Status); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminFlashSaleDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
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
		if err := manage.AdminFlashSaleDelete(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminAuditList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.PageReq
		_ = httpx.ParseForm(r, &req)
		resp, err := manage.AdminAuditList(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminSensitiveList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.PageReq
		_ = httpx.ParseForm(r, &req)
		resp, err := manage.AdminSensitiveList(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminSensitiveSave(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.SensitiveWordSaveReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.AdminSensitiveSave(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminSensitiveStatus(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var req types.BannerStatusReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		id, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.AdminSensitiveStatus(sc, id, req.Status); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminSensitiveDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
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
		if err := manage.AdminSensitiveDelete(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

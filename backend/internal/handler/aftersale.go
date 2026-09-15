package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/aftersale"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// AfterSaleApply POST /api/aftersale —— 会员发起售后
func AfterSaleApply(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.AfterSaleApplyReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		out, err := aftersale.Apply(sc, memberID(r), &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, out)
	})
}

// AfterSaleByOrder GET /api/aftersale?orderNo= —— 查询订单的售后单（本人）
func AfterSaleByOrder(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.AfterSaleOrderReq
		if err := httpx.ParseForm(r, &req); err != nil || req.OrderNo == "" {
			common.Err(w, common.ErrParam)
			return
		}
		out, err := aftersale.GetByOrder(sc, memberID(r), req.OrderNo)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, out)
	})
}

// AfterSaleCancel POST /api/aftersale/:afterSaleNo/cancel —— 会员撤销（仅待审核）
func AfterSaleCancel(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.AfterSaleNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := aftersale.Cancel(sc, memberID(r), path.AfterSaleNo); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// ─────────────────────────── 平台端 ───────────────────────────

// AdminAfterSales GET /api/admin/aftersales —— 售后列表
func AdminAfterSales(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.AfterSaleAdminListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := aftersale.AdminList(sc, req.Status, req.PageReq)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// AdminAfterSaleAudit POST /api/admin/aftersales/:afterSaleNo/audit —— 审核（同意可调金额 / 拒绝需备注）
func AdminAfterSaleAudit(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.AfterSaleNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var req types.AfterSaleAuditReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := aftersale.Audit(sc, path.AfterSaleNo, req.Agree, req.Amount, req.Note); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// commerce_handlers.go 购物车 / 拼团 / 宠物档案 / 自提核销 / 库存预警 / 风控记录。
package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/growth"
	"pet/backend/internal/logic/manage"
	"pet/backend/internal/logic/trade"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// GroupBuys 拼团活动列表（公开）
func GroupBuys(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := trade.GroupBuys(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

// CartList 我的购物车
func CartList(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := trade.CartList(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// CartAdd 加入购物车
func CartAdd(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.CartAddReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := trade.CartAdd(sc, memberID(r), &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// CartUpdate 勾选/取消勾选
func CartUpdate(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.CartUpdateReq
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var body struct {
			Checked *int `json:"checked,optional"`
		}
		_ = common.ParseBody(r, &body)
		cartID, err := parseID64(req.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := trade.CartUpdate(sc, memberID(r), cartID, body.Checked); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// CartDelete 删除单项（带 id）或清空（不带）
func CartDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var cartID int64
		var req types.CartUpdateReq
		if err := httpx.ParsePath(r, &req); err == nil && req.ID != "" {
			id, perr := parseID64(req.ID)
			if perr != nil {
				common.Err(w, perr)
				return
			}
			cartID = id
		}
		if err := trade.CartDelete(sc, memberID(r), cartID); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// UserPetList 我的宠物档案
func UserPetList(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := growth.UserPetList(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// UserPetUpsert 新建/编辑宠物档案
func UserPetUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.UserPetUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := growth.UserPetUpsert(sc, memberID(r), &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// UserPetUpdate 编辑宠物档案（path id 优先）
func UserPetUpdate(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.UserPetUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := growth.UserPetUpsert(sc, memberID(r), &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// UserPetDelete 删除宠物档案
func UserPetDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `path:"id"`
		}
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		id, err := parseID64(req.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := growth.UserPetDelete(sc, memberID(r), id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// AdminStockAlerts 库存预警清单
func AdminStockAlerts(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := manage.AdminStockAlerts(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// AdminRiskLogs 风控拦截记录
func AdminRiskLogs(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.RiskLogListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := manage.AdminRiskLogs(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// AdminPickupVerify 自提核销
func AdminPickupVerify(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.PickupVerifyReq
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var body struct {
			Code string `json:"code"`
		}
		if err := common.ParseBody(r, &body); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.AdminPickupVerify(sc, req.OrderNo, body.Code); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// AdminMemberBlacklist 拉黑/解除
func AdminMemberBlacklist(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.MemberBlacklistReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.UpdateMemberBlacklist(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

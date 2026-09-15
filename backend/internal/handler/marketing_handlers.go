package handler

import (
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/marketing"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func parseID64(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, common.ErrParam
	}
	return id, nil
}

// ─────────────────────────── 用户端 · 增值服务/优惠券 ───────────────────────────

func ServiceList(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := marketing.ServiceList(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	}
}

func CouponCenter(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		list, err := marketing.CouponCenter(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	})
}

func ClaimCoupon(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		id, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := marketing.Claim(sc, memberID(r), id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func MyCoupons(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Status int `form:"status,optional"`
		}
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		list, err := marketing.MyCoupons(sc, memberID(r), req.Status)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	})
}

func UsableCoupons(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.UsableCouponsReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		productID, err := parseID64(req.ProductID)
		if err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		serviceIDs, err := marketing.ParseIDs(req.ServiceIDs)
		if err != nil {
			common.Err(w, err)
			return
		}
		list, err := marketing.UsableCoupons(sc, memberID(r), productID, serviceIDs)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	})
}

// ─────────────────────────── 平台端 · 营销管理 ───────────────────────────

func AdminServiceList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		list, err := marketing.AdminServiceList(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	})
}

func AdminServiceUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.ServiceUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := marketing.AdminServiceUpsert(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminServiceDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		id, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := marketing.AdminServiceDelete(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminCouponList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.CouponAdminListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := marketing.AdminCouponList(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminCouponUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.CouponUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := marketing.AdminCouponUpsert(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminCouponDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		id, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := marketing.AdminCouponDelete(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminCouponIssue(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		templateID, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var req types.CouponIssueReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		memberIDs, err := marketing.ParseIDs(req.MemberIDs)
		if err != nil {
			common.Err(w, err)
			return
		}
		if len(memberIDs) == 0 {
			common.Err(w, common.ErrParam)
			return
		}
		issued, failed := marketing.IssueToMembers(sc, templateID, memberIDs)
		common.OK(w, map[string]int{"issued": issued, "failed": failed})
	})
}

// ─────────────────────────── 平台端 · Banner 运营位 ───────────────────────────

func AdminBannerList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.BannerListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := marketing.AdminBannerList(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminBannerUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.BannerUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err == nil && path.ID != "" {
			req.ID = path.ID
		}
		if err := marketing.AdminBannerUpsert(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminBannerStatus(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		id, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var req types.BannerStatusReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := marketing.AdminBannerStatus(sc, id, req.Status); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminBannerDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		id, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := marketing.AdminBannerDelete(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

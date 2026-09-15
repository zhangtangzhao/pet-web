package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/review"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// ProductReviews GET /api/products/:id/reviews —— 商品评价列表（公开，仅显示中）
func ProductReviews(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		productID, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		var req types.ReviewListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := review.ListByProduct(sc, productID, req.Cursor, req.Limit)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

// ProductReviewSummary GET /api/products/:id/review-summary —— 评价摘要（公开）
func ProductReviewSummary(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		productID, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		resp, err := review.ProductReviewSummary(sc, productID)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

// OrderReview POST /api/orders/:orderNo/review —— 会员提交评价
func OrderReview(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.OrderNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var req types.ReviewCreateReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		req.OrderNo = path.OrderNo
		out, err := review.Create(sc, memberID(r), &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, out)
	})
}

// ─────────────────────────── 平台端 ───────────────────────────

// AdminReviews GET /api/admin/reviews —— 评价列表（含隐藏）
func AdminReviews(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.ReviewListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := review.AdminList(sc, req.Cursor, req.Limit)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// AdminReviewStatus PUT /api/admin/reviews/:id/status —— 隐藏/显示
func AdminReviewStatus(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		id, err := parseCsPath(r)
		if err != nil {
			common.Err(w, err)
			return
		}
		var req types.ReviewStatusReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := review.AdminSetStatus(sc, id, req.Status); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// AdminReviewDelete DELETE /api/admin/reviews/:id
func AdminReviewDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		id, err := parseCsPath(r)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := review.AdminDelete(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

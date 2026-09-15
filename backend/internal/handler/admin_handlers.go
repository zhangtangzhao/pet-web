package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/manage"
	"pet/backend/internal/logic/trade"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func AdminCaptcha(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := manage.GenCaptcha()
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

func AdminLogin(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminLoginReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		resp, err := manage.AdminLogin(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

func AdminRefresh(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshTokenReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		resp, err := manage.AdminRefresh(sc, req.RefreshToken)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

func AdminLogout(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshTokenReq
		_ = common.ParseBody(r, &req)
		manage.AdminLogout(sc, adminID(r), req.RefreshToken)
		common.OK(w, nil)
	})
}

func AdminChangePassword(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.ChangePasswordReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.ChangeAdminPassword(sc, adminID(r), &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminOverview(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := manage.Overview(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// ─────────────────────────── 分类/品种 ───────────────────────────

func AdminCategoryUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.CategoryUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.UpsertCategory(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// AdminCategoryUpdate 编辑分类（path id 优先于 body id）
func AdminCategoryUpdate(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var req types.CategoryUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		req.ID = path.ID
		if err := manage.UpsertCategory(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminCategoryDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := manage.DeleteCategory(sc, path.ID); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminBreedUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.BreedUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.UpsertBreed(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminBreedUpdate(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var req types.BreedUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		req.ID = path.ID
		if err := manage.UpsertBreed(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminBreedDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := manage.DeleteBreed(sc, path.ID); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// ─────────────────────────── 商品 ───────────────────────────

func AdminProductList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminProductListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := manage.AdminProductList(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminProductDetail(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := manage.AdminProductDetail(sc, path.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminProductCreate(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.ProductUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.UpsertProduct(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminProductUpdate(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var req types.ProductUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		req.ID = path.ID
		if err := manage.UpsertProduct(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminProductStatus(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.ProductStatusReq
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.UpdateProductStatus(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminUploadToken(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadTokenReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		resp, err := manage.UploadToken(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// ─────────────────────────── 订单 ───────────────────────────

func AdminOrderList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminOrderListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := manage.AdminOrderList(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminOrderDetail(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.OrderNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := manage.AdminOrderDetail(sc, path.OrderNo)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminRefund(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.RefundReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		var path types.OrderNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		req.OrderNo = path.OrderNo
		if err := manage.AdminRefund(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminShip(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.OrderNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var req types.ShipReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := trade.AdminShip(sc, path.OrderNo, req.ShipNo); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminDeliver(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.OrderNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := trade.AdminDeliver(sc, path.OrderNo); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminShipMethodList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.ShipMethodListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := manage.AdminShipMethodList(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminShipMethodUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.ShipMethodUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err == nil && path.ID != "" {
			req.ID = path.ID
		}
		if err := manage.AdminShipMethodUpsert(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminShipMethodStatus(sc *svc.ServiceContext) http.HandlerFunc {
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
		var req types.BannerStatusReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.AdminShipMethodStatus(sc, id, req.Status); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminShipMethodDelete(sc *svc.ServiceContext) http.HandlerFunc {
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
		if err := manage.AdminShipMethodDelete(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// ─────────────────────────── 会员 ───────────────────────────

func AdminMemberList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.MemberListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := manage.MemberList(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminMemberStatus(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.MemberStatusReq
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.UpdateMemberStatus(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// ─────────────────────────── AI 知识库 ───────────────────────────

func AdminKnowledgeList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.KnowledgeListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := manage.KnowledgeList(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminKnowledgeUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.KnowledgeUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.UpsertKnowledge(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminKnowledgeDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := manage.DeleteKnowledge(sc, path.ID); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

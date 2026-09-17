// engagement_handlers.go 合规注销/协议/门店/签到日历/积分商城/晒单广场/百科/日历大屏。
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/aftersale"
	"pet/backend/internal/logic/auth"
	"pet/backend/internal/logic/community"
	"pet/backend/internal/logic/growth"
	"pet/backend/internal/logic/manage"
	"pet/backend/internal/logic/marketing"
	"pet/backend/internal/logic/pet"
	"pet/backend/internal/logic/trade"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// optionalMemberID 公开接口中的可选登录态（用于点赞态等个性化字段）
func optionalMemberID(sc *svc.ServiceContext, r *http.Request) int64 {
	uid, _ := parseAs(r, sc.Config.Auth.MemberAccessSecret, common.TokenTypeMember)
	return uid
}

// timeParseRFC3339 解析时间参数
func timeParseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// ─────────────────────────── 合规：注销 / 协议 ───────────────────────────

func Agreement(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		common.OK(w, auth.AgreementCurrent())
	}
}

func AccountDeleteStatus(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := auth.DeleteStatus(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AccountDeleteRequest(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.AccountDeleteReq
		_ = common.ParseBody(r, &req)
		if err := auth.DeleteRequest(sc, memberID(r)); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AccountDeleteCancel(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		if err := auth.DeleteCancel(sc, memberID(r)); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// ─────────────────────────── 门店 / 轨迹 ───────────────────────────

func StoreList(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := trade.StoreList(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	}
}

func AdminStoreList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		list, err := manage.AdminStoreList(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	})
}

func AdminStoreUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.StoreUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.AdminStoreUpsert(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminStoreStatus(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var body types.BannerStatusReq
		if err := common.ParseBody(r, &body); err != nil {
			common.Err(w, err)
			return
		}
		id, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.AdminStoreStatus(sc, id, body.Status); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminStoreDelete(sc *svc.ServiceContext) http.HandlerFunc {
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
		if err := manage.AdminStoreDelete(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminTraceAdd(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.OrderTraceAddReq
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var body types.OrderTraceAddReq
		if err := common.ParseBody(r, &body); err != nil {
			common.Err(w, err)
			return
		}
		var at *time.Time
		if body.HappenedAt != "" {
			t, err := timeParseRFC3339(body.HappenedAt)
			if err != nil {
				common.Err(w, common.ErrParam)
				return
			}
			at = &t
		}
		if err := trade.AddTrace(sc, req.OrderNo, at, body.StatusDesc, body.Detail); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// ─────────────────────────── 签到日历 / 补签 ───────────────────────────

func SignCalendar(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Month string `form:"month,optional"`
		}
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := growth.SignCalendar(sc, memberID(r), req.Month)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func SignMakeup(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.SignMakeupReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := growth.SignMakeup(sc, memberID(r), req.Date); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// ─────────────────────────── 积分商城 ───────────────────────────

func PointsShop(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := marketing.PointsShop(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

func PointsExchange(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.PointsExchangeReq
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		id, err := parseID64(req.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := marketing.PointsExchange(sc, memberID(r), id, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func MyPointsOrders(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := marketing.MyPointsOrders(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminPointsProductUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.PointsProductUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := marketing.PointsProductUpsert(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminPointsOrders(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.PageReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := marketing.AdminPointsOrders(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminPointsOrderShip(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.PointsOrderShipReq
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		id, err := parseID64(req.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := marketing.PointsOrderShip(sc, id, req.ShipNo); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// ─────────────────────────── 晒单广场 ───────────────────────────

func PostList(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PostListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var uid int64
		uid = optionalMemberID(sc, r)
		resp, err := community.List(sc, uid, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

func PostCreate(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.PostCreateReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := community.Create(sc, memberID(r), &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func PostLike(sc *svc.ServiceContext) http.HandlerFunc {
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
		liked, err := community.LikeToggle(sc, memberID(r), id)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, map[string]bool{"liked": liked})
	})
}

func PostDelete(sc *svc.ServiceContext) http.HandlerFunc {
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
		if err := community.Delete(sc, memberID(r), id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func MyPosts(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		list, err := community.MyList(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, map[string]any{"list": list})
	})
}

func AdminPostList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Status int `form:"status,optional"`
		}
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		list, err := community.AdminList(sc, req.Status)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	})
}

func AdminPostStatus(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.PostStatusReq
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		id, err := parseID64(req.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := community.AdminSetStatus(sc, id, req.Status); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminPostDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
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
		if err := community.AdminDelete(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// ─────────────────────────── 百科 / 日历 / 大屏 ───────────────────────────

func EncyclopediaBreeds(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := pet.EncyclopediaBreeds(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	}
}

func EncyclopediaArticles(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			BreedID string `path:"breedId"`
		}
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		id, err := strconv.ParseInt(req.BreedID, 10, 64)
		if err != nil || id <= 0 {
			common.Err(w, common.ErrParam)
			return
		}
		list, err := pet.EncyclopediaArticles(sc, id)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	}
}

func AdminActivityCalendar(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Days int `form:"days,optional"`
		}
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := manage.ActivityCalendar(sc, req.Days)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminRealtime(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := manage.RealtimeDashboard(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminPointsProductList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := marketing.AdminPointsProducts(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AdminExchangeConfirm(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.AfterSaleNoPathReq
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var body struct {
			ExchangeShipNo string `json:"exchangeShipNo"`
		}
		if err := common.ParseBody(r, &body); err != nil {
			common.Err(w, err)
			return
		}
		if err := aftersale.ConfirmExchange(sc, req.AfterSaleNo, body.ExchangeShipNo); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

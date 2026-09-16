// discovery_handlers.go 搜索增强 / 浏览历史 / 相关推荐 / 会员等级 / 电子健康证书（用户端）。
package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/growth"
	"pet/backend/internal/logic/pet"
	"pet/backend/internal/logic/trade"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// SearchHot 热搜榜（公开）
func SearchHot(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := pet.SearchHot(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

// SearchComplete 搜索联想（公开）
func SearchComplete(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchCompleteReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := pet.SearchComplete(sc, req.Prefix)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

// SearchTrace 记录搜索词（登录后写个人历史，游客只计热搜）
func SearchTrace(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchTraceReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := pet.TraceSearch(sc, memberID(r), req.Keyword); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// SearchHistory 我的搜索历史
func SearchHistory(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := pet.SearchHistory(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// SearchHistoryClear 清空搜索历史
func SearchHistoryClear(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		if err := pet.ClearSearchHistory(sc, memberID(r)); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// ViewHistory 我的浏览历史
func ViewHistory(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := pet.ViewHistory(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// RelatedProducts 相关推荐（详情页底部，公开）
func RelatedProducts(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := pet.RelatedProducts(sc, path.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

// MyLevel 我的会员等级
func MyLevel(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := growth.LevelView(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// OrderCertificate 电子健康证书
func OrderCertificate(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.OrderNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := trade.Certificate(sc, memberID(r), path.OrderNo)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

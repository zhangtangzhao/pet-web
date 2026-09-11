package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/ai"
	"pet/backend/internal/logic/auth"
	"pet/backend/internal/logic/pet"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// optionalMember 有 token 则解析 uid，游客返回 0
func optionalMember(sc *svc.ServiceContext, r *http.Request) int64 {
	uid, ok := parseAs(r, sc.Config.Auth.MemberAccessSecret, common.TokenTypeMember)
	if !ok {
		return 0
	}
	return uid
}

// ─────────────────────────── 认证 ───────────────────────────

func WechatMiniLogin(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WechatMiniLoginReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		resp, err := auth.WechatMiniLogin(sc, req.Code)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

func WechatH5OAuthURL(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WechatH5OAuthUrlReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		url, err := auth.H5OAuthURL(sc, req.Redirect)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, &types.OAuthUrlResp{URL: url})
	}
}

func WechatH5Login(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WechatH5LoginReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		resp, err := auth.WechatH5Login(sc, req.Code)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

func SmsSend(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SmsSendReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		interval, err := auth.SendSmsCode(sc, req.Phone)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, &types.SmsSendResp{Interval: interval})
	}
}

func SmsLogin(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SmsLoginReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		resp, err := auth.SmsLogin(sc, req.Phone, req.Code)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

func RefreshToken(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshTokenReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		resp, err := auth.Refresh(sc, req.RefreshToken)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

func Logout(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshTokenReq
		_ = common.ParseBody(r, &req)
		auth.Logout(sc, memberID(r), req.RefreshToken)
		common.OK(w, nil)
	})
}

// ─────────────────────────── 会员 ───────────────────────────

func Profile(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := auth.Profile(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func UpdateProfile(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateProfileReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := auth.UpdateProfile(sc, memberID(r), &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// ─────────────────────────── 宠物商品 ───────────────────────────

func Home(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := pet.Home(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

func Categories(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := pet.Categories(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	}
}

func Breeds(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			CategoryID string `form:"categoryId,optional"`
		}
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		list, err := pet.Breeds(sc, req.CategoryID)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	}
}

func ProductList(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ProductListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := pet.ListProducts(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

func ProductDetail(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := pet.ProductDetail(sc, optionalMember(sc, r), path.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	}
}

// ─────────────────────────── 收藏 ───────────────────────────

func Favorite(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := pet.Favorite(sc, memberID(r), path.ID); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func Unfavorite(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := pet.Unfavorite(sc, memberID(r), path.ID); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func FavoriteList(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.FavoriteListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := pet.FavoriteList(sc, memberID(r), &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// ─────────────────────────── AI 客服 ───────────────────────────

func AIAsk(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.AIAskReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		resp, err := ai.PetAIAsk(sc, memberID(r), &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

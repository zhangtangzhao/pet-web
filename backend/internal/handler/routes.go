package handler

import (
	"github.com/zeromicro/go-zero/rest"

	"pet/backend/internal/svc"
)

// RegisterHandlers 注册全部路由
func RegisterHandlers(server *rest.Server, sc *svc.ServiceContext) {
	// ───────── 用户端 · 无需登录 ─────────
	server.AddRoutes([]rest.Route{
		{Method: "POST", Path: "/api/auth/wechat/mini-login", Handler: WechatMiniLogin(sc)},
		{Method: "GET", Path: "/api/auth/wechat/h5-url", Handler: WechatH5OAuthURL(sc)},
		{Method: "POST", Path: "/api/auth/wechat/h5-login", Handler: WechatH5Login(sc)},
		{Method: "POST", Path: "/api/auth/sms/send", Handler: SmsSend(sc)},
		{Method: "POST", Path: "/api/auth/sms/login", Handler: SmsLogin(sc)},
		{Method: "POST", Path: "/api/auth/refresh", Handler: RefreshToken(sc)},
		{Method: "GET", Path: "/api/home", Handler: Home(sc)},
		{Method: "GET", Path: "/api/categories", Handler: Categories(sc)},
		{Method: "GET", Path: "/api/breeds", Handler: Breeds(sc)},
		{Method: "GET", Path: "/api/products", Handler: ProductList(sc)},
		{Method: "GET", Path: "/api/products/:id", Handler: ProductDetail(sc)},
		{Method: "POST", Path: "/api/pay/notify/wx", Handler: WxPayNotify(sc)},
	})

	// ───────── 用户端 · 需登录 ─────────
	server.AddRoutes([]rest.Route{
		{Method: "POST", Path: "/api/auth/logout", Handler: Logout(sc)},
		{Method: "GET", Path: "/api/member/profile", Handler: Profile(sc)},
		{Method: "PUT", Path: "/api/member/profile", Handler: UpdateProfile(sc)},
		{Method: "POST", Path: "/api/favorites/:id", Handler: Favorite(sc)},
		{Method: "DELETE", Path: "/api/favorites/:id", Handler: Unfavorite(sc)},
		{Method: "GET", Path: "/api/favorites", Handler: FavoriteList(sc)},
		{Method: "POST", Path: "/api/orders", Handler: CreateOrder(sc)},
		{Method: "GET", Path: "/api/orders", Handler: OrderList(sc)},
		{Method: "GET", Path: "/api/orders/:orderNo", Handler: OrderDetail(sc)},
		{Method: "POST", Path: "/api/orders/:orderNo/cancel", Handler: CancelOrder(sc)},
		{Method: "POST", Path: "/api/orders/:orderNo/confirm", Handler: ConfirmOrder(sc)},
		{Method: "POST", Path: "/api/orders/:orderNo/prepay", Handler: Prepay(sc)},
		{Method: "GET", Path: "/api/payments/:paymentNo/status", Handler: PaymentStatus(sc)},
		{Method: "POST", Path: "/api/ai/ask", Handler: AIAsk(sc)},
	})

	// ───────── 平台端 · 登录/刷新（无需 token）─────────
	server.AddRoutes([]rest.Route{
		{Method: "GET", Path: "/api/admin/captcha", Handler: AdminCaptcha(sc)},
		{Method: "POST", Path: "/api/admin/login", Handler: AdminLogin(sc)},
		{Method: "POST", Path: "/api/admin/refresh", Handler: AdminRefresh(sc)},
	})

	// ───────── 平台端 · 需登录 ─────────
	server.AddRoutes([]rest.Route{
		{Method: "POST", Path: "/api/admin/logout", Handler: AdminLogout(sc)},
		{Method: "PUT", Path: "/api/admin/password", Handler: AdminChangePassword(sc)},
		{Method: "GET", Path: "/api/admin/overview", Handler: AdminOverview(sc)},

		{Method: "POST", Path: "/api/admin/categories", Handler: AdminCategoryUpsert(sc)},
		{Method: "PUT", Path: "/api/admin/categories/:id", Handler: AdminCategoryUpdate(sc)},
		{Method: "DELETE", Path: "/api/admin/categories/:id", Handler: AdminCategoryDelete(sc)},
		{Method: "POST", Path: "/api/admin/breeds", Handler: AdminBreedUpsert(sc)},
		{Method: "PUT", Path: "/api/admin/breeds/:id", Handler: AdminBreedUpdate(sc)},
		{Method: "DELETE", Path: "/api/admin/breeds/:id", Handler: AdminBreedDelete(sc)},

		{Method: "GET", Path: "/api/admin/products", Handler: AdminProductList(sc)},
		{Method: "GET", Path: "/api/admin/products/:id", Handler: AdminProductDetail(sc)},
		{Method: "POST", Path: "/api/admin/products", Handler: AdminProductCreate(sc)},
		{Method: "PUT", Path: "/api/admin/products/:id", Handler: AdminProductUpdate(sc)},
		{Method: "PUT", Path: "/api/admin/products/:id/status", Handler: AdminProductStatus(sc)},
		{Method: "POST", Path: "/api/admin/upload-token", Handler: AdminUploadToken(sc)},

		{Method: "GET", Path: "/api/admin/orders", Handler: AdminOrderList(sc)},
		{Method: "GET", Path: "/api/admin/orders/:orderNo", Handler: AdminOrderDetail(sc)},
		{Method: "POST", Path: "/api/admin/orders/:orderNo/refund", Handler: AdminRefund(sc)},

		{Method: "GET", Path: "/api/admin/members", Handler: AdminMemberList(sc)},
		{Method: "PUT", Path: "/api/admin/members/:id/status", Handler: AdminMemberStatus(sc)},

		{Method: "GET", Path: "/api/admin/ai/knowledge", Handler: AdminKnowledgeList(sc)},
		{Method: "POST", Path: "/api/admin/ai/knowledge", Handler: AdminKnowledgeUpsert(sc)},
		{Method: "PUT", Path: "/api/admin/ai/knowledge/:id", Handler: AdminKnowledgeUpsert(sc)},
		{Method: "DELETE", Path: "/api/admin/ai/knowledge/:id", Handler: AdminKnowledgeDelete(sc)},
	})
}

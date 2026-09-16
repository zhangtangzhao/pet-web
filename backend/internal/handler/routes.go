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
		{Method: "GET", Path: "/api/products/:id/reviews", Handler: ProductReviews(sc)},
		{Method: "GET", Path: "/api/products/:id/review-summary", Handler: ProductReviewSummary(sc)},
		{Method: "GET", Path: "/api/services", Handler: ServiceList(sc)},
		{Method: "GET", Path: "/api/flash-sales", Handler: FlashSales(sc)},
		{Method: "GET", Path: "/api/trade-config", Handler: TradeConfig(sc)},
		{Method: "POST", Path: "/api/pay/notify/wx", Handler: WxPayNotify(sc)},
	})

	// ───────── 用户端 · 需登录 ─────────
	server.AddRoutes([]rest.Route{
		{Method: "POST", Path: "/api/auth/logout", Handler: Logout(sc)},
		{Method: "GET", Path: "/api/member/profile", Handler: Profile(sc)},
		{Method: "PUT", Path: "/api/member/profile", Handler: UpdateProfile(sc)},
		{Method: "GET", Path: "/api/member/reviews", Handler: MemberReviews(sc)},
		{Method: "POST", Path: "/api/favorites/:id", Handler: Favorite(sc)},
		{Method: "DELETE", Path: "/api/favorites/:id", Handler: Unfavorite(sc)},
		{Method: "GET", Path: "/api/favorites", Handler: FavoriteList(sc)},
		{Method: "POST", Path: "/api/orders", Handler: CreateOrder(sc)},
		{Method: "GET", Path: "/api/orders", Handler: OrderList(sc)},
		{Method: "GET", Path: "/api/orders/:orderNo", Handler: OrderDetail(sc)},
		{Method: "POST", Path: "/api/orders/:orderNo/cancel", Handler: CancelOrder(sc)},
		{Method: "POST", Path: "/api/orders/:orderNo/confirm", Handler: ConfirmOrder(sc)},
		{Method: "POST", Path: "/api/orders/:orderNo/prepay", Handler: Prepay(sc)},
		{Method: "POST", Path: "/api/orders/:orderNo/review", Handler: OrderReview(sc)},
		{Method: "GET", Path: "/api/ship/methods", Handler: ShipMethods(sc)},
		{Method: "GET", Path: "/api/payments/:paymentNo/status", Handler: PaymentStatus(sc)},
		{Method: "POST", Path: "/api/aftersale", Handler: AfterSaleApply(sc)},
		{Method: "GET", Path: "/api/aftersale", Handler: AfterSaleByOrder(sc)},
		{Method: "GET", Path: "/api/aftersale/list", Handler: AfterSaleMyList(sc)},
		{Method: "POST", Path: "/api/aftersale/:afterSaleNo/cancel", Handler: AfterSaleCancel(sc)},
		{Method: "GET", Path: "/api/notify/tmpl", Handler: NotifyTmpl(sc)},
		{Method: "GET", Path: "/api/notify/list", Handler: NotifyList(sc)},
		{Method: "POST", Path: "/api/notify/read", Handler: NotifyRead(sc)},
		{Method: "POST", Path: "/api/notify/:id/read", Handler: NotifyReadOne(sc)},
		{Method: "POST", Path: "/api/ai/ask", Handler: AIAsk(sc)},
		{Method: "POST", Path: "/api/ai/recommend", Handler: AIRecommend(sc)},
		{Method: "GET", Path: "/api/coupons", Handler: MyCoupons(sc)},
		{Method: "GET", Path: "/api/coupons/center", Handler: CouponCenter(sc)},
		{Method: "GET", Path: "/api/coupons/usable", Handler: UsableCoupons(sc)},
		{Method: "POST", Path: "/api/coupons/usable", Handler: UsableCoupons(sc)},
		{Method: "POST", Path: "/api/coupons/:id/claim", Handler: ClaimCoupon(sc)},
		{Method: "GET", Path: "/api/ws/cs", Handler: CsWS(sc)},
		{Method: "POST", Path: "/api/cs/messages", Handler: CsMemberSend(sc)},
		{Method: "GET", Path: "/api/cs/messages", Handler: CsMemberMessages(sc)},
		{Method: "POST", Path: "/api/cs/read", Handler: CsMemberRead(sc)},
		{Method: "POST", Path: "/api/upload-token", Handler: MemberUploadToken(sc)},
		{Method: "GET", Path: "/api/addresses", Handler: AddressList(sc)},
		{Method: "POST", Path: "/api/addresses", Handler: AddressSave(sc)},
		{Method: "PUT", Path: "/api/addresses/:id/default", Handler: AddressSetDefault(sc)},
		{Method: "DELETE", Path: "/api/addresses/:id", Handler: AddressDelete(sc)},
		{Method: "POST", Path: "/api/points/signin", Handler: SignIn(sc)},
		{Method: "GET", Path: "/api/points", Handler: MyPoints(sc)},
		{Method: "GET", Path: "/api/invite", Handler: Invite(sc)},
		{Method: "POST", Path: "/api/orders/:orderNo/mock-pay", Handler: MockPay(sc)},
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
		{Method: "GET", Path: "/api/admin/report", Handler: AdminReport(sc)},

		{Method: "GET", Path: "/api/admin/flash-sales", Handler: AdminFlashSaleList(sc)},
		{Method: "POST", Path: "/api/admin/flash-sales", Handler: AdminFlashSaleUpsert(sc)},
		{Method: "PUT", Path: "/api/admin/flash-sales/:id", Handler: AdminFlashSaleUpsert(sc)},
		{Method: "PUT", Path: "/api/admin/flash-sales/:id/status", Handler: AdminFlashSaleStatus(sc)},
		{Method: "DELETE", Path: "/api/admin/flash-sales/:id", Handler: AdminFlashSaleDelete(sc)},

		{Method: "GET", Path: "/api/admin/audit-logs", Handler: AdminAuditList(sc)},
		{Method: "GET", Path: "/api/admin/sensitive-words", Handler: AdminSensitiveList(sc)},
		{Method: "POST", Path: "/api/admin/sensitive-words", Handler: AdminSensitiveSave(sc)},
		{Method: "PUT", Path: "/api/admin/sensitive-words/:id/status", Handler: AdminSensitiveStatus(sc)},
		{Method: "DELETE", Path: "/api/admin/sensitive-words/:id", Handler: AdminSensitiveDelete(sc)},

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
		{Method: "POST", Path: "/api/admin/orders/:orderNo/ship", Handler: AdminShip(sc)},
		{Method: "POST", Path: "/api/admin/orders/:orderNo/deliver", Handler: AdminDeliver(sc)},

		{Method: "GET", Path: "/api/admin/ship-methods", Handler: AdminShipMethodList(sc)},
		{Method: "POST", Path: "/api/admin/ship-methods", Handler: AdminShipMethodUpsert(sc)},
		{Method: "PUT", Path: "/api/admin/ship-methods/:id", Handler: AdminShipMethodUpsert(sc)},
		{Method: "PUT", Path: "/api/admin/ship-methods/:id/status", Handler: AdminShipMethodStatus(sc)},
		{Method: "DELETE", Path: "/api/admin/ship-methods/:id", Handler: AdminShipMethodDelete(sc)},

		{Method: "GET", Path: "/api/admin/members", Handler: AdminMemberList(sc)},
		{Method: "PUT", Path: "/api/admin/members/:id/status", Handler: AdminMemberStatus(sc)},

		{Method: "GET", Path: "/api/admin/ai/knowledge", Handler: AdminKnowledgeList(sc)},
		{Method: "POST", Path: "/api/admin/ai/knowledge", Handler: AdminKnowledgeUpsert(sc)},
		{Method: "PUT", Path: "/api/admin/ai/knowledge/:id", Handler: AdminKnowledgeUpsert(sc)},
		{Method: "DELETE", Path: "/api/admin/ai/knowledge/:id", Handler: AdminKnowledgeDelete(sc)},

		{Method: "GET", Path: "/api/admin/marketing/services", Handler: AdminServiceList(sc)},
		{Method: "POST", Path: "/api/admin/marketing/services", Handler: AdminServiceUpsert(sc)},
		{Method: "PUT", Path: "/api/admin/marketing/services/:id", Handler: AdminServiceUpsert(sc)},
		{Method: "DELETE", Path: "/api/admin/marketing/services/:id", Handler: AdminServiceDelete(sc)},

		{Method: "GET", Path: "/api/admin/marketing/coupons", Handler: AdminCouponList(sc)},
		{Method: "POST", Path: "/api/admin/marketing/coupons", Handler: AdminCouponUpsert(sc)},
		{Method: "PUT", Path: "/api/admin/marketing/coupons/:id", Handler: AdminCouponUpsert(sc)},
		{Method: "DELETE", Path: "/api/admin/marketing/coupons/:id", Handler: AdminCouponDelete(sc)},
		{Method: "POST", Path: "/api/admin/marketing/coupons/:id/issue", Handler: AdminCouponIssue(sc)},

		{Method: "GET", Path: "/api/admin/banners", Handler: AdminBannerList(sc)},
		{Method: "POST", Path: "/api/admin/banners", Handler: AdminBannerUpsert(sc)},
		{Method: "PUT", Path: "/api/admin/banners/:id", Handler: AdminBannerUpsert(sc)},
		{Method: "PUT", Path: "/api/admin/banners/:id/status", Handler: AdminBannerStatus(sc)},
		{Method: "DELETE", Path: "/api/admin/banners/:id", Handler: AdminBannerDelete(sc)},

		{Method: "GET", Path: "/api/admin/cs/sessions", Handler: AdminCsSessions(sc)},
		{Method: "GET", Path: "/api/admin/cs/sessions/:id/messages", Handler: AdminCsMessages(sc)},
		{Method: "POST", Path: "/api/admin/cs/sessions/:id/messages", Handler: AdminCsMessages(sc)},
		{Method: "POST", Path: "/api/admin/cs/sessions/:id/read", Handler: AdminCsRead(sc)},
		{Method: "POST", Path: "/api/admin/cs/sessions/:id/close", Handler: AdminCsClose(sc)},

		{Method: "GET", Path: "/api/admin/reviews", Handler: AdminReviews(sc)},
		{Method: "PUT", Path: "/api/admin/reviews/:id/status", Handler: AdminReviewStatus(sc)},
		{Method: "POST", Path: "/api/admin/reviews/:id/reply", Handler: AdminReviewReply(sc)},
		{Method: "DELETE", Path: "/api/admin/reviews/:id", Handler: AdminReviewDelete(sc)},
		{Method: "GET", Path: "/api/admin/aftersales", Handler: AdminAfterSales(sc)},
		{Method: "POST", Path: "/api/admin/aftersales/:afterSaleNo/audit", Handler: AdminAfterSaleAudit(sc)},
	})
}

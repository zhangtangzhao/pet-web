package common

type ApiError struct {
	HTTP int
	Code int
	Msg  string
}

func (e *ApiError) Error() string { return e.Msg }

func NewErr(httpStatus, code int, msg string) *ApiError {
	return &ApiError{HTTP: httpStatus, Code: code, Msg: msg}
}

var (
	ErrParam        = NewErr(400, 40000, "参数错误")
	ErrUnauthorized = NewErr(401, 40100, "未登录")
	ErrTokenExpired = NewErr(401, 40101, "登录已过期")
	ErrForbidden    = NewErr(403, 40300, "无权限")
	ErrNotFound     = NewErr(404, 40400, "资源不存在")

	ErrGoodsOffline  = NewErr(400, 41001, "商品已下架")
	ErrGoodsLocked   = NewErr(400, 41002, "该宠物已被下单，请看看其他宝贝")
	ErrOrderState    = NewErr(400, 41003, "订单状态不允许此操作")
	ErrAlreadyFav    = NewErr(400, 41004, "已收藏该商品")
	ErrSmsCode       = NewErr(400, 41101, "验证码错误或已过期")
	ErrSmsFrequency  = NewErr(400, 41102, "发送过于频繁，请稍后再试")
	ErrSmsDailyLimit = NewErr(400, 41103, "超过当日发送上限")
	ErrSmsTooManyTry = NewErr(400, 41104, "验证码错误次数过多，请重新获取")

	ErrAdminLogin     = NewErr(401, 40102, "账号或密码错误")
	ErrCaptcha        = NewErr(400, 41203, "图形验证码错误")
	ErrPasswordLength = NewErr(400, 41204, "新密码长度至少 8 位")
	ErrOldPassword    = NewErr(400, 41205, "原密码错误")
	ErrReferenced     = NewErr(400, 41206, "存在关联数据，无法删除")
	ErrGoodsState     = NewErr(400, 41207, "商品当前状态不允许此操作")

	ErrAIAskFrequency = NewErr(429, 41301, "提问太频繁，请稍后再试")
	ErrAIAskDaily     = NewErr(429, 41302, "今日提问次数已达上限")
	ErrAIRecFrequency = NewErr(429, 41303, "推荐太频繁，请稍后再试")
	ErrAIRecDaily     = NewErr(429, 41304, "今日推荐次数已达上限")
	ErrAIReferenced   = NewErr(400, 41209, "知识条目所属品种不存在")

	ErrCouponUnusable  = NewErr(400, 41401, "优惠券不可用")
	ErrCouponThreshold = NewErr(400, 41402, "未达到优惠券使用门槛")
	ErrCouponSoldOut   = NewErr(400, 41403, "优惠券已领完")
	ErrCouponPerLimit  = NewErr(400, 41404, "已达到该券的领取上限")
	ErrCouponNotInTime = NewErr(400, 41405, "不在优惠券可领取时段")
	ErrServiceInvalid  = NewErr(400, 41406, "包含不可用的增值服务")

	ErrCsSession = NewErr(404, 41501, "会话不存在")

	ErrReview      = NewErr(400, 41601, "该订单已评价过")
	ErrReviewOrder = NewErr(400, 41602, "订单当前不可评价")

	ErrAfterSale       = NewErr(404, 41701, "售后单不存在")
	ErrAfterSaleActive = NewErr(400, 41702, "该订单已有进行中的售后")
	ErrAfterSaleOrder  = NewErr(400, 41703, "订单当前不可申请售后")
	ErrAfterSaleState  = NewErr(400, 41704, "售后状态不允许此操作")
	ErrAfterSaleAudit  = NewErr(409, 41705, "审核状态已变化，请刷新后重试")

	ErrTooManyRequests = NewErr(429, 41801, "请求过于频繁，请稍后再试")
	ErrSensitive       = NewErr(400, 41802, "内容包含敏感词，无法提交")
	ErrPointsNotEnough = NewErr(400, 41803, "积分不足")
	ErrDepositCoupon   = NewErr(400, 41804, "定金模式暂不支持使用优惠券")
	ErrInviteCode      = NewErr(400, 41805, "邀请码无效")
	ErrCertState       = NewErr(400, 41806, "订单完成后方可查看健康证书")

	ErrSupplierReferenced = NewErr(400, 41901, "该供货商已关联商品，请先调整商品")

	ErrInternal  = NewErr(500, 50000, "系统异常")
	ErrWxAPI     = NewErr(502, 50001, "微信接口异常")
	ErrPayConfig = NewErr(500, 50002, "支付未配置")
	ErrSmsSend   = NewErr(500, 50003, "短信发送失败")
	ErrAIService = NewErr(500, 50005, "AI 服务暂时不可用，请稍后再试")
)

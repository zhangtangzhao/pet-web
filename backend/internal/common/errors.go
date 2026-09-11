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
	ErrAIReferenced   = NewErr(400, 41209, "知识条目所属品种不存在")

	ErrInternal  = NewErr(500, 50000, "系统异常")
	ErrWxAPI     = NewErr(502, 50001, "微信接口异常")
	ErrPayConfig = NewErr(500, 50002, "支付未配置")
	ErrSmsSend   = NewErr(500, 50003, "短信发送失败")
	ErrAIService = NewErr(500, 50005, "AI 服务暂时不可用，请稍后再试")
)

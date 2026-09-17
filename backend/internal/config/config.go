package config

import (
	"errors"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	// 用户端 / 平台端双 secret 隔离
	Auth struct {
		MemberAccessSecret string
		MemberAccessExpire int // 秒
		AdminAccessSecret  string
		AdminAccessExpire  int // 秒
	}
	RefreshTokenExpireDays int `json:",default=30"`

	Database struct {
		DataSource string
	}

	Redis struct {
		Addr string
		Pass string `json:",optional"`
		DB   int    `json:",default=0"`
	}

	Snowflake struct {
		Node int64 `json:",default=1"`
	}

	WeChat struct {
		MiniAppID  string `json:",optional"`
		MiniSecret string `json:",optional"`
		H5AppID    string `json:",optional"` // 公众号（H5 网页授权）
		H5Secret   string `json:",optional"`
	}

	WeChatPay struct {
		MchID          string `json:",optional"`
		MchAPIv3Key    string `json:",optional"`
		MchSerialNo    string `json:",optional"`
		PrivateKeyPath string `json:",optional"`
		NotifyURL      string `json:",optional"`
	}

	Sms struct {
		Provider            string `json:",optional"` // 预留：aliyun/tencent；空=不发真实短信
		TestCode            string `json:",optional"` // 测试公共验证码，仅 Mode != prod 时生效
		CodeExpireSeconds   int    `json:",default=300"`
		SendIntervalSeconds int    `json:",default=60"`
		DailyLimit          int    `json:",default=10"`
	}

	COS struct {
		Bucket    string `json:",optional"` // bucket-appid
		Region    string `json:",optional"`
		SecretID  string `json:",optional"`
		SecretKey string `json:",optional"`
		BaseURL   string `json:",optional"`
	}

	// AI 客服：OpenAI 兼容接口（DeepSeek/通义/智谱等均可），三者全非空即启用真实大模型
	AI struct {
		BaseURL                  string `json:",optional"` // 如 https://api.deepseek.com/v1
		ApiKey                   string `json:",optional"`
		Model                    string `json:",optional"` // 如 deepseek-chat
		TimeoutSeconds           int    `json:",default=60"`
		AskIntervalSeconds       int    `json:",default=5"`
		DailyLimit               int    `json:",default=20"`
		RecommendIntervalSeconds int    `json:",default=10"` // 智能选宠 LLM 推荐最小间隔；≤0 回落 AskIntervalSeconds
		RecommendDailyLimit      int    `json:",default=5"`  // 智能选宠 LLM 推荐每日限额；≤0 回落 DailyLimit
	}

	// 微信通知模板：小程序订阅消息 / 公众号模板消息；未配置时通知降级为仅落库留痕
	Notify struct {
		MiniTmplCsReply string `json:",optional"`
		MiniTmplOrder   string `json:",optional"`
		MiniTmplCoupon  string `json:",optional"`
		H5TmplCsReply   string `json:",optional"`
		H5TmplOrder     string `json:",optional"`
		H5TmplCoupon    string `json:",optional"`
	}

	Trade struct {
		AutoConfirmDays   int `json:",default=7"`  // 已支付订单 N 天后自动确认完成；≤0 关闭
		PayTimeoutMinutes int `json:",default=15"` // 待支付订单 N 分钟后自动关单；≤0 取 15
	}

	// 增长运营：积分 / 邀请 / 定金
	Growth struct {
		SignPoints         int `json:",default=5"`  // 每日签到积分
		ReviewPoints       int `json:",default=10"` // 首次评价积分
		OrderPointsPerYuan int `json:",default=1"`  // 每实付 1 元得积分（向下取整）
		InviteRewardPoints int `json:",default=50"` // 邀请成功双方各得积分
		DepositPercent     int `json:",default=10"` // 定金比例（%实付金额，向上保底 0.01）
		DepositHoldDays    int `json:",default=3"`  // 付定金后 N 天内需补尾款；≤0 取 3
		SignBonus3         int `json:",default=5"`  // 连续签到 3 天额外奖励
		SignBonus7         int `json:",default=20"` // 连续签到 7 天额外奖励
		SignMakeupCost     int `json:",default=20"` // 补签消耗积分
	}

	// 接口限流（Redis 计数器；≤0 关闭对应项）
	RateLimit struct {
		LoginPerMinute int `json:",default=10"` // 每 IP 每分钟登录/验证码尝试上限
		OrderPerMinute int `json:",default=5"`  // 每用户每分钟下单上限
		SmsIPDaily     int `json:",default=20"` // 每 IP 每日短信发送上限
	}

	// 营销自动化（≤0 关闭对应项）
	Automation struct {
		CartRemindHours  int   `json:",default=24"`   // 购物车加入 N 小时未结算 → 提醒
		DormantDays      int   `json:",default=30"`   // 距最近支付订单 N 天未下单 → 召回
		DormantCouponID  int64 `json:",default=5120"` // 召回券模板 ID
		BirthdayCouponID int64 `json:",default=5130"` // 生日礼包券模板 ID
	}

	// 账号注销（合规：冷静期后匿名化）
	Account struct {
		DeletionCooldownDays int `json:",default=7"`
	}
}

// IsProd 生产环境判定（go-zero Mode 取值 pro；兼容 prod）
func (c Config) IsProd() bool {
	return c.Mode == "pro" || c.Mode == "prod"
}

// AIEnabled 是否已接入真实大模型
func (c Config) AIEnabled() bool {
	return c.AI.BaseURL != "" && c.AI.ApiKey != "" && c.AI.Model != ""
}

// Validate 生产模式关键配置校验（配合 yaml ${VAR} 占位：环境变量缺失时不允许带病启动）
func (c Config) Validate() error {
	if !c.IsProd() {
		return nil
	}
	if c.Auth.MemberAccessSecret == "" || c.Auth.AdminAccessSecret == "" {
		return errors.New("生产环境必须配置 Auth.MemberAccessSecret / AdminAccessSecret（经环境变量注入）")
	}
	if c.Database.DataSource == "" {
		return errors.New("生产环境必须配置 Database.DataSource")
	}
	if c.Redis.Addr == "" {
		return errors.New("生产环境必须配置 Redis.Addr")
	}
	if c.Sms.TestCode != "" {
		return errors.New("生产环境 Sms.TestCode 必须留空")
	}
	return nil
}

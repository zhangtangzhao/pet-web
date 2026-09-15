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
		BaseURL                 string `json:",optional"` // 如 https://api.deepseek.com/v1
		ApiKey                  string `json:",optional"`
		Model                   string `json:",optional"` // 如 deepseek-chat
		TimeoutSeconds          int    `json:",default=60"`
		AskIntervalSeconds      int    `json:",default=5"`
		DailyLimit              int    `json:",default=20"`
		RecommendIntervalSeconds int   `json:",default=10"` // 智能选宠 LLM 推荐最小间隔；≤0 回落 AskIntervalSeconds
		RecommendDailyLimit      int   `json:",default=5"`  // 智能选宠 LLM 推荐每日限额；≤0 回落 DailyLimit
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

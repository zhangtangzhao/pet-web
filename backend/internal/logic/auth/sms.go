package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/marketing"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func normalizePhone(p string) (string, error) {
	p = strings.TrimSpace(p)
	if len(p) != 11 || !strings.HasPrefix(p, "1") {
		return "", errors.New("手机号格式错误")
	}
	for _, ch := range p {
		if ch < '0' || ch > '9' {
			return "", errors.New("手机号格式错误")
		}
	}
	return p, nil
}

func randCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}

// SendSmsCode 发送短信验证码（频控 + 测试公共码环境说明见 docs/02 §3.4）
func SendSmsCode(sc *svc.ServiceContext, rawPhone string) (int, error) {
	phone, err := normalizePhone(rawPhone)
	if err != nil {
		return 0, common.ErrParam
	}
	cfg := sc.Config.Sms
	ctx := context.Background()

	ok, err := sc.Rdb.SetNX(ctx, "sms:interval:"+phone, 1, time.Duration(cfg.SendIntervalSeconds)*time.Second).Result()
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, common.ErrSmsFrequency
	}

	dayKey := "sms:day:" + phone + ":" + time.Now().Format("20060102")
	cnt, err := sc.Rdb.Incr(ctx, dayKey).Result()
	if err != nil {
		return 0, err
	}
	if cnt == 1 {
		sc.Rdb.Expire(ctx, dayKey, 24*time.Hour)
	}
	if cnt > int64(cfg.DailyLimit) {
		return 0, common.ErrSmsDailyLimit
	}

	code := randCode()
	expire := time.Duration(cfg.CodeExpireSeconds) * time.Second
	sc.Rdb.Set(ctx, "sms:code:"+phone, code, expire)
	sc.Rdb.Del(ctx, "sms:try:"+phone)

	if cfg.Provider == "" {
		// 短信通道联调时接入 aliyun/tencent；开发环境直接看日志验证
		logx.Infof("[sms] provider 未配置，跳过真实发送 phone=%s code=%s", phone, code)
	} else {
		logx.Infof("[sms] provider=%s 发送 phone=%s（发送实现联调时接入）", cfg.Provider, phone)
	}
	return cfg.SendIntervalSeconds, nil
}

// verifySmsCode 校验验证码（成功即消费，防重放；测试公共码仅非生产生效）
func verifySmsCode(sc *svc.ServiceContext, phone, code string) error {
	cfg := sc.Config.Sms
	if !sc.Config.IsProd() && cfg.TestCode != "" && code == cfg.TestCode {
		return nil
	}
	ctx := context.Background()
	stored, err := sc.Rdb.Get(ctx, "sms:code:"+phone).Result()
	if err != nil {
		return common.ErrSmsCode
	}
	if stored != code {
		tries, _ := sc.Rdb.Incr(ctx, "sms:try:"+phone).Result()
		sc.Rdb.Expire(ctx, "sms:try:"+phone, time.Duration(cfg.CodeExpireSeconds)*time.Second)
		if tries >= 5 {
			sc.Rdb.Del(ctx, "sms:code:"+phone)
			return common.ErrSmsTooManyTry
		}
		return common.ErrSmsCode
	}
	sc.Rdb.Del(ctx, "sms:code:"+phone, "sms:try:"+phone)
	return nil
}

// SmsLogin 手机号验证码登录，未注册自动注册
func SmsLogin(sc *svc.ServiceContext, rawPhone, code string) (*types.LoginResp, error) {
	phone, err := normalizePhone(rawPhone)
	if err != nil {
		return nil, common.ErrParam
	}
	if code == "" {
		return nil, common.ErrSmsCode
	}
	if err := verifySmsCode(sc, phone, code); err != nil {
		return nil, err
	}
	return findOrCreateMemberByPhone(sc, phone)
}

func findOrCreateMemberByPhone(sc *svc.ServiceContext, phone string) (*types.LoginResp, error) {
	var m model.Member
	isNew := false
	err := sc.DB.Where("phone = ?", phone).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		m = model.Member{
			ID:       common.NewID(),
			Nickname: "用户" + phone[7:],
			Phone:    phone,
			Status:   1,
		}
		if err := sc.DB.Create(&m).Error; err != nil {
			// 并发下唯一索引冲突则回查
			if e := sc.DB.Where("phone = ?", phone).First(&m).Error; e != nil {
				return nil, err
			}
		} else {
			isNew = true
		}
	} else if err != nil {
		return nil, err
	}
	if m.Status != 1 {
		return nil, common.ErrForbidden
	}
	if isNew {
		go marketing.GrantNewUserCoupons(sc, m.ID)
	}
	touchLastLogin(sc.DB, m.ID)
	return IssueTokens(sc, &m, isNew)
}

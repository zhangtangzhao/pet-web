// Package notify 微信通知投递：biz_key 幂等入队 + 后台投递器（订阅消息/模板消息）。
// 模板未配置时降级为仅落库留痕（仿短信降级），开发环境零配置可跑。
package notify

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
)

const (
	maxRetry   = 3  // 失败重试上限（超过置终态失败）
	thingLimit = 20 // 模板 thing 字段上限（微信规定 20 字符）
	titleLimit = 64
)

// Enqueue 幂等入队：biz_key 冲突时忽略（重复事件只投一次）
func Enqueue(sc *svc.ServiceContext, memberID int64, scene int, bizKey, title, content, orderNo string) error {
	n := model.Notification{
		ID:       common.NewID(),
		MemberID: memberID,
		Scene:    scene,
		BizKey:   bizKey,
		Title:    truncate(title, titleLimit),
		Content:  truncate(content, thingLimit),
		OrderNo:  orderNo,
		Status:   model.NotifyPending,
	}
	if err := sc.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&n).Error; err != nil {
		return err
	}
	return nil
}

// StartNotifier 启动投递器（每 10s 扫描待投递行）
func StartNotifier(ctx context.Context, sc *svc.ServiceContext) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deliverBatch(ctx, sc)
		}
	}
}

// deliverBatch 单行事务逐条取队首（FOR UPDATE SKIP LOCKED，多实例不重复投递），直到取空
func deliverBatch(ctx context.Context, sc *svc.ServiceContext) {
	for {
		found, err := deliverOne(ctx, sc)
		if err != nil {
			logx.Errorf("notify: 投递异常: %v", err)
			return
		}
		if !found {
			return
		}
	}
}

func deliverOne(ctx context.Context, sc *svc.ServiceContext) (bool, error) {
	found := false
	err := sc.DB.Transaction(func(tx *gorm.DB) error {
		var ids []int64
		// 先 Pluck 主键（空队列不产生 ErrRecordNotFound），锁行防止多实例重复投递
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Model(&model.Notification{}).
			Where("status = ? AND retry < ?", model.NotifyPending, maxRetry).
			Order("id").Limit(1).Pluck("id", &ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil // found=false，批次结束
		}
		var n model.Notification
		if err := tx.First(&n, ids[0]).Error; err != nil {
			return err
		}
		found = true
		status, channel := deliver(ctx, sc, &n)
		updates := map[string]any{
			"status":     status,
			"channel":    channel,
			"updated_at": time.Now(),
		}
		if status == model.NotifyPending { // 本轮失败，等待下轮重试
			updates["retry"] = n.Retry + 1
			if n.Retry+1 >= maxRetry { // 重试耗尽 → 终态失败
				updates["status"] = model.NotifyFailed
			}
		}
		return tx.Model(&model.Notification{}).Where("id = ?", n.ID).Updates(updates).Error
	})
	return found, err
}

// deliver 单条投递：按场景选模板，小程序订阅消息优先、公众号模板消息兜底。
// 返回终态（已投递/失败/降级）或 NotifyPending（本轮失败待重试）。
func deliver(ctx context.Context, sc *svc.ServiceContext, n *model.Notification) (int, int) {
	cfg := sc.Config.Notify
	miniTmpl, h5Tmpl := cfg.MiniTmplOrder, cfg.H5TmplOrder
	if n.Scene == model.NotifySceneCsReply {
		miniTmpl, h5Tmpl = cfg.MiniTmplCsReply, cfg.H5TmplCsReply
	} else if n.Scene == model.NotifySceneCoupon {
		miniTmpl, h5Tmpl = cfg.MiniTmplCoupon, cfg.H5TmplCoupon
	}
	if miniTmpl == "" && h5Tmpl == "" {
		logx.Infof("notify: 模板未配置，降级落库 bizKey=%s scene=%d", n.BizKey, n.Scene)
		return model.NotifyDegrade, model.NotifyChannelNone
	}

	if miniTmpl != "" {
		if openID := authOpenID(sc, n.MemberID, model.WxAppMini); openID != "" {
			err := sendMiniSubscribe(ctx, sc, miniTmpl, openID, n)
			if err == nil {
				return model.NotifyDelivered, model.NotifyChannelMini
			}
			var we *wxAPIError
			if errors.As(err, &we) && we.ErrCode == 43101 {
				logx.Infof("notify: 用户未订阅小程序消息 bizKey=%s", n.BizKey) // 43101 不重试
				return model.NotifyFailed, model.NotifyChannelMini
			}
			logx.Errorf("notify: 小程序订阅消息发送失败 bizKey=%s: %v", n.BizKey, err)
			// 小程序渠道失败不直接终态，继续尝试公众号
		}
	}
	if h5Tmpl != "" {
		if openID := authOpenID(sc, n.MemberID, model.WxAppH5); openID != "" {
			err := sendH5Template(ctx, sc, h5Tmpl, openID, n)
			if err == nil {
				return model.NotifyDelivered, model.NotifyChannelH5
			}
			logx.Errorf("notify: 公众号模板消息发送失败 bizKey=%s: %v", n.BizKey, err)
		}
	}
	return model.NotifyPending, model.NotifyChannelNone
}

// authOpenID 取会员在指定端的微信 openid（未授权返回空）
func authOpenID(sc *svc.ServiceContext, memberID, appType int64) string {
	var auth model.WechatAuth
	if err := sc.DB.Where("member_id = ? AND app_type = ?", memberID, appType).First(&auth).Error; err != nil {
		return ""
	}
	return auth.OpenID
}

func truncate(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	r := []rune(s)
	return string(r[:limit])
}

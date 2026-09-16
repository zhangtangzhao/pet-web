package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/growth"
	"pet/backend/internal/logic/trade"
	"pet/backend/internal/metrics"
	"pet/backend/internal/ratelimit"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func CreateOrder(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateOrderReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		// 下单频控（防刷单/脚本抢单）
		if !ratelimit.Allow(sc.Rdb, "rl:order:"+strconv.FormatInt(memberID(r), 10),
			sc.Config.RateLimit.OrderPerMinute, time.Minute) {
			metrics.LimitRejected.WithLabelValues("order").Inc()
			common.Err(w, common.ErrTooManyRequests)
			return
		}
		resp, err := trade.CreateOrder(sc, memberID(r), &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// TradeConfig 下单相关公开配置（定金比例/尾款期限），供 checkout 预估展示。
// 携带用户 token 时附带会员等级折扣信息（游客仅返回公共配置）。
func TradeConfig(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		holdDays := sc.Config.Growth.DepositHoldDays
		if holdDays <= 0 {
			holdDays = 3
		}
		resp := &types.TradeConfigResp{
			DepositPercent:  sc.Config.Growth.DepositPercent,
			DepositHoldDays: holdDays,
		}
		if uid, ok := parseAs(r, sc.Config.Auth.MemberAccessSecret, common.TokenTypeMember); ok {
			if lv, err := growth.LevelView(sc, uid); err == nil {
				resp.MyLevelName = lv.LevelName
				resp.MyDiscount = lv.Discount
				resp.MyGrowthValue = lv.GrowthValue
			}
		}
		common.OK(w, resp)
	}
}

func ShipMethods(sc *svc.ServiceContext) http.HandlerFunc {	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		list, err := trade.ListMethods(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	})
}

func OrderList(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.OrderListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := trade.OrderList(sc, memberID(r), &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func OrderDetail(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.OrderNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := trade.OrderDetail(sc, memberID(r), path.OrderNo)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func CancelOrder(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.OrderNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var req types.CancelOrderReq
		_ = common.ParseBody(r, &req)
		if err := trade.CancelOrder(sc, memberID(r), path.OrderNo, req.Reason); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func ConfirmOrder(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.OrderNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := trade.ConfirmOrder(sc, memberID(r), path.OrderNo); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func Prepay(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.OrderNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := trade.Prepay(sc, memberID(r), path.OrderNo)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func PaymentStatus(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.PaymentStatusReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := trade.PaymentStatus(sc, memberID(r), path.PaymentNo)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// WxPayNotify 微信支付回调（验签解密 → 幂等落账）
// 应答规范：成功 {"code":"SUCCESS"}；失败 {"code":"FAIL","message":"..."} 触发微信重试
func WxPayNotify(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		write := func(code, msg string) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": msg})
		}
		parser, err := sc.NotifyParser()
		if err != nil {
			write("FAIL", "支付回调未配置")
			return
		}
		transaction := new(payments.Transaction)
		if _, err := parser.ParseNotifyRequest(context.Background(), r, transaction); err != nil {
			logx.Errorf("支付回调验签失败: %v", err)
			write("FAIL", "验签失败")
			return
		}
		if err := trade.HandleWxPayNotify(sc, transaction); err != nil {
			logx.Errorf("支付回调落账失败: %v", err)
			write("FAIL", "落账失败，等待重试")
			return
		}
		write("SUCCESS", "成功")
	}
}

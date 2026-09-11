package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/trade"
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
		resp, err := trade.CreateOrder(sc, memberID(r), &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
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

package svc

import (
	"context"
	"fmt"
	"sync"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"

	"pet/backend/internal/common"
)

// PayClient 懒加载微信支付 V3 客户端（配置为空时返回 ErrPayConfig）
func (sc *ServiceContext) PayClient(ctx context.Context) (*core.Client, error) {
	sc.payMtx.Lock()
	defer sc.payMtx.Unlock()
	if sc.payReady {
		return sc.payClient, sc.payErr
	}
	sc.payReady = true
	c := sc.Config.WeChatPay
	if c.MchID == "" || c.MchSerialNo == "" || c.MchAPIv3Key == "" || c.PrivateKeyPath == "" {
		sc.payErr = common.ErrPayConfig
		return nil, sc.payErr
	}
	priv, err := utils.LoadPrivateKeyWithPath(c.PrivateKeyPath)
	if err != nil {
		sc.payErr = fmt.Errorf("加载商户私钥失败: %w", err)
		return nil, sc.payErr
	}
	client, err := core.NewClient(ctx,
		option.WithWechatPayAutoAuthCipher(c.MchID, c.MchSerialNo, priv, c.MchAPIv3Key))
	if err != nil {
		sc.payErr = fmt.Errorf("初始化微信支付客户端失败: %w", err)
		return nil, sc.payErr
	}
	sc.payClient = client
	return client, nil
}

// Prepay JSAPI 下单并生成小程序拉起支付所需参数
func (sc *ServiceContext) Prepay(ctx context.Context, outTradeNo, description string, amountFen int64, openID string) (*jsapi.PrepayWithRequestPaymentResponse, error) {
	client, err := sc.PayClient(ctx)
	if err != nil {
		return nil, err
	}
	c := sc.Config.WeChatPay
	resp, _, err := (&jsapi.JsapiApiService{Client: client}).PrepayWithRequestPayment(ctx, jsapi.PrepayRequest{
		Appid:       core.String(sc.Config.WeChat.MiniAppID),
		Mchid:       core.String(c.MchID),
		Description: core.String(description),
		OutTradeNo:  core.String(outTradeNo),
		NotifyUrl:   core.String(c.NotifyURL),
		Amount:      &jsapi.Amount{Total: core.Int64(amountFen)},
		Payer:       &jsapi.Payer{Openid: core.String(openID)},
	})
	if err != nil {
		return nil, fmt.Errorf("微信下单失败: %w", err)
	}
	return resp, nil
}

// QueryOrder 主动查单（对账兜底）
func (sc *ServiceContext) QueryOrder(ctx context.Context, outTradeNo string) (*payments.Transaction, error) {
	client, err := sc.PayClient(ctx)
	if err != nil {
		return nil, err
	}
	txn, _, err := (&jsapi.JsapiApiService{Client: client}).QueryOrderByOutTradeNo(ctx, jsapi.QueryOrderByOutTradeNoRequest{
		Mchid:      core.String(sc.Config.WeChatPay.MchID),
		OutTradeNo: core.String(outTradeNo),
	})
	if err != nil {
		return nil, err
	}
	return txn, nil
}

// Refund 发起退款（全额或指定金额）
func (sc *ServiceContext) Refund(ctx context.Context, outTradeNo, outRefundNo, reason string, refundFen, totalFen int64) (*refunddomestic.Refund, error) {
	client, err := sc.PayClient(ctx)
	if err != nil {
		return nil, err
	}
	refund, _, err := (&refunddomestic.RefundsApiService{Client: client}).Create(ctx, refunddomestic.CreateRequest{
		OutTradeNo:  core.String(outTradeNo),
		OutRefundNo: core.String(outRefundNo),
		Reason:      core.String(reason),
		Amount: &refunddomestic.AmountReq{
			Refund:   core.Int64(refundFen),
			Total:    core.Int64(totalFen),
			Currency: core.String("CNY"),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("微信退款失败: %w", err)
	}
	return refund, nil
}

var (
	notifyOnce    sync.Once
	notifyHandler *notify.Handler
	notifyErr     error
)

// NotifyParser 支付回调验签+解密器
func (sc *ServiceContext) NotifyParser() (*notify.Handler, error) {
	notifyOnce.Do(func() {
		c := sc.Config.WeChatPay
		if c.MchID == "" || c.MchAPIv3Key == "" {
			notifyErr = common.ErrPayConfig
			return
		}
		certVisitor := downloader.MgrInstance().GetCertificateVisitor(c.MchID)
		notifyHandler = notify.NewNotifyHandler(c.MchAPIv3Key, verifiers.NewSHA256WithRSAVerifier(certVisitor))
	})
	return notifyHandler, notifyErr
}

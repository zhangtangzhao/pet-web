package trade

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/growth"
	"pet/backend/internal/logic/notify"
	"pet/backend/internal/metrics"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func decimal100() decimal.Decimal {
	return decimal.NewFromInt(100)
}

// prepayOrder 对待支付订单发起 JSAPI 预支付，返回小程序拉起支付参数
func prepayOrder(sc *svc.ServiceContext, memberID int64, o *model.Order) (*types.WxPayParams, error) {
	if sc.Config.WeChatPay.MchID == "" {
		return nil, common.ErrPayConfig
	}
	// 小程序 JSAPI 需要 openid：取会员绑定的小程序授权
	var auth model.WechatAuth
	if err := sc.DB.Where("member_id = ? AND app_type = ?", memberID, model.WxAppMini).
		First(&auth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewErr(400, 41201, "请先使用微信登录后再支付")
		}
		return nil, err
	}
	desc := "宠物订单-" + o.OrderNo
	fen := o.PayAmount.Mul(decimal100()).IntPart()
	resp, err := sc.Prepay(context.Background(), o.OrderNo, desc, fen, auth.OpenID)
	if err != nil {
		return nil, err
	}
	return &types.WxPayParams{
		TimeStamp: *resp.TimeStamp,
		NonceStr:  *resp.NonceStr,
		Package:   *resp.Package,
		SignType:  *resp.SignType,
		PaySign:   *resp.PaySign,
	}, nil
}

// Prepay 待支付单重新拉起支付
func Prepay(sc *svc.ServiceContext, memberID int64, orderNo string) (*types.CreateOrderResp, error) {
	var o model.Order
	if err := sc.DB.Where("order_no = ? AND member_id = ?", orderNo, memberID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	if o.Status == model.OrderDepositPaid {
		return PayTail(sc, memberID, orderNo)
	}
	if o.Status != model.OrderPending {
		return nil, common.ErrOrderState
	}
	if time.Now().After(o.ExpireAt) {
		return nil, common.ErrOrderState
	}
	params, err := prepayOrder(sc, memberID, &o)
	if err != nil {
		return nil, err
	}
	return &types.CreateOrderResp{
		OrderNo:   o.OrderNo,
		PayAmount: o.PayAmount.StringFixed(2),
		ExpireAt:  o.ExpireAt.Format(time.RFC3339),
		PayParams: params,
	}, nil
}

// PaymentStatus 查询支付结果（轮询接口，附带主动查单兜底回调丢失）
func PaymentStatus(sc *svc.ServiceContext, memberID int64, paymentNo string) (*types.PaymentStatusResp, error) {
	var p model.Payment
	if err := sc.DB.Where("payment_no = ? AND member_id = ?", paymentNo, memberID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	if p.Status == model.PayStatusPending && p.PayType == model.PayTypePurchase {
		if txn, err := sc.QueryOrder(context.Background(), p.OrderNo); err == nil &&
			txn.TradeState != nil && *txn.TradeState == "SUCCESS" {
			_ = markPaid(sc, &p, txn)
			p.Status = model.PayStatusSuccess
		}
	}
	text := "待支付"
	switch p.Status {
	case model.PayStatusSuccess:
		text = "支付成功"
	case model.PayStatusFail:
		text = "支付失败"
	case model.PayStatusRefund:
		text = "已退款"
	}
	return &types.PaymentStatusResp{Status: p.Status, PayStatus: text}, nil
}

// HandleWxPayNotify 微信支付回调：已验签解密 → 金额比对 → 幂等落账
func HandleWxPayNotify(sc *svc.ServiceContext, txn *payments.Transaction) error {
	if txn.TradeState == nil || *txn.TradeState != "SUCCESS" || txn.OutTradeNo == nil {
		return nil // 非成功态忽略，应答 SUCCESS 防止反复通知
	}
	var p model.Payment
	if err := sc.DB.Where("payment_no = ?", *txn.OutTradeNo).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // 非本系统单号
		}
		return err
	}
	if txn.Amount != nil && txn.Amount.Total != nil {
		expect := p.Amount.Mul(decimal100()).IntPart()
		if int64(*txn.Amount.Total) != expect {
			return common.NewErr(500, 50005, "回调金额不一致")
		}
	}
	return markPaid(sc, &p, txn)
}

// markPaid 幂等落账：payment 成功 + 订单状态按分支推进
// 全款（10→20）/ 定金（10→15，商品保持锁定）/ 尾款（15→20，商品售出）
func markPaid(sc *svc.ServiceContext, p *model.Payment, txn *payments.Transaction) error {
	now := time.Now()
	txnID := ""
	if txn.TransactionId != nil {
		txnID = *txn.TransactionId
	}
	err := sc.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Payment{}).
			Where("payment_no = ? AND status = ?", p.PaymentNo, model.PayStatusPending).
			Updates(map[string]any{
				"status":         model.PayStatusSuccess,
				"callback_at":    &now,
				"transaction_id": txnID,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil // 已处理（幂等）
		}
		var o model.Order
		if err := tx.Where("order_no = ?", p.OrderNo).First(&o).Error; err != nil {
			return err
		}
		switch {
		case p.PayType == model.PayTypeTail:
			// 尾款到账 → 待补尾款 15 → 已支付 20（paid_at 保留定金支付时间）
			if err := tx.Model(&model.Order{}).
				Where("order_no = ? AND status = ?", p.OrderNo, model.OrderDepositPaid).
				Updates(map[string]any{"status": model.OrderPaid, "updated_at": now}).
				Error; err != nil {
				return err
			}
		case o.DepositAmount.GreaterThan(decimal.Zero):
			// 定金到账 → 待支付 10 → 待补尾款 15，商品保持锁定，尾款截止 = now + N 天
			holdDays := sc.Config.Growth.DepositHoldDays
			if holdDays <= 0 {
				holdDays = 3
			}
			tailDeadline := now.AddDate(0, 0, holdDays)
			if err := tx.Model(&model.Order{}).
				Where("order_no = ? AND status = ?", p.OrderNo, model.OrderPending).
				Updates(map[string]any{
					"status":         model.OrderDepositPaid,
					"paid_at":        &now,
					"tail_expire_at": &tailDeadline,
					"updated_at":     now,
				}).Error; err != nil {
				return err
			}
			return nil // 定金单禁用券，商品未售出
		default:
			if err := tx.Model(&model.Order{}).
				Where("order_no = ? AND status = ?", p.OrderNo, model.OrderPending).
				Updates(map[string]any{"status": model.OrderPaid, "paid_at": &now, "updated_at": now}).
				Error; err != nil {
				return err
			}
		}
		// 核销锁定的优惠券
		if err := tx.Exec(
			`UPDATE member_coupon SET status = ?, used_at = ? WHERE order_id = ? AND status = ?`,
			model.CouponUsed, now, p.OrderID, model.CouponLocked).Error; err != nil {
			return err
		}
		// 商品锁定 → 已售出，累计销量
		if err := tx.Exec(`
			UPDATE pet_product SET status = ?, sales = sales + 1, updated_at = now()
			WHERE id IN (SELECT product_id FROM order_item WHERE order_id =
				(SELECT id FROM orders WHERE order_no = ?)) AND status = ?`,
			model.ProductSold, p.OrderNo, model.ProductLocked).Error; err != nil {
			return err
		}
		// 支付完成返积分（随本事务仅一次；定金阶段在上分支已返回不返）
		growth.RewardOrderTx(tx, sc, o.MemberID, o.PayAmount, o.OrderNo)
		return nil
	})
	if err == nil {
		metrics.OrdersPaid.Inc()
	}
	return err
}

// PayTail 补尾款：复用/创建尾款支付流水并预支付（定金单 status=15 专用）
func PayTail(sc *svc.ServiceContext, memberID int64, orderNo string) (*types.CreateOrderResp, error) {
	var o model.Order
	if err := sc.DB.Where("order_no = ? AND member_id = ?", orderNo, memberID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	if o.Status != model.OrderDepositPaid {
		return nil, common.ErrOrderState
	}
	if o.TailExpireAt != nil && time.Now().After(*o.TailExpireAt) {
		return nil, common.ErrOrderState
	}
	tailAmount := o.PayAmount.Sub(o.DepositAmount)
	if tailAmount.LessThanOrEqual(decimal.Zero) {
		return nil, common.ErrOrderState
	}
	var pay model.Payment
	err := sc.DB.Where("order_no = ? AND pay_type = ?", orderNo, model.PayTypeTail).
		Order("id DESC").First(&pay).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		pay = model.Payment{
			ID:        common.NewID(),
			PaymentNo: common.NewBizNo("PAY"),
			OrderID:   o.ID,
			OrderNo:   o.OrderNo,
			MemberID:  memberID,
			Amount:    tailAmount,
			Channel:   model.PayChannelMini,
			PayType:   model.PayTypeTail,
			Status:    model.PayStatusPending,
		}
		if err := sc.DB.Create(&pay).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if pay.Status != model.PayStatusPending {
		return nil, common.ErrOrderState
	}
	params, err := prepayOrder(sc, memberID, &o)
	if err != nil && sc.Config.WeChatPay.MchID != "" {
		return nil, err
	}
	expire := ""
	if o.TailExpireAt != nil {
		expire = o.TailExpireAt.Format(time.RFC3339)
	}
	return &types.CreateOrderResp{
		OrderNo:    o.OrderNo,
		PaymentNo:  pay.PaymentNo,
		PayAmount:  tailAmount.StringFixed(2),
		ExpireAt:   expire,
		PayParams:  params,
		IsDeposit:  true,
		TailAmount: tailAmount.StringFixed(2),
	}, nil
}

// MockPay 开发演示支付：未配置商户号且非生产时直接落账（生产返回 404）
func MockPay(sc *svc.ServiceContext, memberID int64, orderNo string) error {
	if sc.Config.WeChatPay.MchID != "" || sc.Config.IsProd() {
		return common.ErrNotFound
	}
	var o model.Order
	if err := sc.DB.Where("order_no = ? AND member_id = ?", orderNo, memberID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNotFound
		}
		return err
	}
	if o.Status != model.OrderPending && o.Status != model.OrderDepositPaid {
		return common.ErrOrderState
	}
	var pay model.Payment
	err := sc.DB.Where("order_no = ? AND status = ?", orderNo, model.PayStatusPending).
		Order("id DESC").First(&pay).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return common.NewErr(400, 41202, "未找到待支付流水")
	}
	if err != nil {
		return err
	}
	return markPaid(sc, &pay, &payments.Transaction{})
}

// RefundOrder 发起退款（平台端整单退款 / 售后审核通过共用）。
// outRefundNo 由调用方提供（微信按 out_refund_no 幂等，售后路径传 "RF"+售后单号保证重试安全）；
// amount 为 nil 表示全额。简化：退款受理成功即落账（订单→已退款、退款流水→已退款）。
// 生产强化项：接入退款结果回调，按微信最终状态落账。
func RefundOrder(sc *svc.ServiceContext, orderNo, reason string, amount *decimal.Decimal, outRefundNo string) error {
	var o model.Order
	if err := sc.DB.Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNotFound
		}
		return err
	}
	if o.Status != model.OrderPaid && o.Status != model.OrderCompleted && o.Status != model.OrderDepositPaid {
		return common.ErrOrderState
	}
	var pay model.Payment
	if err := sc.DB.Where("order_no = ? AND pay_type = ? AND status = ?",
		orderNo, model.PayTypePurchase, model.PayStatusSuccess).First(&pay).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewErr(400, 41202, "未找到成功的支付流水")
		}
		return err
	}
	// 幂等短路：已有成功退款流水则视为已退款（重复审核/重复回调安全）
	var exist model.Payment
	if err := sc.DB.Where("order_no = ? AND pay_type = ? AND status = ?",
		orderNo, model.PayTypeRefund, model.PayStatusRefund).First(&exist).Error; err == nil {
		return nil
	}
	totalFen := pay.Amount.Mul(decimal100()).IntPart()
	refundAmount := pay.Amount
	if amount != nil {
		refundAmount = *amount
	}
	refundFen := refundAmount.Mul(decimal100()).IntPart()
	if refundFen <= 0 || refundFen > totalFen {
		return common.ErrParam
	}

	// 开发演示：未配置商户号且非生产 → 直接落 DEMO 退款（不调微信）
	demo := sc.Config.WeChatPay.MchID == "" && !sc.Config.IsProd()
	var refundID string
	if !demo {
		refund, err := sc.Refund(context.Background(), orderNo, outRefundNo, reason, refundFen, totalFen)
		if err != nil {
			return err
		}
		if refund.RefundId != nil {
			refundID = *refund.RefundId
		}
	} else {
		refundID = "DEMO"
	}
	now := time.Now()
	err := sc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.Payment{
			ID:            common.NewID(),
			PaymentNo:     outRefundNo,
			OrderID:       o.ID,
			OrderNo:       orderNo,
			MemberID:      o.MemberID,
			Amount:        refundAmount,
			Channel:       pay.Channel,
			PayType:       model.PayTypeRefund,
			TransactionID: refundID,
			Status:        model.PayStatusRefund,
			CallbackAt:    &now,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Order{}).
			Where("id = ? AND status IN ?", o.ID, []int{model.OrderPaid, model.OrderCompleted, model.OrderDepositPaid}).
			Updates(map[string]any{"status": model.OrderRefunded, "cancel_reason": reason, "updated_at": now}).Error
	})
	if err != nil {
		return err
	}
	// 定金单退款：商品仍在锁定态，回在售并回补秒杀名额
	if o.Status == model.OrderDepositPaid {
		releaseOrderStock(sc.DB, o.ID, o.FlashSaleID)
	}
	// 退款到账通知（biz_key 幂等：售后与整单退款同单只投一次）
	if err := notify.Enqueue(sc, o.MemberID, model.NotifySceneOrder,
		"refund:"+orderNo, "退款已到账", "订单退款已原路退回", orderNo); err != nil {
		logx.Errorf("退款通知入队失败 orderNo=%s: %v", orderNo, err)
	}
	return nil
}

// vip.go 付费会员卡：年卡购买（dev 直落成功）+ 每月权益发放。
package auth

import (
	"time"

	"github.com/shopspring/decimal"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// VipBuy 购买年卡：创建支付流水；dev（未配置商户号且非生产）直接落成功并延长会员
func VipBuy(sc *svc.ServiceContext, memberID int64) (*types.VipBuyResp, error) {
	price := sc.Config.Vip.PriceYuan
	if price <= 0 {
		price = 99
	}
	amount := decimal.NewFromInt(int64(price))
	pay := model.Payment{
		ID: common.NewID(), PaymentNo: common.NewBizNo("PAY"),
		OrderNo: "VIP" + intStr(memberID), MemberID: memberID,
		Amount: amount, Channel: model.PayChannelMini,
		PayType: model.PayTypeVip, Status: model.PayStatusPending,
	}
	if err := sc.DB.Create(&pay).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	if err := sc.DB.Model(&model.Payment{}).Where("payment_no = ?", pay.PaymentNo).
		Updates(map[string]any{"status": model.PayStatusSuccess, "callback_at": &now}).Error; err != nil {
		return nil, err
	}
	if err := extendVip(sc, memberID); err != nil {
		return nil, err
	}
	return &types.VipBuyResp{PaymentNo: pay.PaymentNo, Price: amount.StringFixed(2)}, nil
}

// extendVip 到期续延一年
func extendVip(sc *svc.ServiceContext, memberID int64) error {
	return sc.DB.Exec(`
		UPDATE member SET vip_expire_at = CASE
			WHEN vip_expire_at IS NOT NULL AND vip_expire_at > now()
			THEN vip_expire_at + INTERVAL '1 year'
			ELSE now() + INTERVAL '1 year' END,
		updated_at = now() WHERE id = ?`, memberID).Error
}

func intStr(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	var b []byte
	if neg {
		v = -v
	}
	for v > 0 {
		b = append([]byte{byte('0' + v%10)}, b...)
		v /= 10
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}

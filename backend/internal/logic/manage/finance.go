// finance.go 财务对账：支付流水流水账（payment 表直查）+ 按日收支汇总（支付/退款/净额）。
package manage

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func financePaymentView(m *model.Payment) types.FinancePaymentView {
	v := types.FinancePaymentView{
		ID:            strconvI64(m.ID),
		PaymentNo:     m.PaymentNo,
		OrderNo:       m.OrderNo,
		Amount:        m.Amount.StringFixed(2),
		Channel:       m.Channel,
		PayType:       m.PayType,
		Status:        m.Status,
		TransactionID: m.TransactionID,
		CreatedAt:     m.CreatedAt.Format(time.RFC3339),
	}
	if m.CallbackAt != nil {
		v.CallbackAt = m.CallbackAt.Format(time.RFC3339)
	}
	return v
}

// AdminFinancePayments 支付流水列表（含退款/失败），orderNo 模糊、status/payType 精确
func AdminFinancePayments(sc *svc.ServiceContext, req *types.FinancePaymentListReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	q := sc.DB.Model(&model.Payment{})
	if no := strings.TrimSpace(req.OrderNo); no != "" {
		q = q.Where("order_no LIKE ?", "%"+no+"%")
	}
	if req.Status >= 0 {
		q = q.Where("status = ?", req.Status)
	}
	if req.PayType >= 0 {
		q = q.Where("pay_type = ?", req.PayType)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.Payment
	if err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.FinancePaymentView, 0, len(rows))
	for i := range rows {
		list = append(list, financePaymentView(&rows[i]))
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// AdminFinanceDaily 按日汇总：pay = 支付成功（全款/定金/尾款），refund = 退款成功
func AdminFinanceDaily(sc *svc.ServiceContext, days int) (*types.FinanceDailyResp, error) {
	if days < 1 || days > 90 {
		days = 30
	}
	start := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")

	type row struct {
		Date      string
		PayCount  int64
		PayAmount decimal.Decimal
		RefCnt    int64
		RefAmount decimal.Decimal
	}
	var rows []row
	if err := sc.DB.Model(&model.Payment{}).
		Select(`to_char(created_at, 'YYYY-MM-DD') AS date,
			COUNT(*) FILTER (WHERE status = 1 AND pay_type IN (1,3)) AS pay_count,
			COALESCE(SUM(amount) FILTER (WHERE status = 1 AND pay_type IN (1,3)), 0) AS pay_amount,
			COUNT(*) FILTER (WHERE status = 3 AND pay_type = 2) AS ref_cnt,
			COALESCE(SUM(amount) FILTER (WHERE status = 3 AND pay_type = 2), 0) AS ref_amount`).
		Where("created_at >= ?::date", start).
		Group("date").Order("date DESC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.FinanceDailyRow, 0, len(rows))
	for _, r := range rows {
		ref := r.RefAmount
		list = append(list, types.FinanceDailyRow{
			Date:         r.Date,
			PayCount:     r.PayCount,
			PayAmount:    r.PayAmount.StringFixed(2),
			RefundCount:  r.RefCnt,
			RefundAmount: ref.StringFixed(2),
			NetAmount:    r.PayAmount.Sub(ref).StringFixed(2),
		})
	}
	return &types.FinanceDailyResp{List: list}, nil
}

// invoice.go 电子发票申请流（真实开票需税控对接，管理端回传链接）。
package trade

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// InvoiceApply 申请开票（本人已支付订单）
func InvoiceApply(sc *svc.ServiceContext, memberID int64, req *types.InvoiceApplyReq) error {
	title := strings.TrimSpace(req.Title)
	if title == "" || len([]rune(title)) > 128 {
		return common.ErrInvoiceInvalid
	}
	if req.TitleType != 1 && req.TitleType != 2 {
		req.TitleType = 1
	}
	var o model.Order
	if err := sc.DB.Where("order_no = ? AND member_id = ?", req.OrderNo, memberID).
		First(&o).Error; err != nil {
		return common.ErrInvoiceInvalid
	}
	if o.Status != model.OrderPaid && o.Status != model.OrderCompleted {
		return common.ErrInvoiceInvalid
	}
	inv := model.Invoice{
		ID: common.NewID(), MemberID: memberID, OrderNo: o.OrderNo,
		TitleType: req.TitleType, Title: title, TaxNo: strings.TrimSpace(req.TaxNo),
		Amount: o.PayAmount, Status: model.InvoicePending,
	}
	if err := sc.DB.Create(&inv).Error; err != nil {
		return common.ErrInvoiceInvalid // 唯一键 = 一单一张
	}
	return nil
}

// MyInvoices 我的发票
func MyInvoices(sc *svc.ServiceContext, memberID int64) ([]types.InvoiceView, error) {
	var rows []model.Invoice
	if err := sc.DB.Where("member_id = ?", memberID).Order("id DESC").Limit(50).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]types.InvoiceView, 0, len(rows))
	for _, r := range rows {
		out = append(out, invoiceView(r))
	}
	return out, nil
}

func invoiceView(r model.Invoice) types.InvoiceView {
	return types.InvoiceView{
		OrderNo: r.OrderNo, TitleType: r.TitleType, Title: r.Title, TaxNo: r.TaxNo,
		Amount: r.Amount.StringFixed(2), Status: r.Status, Link: r.Link,
		CreatedAt: r.CreatedAt.Format("2006-01-02 15:04"),
	}
}

// AdminInvoices 平台端发票列表
func AdminInvoices(sc *svc.ServiceContext, status int) ([]types.InvoiceView, error) {
	q := sc.DB.Model(&model.Invoice{})
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	var rows []model.Invoice
	if err := q.Order("id DESC").Limit(100).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]types.InvoiceView, 0, len(rows))
	for _, r := range rows {
		out = append(out, invoiceView(r))
	}
	return out, nil
}

// AdminInvoiceIssue 标记已开票（可回传下载链接）
func AdminInvoiceIssue(sc *svc.ServiceContext, id int64, link string) error {
	res := sc.DB.Model(&model.Invoice{}).Where("id = ? AND status = ?", id, model.InvoicePending).
		Updates(map[string]any{"status": model.InvoiceIssued, "link": strings.TrimSpace(link), "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrInvoiceInvalid
	}
	return nil
}

var _ = decimal.Zero

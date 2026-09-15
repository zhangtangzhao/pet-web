// Package aftersale 用户侧售后申请：已支付/已完成订单可发起退款售后（默认全额），
// 管理端审核（金额可调）；同单同时仅一个进行中售后（部分唯一索引 + 23505 兜底）。
// 审核采用 CAS 序列（HTTP 请求不在事务内调微信）：CAS 1→2 → 退款 → 失败则 CAS 2→1 回滚。
package aftersale

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/notify"
	"pet/backend/internal/logic/trade"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

const (
	maxReasonRunes = 500
	// pg unique_violation：并发重复申请撞 uk_aftersale_active
	pgUniqueViolation = "23505"
)

// ─────────────────────────── 会员侧 ───────────────────────────

// Apply 发起售后：订单归属 + 已支付/已完成 + 无进行中售后；金额默认全额
func Apply(sc *svc.ServiceContext, memberID int64, req *types.AfterSaleApplyReq) (*types.AfterSaleView, error) {
	reason := trimSpace(req.Reason)
	if reason == "" || utf8.RuneCountInString(reason) > maxReasonRunes {
		return nil, common.ErrParam
	}
	var o model.Order
	if err := sc.DB.Where("order_no = ? AND member_id = ?", req.OrderNo, memberID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	if o.Status != model.OrderPaid && o.Status != model.OrderCompleted {
		return nil, common.ErrAfterSaleOrder
	}
	var cnt int64
	if err := sc.DB.Model(&model.AfterSale{}).
		Where("order_no = ? AND status = ?", o.OrderNo, model.AfterSalePending).Count(&cnt).Error; err != nil {
		return nil, err
	}
	if cnt > 0 {
		return nil, common.ErrAfterSaleActive
	}

	as := model.AfterSale{
		ID:           common.NewID(),
		AfterSaleNo:  common.NewBizNo("AS"),
		OrderNo:      o.OrderNo,
		MemberID:     memberID,
		Reason:       reason,
		RefundAmount: o.PayAmount, // 默认全额，审核时可调
		Status:       model.AfterSalePending,
	}
	if err := sc.DB.Create(&as).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return nil, common.ErrAfterSaleActive // 并发兜底
		}
		return nil, err
	}
	return view(as, "", ""), nil
}

// Cancel 会员撤销售后（仅本人待审核单）
func Cancel(sc *svc.ServiceContext, memberID int64, afterSaleNo string) error {
	var row model.AfterSale
	if err := sc.DB.Where("after_sale_no = ?", afterSaleNo).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrAfterSale // 不存在
		}
		return err
	}
	if row.MemberID != memberID {
		return common.ErrAfterSale // 非本人
	}
	res := sc.DB.Model(&model.AfterSale{}).
		Where("after_sale_no = ? AND status = ?", afterSaleNo, model.AfterSalePending).
		Updates(map[string]any{"status": model.AfterSaleCancel, "audit_at": time.Now(), "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrAfterSaleState // 状态不允许
	}
	return nil
}

// MyList 我的售后单列表（分页）
func MyList(sc *svc.ServiceContext, memberID int64, pageReq types.PageReq) (*types.PageResp, error) {
	page, size := pageReq.Page, pageReq.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}
	q := sc.DB.Model(&model.AfterSale{}).Where("member_id = ?", memberID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.AfterSale
	if err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.AfterSaleView, 0, len(rows))
	for _, r := range rows {
		list = append(list, *view(r, "", ""))
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// GetByOrder 查询订单的售后单（本人，取最新一条）
func GetByOrder(sc *svc.ServiceContext, memberID int64, orderNo string) (*types.AfterSaleView, error) {
	var row model.AfterSale
	if err := sc.DB.Where("order_no = ? AND member_id = ?", orderNo, memberID).
		Order("id DESC").First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrAfterSale
		}
		return nil, err
	}
	return view(row, "", ""), nil
}

// ─────────────────────────── 平台端 ───────────────────────────

// AdminList 售后列表（含会员昵称/手机号）
func AdminList(sc *svc.ServiceContext, status int, pageReq types.PageReq) (*types.PageResp, error) {
	page, size := pageReq.Page, pageReq.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	q := sc.DB.Model(&model.AfterSale{})
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.AfterSale
	if err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, err
	}
	memberNames := memberTexts(sc, rows)
	list := make([]types.AfterSaleView, 0, len(rows))
	for _, r := range rows {
		list = append(list, *view(r, memberNames[r.MemberID][0], memberNames[r.MemberID][1]))
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// Audit 审核：同意（金额可调 → 退款）或拒绝（备注必填）。
// CAS 序列：①售后单 1→2 → ②退款（demo 或微信）→ ③退款失败 CAS 2→1 回滚并报错。
func Audit(sc *svc.ServiceContext, afterSaleNo string, agree bool, amountStr, note string) error {
	var as model.AfterSale
	if err := sc.DB.Where("after_sale_no = ?", afterSaleNo).First(&as).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrAfterSale
		}
		return err
	}
	if as.Status != model.AfterSalePending {
		return common.ErrAfterSaleState
	}

	now := time.Now()
	if !agree {
		note = trimSpace(note)
		if note == "" {
			return common.NewErr(400, 40001, "拒绝时必须填写备注")
		}
		res := sc.DB.Model(&model.AfterSale{}).
			Where("id = ? AND status = ?", as.ID, model.AfterSalePending).
			Updates(map[string]any{
				"status": model.AfterSaleRefused, "admin_note": note,
				"audit_at": now, "updated_at": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return common.ErrAfterSaleAudit
		}
		// 售后结果通知（同意路径由 RefundOrder 落 refund 事件，避免重复打扰）
		if err := notify.Enqueue(sc, as.MemberID, model.NotifySceneOrder,
			"aftersale:"+as.AfterSaleNo, "售后已拒绝", "您的售后申请未通过审核", as.OrderNo); err != nil {
			logx.Errorf("售后拒绝通知入队失败 %s: %v", as.AfterSaleNo, err)
		}
		return nil
	}

	// 同意：确定退款金额（默认申请金额，可调；上限为实付）
	refundAmount := as.RefundAmount
	if trimSpace(amountStr) != "" {
		d, err := decimal.NewFromString(amountStr)
		if err != nil {
			return common.ErrParam
		}
		refundAmount = d
	}
	var pay model.Payment
	if err := sc.DB.Where("order_no = ? AND pay_type = ? AND status = ?",
		as.OrderNo, model.PayTypePurchase, model.PayStatusSuccess).First(&pay).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewErr(400, 41202, "未找到成功的支付流水")
		}
		return err
	}
	if refundAmount.LessThanOrEqual(decimal.Zero) || refundAmount.GreaterThan(pay.Amount) {
		return common.ErrParam
	}

	// ① CAS 待审核 → 已同意（携带最终退款金额）
	res := sc.DB.Model(&model.AfterSale{}).
		Where("id = ? AND status = ?", as.ID, model.AfterSalePending).
		Updates(map[string]any{
			"status": model.AfterSaleAgreed, "refund_amount": refundAmount, "admin_note": note,
			"audit_at": now, "updated_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrAfterSaleAudit
	}

	// ② 退款（确定性退款单号 RF+售后单号，微信按其幂等）
	outRefundNo := "RF" + as.AfterSaleNo
	if err := trade.RefundOrder(sc, as.OrderNo, "售后退款 "+as.AfterSaleNo, &refundAmount, outRefundNo); err != nil {
		// ③ 回滚审核状态，管理员可修正后重试（outRefundNo 确定性保证微信侧安全）
		if e := sc.DB.Model(&model.AfterSale{}).
			Where("id = ? AND status = ?", as.ID, model.AfterSaleAgreed).
			Updates(map[string]any{
				"status": model.AfterSalePending, "refund_amount": as.RefundAmount,
				"updated_at": time.Now(),
			}).Error; e != nil {
			logx.Errorf("售后审核回滚失败 %s: %v", as.AfterSaleNo, e)
		}
		return err
	}
	_ = sc.DB.Model(&model.AfterSale{}).Where("id = ?", as.ID).
		Update("refund_payment_no", outRefundNo).Error
	return nil
}

// ─────────────────────────── 内部 ───────────────────────────

func view(r model.AfterSale, nickname, phone string) *types.AfterSaleView {
	v := &types.AfterSaleView{
		ID:              strconv.FormatInt(r.ID, 10),
		AfterSaleNo:     r.AfterSaleNo,
		OrderNo:         r.OrderNo,
		MemberID:        strconv.FormatInt(r.MemberID, 10),
		Nickname:        nickname,
		Phone:           phone,
		Reason:          r.Reason,
		RefundAmount:    r.RefundAmount.StringFixed(2),
		Status:          r.Status,
		StatusText:      model.AfterSaleStatusText(r.Status),
		AdminNote:       r.AdminNote,
		RefundPaymentNo: r.RefundPaymentNo,
		CreatedAt:       r.CreatedAt.Format(time.RFC3339),
	}
	if r.AuditAt != nil {
		v.AuditAt = r.AuditAt.Format(time.RFC3339)
	}
	return v
}

// memberTexts 批量取会员昵称/手机号：memberID → [nickname, phone]
func memberTexts(sc *svc.ServiceContext, rows []model.AfterSale) map[int64][2]string {
	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.MemberID)
	}
	out := map[int64][2]string{}
	if len(ids) == 0 {
		return out
	}
	var members []model.Member
	if err := sc.DB.Where("id IN ?", ids).Find(&members).Error; err != nil {
		return out
	}
	for _, m := range members {
		out[m.ID] = [2]string{m.Nickname, m.Phone}
	}
	return out
}

func trimSpace(s string) string { return strings.TrimSpace(s) }

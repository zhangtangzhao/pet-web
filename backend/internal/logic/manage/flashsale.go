// flashsale.go 秒杀活动管理：每商品同时至多一个启用中（部分唯一索引兜底）。
package manage

import (
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

const pgUniqueViolation = "23505"

type flashRow struct {
	model.FlashSale
	ProductTitle string
}

func flashView(r flashRow) types.FlashSaleView {
	return types.FlashSaleView{
		ID:           strconvI64(r.ID),
		ProductID:    strconvI64(r.ProductID),
		ProductTitle: r.ProductTitle,
		SalePrice:    r.SalePrice.StringFixed(2),
		Stock:        r.Stock,
		Sold:         r.Sold,
		StartAt:      r.StartAt.Format(time.RFC3339),
		EndAt:        r.EndAt.Format(time.RFC3339),
		Status:       r.Status,
		CreatedAt:    r.CreatedAt.Format(time.RFC3339),
	}
}

// AdminFlashSaleList 秒杀活动列表（含未启用，id DESC 分页，带商品标题）
func AdminFlashSaleList(sc *svc.ServiceContext, req *types.PageReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	q := sc.DB.Model(&model.FlashSale{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []flashRow
	if err := q.Select("flash_sale.*, p.title AS product_title").
		Joins("LEFT JOIN pet_product p ON p.id = flash_sale.product_id").
		Order("flash_sale.id DESC").Offset((page - 1) * size).Limit(size).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.FlashSaleView, 0, len(rows))
	for i := range rows {
		list = append(list, flashView(rows[i]))
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// AdminFlashSaleUpsert 新建 / 编辑；启用中活动每商品唯一（并发冲突报 23505 → 明确错误）
func AdminFlashSaleUpsert(sc *svc.ServiceContext, req *types.FlashSaleUpsertReq) error {
	productID, err := strconvI64Err(req.ProductID)
	if err != nil {
		return err
	}
	var p model.PetProduct
	if err := sc.DB.First(&p, productID).Error; err != nil {
		return common.ErrNotFound
	}
	salePrice, err := decimal.NewFromString(strings.TrimSpace(req.SalePrice))
	if err != nil || salePrice.LessThanOrEqual(decimal.Zero) {
		return common.ErrParam
	}
	if req.Stock < 1 || req.Stock > 9999 {
		return common.ErrParam
	}
	startAt, err := parseTimePtr(req.StartAt)
	if err != nil || startAt == nil {
		return common.ErrParam
	}
	endAt, err := parseTimePtr(req.EndAt)
	if err != nil || endAt == nil {
		return common.ErrParam
	}
	if !endAt.After(*startAt) {
		return common.NewErr(400, 40001, "结束时间必须晚于开始时间")
	}
	if salePrice.GreaterThan(p.Price) {
		return common.NewErr(400, 40001, "秒杀价不能高于商品原价")
	}
	status := req.Status
	if status == 0 {
		status = model.FlashSaleOn
	}
	if status != model.FlashSaleOff && status != model.FlashSaleOn {
		return common.ErrParam
	}

	save := func(id int64) error {
		err := sc.DB.Transaction(func(tx *gorm.DB) error {
			fs := model.FlashSale{
				ID: id, ProductID: productID, SalePrice: salePrice,
				Stock: req.Stock, StartAt: *startAt, EndAt: *endAt, Status: status,
			}
			if id == 0 {
				fs.ID = common.NewID()
				return tx.Create(&fs).Error
			}
			res := tx.Model(&model.FlashSale{}).Where("id = ?", id).Updates(map[string]any{
				"product_id": productID, "sale_price": salePrice, "stock": req.Stock,
				"start_at": startAt, "end_at": endAt, "status": status, "updated_at": time.Now(),
			})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return common.ErrNotFound
			}
			return nil
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
				return common.NewErr(400, 40002, "该商品已有启用中的秒杀活动")
			}
		}
		return err
	}
	if req.ID == "" {
		return save(0)
	}
	id, err := strconvI64Err(req.ID)
	if err != nil {
		return err
	}
	return save(id)
}

// AdminFlashSaleStatus 启用 / 停用（启用撞唯一索引时给明确错误）
func AdminFlashSaleStatus(sc *svc.ServiceContext, id int64, status int) error {
	if status != model.FlashSaleOff && status != model.FlashSaleOn {
		return common.ErrParam
	}
	err := sc.DB.Model(&model.FlashSale{}).Where("id = ?", id).Update("status", status).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return common.NewErr(400, 40002, "该商品已有启用中的秒杀活动")
		}
		return err
	}
	return nil
}

// AdminFlashSaleDelete 删除活动（订单已快照秒杀价，名额回补按 ID 静默失效）
func AdminFlashSaleDelete(sc *svc.ServiceContext, id int64) error {
	res := sc.DB.Delete(&model.FlashSale{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func parseTimePtr(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return &t, nil
		}
	}
	return nil, common.ErrParam
}

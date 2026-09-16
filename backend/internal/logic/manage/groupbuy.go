// groupbuy.go 拼团活动管理：CRUD + 启停，活动价/成团人数/成团时限。
package manage

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

type groupBuyRow struct {
	model.GroupBuy
	ProductTitle string `gorm:"column:product_title"`
}

func groupBuyView(r groupBuyRow) types.GroupBuyView {
	return types.GroupBuyView{
		ID:           strconvI64(r.ID),
		ProductID:    strconvI64(r.ProductID),
		ProductTitle: r.ProductTitle,
		Price:        r.Price.StringFixed(2),
		Size:         r.Size,
		Hours:        r.Hours,
		Status:       r.Status,
		CreatedAt:    r.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// AdminGroupBuyList 拼团活动列表（分页 + 商品标题）
func AdminGroupBuyList(sc *svc.ServiceContext, req *types.PageReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	q := sc.DB.Model(&model.GroupBuy{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []groupBuyRow
	if err := q.Select("group_buy.*, p.title AS product_title").
		Joins("LEFT JOIN pet_product p ON p.id = group_buy.product_id").
		Order("group_buy.id DESC").Offset((page - 1) * size).Limit(size).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.GroupBuyView, 0, len(rows))
	for _, r := range rows {
		list = append(list, groupBuyView(r))
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// AdminGroupBuyUpsert 新建/编辑拼团活动
func AdminGroupBuyUpsert(sc *svc.ServiceContext, req *types.GroupBuyUpsertReq) error {
	productID, err := strconvI64Err(req.ProductID)
	if err != nil {
		return err
	}
	if !existsByID(sc.DB, &model.PetProduct{}, productID) {
		return common.NewErr(400, 41208, "商品不存在")
	}
	price, err := decimal.NewFromString(strings.TrimSpace(req.Price))
	if err != nil || price.Sign() <= 0 || price.GreaterThan(decimal.NewFromInt(99999999)) {
		return common.NewErr(400, 40001, "拼团价不合法")
	}
	size := req.Size
	if size < 2 || size > 10 {
		return common.NewErr(400, 40001, "成团人数应为 2-10 人")
	}
	hours := req.Hours
	if hours < 1 || hours > 72 {
		return common.NewErr(400, 40001, "成团时限应为 1-72 小时")
	}
	status := 1
	if req.Status != nil && *req.Status == 0 {
		status = 0
	}
	if req.ID == "" {
		return sc.DB.Create(&model.GroupBuy{
			ID:        common.NewID(),
			ProductID: productID,
			Price:     price,
			Size:      size,
			Hours:     hours,
			Status:    status,
		}).Error
	}
	id, err := strconvI64Err(req.ID)
	if err != nil {
		return err
	}
	res := sc.DB.Model(&model.GroupBuy{}).Where("id = ?", id).Updates(map[string]any{
		"product_id": productID, "price": price, "size": size,
		"hours": hours, "status": status, "updated_at": time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// AdminGroupBuyStatus 启用/停用
func AdminGroupBuyStatus(sc *svc.ServiceContext, id int64, status int) error {
	if status != 0 && status != 1 {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.GroupBuy{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// AdminGroupBuyDelete 删除活动（进行中的团不受影响，由 closer 收尾）
func AdminGroupBuyDelete(sc *svc.ServiceContext, id int64) error {
	res := sc.DB.Delete(&model.GroupBuy{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

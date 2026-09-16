// Package review 订单评价：仅已完成订单可评，一单一评；商品详情页展示，管理端可隐藏/删除。
package review

import (
	"encoding/json"
	"errors"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/growth"
	"pet/backend/internal/logic/risk"
	"pet/backend/internal/logic/sensitive"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

const (
	maxContentRunes = 500
	maxImages       = 9
	maxReplyRunes   = 200
	defaultLimit    = 10
	maxLimit        = 50
)

type reviewRow struct {
	model.OrderReview
	Nickname string
	Avatar   string
}

func view(r reviewRow) types.ReviewView {
	images := []string{}
	_ = json.Unmarshal([]byte(r.Images), &images)
	v := types.ReviewView{
		ID:           strconvI64(r.ID),
		OrderNo:      r.OrderNo,
		MemberID:     strconvI64(r.MemberID),
		Nickname:     r.Nickname,
		Avatar:       r.Avatar,
		ProductID:    strconvI64(r.ProductID),
		ProductTitle: r.ProductTitle,
		Rating:       r.Rating,
		Content:      r.Content,
		Images:       images,
		HealthScore:  r.HealthScore,
		LookScore:    r.LookScore,
		ServiceScore: r.ServiceScore,
		Status:       r.Status,
		CreatedAt:    r.CreatedAt.Format(time.RFC3339),
		Reply:        r.Reply,
	}
	if r.RepliedAt != nil {
		v.RepliedAt = r.RepliedAt.Format(time.RFC3339)
	}
	return v
}

// Create 会员提交评价：订单归属 + 已完成 + 未评过 + 内容校验
func Create(sc *svc.ServiceContext, memberID int64, req *types.ReviewCreateReq) (*types.ReviewView, error) {
	if req.Rating < 1 || req.Rating > 5 {
		return nil, common.ErrParam
	}
	if err := risk.CheckReviewContent(sc, memberID, trimSpace(req.Content)); err != nil {
		return nil, err
	}
	content := sensitive.Filter(sc.DB, trimSpace(req.Content))
	if utf8.RuneCountInString(content) > maxContentRunes {
		return nil, common.ErrParam
	}
	if len(req.Images) > maxImages {
		return nil, common.ErrParam
	}
	for _, img := range req.Images {
		if img == "" || len(img) > 512 {
			return nil, common.ErrParam
		}
	}

	var o model.Order
	if err := sc.DB.Where("order_no = ? AND member_id = ?", req.OrderNo, memberID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	if o.Status != model.OrderCompleted {
		return nil, common.ErrReviewOrder
	}
	var cnt int64
	if err := sc.DB.Model(&model.OrderReview{}).Where("order_no = ?", o.OrderNo).Count(&cnt).Error; err != nil {
		return nil, err
	}
	if cnt > 0 {
		return nil, common.ErrReview
	}

	var item model.OrderItem
	if err := sc.DB.Where("order_id = ?", o.ID).First(&item).Error; err != nil {
		return nil, err
	}
	images, _ := json.Marshal(req.Images)
	health := clampScore(req.HealthScore)
	look := clampScore(req.LookScore)
	service := clampScore(req.ServiceScore)
	r := model.OrderReview{
		ID:           common.NewID(),
		OrderNo:      o.OrderNo,
		MemberID:     memberID,
		ProductID:    item.ProductID,
		ProductTitle: item.ProductTitle,
		Rating:       req.Rating,
		Content:      content,
		Images:       string(images),
		HealthScore:  health,
		LookScore:    look,
		ServiceScore: service,
		Status:       model.ReviewShown,
	}
	if err := sc.DB.Create(&r).Error; err != nil {
		return nil, err
	}
	// 首评奖励积分（失败仅日志，不影响评价提交）
	growth.RewardReview(sc, memberID, o.OrderNo)
	return &types.ReviewView{
		ID:           strconvI64(r.ID),
		OrderNo:      r.OrderNo,
		MemberID:     strconvI64(r.MemberID),
		ProductID:    strconvI64(r.ProductID),
		ProductTitle: r.ProductTitle,
		Rating:       r.Rating,
		Content:      r.Content,
		Images:       req.Images,
		HealthScore:  health,
		LookScore:    look,
		ServiceScore: service,
		Status:       r.Status,
		CreatedAt:    r.CreatedAt.Format(time.RFC3339),
	}, nil
}

func clampScore(v int) int {
	if v < 1 || v > 5 {
		return 5
	}
	return v
}

// ListByProduct 商品评价列表（仅显示中，游标 id DESC）
func ListByProduct(sc *svc.ServiceContext, productID int64, cursor string, limit int) (*types.ReviewListResp, error) {
	limit = normLimit(limit)
	q := sc.DB.Table("order_review r").
		Select("r.*, m.nickname AS nickname, m.avatar AS avatar").
		Joins("LEFT JOIN member m ON m.id = r.member_id").
		Where("r.product_id = ? AND r.status = ?", productID, model.ReviewShown)
	rows, hasMore, err := pageRows(q, cursor, limit)
	if err != nil {
		return nil, err
	}
	list := make([]types.ReviewView, 0, len(rows))
	for _, r := range rows {
		list = append(list, view(r))
	}
	return &types.ReviewListResp{List: list, HasMore: hasMore}, nil
}

// ListByMember 我的评价列表（本人全部，含被隐藏的）
func ListByMember(sc *svc.ServiceContext, memberID int64, cursor string, limit int) (*types.ReviewListResp, error) {
	limit = normLimit(limit)
	q := sc.DB.Table("order_review r").
		Select("r.*, m.nickname AS nickname, m.avatar AS avatar").
		Joins("LEFT JOIN member m ON m.id = r.member_id").
		Where("r.member_id = ?", memberID)
	rows, hasMore, err := pageRows(q, cursor, limit)
	if err != nil {
		return nil, err
	}
	list := make([]types.ReviewView, 0, len(rows))
	for _, r := range rows {
		list = append(list, view(r))
	}
	return &types.ReviewListResp{List: list, HasMore: hasMore}, nil
}

// ProductReviewSummary 详情页评价摘要（评分均值 + 总数 + 最新 3 条）
func ProductReviewSummary(sc *svc.ServiceContext, productID int64) (*types.ReviewSummaryResp, error) {
	var agg struct {
		Avg       float64
		AvgHealth float64
		AvgLook   float64
		AvgSvc    float64
		Total     int64
	}
	if err := sc.DB.Model(&model.OrderReview{}).
		Select("COALESCE(AVG(rating), 0) AS avg, COALESCE(AVG(health_score), 0) AS avg_health,"+
			"COALESCE(AVG(look_score), 0) AS avg_look, COALESCE(AVG(service_score), 0) AS avg_svc, COUNT(*) AS total").
		Where("product_id = ? AND status = ?", productID, model.ReviewShown).
		Scan(&agg).Error; err != nil {
		return nil, err
	}
	resp := &types.ReviewSummaryResp{
		AvgRating:  "0.0",
		AvgHealth:  "0.0",
		AvgLook:    "0.0",
		AvgService: "0.0",
		Total:      agg.Total,
		Latest:     []types.ReviewView{},
	}
	if agg.Total > 0 {
		resp.AvgRating = trimFloat(agg.Avg)
		resp.AvgHealth = trimFloat(agg.AvgHealth)
		resp.AvgLook = trimFloat(agg.AvgLook)
		resp.AvgService = trimFloat(agg.AvgSvc)
		page, err := ListByProduct(sc, productID, "", 3)
		if err != nil {
			return nil, err
		}
		resp.Latest = page.List
	}
	return resp, nil
}

// AdminList 平台端评价列表（含隐藏）
func AdminList(sc *svc.ServiceContext, cursor string, limit int) (*types.ReviewListResp, error) {
	limit = normLimit(limit)
	q := sc.DB.Table("order_review r").
		Select("r.*, m.nickname AS nickname, m.avatar AS avatar").
		Joins("LEFT JOIN member m ON m.id = r.member_id")
	rows, hasMore, err := pageRows(q, cursor, limit)
	if err != nil {
		return nil, err
	}
	list := make([]types.ReviewView, 0, len(rows))
	for _, r := range rows {
		list = append(list, view(r))
	}
	return &types.ReviewListResp{List: list, HasMore: hasMore}, nil
}

// AdminSetStatus 平台端隐藏/显示评价
func AdminSetStatus(sc *svc.ServiceContext, id int64, status int) error {
	if status != model.ReviewShown && status != model.ReviewHidden {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.OrderReview{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// AdminReply 平台端官方回复（可重复调用覆盖，replied_at 刷新）
func AdminReply(sc *svc.ServiceContext, id int64, reply string) error {
	reply = trimSpace(reply)
	if reply == "" || utf8.RuneCountInString(reply) > maxReplyRunes {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.OrderReview{}).Where("id = ?", id).
		Updates(map[string]any{"reply": reply, "replied_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// AdminDelete 平台端删除评价
func AdminDelete(sc *svc.ServiceContext, id int64) error {
	res := sc.DB.Where("id = ?", id).Delete(&model.OrderReview{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// pageRows 游标分页：cursor 为上一页末条 id，取更早一页
func pageRows(q *gorm.DB, cursor string, limit int) ([]reviewRow, bool, error) {
	if cursor != "" {
		id, err := parseID(cursor)
		if err != nil {
			return nil, false, err
		}
		q = q.Where("r.id < ?", id)
	}
	var rows []reviewRow
	if err := q.Order("r.id DESC").Limit(limit + 1).Scan(&rows).Error; err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, nil
}

func normLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

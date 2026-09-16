// history.go 浏览历史（Redis LIST，最近在后、展示倒序取最新 50）+ 详情页相关推荐（同品种 → 同分类 → 热销兜底）。
package pet

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

const (
	redisViewFmt = "rl:view:%d"
	viewKeep     = 50
	redisViewTTL = 30 * 24 * time.Hour
	relatedSize  = 4
)

// recordView 记录浏览：去重后追加到队尾，超长从头裁掉（异步调用，不阻塞）
func recordView(sc *svc.ServiceContext, memberID, productID int64) {
	key := fmt.Sprintf(redisViewFmt, memberID)
	ctx := context.Background()
	pipe := sc.Rdb.Pipeline()
	pipe.LRem(ctx, key, 0, strconv.FormatInt(productID, 10))
	pipe.RPush(ctx, key, strconv.FormatInt(productID, 10))
	pipe.LTrim(ctx, key, -viewKeep, -1)
	pipe.Expire(ctx, key, redisViewTTL)
	_, _ = pipe.Exec(ctx)
}

// ViewHistory 我的浏览历史（最近在前，仅含仍在售商品）
func ViewHistory(sc *svc.ServiceContext, memberID int64) (*types.PageResp, error) {
	if memberID <= 0 {
		return &types.PageResp{List: []types.ProductCard{}}, nil
	}
	key := fmt.Sprintf(redisViewFmt, memberID)
	rows, err := sc.Rdb.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		if id, e := strconv.ParseInt(r, 10, 64); e == nil {
			ids = append(ids, id)
		}
	}
	empty := &types.PageResp{List: []types.ProductCard{}}
	if len(ids) == 0 {
		return empty, nil
	}
	var products []model.PetProduct
	if err := sc.DB.Where("id IN ? AND status = ?", ids, model.ProductOnSale).Find(&products).Error; err != nil {
		return nil, err
	}
	byID := map[int64]model.PetProduct{}
	for _, p := range products {
		byID[p.ID] = p
	}
	ordered := make([]model.PetProduct, 0, len(products))
	seen := map[int64]bool{}
	for i := len(ids) - 1; i >= 0; i-- {
		id := ids[i]
		if !seen[id] {
			seen[id] = true
			if p, ok := byID[id]; ok {
				ordered = append(ordered, p)
			}
		}
	}
	list, err := buildCards(sc, ordered)
	if err != nil {
		return nil, err
	}
	return &types.PageResp{Total: int64(len(list)), List: list}, nil
}

// RelatedProducts 相关推荐：同品种优先，不足补同分类，再不足补全站热销；均排除自身
func RelatedProducts(sc *svc.ServiceContext, idStr string) (*types.PageResp, error) {
	id, err := parseID(idStr)
	if err != nil {
		return nil, err
	}
	var p model.PetProduct
	if err := sc.DB.Select("id", "breed_id", "category_id").First(&p, id).Error; err != nil {
		return nil, common.ErrNotFound
	}
	picked := map[int64]bool{p.ID: true}
	var pickedRows []model.PetProduct
	fill := func(cond string, args ...any) {
		if len(pickedRows) >= relatedSize {
			return
		}
		var rows []model.PetProduct
		q := sc.DB.Where("status = ?", model.ProductOnSale).Where(cond, args...)
		if err := q.Order("sales DESC, id DESC").Limit(relatedSize * 2).Find(&rows).Error; err != nil {
			return
		}
		for _, r := range rows {
			if !picked[r.ID] {
				picked[r.ID] = true
				pickedRows = append(pickedRows, r)
				if len(pickedRows) >= relatedSize {
					break
				}
			}
		}
	}
	fill("breed_id = ?", p.BreedID)
	fill("category_id = ?", p.CategoryID)
	fill("id > 0")
	if pickedRows == nil {
		pickedRows = []model.PetProduct{}
	}
	list, err := buildCards(sc, pickedRows)
	if err != nil {
		return nil, err
	}
	return &types.PageResp{Total: int64(len(list)), List: list}, nil
}

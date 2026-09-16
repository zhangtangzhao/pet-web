// search.go 搜索增强：热搜榜（ZSET 7 天滚动）+ 个人搜索历史（去重留 10 条）+ 搜索联想。
package pet

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

const (
	redisSearchHot     = "rl:search:hot"
	redisSearchHistFmt = "rl:search:hist:%d"
	redisSearchTTL     = 7 * 24 * time.Hour
	historyKeep        = 10
)

func normKeyword(kw string) (string, error) {
	kw = strings.Join(strings.Fields(strings.TrimSpace(kw)), " ")
	if kw == "" || len([]rune(kw)) > 30 {
		return "", common.ErrParam
	}
	return kw, nil
}

// TraceSearch 记录一次搜索：热搜 +1，个人历史去重置顶（游客只记热搜）
func TraceSearch(sc *svc.ServiceContext, memberID int64, keyword string) error {
	kw, err := normKeyword(keyword)
	if err != nil {
		return err
	}
	ctx := context.Background()
	pipe := sc.Rdb.Pipeline()
	pipe.ZIncrBy(ctx, redisSearchHot, 1, kw)
	pipe.Expire(ctx, redisSearchHot, redisSearchTTL)
	if memberID > 0 {
		key := fmt.Sprintf(redisSearchHistFmt, memberID)
		pipe.LRem(ctx, key, 0, kw)
		pipe.LPush(ctx, key, kw)
		pipe.LTrim(ctx, key, 0, historyKeep-1)
		pipe.Expire(ctx, key, redisSearchTTL)
	}
	_, err = pipe.Exec(ctx)
	return err
}

// SearchHistory 我的搜索历史（最新在前）
func SearchHistory(sc *svc.ServiceContext, memberID int64) (*types.SearchListResp, error) {
	if memberID <= 0 {
		return &types.SearchListResp{List: []string{}}, nil
	}
	key := fmt.Sprintf(redisSearchHistFmt, memberID)
	rows, err := sc.Rdb.LRange(context.Background(), key, 0, historyKeep-1).Result()
	if err != nil {
		if err == redis.Nil {
			return &types.SearchListResp{List: []string{}}, nil
		}
		return nil, err
	}
	if rows == nil {
		rows = []string{}
	}
	return &types.SearchListResp{List: rows}, nil
}

// ClearSearchHistory 清空我的搜索历史
func ClearSearchHistory(sc *svc.ServiceContext, memberID int64) error {
	return sc.Rdb.Del(context.Background(), fmt.Sprintf(redisSearchHistFmt, memberID)).Err()
}

// SearchHot 热搜榜（前 10）
func SearchHot(sc *svc.ServiceContext) (*types.SearchHotResp, error) {
	rows, err := sc.Rdb.ZRevRangeWithScores(context.Background(), redisSearchHot, 0, 9).Result()
	if err != nil {
		return nil, err
	}
	list := make([]types.SearchHotRow, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.SearchHotRow{Keyword: r.Member.(string), Score: r.Score})
	}
	return &types.SearchHotResp{List: list}, nil
}

// SearchComplete 搜索联想：品种名 + 分类名 + 商品标题，各取若干合并去重
func SearchComplete(sc *svc.ServiceContext, prefix string) (*types.SearchListResp, error) {
	p := strings.TrimSpace(prefix)
	if p == "" {
		return &types.SearchListResp{List: []string{}}, nil
	}
	like := p + "%"
	seen := map[string]bool{}
	list := make([]string, 0, 10)
	add := func(rows []string) {
		for _, r := range rows {
			r = strings.TrimSpace(r)
			if r == "" || seen[r] {
				continue
			}
			seen[r] = true
			list = append(list, r)
		}
	}
	var breedNames []string
	if err := sc.DB.Model(&model.Breed{}).Where("status = 1 AND name ILIKE ?", like).
		Order("sort ASC, id ASC").Limit(5).Pluck("name", &breedNames).Error; err != nil {
		return nil, err
	}
	var categoryNames []string
	if err := sc.DB.Model(&model.Category{}).Where("status = 1 AND name ILIKE ?", like).
		Order("sort ASC, id ASC").Limit(3).Pluck("name", &categoryNames).Error; err != nil {
		return nil, err
	}
	var titles []string
	if err := sc.DB.Model(&model.PetProduct{}).Where("status = ? AND title ILIKE ?", model.ProductOnSale, like).
		Order("sales DESC, id DESC").Limit(5).Pluck("title", &titles).Error; err != nil {
		return nil, err
	}
	add(breedNames)
	add(categoryNames)
	add(titles)
	if list == nil {
		list = []string{}
	}
	return &types.SearchListResp{List: list}, nil
}

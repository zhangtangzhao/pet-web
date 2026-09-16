// Package ratelimit Redis 固定窗口计数限流（INCR + EXPIRE）。
// limit ≤0 视为关闭直接放行；Redis 异常时放行（限流是保护措施，不应放大故障）。
package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Allow 记一次并判断是否放行；window 内累计 key 计数超过 limit 返回 false
func Allow(rdb redis.UniversalClient, key string, limit int, window time.Duration) bool {
	if limit <= 0 {
		return true
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	cnt, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return true
	}
	if cnt == 1 {
		rdb.Expire(ctx, key, window)
	}
	return cnt <= int64(limit)
}

// Package sensitive 敏感词过滤：进程内缓存词表（60s 刷新），
// Filter 用于展示文本替换（评价/售后原因），Contains 用于发送拦截（客服聊天）。
package sensitive

import (
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"pet/backend/internal/model"
)

const (
	cacheTTL   = time.Minute
	maskSymbol = "***"
)

var (
	mu       sync.RWMutex
	words    []string
	loadedAt time.Time
)

func load(db *gorm.DB) {
	var rows []model.SensitiveWord
	if err := db.Where("status = ?", model.SensitiveWordOn).Find(&rows).Error; err != nil {
		return // 失败沿用旧缓存
	}
	list := make([]string, 0, len(rows))
	for _, r := range rows {
		if w := strings.TrimSpace(r.Word); w != "" {
			list = append(list, w)
		}
	}
	mu.Lock()
	words, loadedAt = list, time.Now()
	mu.Unlock()
}

func get(db *gorm.DB) []string {
	mu.RLock()
	if time.Since(loadedAt) < cacheTTL {
		defer mu.RUnlock()
		return words
	}
	mu.RUnlock()
	load(db)
	mu.RLock()
	defer mu.RUnlock()
	return words
}

// Contains 是否命中启用中的敏感词
func Contains(db *gorm.DB, text string) bool {
	if text == "" {
		return false
	}
	for _, w := range get(db) {
		if strings.Contains(strings.ToLower(text), strings.ToLower(w)) {
			return true
		}
	}
	return false
}

// Filter 命中部分替换为 ***（用于评价/售后原因等展示文本）
func Filter(db *gorm.DB, text string) string {
	if text == "" {
		return text
	}
	for _, w := range get(db) {
		if strings.Contains(strings.ToLower(text), strings.ToLower(w)) {
			text = strings.ReplaceAll(text, w, maskSymbol)
		}
	}
	return text
}

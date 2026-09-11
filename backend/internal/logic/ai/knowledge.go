package ai

import (
	"strings"

	"pet/backend/internal/model"
	"pet/backend/internal/svc"
)

// retrieveKnowledge 知识库检索 v1：品种归属 + 关键词命中打分（不依赖 embedding 服务）。
// 候选 = 启用且（属于该品种 或 平台通用），截取候选集后内存打分取 topN；
// 无任何命中时回退为「品种专属优先」的默认排序，保证提示词始终有据可依。
func retrieveKnowledge(sc *svc.ServiceContext, breedID int64, question string, limit int) []model.AIKnowledge {
	if limit <= 0 {
		limit = 4
	}
	var candidates []model.AIKnowledge
	if err := sc.DB.Model(&model.AIKnowledge{}).
		Where("status = 1 AND (breed_id = ? OR breed_id IS NULL)", breedID).
		Order("sort ASC, id DESC").Limit(100).
		Find(&candidates).Error; err != nil {
		return nil
	}

	scored := make([]model.AIKnowledge, 0, len(candidates))
	for _, k := range candidates {
		s := knowledgeScore(&k, question)
		if k.BreedID != nil {
			s += 2 // 品种专属加权
		}
		if s > 0 {
			scored = append(scored, k)
		}
	}
	if len(scored) == 0 {
		for _, k := range candidates {
			if k.BreedID != nil {
				scored = append(scored, k)
			}
		}
	}
	if len(scored) > limit {
		scored = scored[:limit]
	}
	return scored
}

// knowledgeScore 关键词命中×3 + 标题命中×2 + 内容包含整句问题×1
func knowledgeScore(k *model.AIKnowledge, question string) int {
	if question == "" {
		return 0
	}
	s := 0
	for _, kw := range strings.Split(k.Keywords, ",") {
		if kw = strings.TrimSpace(kw); kw != "" && strings.Contains(question, kw) {
			s += 3
		}
	}
	if k.Title != "" && strings.Contains(question, k.Title) {
		s += 2
	}
	if len([]rune(question)) <= 30 && strings.Contains(k.Content, question) {
		s += 1
	}
	return s
}

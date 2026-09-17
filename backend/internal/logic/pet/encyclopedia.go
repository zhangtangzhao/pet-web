// encyclopedia.go 养宠百科中心：前端化展示 ai_knowledge（按品种聚合，公开）。
package pet

import (
	"strconv"

	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// EncyclopediaBreeds 百科首页：有专属知识的品种 + 文章数
func EncyclopediaBreeds(sc *svc.ServiceContext) ([]types.EncyclopediaBreed, error) {
	var breeds []model.Breed
	if err := sc.DB.Where("status = ?", 1).Order("sort ASC, id ASC").Limit(50).Find(&breeds).Error; err != nil {
		return nil, err
	}
	type cntRow struct {
		BreedID int64
		Cnt     int64
	}
	var cnts []cntRow
	if err := sc.DB.Model(&model.AIKnowledge{}).
		Select("breed_id AS breed_id, COUNT(*) AS cnt").
		Where("status = 1 AND breed_id IS NOT NULL").
		Group("breed_id").Scan(&cnts).Error; err != nil {
		return nil, err
	}
	cntMap := map[int64]int64{}
	for _, c := range cnts {
		cntMap[c.BreedID] = c.Cnt
	}
	out := make([]types.EncyclopediaBreed, 0, len(breeds))
	for _, b := range breeds {
		if cntMap[b.ID] == 0 {
			continue
		}
		out = append(out, types.EncyclopediaBreed{
			BreedID:    strconv.FormatInt(b.ID, 10),
			BreedName:  b.Name,
			Cover:      b.Cover,
			ArticleCnt: int(cntMap[b.ID]),
		})
	}
	return out, nil
}

// EncyclopediaArticles 某品种的文章列表（含正文）
func EncyclopediaArticles(sc *svc.ServiceContext, breedID int64) ([]types.EncyclopediaArticle, error) {
	var articles []model.AIKnowledge
	if err := sc.DB.Where("breed_id = ? AND status = ?", breedID, 1).
		Order("sort ASC, id ASC").Limit(20).Find(&articles).Error; err != nil {
		return nil, err
	}
	out := make([]types.EncyclopediaArticle, 0, len(articles))
	for _, a := range articles {
		out = append(out, types.EncyclopediaArticle{
			Title:   a.Title,
			Content: a.Content,
		})
	}
	return out, nil
}

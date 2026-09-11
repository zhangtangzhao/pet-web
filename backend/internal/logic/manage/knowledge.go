package manage

import (
	"strconv"
	"strings"
	"time"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// KnowledgeList 平台端知识库分页列表
func KnowledgeList(sc *svc.ServiceContext, req *types.KnowledgeListReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}
	query := sc.DB.Model(&model.AIKnowledge{})
	if req.BreedID != "" {
		id, err := parseID(req.BreedID)
		if err != nil {
			return nil, err
		}
		query = query.Where("breed_id = ?", id)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		query = query.Where("title ILIKE ? OR keywords ILIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.AIKnowledge
	if err := query.Order("sort ASC, id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, err
	}

	breedName := map[int64]string{}
	ids := make([]int64, 0, len(list))
	for _, k := range list {
		if k.BreedID != nil {
			ids = append(ids, *k.BreedID)
		}
	}
	if len(ids) > 0 {
		var breeds []model.Breed
		if err := sc.DB.Where("id IN ?", ids).Find(&breeds).Error; err != nil {
			return nil, err
		}
		for _, b := range breeds {
			breedName[b.ID] = b.Name
		}
	}

	items := make([]types.KnowledgeItem, 0, len(list))
	for _, k := range list {
		item := types.KnowledgeItem{
			ID:        formatID(k.ID),
			BreedName: "平台通用",
			Title:     k.Title,
			Keywords:  k.Keywords,
			Content:   k.Content,
			Sort:      k.Sort,
			Status:    k.Status,
			UpdatedAt: k.UpdatedAt.Format(time.DateTime),
		}
		if k.BreedID != nil {
			item.BreedID = formatID(*k.BreedID)
			item.BreedName = breedName[*k.BreedID]
		}
		items = append(items, item)
	}
	return &types.PageResp{Total: total, List: items}, nil
}

// UpsertKnowledge 新建/编辑知识条目
func UpsertKnowledge(sc *svc.ServiceContext, req *types.KnowledgeUpsertReq) error {
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" {
		return common.ErrParam
	}
	var breedID *int64
	if req.BreedID != "" {
		id, err := parseID(req.BreedID)
		if err != nil {
			return err
		}
		if !existsByID(sc.DB, &model.Breed{}, id) {
			return common.ErrAIReferenced
		}
		breedID = &id
	}
	status := req.Status
	if status != 0 {
		status = 1
	}

	if req.ID == "" {
		return sc.DB.Create(&model.AIKnowledge{
			ID: common.NewID(), BreedID: breedID, Title: req.Title,
			Keywords: req.Keywords, Content: req.Content, Sort: req.Sort, Status: status,
		}).Error
	}
	id, err := parseID(req.ID)
	if err != nil {
		return err
	}
	// breed_id 可空，Updates map 精确覆盖
	res := sc.DB.Model(&model.AIKnowledge{}).Where("id = ?", id).
		Updates(map[string]any{
			"breed_id": breedID, "title": req.Title, "keywords": req.Keywords,
			"content": req.Content, "sort": req.Sort, "status": status,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// DeleteKnowledge 删除知识条目
func DeleteKnowledge(sc *svc.ServiceContext, idStr string) error {
	id, err := parseID(idStr)
	if err != nil {
		return err
	}
	res := sc.DB.Delete(&model.AIKnowledge{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

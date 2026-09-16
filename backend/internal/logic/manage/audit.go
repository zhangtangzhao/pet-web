// audit.go 操作审计日志查询 + 敏感词管理。
package manage

import (
	"strings"
	"time"
	"unicode/utf8"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// ─────────────────────────── 审计日志 ───────────────────────────

// AdminAuditList 审计日志分页（id DESC）
func AdminAuditList(sc *svc.ServiceContext, req *types.PageReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	q := sc.DB.Model(&model.AdminAuditLog{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.AdminAuditLog
	if err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.AuditLogRow, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.AuditLogRow{
			ID:        strconvI64(r.ID),
			AdminName: r.AdminName,
			Method:    r.Method,
			Path:      r.Path,
			IP:        r.IP,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		})
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// ─────────────────────────── 敏感词 ───────────────────────────

// AdminSensitiveList 敏感词分页
func AdminSensitiveList(sc *svc.ServiceContext, req *types.PageReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	q := sc.DB.Model(&model.SensitiveWord{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.SensitiveWord
	if err := q.Order("id ASC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.SensitiveWordRow, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.SensitiveWordRow{
			ID:        strconvI64(r.ID),
			Word:      r.Word,
			Status:    r.Status,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		})
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// AdminSensitiveSave 新增敏感词（进程内词表 60s 自动刷新）
func AdminSensitiveSave(sc *svc.ServiceContext, req *types.SensitiveWordSaveReq) error {
	word := strings.TrimSpace(req.Word)
	if word == "" || utf8.RuneCountInString(word) > 64 {
		return common.ErrParam
	}
	var cnt int64
	if err := sc.DB.Model(&model.SensitiveWord{}).Where("word = ?", word).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return common.NewErr(400, 40002, "敏感词已存在")
	}
	return sc.DB.Create(&model.SensitiveWord{ID: common.NewID(), Word: word, Status: model.SensitiveWordOn}).Error
}

// AdminSensitiveStatus 启用 / 停用
func AdminSensitiveStatus(sc *svc.ServiceContext, id int64, status int) error {
	if status != model.SensitiveWordOff && status != model.SensitiveWordOn {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.SensitiveWord{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// AdminSensitiveDelete 删除敏感词
func AdminSensitiveDelete(sc *svc.ServiceContext, id int64) error {
	res := sc.DB.Delete(&model.SensitiveWord{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

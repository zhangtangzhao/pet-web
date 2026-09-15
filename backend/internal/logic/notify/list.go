// list.go 站内消息中心：列表（分页 + 未读数）与全部已读。
// 站内已读（read_at）与微信投递状态（status）解耦，列表展示全部事件行。
package notify

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// List 消息列表（id DESC 分页）+ 未读数
func List(sc *svc.ServiceContext, memberID int64, req *types.PageReq) (*types.NotifyListResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}
	// 每次取全新查询链，避免 Where 条件在同一条语句上累积污染后续查询
	base := func() *gorm.DB {
		return sc.DB.Model(&model.Notification{}).Where("member_id = ?", memberID)
	}

	var total int64
	if err := base().Count(&total).Error; err != nil {
		return nil, err
	}
	var unread int64
	if err := base().Where("read_at IS NULL").Count(&unread).Error; err != nil {
		return nil, err
	}
	var rows []model.Notification
	if err := base().Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, err
	}

	list := make([]types.NotifyView, 0, len(rows))
	for _, n := range rows {
		v := types.NotifyView{
			ID:        strconv.FormatInt(n.ID, 10),
			Scene:     n.Scene,
			Title:     n.Title,
			Content:   n.Content,
			OrderNo:   n.OrderNo,
			CreatedAt: n.CreatedAt.Format(time.RFC3339),
		}
		if n.ReadAt != nil {
			v.ReadAt = n.ReadAt.Format(time.RFC3339)
		}
		list = append(list, v)
	}
	return &types.NotifyListResp{PageResp: types.PageResp{Total: total, List: list}, Unread: int(unread)}, nil
}

// ReadAll 全部已读：未读行置 read_at=now
func ReadAll(sc *svc.ServiceContext, memberID int64) error {
	return sc.DB.Model(&model.Notification{}).
		Where("member_id = ? AND read_at IS NULL", memberID).
		Update("read_at", time.Now()).Error
}

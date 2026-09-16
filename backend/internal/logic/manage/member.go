package manage

import (
	"strconv"
	"time"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/growth"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// MemberList 会员列表（含订单数/收藏数子查询统计）
func MemberList(sc *svc.ServiceContext, req *types.MemberListReq) (*types.PageResp, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	query := sc.DB.Model(&model.Member{})
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where("nickname ILIKE ? OR phone ILIKE ?", kw, kw)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var members []model.Member
	if err := query.Order("created_at DESC, id DESC").Offset((page - 1) * size).Limit(size).Find(&members).Error; err != nil {
		return nil, err
	}
	list := make([]types.MemberAdminView, 0, len(members))
	for _, m := range members {
		var orderCnt, favCnt int64
		_ = sc.DB.Model(&model.Order{}).Where("member_id = ?", m.ID).Count(&orderCnt).Error
		_ = sc.DB.Model(&model.MemberFavorite{}).Where("member_id = ?", m.ID).Count(&favCnt).Error
		list = append(list, types.MemberAdminView{
			ID:          strconv.FormatInt(m.ID, 10),
			Nickname:    m.Nickname,
			Avatar:      m.Avatar,
			Phone:       maskPhone(m.Phone),
			Gender:      m.Gender,
			Status:      m.Status,
			OrderCount:  orderCnt,
			FavCount:    favCnt,
			GrowthValue: m.GrowthValue,
			LevelName:   growth.LevelName(growth.LevelOf(m.GrowthValue)),
			CreatedAt:   m.CreatedAt.Format(time.RFC3339),
		})
	}
	return &types.PageResp{Total: total, List: list}, nil
}

// UpdateMemberStatus 启用/禁用会员（禁用后无法登录/刷新 token）
func UpdateMemberStatus(sc *svc.ServiceContext, req *types.MemberStatusReq) error {
	id, err := parseID(req.ID)
	if err != nil {
		return err
	}
	if req.Status != 1 && req.Status != 2 {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.Member{}).Where("id = ?", id).Update("status", req.Status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// UpdateMemberBlacklist 拉黑/解除（可登录但禁交易/评价/领券）
func UpdateMemberBlacklist(sc *svc.ServiceContext, req *types.MemberBlacklistReq) error {
	id, err := parseID(req.ID)
	if err != nil {
		return err
	}
	if req.Blacklist != 0 && req.Blacklist != 1 {
		return common.ErrParam
	}
	res := sc.DB.Model(&model.Member{}).Where("id = ?", id).Update("blacklist", req.Blacklist)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func maskPhone(phone string) string {
	if len(phone) == 11 {
		return phone[:3] + "****" + phone[7:]
	}
	return phone
}

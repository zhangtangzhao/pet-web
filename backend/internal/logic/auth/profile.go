package auth

import (
	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func Profile(sc *svc.ServiceContext, memberID int64) (*types.MemberInfo, error) {
	var m model.Member
	if err := sc.DB.First(&m, memberID).Error; err != nil {
		return nil, common.ErrNotFound
	}
	v := MemberView(sc, &m, false)
	return &v, nil
}

func UpdateProfile(sc *svc.ServiceContext, memberID int64, req *types.UpdateProfileReq) error {
	updates := map[string]any{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Gender > 0 {
		updates["gender"] = req.Gender
	}
	if len(updates) == 0 {
		return nil
	}
	return sc.DB.Model(&model.Member{}).Where("id = ?", memberID).Updates(updates).Error
}

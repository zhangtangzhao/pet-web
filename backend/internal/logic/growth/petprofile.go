// Package growth 用户宠物档案：品种/生日/体重/疫苗记录 + 护理提醒数据源。
package growth

import (
	"strconv"
	"strings"
	"time"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func parseDay(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return nil
	}
	return &t
}

func dayStr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// UserPetList 我的宠物档案
func UserPetList(sc *svc.ServiceContext, memberID int64) (*types.UserPetListResp, error) {
	var pets []model.PetProfile
	if err := sc.DB.Where("member_id = ?", memberID).Order("id DESC").Limit(50).Find(&pets).Error; err != nil {
		return nil, err
	}
	list := make([]types.UserPetView, 0, len(pets))
	for _, p := range pets {
		list = append(list, petView(p))
	}
	return &types.UserPetListResp{List: list}, nil
}

func petView(p model.PetProfile) types.UserPetView {
	v := types.UserPetView{
		ID:              strconv.FormatInt(p.ID, 10),
		Name:            p.Name,
		BreedName:       p.BreedName,
		Gender:          p.Gender,
		Birthday:        dayStr(p.Birthday),
		Avatar:          p.Avatar,
		VaccineAt:       dayStr(p.VaccineAt),
		NextVaccineDate: dayStr(p.NextVaccineDate),
		NextDewormDate:  dayStr(p.NextDewormDate),
		Remark:          p.Remark,
		CreatedAt:       p.CreatedAt.Format(time.RFC3339),
	}
	if p.Weight != nil {
		v.Weight = strconv.FormatFloat(*p.Weight, 'f', 2, 64)
	}
	return v
}

// UserPetUpsert 新建/编辑宠物档案
func UserPetUpsert(sc *svc.ServiceContext, memberID int64, req *types.UserPetUpsertReq) error {
	name := strings.TrimSpace(req.Name)
	if name == "" || len([]rune(name)) > 32 {
		return common.NewErr(400, 40001, "请填写宠物昵称")
	}
	breed := strings.TrimSpace(req.BreedName)
	if len([]rune(breed)) > 32 {
		return common.NewErr(400, 40001, "品种名过长")
	}
	remark := strings.TrimSpace(req.Remark)
	if len([]rune(remark)) > 255 {
		return common.NewErr(400, 40001, "备注过长")
	}
	gender := req.Gender
	if gender != 1 && gender != 2 {
		gender = 0
	}
	var weight *float64
	if req.Weight != "" {
		if w, err := strconv.ParseFloat(req.Weight, 64); err == nil && w > 0 && w < 200 {
			weight = &w
		}
	}
	fields := map[string]any{
		"name":              name,
		"breed_name":        breed,
		"gender":            gender,
		"birthday":          parseDay(req.Birthday),
		"weight":            weight,
		"avatar":            truncateStr(req.Avatar, 512),
		"vaccine_at":        parseDay(req.VaccineAt),
		"next_vaccine_date": parseDay(req.NextVaccineDate),
		"next_deworm_date":  parseDay(req.NextDewormDate),
		"remark":            remark,
		"updated_at":        time.Now(),
	}
	if req.ID != "" {
		id, err := strconv.ParseInt(req.ID, 10, 64)
		if err != nil || id <= 0 {
			return common.ErrParam
		}
		res := sc.DB.Model(&model.PetProfile{}).
			Where("id = ? AND member_id = ?", id, memberID).Updates(fields)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return common.ErrNotFound
		}
		return nil
	}
	pet := model.PetProfile{
		ID:       common.NewID(),
		MemberID: memberID,
	}
	pet.Name = name
	pet.BreedName = breed
	pet.Gender = gender
	pet.Birthday = parseDay(req.Birthday)
	pet.Weight = weight
	pet.Avatar = truncateStr(req.Avatar, 512)
	pet.VaccineAt = parseDay(req.VaccineAt)
	pet.NextVaccineDate = parseDay(req.NextVaccineDate)
	pet.NextDewormDate = parseDay(req.NextDewormDate)
	pet.Remark = remark
	return sc.DB.Create(&pet).Error
}

// UserPetDelete 删除档案
func UserPetDelete(sc *svc.ServiceContext, memberID, petID int64) error {
	res := sc.DB.Where("id = ? AND member_id = ?", petID, memberID).Delete(&model.PetProfile{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

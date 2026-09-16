package trade

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func addressView(a model.MemberAddress) *types.AddressView {
	return &types.AddressView{
		ID:        strconv.FormatInt(a.ID, 10),
		Name:      a.Name,
		Phone:     a.Phone,
		Address:   a.Address,
		IsDefault: a.IsDefault,
	}
}

// AddressList 我的地址簿（默认地址置顶，其余按更新时间倒序）
func AddressList(sc *svc.ServiceContext, memberID int64) ([]*types.AddressView, error) {
	var rows []model.MemberAddress
	if err := sc.DB.Where("member_id = ?", memberID).
		Order("is_default DESC, updated_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]*types.AddressView, 0, len(rows))
	for _, a := range rows {
		list = append(list, addressView(a))
	}
	return list, nil
}

// AddressSave 新建 / 编辑（编辑时校验归属）；设为默认会清掉原默认
func AddressSave(sc *svc.ServiceContext, memberID int64, req *types.AddressSaveReq) error {
	name := strings.TrimSpace(req.Name)
	address := strings.TrimSpace(req.Address)
	phone := strings.TrimSpace(req.Phone)
	if name == "" || phone == "" || address == "" ||
		utf8.RuneCountInString(name) > 32 || utf8.RuneCountInString(address) > 255 ||
		len(phone) > 20 {
		return common.ErrParam
	}
	isDefault := 0
	if req.IsDefault {
		isDefault = 1
	}

	return sc.DB.Transaction(func(tx *gorm.DB) error {
		if isDefault == 1 {
			if err := tx.Model(&model.MemberAddress{}).
				Where("member_id = ? AND is_default = 1", memberID).
				Update("is_default", 0).Error; err != nil {
				return err
			}
		}
		if req.ID == "" {
			return tx.Create(&model.MemberAddress{
				ID: common.NewID(), MemberID: memberID,
				Name: name, Phone: phone, Address: address, IsDefault: isDefault,
			}).Error
		}
		id, err := parseID(req.ID)
		if err != nil {
			return err
		}
		res := tx.Model(&model.MemberAddress{}).
			Where("id = ? AND member_id = ?", id, memberID).
			Updates(map[string]any{
				"name": name, "phone": phone, "address": address,
				"is_default": isDefault, "updated_at": time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return common.ErrNotFound
		}
		return nil
	})
}

// AddressSetDefault 设默认地址
func AddressSetDefault(sc *svc.ServiceContext, memberID int64, id string) error {
	addrID, err := parseID(id)
	if err != nil {
		return err
	}
	return sc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.MemberAddress{}).
			Where("member_id = ? AND is_default = 1", memberID).
			Update("is_default", 0).Error; err != nil {
			return err
		}
		res := tx.Model(&model.MemberAddress{}).
			Where("id = ? AND member_id = ?", addrID, memberID).
			Updates(map[string]any{"is_default": 1, "updated_at": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return common.ErrNotFound
		}
		return nil
	})
}

// AddressDelete 删除地址（归属校验）
func AddressDelete(sc *svc.ServiceContext, memberID int64, id string) error {
	addrID, err := parseID(id)
	if err != nil {
		return err
	}
	res := sc.DB.Where("id = ? AND member_id = ?", addrID, memberID).Delete(&model.MemberAddress{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// certificate.go 电子健康证书：已完成订单生成健康档案卡（编号 HC-订单号），持有人昵称/打码手机号。
package trade

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// Certificate 查看订单电子健康证书（归属校验 + 仅已完成）
func Certificate(sc *svc.ServiceContext, memberID int64, orderNo string) (*types.CertView, error) {
	var o model.Order
	if err := sc.DB.Where("order_no = ? AND member_id = ?", orderNo, memberID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	if o.Status != model.OrderCompleted || o.CompletedAt == nil {
		return nil, common.ErrCertState
	}
	var item model.OrderItem
	if err := sc.DB.Where("order_id = ?", o.ID).Order("id ASC").First(&item).Error; err != nil {
		return nil, err
	}
	var m model.Member
	if err := sc.DB.Select("nickname", "phone").First(&m, o.MemberID).Error; err != nil {
		return nil, err
	}
	holder := m.Nickname
	if holder == "" {
		holder = maskPhone(m.Phone)
	}
	v := &types.CertView{
		CertNo:        "HC-" + o.OrderNo,
		OrderNo:       o.OrderNo,
		MemberMasked:  holder,
		ProductTitle:  item.ProductTitle,
		BreedName:     item.BreedName,
		CompletedAt:   o.CompletedAt.Format(time.RFC3339),
		GuaranteeDays: o.GuaranteeDays,
	}
	if o.GuaranteeDays > 0 {
		v.GuaranteeEndAt = o.CompletedAt.AddDate(0, 0, o.GuaranteeDays).Format("2006-01-02")
	}
	var p model.PetProduct
	if err := sc.DB.Select("quarantine_cert_url", "supplier_id").First(&p, item.ProductID).Error; err == nil {
		v.QuarantineURL = p.QuarantineCertURL
		if p.SupplierID > 0 {
			var s model.Supplier
			if err := sc.DB.Select("name").First(&s, p.SupplierID).Error; err == nil {
				v.SupplierName = s.Name
			}
		}
	}
	return v, nil
}

func maskPhone(phone string) string {
	rs := []rune(phone)
	if len(rs) != 11 {
		return "宠物家长"
	}
	return string(rs[:3]) + "****" + string(rs[7:])
}

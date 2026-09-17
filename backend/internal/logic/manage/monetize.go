package manage

import (
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func InsuranceList(sc *svc.ServiceContext) ([]types.InsuranceProductView, error) {
	var rows []model.InsuranceProduct
	if err := sc.DB.Where("status = 1").Order("id ASC").Find(&rows).Error; err != nil { return nil, err }
	out := make([]types.InsuranceProductView, 0, len(rows))
	for _, r := range rows {
		out = append(out, types.InsuranceProductView{ID: strconv.FormatInt(r.ID, 10), Name: r.Name, Company: r.Company, CoverDesc: r.CoverDesc, Price: r.Price.StringFixed(2)})
	}
	return out, nil
}

func AdminInsuranceUpsert(sc *svc.ServiceContext, req *types.InsuranceProductView) error {
	id, _ := strconv.ParseInt(req.ID, 10, 64)
	price, _ := decimal.NewFromString(req.Price)
	if id > 0 {
		return sc.DB.Model(&model.InsuranceProduct{}).Where("id = ?", id).Updates(map[string]any{"name": req.Name, "company": req.Company, "cover_desc": req.CoverDesc, "price": price, "updated_at": time.Now()}).Error
	}
	return sc.DB.Create(&model.InsuranceProduct{ID: common.NewID(), Name: req.Name, Company: req.Company, CoverDesc: req.CoverDesc, Price: price, Status: 1}).Error
}

func AdminInsuranceApplyList(sc *svc.ServiceContext) ([]map[string]any, error) {
	var rows []map[string]any
	err := sc.DB.Table("insurance_apply a").Select("a.*, m.nickname").Joins("LEFT JOIN member m ON m.id = a.member_id").Order("a.id DESC").Limit(100).Scan(&rows).Error
	if rows == nil { rows = []map[string]any{} }
	return rows, err
}

func StudList(sc *svc.ServiceContext) ([]types.StudServiceView, error) {
	var rows []model.StudService
	if err := sc.DB.Where("status = 1").Order("id DESC").Limit(50).Find(&rows).Error; err != nil { return nil, err }
	out := make([]types.StudServiceView, 0, len(rows))
	for _, r := range rows {
		out = append(out, types.StudServiceView{ID: strconv.FormatInt(r.ID, 10), BreedName: r.BreedName, PetName: r.PetName, HealthCerts: r.HealthCerts, Price: r.Price.StringFixed(2), Description: r.Description})
	}
	return out, nil
}

func AdminStudUpsert(sc *svc.ServiceContext, req *types.StudUpsertReq) error {
	id, _ := strconv.ParseInt(req.ID, 10, 64)
	price, _ := decimal.NewFromString(req.Price)
	if id > 0 {
		return sc.DB.Model(&model.StudService{}).Where("id = ?", id).Updates(map[string]any{"breed_name": req.BreedName, "pet_name": req.PetName, "health_certs": req.HealthCerts, "price": price, "description": req.Description, "updated_at": time.Now()}).Error
	}
	return sc.DB.Create(&model.StudService{ID: common.NewID(), BreedName: req.BreedName, PetName: req.PetName, HealthCerts: req.HealthCerts, Price: price, Description: req.Description, Status: 1}).Error
}

func HomeConfigGet(sc *svc.ServiceContext) (string, error) {
	var hc model.HomeConfig
	if err := sc.DB.First(&hc, 1).Error; err != nil { return "[]", nil }
	return hc.Config, nil
}

func HomeConfigSet(sc *svc.ServiceContext, config string) error {
	return sc.DB.Exec("INSERT INTO home_config (id, config, updated_at) VALUES (1, ?, now()) ON CONFLICT (id) DO UPDATE SET config = ?, updated_at = now()", config, config).Error
}

func ScheduledOffSale(sc *svc.ServiceContext, productID int64, at *time.Time) error {
	res := sc.DB.Model(&model.PetProduct{}).Where("id = ?", productID).Update("scheduled_off_sale_at", at)
	if res.Error != nil { return res.Error }
	if res.RowsAffected == 0 { return common.ErrNotFound }
	return nil
}

// ScheduledOffSaleScan closer：到达定时下架时间的在售商品自动下架
func ScheduledOffSaleScan(sc *svc.ServiceContext) {
	sc.DB.Exec("UPDATE pet_product SET status = ?, updated_at = now() WHERE status = ? AND scheduled_off_sale_at IS NOT NULL AND scheduled_off_sale_at <= now()",
		model.ProductOffSale, model.ProductOnSale)
}

func InsuranceApplyCreate(sc *svc.ServiceContext, memberID int64, req *types.InsuranceApplyReq) error {
	pid, _ := strconv.ParseInt(req.ProductID, 10, 64)
	var p model.InsuranceProduct
	if err := sc.DB.First(&p, pid).Error; err != nil { return common.ErrParam }
	return sc.DB.Create(&model.InsuranceApply{ID: common.NewID(), MemberID: memberID, ProductID: pid, ProductName: p.Name, Contact: req.Contact, Phone: req.Phone}).Error
}

func DistributorApply(sc *svc.ServiceContext, memberID int64) error {
	var cnt int64
	sc.DB.Model(&model.Distributor{}).Where("member_id = ?", memberID).Count(&cnt)
	if cnt > 0 { return common.NewErr(400, 40001, "已是分销达人") }
	return sc.DB.Create(&model.Distributor{ID: common.NewID(), MemberID: memberID, Level: 1, CommissionRate: decimal.NewFromFloat(5)}).Error
}

func DistributorInfo(sc *svc.ServiceContext, memberID int64) (*model.Distributor, error) {
	var d model.Distributor
	if err := sc.DB.Where("member_id = ?", memberID).First(&d).Error; err != nil { return nil, common.ErrNotFound }
	return &d, nil
}

func DistributorWithdrawal(sc *svc.ServiceContext, memberID int64, amountStr string) error {
	amount, _ := decimal.NewFromString(amountStr)
	var d model.Distributor
	if err := sc.DB.Where("member_id = ? AND status = 1", memberID).First(&d).Error; err != nil { return common.ErrParam }
	if d.Balance.LessThan(amount) || amount.LessThanOrEqual(decimal.Zero) { return common.NewErr(400, 40001, "余额不足") }
	return sc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("UPDATE distributor SET balance = balance - ?, updated_at = now() WHERE id = ? AND balance >= ?", amount, d.ID, amount).Error; err != nil { return err }
		return tx.Create(&model.DistributorWithdrawal{ID: common.NewID(), DistributorID: d.ID, Amount: amount}).Error
	})
}

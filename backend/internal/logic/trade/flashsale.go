package trade

import (
	"gorm.io/gorm"

	"pet/backend/internal/model"
	"pet/backend/internal/svc"
)

// ActiveFlashSaleOf 商品启用中且在窗口内的秒杀（每商品至多一个启用活动，部分唯一索引保证）
func ActiveFlashSaleOf(sc *svc.ServiceContext, productID int64) *model.FlashSale {
	var fs model.FlashSale
	if err := sc.DB.
		Where("product_id = ? AND status = ? AND start_at <= now() AND end_at > now()",
			productID, model.FlashSaleOn).
		First(&fs).Error; err != nil {
		return nil
	}
	return &fs
}

// ClaimFlashSale 原子占名额（并发下 sold<stock 只允许一次成功）
func ClaimFlashSale(db *gorm.DB, id int64) bool {
	res := db.Exec(
		`UPDATE flash_sale SET sold = sold + 1, updated_at = now()
		 WHERE id = ? AND status = ? AND sold < stock AND start_at <= now() AND end_at > now()`,
		id, model.FlashSaleOn)
	return res.Error == nil && res.RowsAffected > 0
}

// ReleaseFlashSale 关单/失败路径回补名额
func ReleaseFlashSale(db *gorm.DB, id int64) {
	if id <= 0 {
		return
	}
	db.Exec(`UPDATE flash_sale SET sold = sold - 1, updated_at = now() WHERE id = ? AND sold > 0`, id)
}

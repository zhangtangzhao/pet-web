// compliance.go 合规与履约增强：门店自提 / 物流轨迹 / 免运费卡 / 协议校验。
package trade

import (
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/auth"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// StoreList 启用中的门店（公开，checkout 自提选择）
func StoreList(sc *svc.ServiceContext) ([]types.StoreView, error) {
	var stores []model.Store
	if err := sc.DB.Where("status = ?", 1).Order("sort ASC, id ASC").Limit(50).Find(&stores).Error; err != nil {
		return nil, err
	}
	list := make([]types.StoreView, 0, len(stores))
	for _, s := range stores {
		list = append(list, types.StoreView{
			ID:            strconv.FormatInt(s.ID, 10),
			Name:          s.Name,
			Address:       s.Address,
			Phone:         s.Phone,
			BusinessHours: s.BusinessHours,
			Status:        s.Status,
			Sort:          s.Sort,
			CreatedAt:     s.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return list, nil
}

// checkAgreement 下单协议校验
func checkAgreement(agree bool) error {
	if !agree {
		return common.ErrAgreementRequired
	}
	return nil
}

// AddTrace 管理端录入轨迹（预留第三方订阅适配点：provider 拉取后复用本表）
func AddTrace(sc *svc.ServiceContext, orderNo string, happenedAt *time.Time, desc, detail string) error {
	if desc == "" {
		return common.ErrParam
	}
	at := time.Now()
	if happenedAt != nil {
		at = *happenedAt
	}
	if len(desc) > 64 {
		desc = desc[:64]
	}
	return sc.DB.Create(&model.OrderTrace{
		ID:         common.NewID(),
		OrderNo:    orderNo,
		HappenedAt: at,
		StatusDesc: desc,
		Detail:     detail,
	}).Error
}

// resolvePickupStore 自提单：解析并校验门店，返回门店（同时作为地址快照来源）
func resolvePickupStore(sc *svc.ServiceContext, storeIDStr string) (*model.Store, error) {
	if storeIDStr == "" {
		return nil, common.NewErr(400, 40003, "请选择自提门店")
	}
	id, err := strconv.ParseInt(storeIDStr, 10, 64)
	if err != nil || id <= 0 {
		return nil, common.ErrParam
	}
	var store model.Store
	if err := sc.DB.Where("id = ? AND status = ?", id, 1).First(&store).Error; err != nil {
		return nil, common.NewErr(400, 40003, "自提门店不可用")
	}
	return &store, nil
}

// memberHasFreeShipCard 预检免运费卡（定价阶段）
func memberHasFreeShipCard(sc *svc.ServiceContext, memberID int64) bool {
	var cards int
	if err := sc.DB.Model(&model.Member{}).
		Select("free_ship_cards").Where("id = ?", memberID).
		Scan(&cards).Error; err != nil {
		return false
	}
	return cards > 0
}

// consumeFreeShipTx 事务内消耗一张免运费卡（并发安全：行数=0 视为卡已用完）
func consumeFreeShipTx(tx *gorm.DB, memberID int64) error {
	res := tx.Exec(
		"UPDATE member SET free_ship_cards = free_ship_cards - 1, updated_at = now() WHERE id = ? AND free_ship_cards > 0",
		memberID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.NewErr(400, 41407, "免运费卡已用完")
	}
	return nil
}

// freeShipApplies 是否可使用免运费卡（有卡 + 运费大于 0 + 用户勾选）
func freeShipApplies(sc *svc.ServiceContext, memberID int64, want bool, fee decimal.Decimal) bool {
	return want && fee.Sign() > 0 && memberHasFreeShipCard(sc, memberID)
}

// stampAgreement 订单协议快照
func stampAgreement(order *model.Order) {
	order.AgreementVersion = auth.AgreementVersion
	now := time.Now()
	order.AgreementSignedAt = &now
}

// resolveStoreAddress 门店自提地址快照
func resolveStoreAddress(s *model.Store) string {
	if s == nil {
		return ""
	}
	return s.Name + "（" + s.Address + "）"
}

func storeIDOf(s *model.Store) int64 {
	if s == nil {
		return 0
	}
	return s.ID
}

var _ = auth.AnonymizeDeleted

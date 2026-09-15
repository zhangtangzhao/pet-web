package trade

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/notify"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// ListMethods 用户端可用配送方式（启用，按 sort 升序）
func ListMethods(sc *svc.ServiceContext) ([]types.ShipMethodView, error) {
	var methods []model.ShipMethod
	if err := sc.DB.Where("status = ?", model.ShipMethodOn).
		Order("sort ASC, id ASC").Find(&methods).Error; err != nil {
		return nil, err
	}
	list := make([]types.ShipMethodView, 0, len(methods))
	for _, m := range methods {
		list = append(list, shipMethodView(m))
	}
	return list, nil
}

func shipMethodView(m model.ShipMethod) types.ShipMethodView {
	return types.ShipMethodView{
		ID:          strconv.FormatInt(m.ID, 10),
		Name:        m.Name,
		Kind:        m.Kind,
		Description: m.Description,
		Fee:         m.Fee.StringFixed(2),
		Sort:        m.Sort,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
	}
}

// AdminShip 管理端发货：已支付且待配送 → 配送中，登记托运单号
func AdminShip(sc *svc.ServiceContext, orderNo, shipNo string) error {
	shipNo = strings.TrimSpace(shipNo)
	if shipNo == "" {
		return common.NewErr(400, 40005, "请填写托运单号")
	}
	var o model.Order
	if err := sc.DB.Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNotFound
		}
		return err
	}
	// 自提单无需发货（配送方式已被删除的历史单放行）
	if o.ShipMethodID > 0 {
		var sm model.ShipMethod
		if err := sc.DB.First(&sm, o.ShipMethodID).Error; err == nil && sm.Kind == model.KindPickup {
			return common.NewErr(400, 40006, "自提订单无需发货")
		}
	}
	now := time.Now()
	res := sc.DB.Model(&model.Order{}).
		Where("id = ? AND status = ? AND ship_status = ?", o.ID, model.OrderPaid, model.ShipPending).
		Updates(map[string]any{
			"ship_status": model.ShipTransit,
			"ship_no":     shipNo,
			"shipped_at":  &now,
			"updated_at":  now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.NewErr(400, 40004, "当前状态不可发货")
	}
	if err := notify.Enqueue(sc, o.MemberID, model.NotifySceneOrder,
		"ship:"+o.OrderNo, "托运已发出", "您的宠物已发出，可查看托运进度", o.OrderNo); err != nil {
		_ = err // 通知失败不影响发货
	}
	return nil
}

// AdminDeliver 管理端标记送达：配送中 → 已送达
func AdminDeliver(sc *svc.ServiceContext, orderNo string) error {
	var o model.Order
	if err := sc.DB.Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNotFound
		}
		return err
	}
	now := time.Now()
	res := sc.DB.Model(&model.Order{}).
		Where("id = ? AND status = ? AND ship_status = ?", o.ID, model.OrderPaid, model.ShipTransit).
		Updates(map[string]any{
			"ship_status":  model.ShipDelivered,
			"delivered_at": &now,
			"updated_at":   now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.NewErr(400, 40004, "当前状态不可标记送达")
	}
	if err := notify.Enqueue(sc, o.MemberID, model.NotifySceneOrder,
		"deliver:"+o.OrderNo, "宠物已送达", "您的宠物已送达，请确认接宠", o.OrderNo); err != nil {
		_ = err
	}
	return nil
}

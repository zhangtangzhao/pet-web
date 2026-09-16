// Package trade 拼团：开团/凑团、支付后计数、超时未成团自动退款。
package trade

import (
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/notify"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// GroupBuys 进行中的拼团活动（含商品快照与进行中团数）
func GroupBuys(sc *svc.ServiceContext) (*types.GroupBuyListResp, error) {
	var gbs []model.GroupBuy
	if err := sc.DB.Where("status = ?", 1).Order("id DESC").Limit(20).Find(&gbs).Error; err != nil {
		return nil, err
	}
	if len(gbs) == 0 {
		return &types.GroupBuyListResp{List: []types.GroupBuyView{}}, nil
	}
	productIDs := make([]int64, 0, len(gbs))
	ids := make([]int64, 0, len(gbs))
	for _, g := range gbs {
		productIDs = append(productIDs, g.ProductID)
		ids = append(ids, g.ID)
	}
	var products []model.PetProduct
	if err := sc.DB.Select("id", "title", "main_image", "price").
		Where("id IN ?", productIDs).Find(&products).Error; err != nil {
		return nil, err
	}
	prodMap := map[int64]model.PetProduct{}
	for _, p := range products {
		prodMap[p.ID] = p
	}
	type teamCnt struct {
		GroupBuyID int64
		Cnt        int64
	}
	var cnts []teamCnt
	_ = sc.DB.Model(&model.GroupTeam{}).
		Select("group_buy_id AS group_buy_id, COUNT(*) AS cnt").
		Where("group_buy_id IN ? AND status = ?", ids, model.GroupTeamOpen).
		Group("group_buy_id").Scan(&cnts).Error
	cntMap := map[int64]int64{}
	for _, c := range cnts {
		cntMap[c.GroupBuyID] = c.Cnt
	}
	list := make([]types.GroupBuyView, 0, len(gbs))
	for _, g := range gbs {
		p := prodMap[g.ProductID]
		list = append(list, types.GroupBuyView{
			ID:           strconv.FormatInt(g.ID, 10),
			ProductID:    strconv.FormatInt(g.ProductID, 10),
			ProductTitle: p.Title,
			ProductImage: p.MainImage,
			Price:        g.Price.StringFixed(2),
			OrigPrice:    p.Price.StringFixed(2),
			Size:         g.Size,
			Hours:        g.Hours,
			OpenTeams:    int(cntMap[g.ID]),
			Status:       g.Status,
		})
	}
	return &types.GroupBuyListResp{List: list}, nil
}

// joinGroupTeam 事务内解析拼团：凑现有进行中的团，凑不上一人开新团。
// 行锁（FOR UPDATE）防并发超员；同一会员不重复进同一团。
func joinGroupTeam(tx *gorm.DB, gb *model.GroupBuy, memberID int64) (int64, error) {
	now := time.Now()
	hours := gb.Hours
	if hours <= 0 {
		hours = 24
	}
	var teams []model.GroupTeam
	if err := tx.Raw(
		`SELECT * FROM group_team WHERE group_buy_id = ? AND status = ? AND expire_at > now()
		 AND member_count < ? ORDER BY id ASC LIMIT 5 FOR UPDATE`,
		gb.ID, model.GroupTeamOpen, gb.Size).Scan(&teams).Error; err != nil {
		return 0, err
	}
	for _, t := range teams {
		var cnt int64
		if err := tx.Model(&model.Order{}).
			Where("group_team_id = ? AND member_id = ? AND status <> ?", t.ID, memberID, model.OrderCanceled).
			Count(&cnt).Error; err != nil {
			return 0, err
		}
		if cnt > 0 {
			continue
		}
		return t.ID, nil
	}
	team := model.GroupTeam{
		ID:             common.NewID(),
		GroupBuyID:     gb.ID,
		LeaderMemberID: memberID,
		MemberCount:    0,
		Status:         model.GroupTeamOpen,
		ExpireAt:       now.Add(time.Duration(hours) * time.Hour),
	}
	if err := tx.Create(&team).Error; err != nil {
		return 0, err
	}
	return team.ID, nil
}

// groupTeamFilledTx 支付成功后计数：凑满则成团（事务内调用）
func groupTeamFilledTx(tx *gorm.DB, teamID int64) (bool, error) {
	res := tx.Exec(`
		UPDATE group_team t
		SET member_count = t.member_count + 1,
		    status = CASE WHEN t.member_count + 1 >= (SELECT g.size FROM group_buy g WHERE g.id = t.group_buy_id)
		                  THEN ? ELSE t.status END,
		    updated_at = now()
		WHERE t.id = ? AND t.status = ?`,
		model.GroupTeamOK, teamID, model.GroupTeamOpen)
	if res.Error != nil {
		return false, res.Error
	}
	var t model.GroupTeam
	if err := tx.First(&t, teamID).Error; err != nil {
		return false, err
	}
	return t.Status == model.GroupTeamOK, nil
}

// notifyGroupOK 成团后通知团员（biz_key 幂等）
func notifyGroupOK(sc *svc.ServiceContext, teamID int64) {
	var orders []model.Order
	if err := sc.DB.Select("member_id", "order_no").
		Where("group_team_id = ? AND status = ?", teamID, model.OrderPaid).
		Find(&orders).Error; err != nil {
		return
	}
	for _, o := range orders {
		_ = notify.Enqueue(sc, o.MemberID, model.NotifySceneOrder,
			"groupok:"+strconv.FormatInt(teamID, 10)+":"+strconv.FormatInt(o.MemberID, 10),
			"拼团成功", "拼团已成团，宝贝尽快安排发货", o.OrderNo)
	}
}

// refundGroupOrder 团失败：已支付拼团单整单退款（DEMO 退款流水）
func refundGroupOrder(sc *svc.ServiceContext, o *model.Order) error {
	now := time.Now()
	return sc.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Order{}).
			Where("id = ? AND status = ?", o.ID, model.OrderPaid).
			Updates(map[string]any{
				"status":        model.OrderRefunded,
				"canceled_at":   &now,
				"cancel_reason": "拼团超时未成团，已原路退款",
				"updated_at":    now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		var pay model.Payment
		err := tx.Where("order_no = ? AND pay_type = ? AND status = ?",
			o.OrderNo, model.PayTypePurchase, model.PayStatusSuccess).First(&pay).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewErr(400, 41202, "未找到成功的支付流水")
		}
		if err != nil {
			return err
		}
		if err := tx.Create(&model.Payment{
			ID:            common.NewID(),
			PaymentNo:     common.NewBizNo("RF"),
			OrderID:       o.ID,
			OrderNo:       o.OrderNo,
			MemberID:      o.MemberID,
			Amount:        pay.Amount,
			Channel:       pay.Channel,
			PayType:       model.PayTypeRefund,
			TransactionID: "DEMO",
			Status:        model.PayStatusRefund,
			CallbackAt:    &now,
		}).Error; err != nil {
			return err
		}
		return tx.Exec(
			"UPDATE pet_product SET status = ?, updated_at = now() WHERE id IN (SELECT product_id FROM order_item WHERE order_id = ?) AND status = ?",
			model.ProductOnSale, o.ID, model.ProductSold).Error
	})
}

// ExpireGroups 扫描超时未成团的团：置失败 + 已支付订单自动退款
func ExpireGroups(sc *svc.ServiceContext) {
	var teams []model.GroupTeam
	if err := sc.DB.Where("status = ? AND expire_at < now()", model.GroupTeamOpen).
		Limit(100).Find(&teams).Error; err != nil {
		return
	}
	for i := range teams {
		res := sc.DB.Model(&model.GroupTeam{}).
			Where("id = ? AND status = ?", teams[i].ID, model.GroupTeamOpen).
			Update("status", model.GroupTeamFailed)
		if res.Error != nil || res.RowsAffected == 0 {
			continue
		}
		var orders []model.Order
		if err := sc.DB.Where("group_team_id = ? AND status = ?", teams[i].ID, model.OrderPaid).
			Find(&orders).Error; err != nil {
			continue
		}
		for j := range orders {
			if err := refundGroupOrder(sc, &orders[j]); err != nil {
				continue
			}
			_ = notify.Enqueue(sc, orders[j].MemberID, model.NotifySceneOrder,
				"groupfail:"+orders[j].OrderNo, "拼团未成团",
				"拼团超时未成团，货款将原路退回", orders[j].OrderNo)
		}
	}
}

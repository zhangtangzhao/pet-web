// play.go 玩法套件：砍价 / 竞拍 / 服务预约。
package marketing

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/notify"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

var bookingSlots = []string{"09:00-11:00", "11:00-13:00", "13:00-15:00", "15:00-17:00"}

// ─────────────────────────── 平台端 ───────────────────────────

func AdminBargainUpsert(sc *svc.ServiceContext, req *types.BargainActivityUpsert) (any, error) {
	productID, err := strconv.ParseInt(req.ProductID, 10, 64)
	if err != nil {
		return nil, common.ErrParam
	}
	hours := req.DurationHours
	if hours <= 0 {
		hours = 24
	}
	helpers := req.MaxHelpers
	if helpers <= 0 {
		helpers = 5
	}
	if req.ID != "" {
		id, _ := strconv.ParseInt(req.ID, 10, 64)
		return nil, sc.DB.Exec("UPDATE bargain_activity SET product_id=?, bottom_price=?, duration_hours=?, max_helpers=?, updated_at=now() WHERE id=?", productID, req.BottomPrice, hours, helpers, id).Error
	}
	return nil, sc.DB.Exec("INSERT INTO bargain_activity (id, product_id, bottom_price, duration_hours, max_helpers) VALUES (?,?,?,?,?)", common.NewID(), productID, req.BottomPrice, hours, helpers).Error
}

func AdminBargainList(sc *svc.ServiceContext) (any, error) {
	var rows []map[string]any
	err := sc.DB.Table("bargain_activity a").
		Select("a.id, a.product_id, p.title AS product_title, p.price, a.bottom_price, a.duration_hours, a.max_helpers, a.status").
		Joins("LEFT JOIN pet_product p ON p.id = a.product_id").
		Order("a.id DESC").Limit(100).Scan(&rows).Error
	if rows == nil {
		rows = []map[string]any{}
	}
	return rows, err
}

func AdminAuctionUpsert(sc *svc.ServiceContext, req *types.AuctionUpsertReq) (any, error) {
	productID, err := strconv.ParseInt(req.ProductID, 10, 64)
	if err != nil {
		return nil, common.ErrParam
	}
	start, e1 := time.Parse("2006-01-02 15:04:05", req.StartAt)
	end, e2 := time.Parse("2006-01-02 15:04:05", req.EndAt)
	if e1 != nil || e2 != nil {
		return nil, common.ErrParam
	}
	if req.ID != "" {
		id, _ := strconv.ParseInt(req.ID, 10, 64)
		return nil, sc.DB.Exec("UPDATE auction SET product_id=?, start_price=?, step_price=?, deposit_amount=?, start_at=?, end_at=?, updated_at=now() WHERE id=?", productID, req.StartPrice, req.StepPrice, req.DepositAmount, start, end, id).Error
	}
	return nil, sc.DB.Exec("INSERT INTO auction (id, product_id, start_price, step_price, deposit_amount, start_at, end_at, status) VALUES (?,?,?,?,?,?,?,1)", common.NewID(), productID, req.StartPrice, req.StepPrice, req.DepositAmount, start, end).Error
}

func AdminAuctionList(sc *svc.ServiceContext) (any, error) {
	var rows []map[string]any
	err := sc.DB.Table("auction a").
		Select("a.*, p.title AS product_title").
		Joins("LEFT JOIN pet_product p ON p.id = a.product_id").
		Order("a.id DESC").Limit(100).Scan(&rows).Error
	if rows == nil {
		rows = []map[string]any{}
	}
	return rows, err
}

// ─────────────────────────── 砍价 ───────────────────────────

func bargProduct(sc *svc.ServiceContext, productID int64) model.PetProduct {
	var p model.PetProduct
	_ = sc.DB.Select("id", "title", "main_image", "price", "status").First(&p, productID).Error
	return p
}

// BargainLaunches 我发起的砍价
func BargainLaunches(sc *svc.ServiceContext, memberID int64) ([]types.BargainLaunchView, error) {
	var launches []model.BargainLaunch
	if err := sc.DB.Where("member_id = ?", memberID).Order("id DESC").Limit(20).Find(&launches).Error; err != nil {
		return nil, err
	}
	out := make([]types.BargainLaunchView, 0, len(launches))
	for _, l := range launches {
		var act model.BargainActivity
		_ = sc.DB.First(&act, l.ActivityID).Error
		p := bargProduct(sc, act.ProductID)
		out = append(out, types.BargainLaunchView{
			ID: strconv.FormatInt(l.ID, 10), ActivityID: strconv.FormatInt(l.ActivityID, 10),
			ProductTitle: p.Title, ProductImage: p.MainImage,
			OriginPrice: p.Price.StringFixed(2), CurrentPrice: l.CurrentPrice.StringFixed(2),
			BottomPrice: act.BottomPrice.StringFixed(2),
			HelperCount: l.HelperCount, MaxHelpers: act.MaxHelpers,
			Status: l.Status, StatusText: bargainStatusText(l.Status),
			ExpireAt: l.ExpireAt.Format("2006-01-02 15:04"),
		})
	}
	return out, nil
}

func bargainStatusText(s int) string {
	switch s {
	case model.BargainRunning:
		return "砍价中"
	case model.BargainReady:
		return "已到底价"
	case model.BargainBought:
		return "已购买"
	case model.BargainExpired:
		return "已过期"
	}
	return "未知"
}

// BargainLaunch 发起砍价
func BargainLaunch(sc *svc.ServiceContext, memberID int64, activityID int64) ([]types.BargainLaunchView, error) {
	var act model.BargainActivity
	if err := sc.DB.Where("id = ? AND status = 1", activityID).First(&act).Error; err != nil {
		return nil, common.ErrBargainInvalid
	}
	p := bargProduct(sc, act.ProductID)
	if p.Status != model.ProductOnSale {
		return nil, common.ErrGoodsOffline
	}
	var cnt int64
	sc.DB.Model(&model.BargainLaunch{}).
		Where("member_id = ? AND activity_id = ? AND status IN ?", memberID, activityID,
			[]int{model.BargainRunning, model.BargainReady}).Count(&cnt)
	if cnt > 0 {
		return nil, common.NewErr(400, 42101, "该活动你已发起过砍价")
	}
	hours := act.DurationHours
	if hours <= 0 {
		hours = 24
	}
	l := model.BargainLaunch{
		ID: common.NewID(), ActivityID: activityID, MemberID: memberID,
		CurrentPrice: p.Price, Status: model.BargainRunning,
		ExpireAt: time.Now().Add(time.Duration(hours) * time.Hour),
	}
	if err := sc.DB.Create(&l).Error; err != nil {
		return nil, err
	}
	_ = notify.Enqueue(sc, memberID, model.NotifySceneOrder,
		"bargainlaunch:"+strconv.FormatInt(l.ID, 10), "砍价已发起",
		"邀请好友助力，砍到底价即可购买", "")
	return BargainLaunches(sc, memberID)
}

// BargainHelp 好友助力：随机砍一刀，到底则 status=1
func BargainHelp(sc *svc.ServiceContext, helperMemberID int64, launchID int64) ([]types.BargainLaunchView, error) {
	err := sc.DB.Transaction(func(tx *gorm.DB) error {
		var l model.BargainLaunch
		if err := tx.Raw(`SELECT * FROM bargain_launch WHERE id = ? FOR UPDATE`, launchID).
			Scan(&l).Error; err != nil || l.ID == 0 {
			return common.ErrBargainInvalid
		}
		if l.Status != model.BargainRunning || time.Now().After(l.ExpireAt) {
			return common.ErrBargainInvalid
		}
		var act model.BargainActivity
		if err := tx.First(&act, l.ActivityID).Error; err != nil {
			return common.ErrBargainInvalid
		}
		if helperMemberID == l.MemberID {
			return common.NewErr(400, 42101, "不能给自己助力")
		}
		var cnt int64
		tx.Model(&model.BargainHelp{}).Where("launch_id = ? AND helper_member_id = ?", launchID, helperMemberID).Count(&cnt)
		if cnt > 0 {
			return common.NewErr(400, 42101, "你已助力过该砍价")
		}
		// 本刀金额 = 剩余空间 / 剩余刀数（±20% 抖动），最后一刀直接到底
		remaining := l.CurrentPrice.Sub(act.BottomPrice)
		leftKnives := act.MaxHelpers - l.HelperCount
		if leftKnives <= 0 {
			return common.ErrBargainInvalid
		}
		amount := remaining.Div(decimal.NewFromInt(int64(leftKnives)))
		if leftKnives > 1 {
			jitter := amount.Mul(decimal.NewFromFloat(0.6 + rand.Float64()*0.8))
			if jitter.GreaterThan(remaining) {
				jitter = remaining
			}
			amount = jitter
		} else {
			amount = remaining
		}
		if err := tx.Create(&model.BargainHelp{
			ID: common.NewID(), LaunchID: launchID, HelperMemberID: helperMemberID, Amount: amount,
		}).Error; err != nil {
			return err
		}
		newPrice := l.CurrentPrice.Sub(amount)
		if newPrice.LessThan(act.BottomPrice) {
			newPrice = act.BottomPrice
		}
		newCount := l.HelperCount + 1
		status := model.BargainRunning
		if newCount >= act.MaxHelpers || newPrice.Equal(act.BottomPrice) {
			status = model.BargainReady
		}
		return tx.Model(&model.BargainLaunch{}).Where("id = ?", launchID).
			Updates(map[string]any{
				"current_price": newPrice, "helper_count": newCount, "status": status, "updated_at": time.Now(),
			}).Error
	})
	if err != nil {
		return nil, err
	}
	return BargainLaunches(sc, l_ownerOf(sc, launchID))
}

func l_ownerOf(sc *svc.ServiceContext, launchID int64) int64 {
	var l model.BargainLaunch
	_ = sc.DB.Select("member_id").First(&l, launchID).Error
	return l.MemberID
}

// ConsumeBargainTx 下单消耗砍价资格（事务内 CAS）
func ConsumeBargainTx(tx *gorm.DB, memberID, launchID int64) (decimal.Decimal, error) {
	var l model.BargainLaunch
	if err := tx.Raw(`SELECT * FROM bargain_launch WHERE id = ? FOR UPDATE`, launchID).Scan(&l).Error; err != nil || l.ID == 0 {
		return decimal.Zero, common.ErrBargainInvalid
	}
	if l.MemberID != memberID || l.Status != model.BargainReady || time.Now().After(l.ExpireAt) {
		return decimal.Zero, common.ErrBargainInvalid
	}
	if err := tx.Model(&model.BargainLaunch{}).Where("id = ?", launchID).
		Update("status", model.BargainBought).Error; err != nil {
		return decimal.Zero, err
	}
	return l.CurrentPrice, nil
}

// ExpireBargains closer：过期砍价置失效
func ExpireBargains(sc *svc.ServiceContext) {
	sc.DB.Model(&model.BargainLaunch{}).
		Where("status IN ? AND expire_at < now()", []int{model.BargainRunning, model.BargainReady}).
		Updates(map[string]any{"status": model.BargainExpired, "updated_at": time.Now()})
}

// ─────────────────────────── 竞拍 ───────────────────────────

func auctionStatusText(s int) string {
	switch s {
	case model.AuctionNotStarted:
		return "未开始"
	case model.AuctionRunning:
		return "进行中"
	case model.AuctionSold:
		return "已成交"
	case model.AuctionFailed:
		return "已流拍"
	}
	return "未知"
}

// Auctions 竞拍列表（公开，含我的保证金状态）
func Auctions(sc *svc.ServiceContext, memberID int64) ([]types.AuctionView, error) {
	var auctions []model.Auction
	if err := sc.DB.Order("id DESC").Limit(20).Find(&auctions).Error; err != nil {
		return nil, err
	}
	depositMap := map[int64]int{}
	if memberID > 0 {
		var deps []model.AuctionDeposit
		sc.DB.Where("member_id = ?", memberID).Find(&deps)
		for _, d := range deps {
			if d.Status == model.AuctionDepositPaid {
				depositMap[d.AuctionID] = 1
			}
		}
	}
	out := make([]types.AuctionView, 0, len(auctions))
	for _, a := range auctions {
		p := bargProduct(sc, a.ProductID)
		highest := ""
		if a.HighestMemberID > 0 {
			highest = "…"
		}
		out = append(out, types.AuctionView{
			ID: strconv.FormatInt(a.ID, 10), ProductID: strconv.FormatInt(a.ProductID, 10),
			ProductTitle: p.Title, ProductImage: p.MainImage,
			StartPrice: a.StartPrice.StringFixed(2), StepPrice: a.StepPrice.StringFixed(2),
			DepositAmount: a.DepositAmount.StringFixed(2),
			StartAt:       a.StartAt.Format("01-02 15:04"), EndAt: a.EndAt.Format("01-02 15:04"),
			Status: a.Status, StatusText: auctionStatusText(a.Status),
			HighestPrice: a.HighestPrice.StringFixed(2), HighestMember: highest,
			DepositPaid: depositMap[a.ID] == 1,
		})
	}
	return out, nil
}

// AuctionDepositPay 交保证金（创建流水；dev 直接落成功）
func AuctionDepositPay(sc *svc.ServiceContext, memberID, auctionID int64) (*types.VipBuyResp, error) {
	var a model.Auction
	if err := sc.DB.Where("id = ? AND status = ? AND end_at > now()", auctionID, model.AuctionRunning).
		First(&a).Error; err != nil {
		return nil, common.ErrAuctionInvalid
	}
	if a.DepositAmount.LessThanOrEqual(decimal.Zero) {
		return nil, common.NewErr(400, 42102, "该竞拍无需保证金")
	}
	var dep model.AuctionDeposit
	err := sc.DB.Where("auction_id = ? AND member_id = ?", auctionID, memberID).First(&dep).Error
	if err == nil {
		if dep.Status == model.AuctionDepositPaid {
			return nil, common.NewErr(400, 42102, "你已缴纳保证金")
		}
	} else {
		dep = model.AuctionDeposit{ID: common.NewID(), AuctionID: auctionID, MemberID: memberID}
		if err := sc.DB.Create(&dep).Error; err != nil {
			return nil, err
		}
	}
	pay := model.Payment{
		ID: common.NewID(), PaymentNo: common.NewBizNo("PAY"),
		OrderNo: "AU" + strconv.FormatInt(auctionID, 10), MemberID: memberID,
		Amount: a.DepositAmount, Channel: model.PayChannelMini,
		PayType: model.PayTypeAuctionDeposit, Status: model.PayStatusPending,
	}
	if err := sc.DB.Create(&pay).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	if err := sc.DB.Model(&model.Payment{}).Where("payment_no = ?", pay.PaymentNo).
		Updates(map[string]any{"status": model.PayStatusSuccess, "callback_at": &now}).Error; err != nil {
		return nil, err
	}
	if err := sc.DB.Model(&model.AuctionDeposit{}).Where("id = ?", dep.ID).
		Updates(map[string]any{"status": model.AuctionDepositPaid, "payment_no": pay.PaymentNo, "updated_at": now}).Error; err != nil {
		return nil, err
	}
	return &types.VipBuyResp{PaymentNo: pay.PaymentNo, Price: a.DepositAmount.StringFixed(2)}, nil
}

// AuctionBid 出价（保证金已付 + 进行中 + 大于当前价+步长）
func AuctionBid(sc *svc.ServiceContext, memberID int64, auctionID int64, price decimal.Decimal) error {
	var dep model.AuctionDeposit
	if err := sc.DB.Where("auction_id = ? AND member_id = ? AND status = ?",
		auctionID, memberID, model.AuctionDepositPaid).First(&dep).Error; err != nil {
		return common.ErrAuctionInvalid
	}
	var a model.Auction
	if err := sc.DB.First(&a, auctionID).Error; err != nil {
		return common.ErrAuctionInvalid
	}
	if a.Status != model.AuctionRunning || time.Now().After(a.EndAt) || time.Now().Before(a.StartAt) {
		return common.ErrAuctionInvalid
	}
	minPrice := a.StartPrice
	if a.HighestMemberID > 0 {
		minPrice = a.HighestPrice
	}
	if price.LessThan(minPrice.Add(a.StepPrice)) {
		return common.NewErr(400, 42102, "出价需不低于当前价加一个加价幅度")
	}
	res := sc.DB.Exec(
		"UPDATE auction SET highest_member_id = ?, highest_price = ?, updated_at = now() WHERE id = ? AND status = ? AND highest_price < ?",
		memberID, price, auctionID, model.AuctionRunning, price)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.NewErr(400, 42102, "出价已被人抢先，请加价后重试")
	}
	return nil
}

// SettleAuctions closer：截止 → 成交生成待付订单 / 流拍退保证金
func SettleAuctions(sc *svc.ServiceContext) {
	var auctions []model.Auction
	if err := sc.DB.Where("status = ? AND end_at < now()", model.AuctionRunning).
		Limit(20).Find(&auctions).Error; err != nil {
		return
	}
	now := time.Now()
	for _, a := range auctions {
		res := sc.DB.Model(&model.Auction{}).
			Where("id = ? AND status = ?", a.ID, model.AuctionRunning).
			Update("status", model.AuctionFailed)
		if res.Error != nil || res.RowsAffected == 0 {
			continue
		}
		var deps []model.AuctionDeposit
		sc.DB.Where("auction_id = ? AND status = ?", a.ID, model.AuctionDepositPaid).Find(&deps)
		if a.HighestMemberID > 0 {
			sc.DB.Model(&model.Auction{}).Where("id = ?", a.ID).Update("status", model.AuctionSold)
			// 成交：生成待支付订单（原价=最高价），保证金全部原路退回（DEMO）
			order := model.Order{
				ID: common.NewID(), OrderNo: common.NewBizNo("P"),
				MemberID: a.HighestMemberID, TotalAmount: a.HighestPrice,
				PayAmount: a.HighestPrice, Status: model.OrderPending,
				ContactName: "竞拍得主", ContactPhone: "", Remark: "竞拍成交订单",
				ExpireAt: now.Add(24 * time.Hour),
			}
			p := bargProduct(sc, a.ProductID)
			if err := sc.DB.Transaction(func(tx *gorm.DB) error {
				if err := tx.Create(&order).Error; err != nil {
					return err
				}
				return tx.Create(&model.OrderItem{
					ID: common.NewID(), OrderID: order.ID, ProductID: a.ProductID,
					ProductTitle: p.Title, ProductImage: p.MainImage,
					Price: a.HighestPrice, Quantity: 1,
				}).Error
			}); err == nil {
				_ = notify.Enqueue(sc, a.HighestMemberID, model.NotifySceneOrder,
					"auctionwin:"+strconv.FormatInt(a.ID, 10), "竞拍成功",
					"恭喜以 "+a.HighestPrice.StringFixed(2)+" 元拍得 "+p.Title+"，请尽快支付", order.OrderNo)
			}
		} else {
			_ = sc.DB.Model(&model.Auction{}).Where("id = ?", a.ID).Update("status", model.AuctionFailed)
		}
		for _, d := range deps {
			sc.DB.Model(&model.AuctionDeposit{}).Where("id = ?", d.ID).
				Updates(map[string]any{"status": model.AuctionDepositRefund, "updated_at": now})
			if d.MemberID != a.HighestMemberID || a.HighestMemberID == 0 {
				_ = notify.Enqueue(sc, d.MemberID, model.NotifySceneOrder,
					"auctiondeposit:"+strconv.FormatInt(d.ID, 10), "保证金退回",
					"竞拍已结束，你的保证金将原路退回", "")
			}
		}
	}
}

// ─────────────────────────── 服务预约 ───────────────────────────

// BookingSlots 可约时段
func BookingSlots() []string { return bookingSlots }

// CreateBooking 创建服务预约（到店付，生成核销码）
func CreateBooking(sc *svc.ServiceContext, memberID int64, req *types.BookingReq) (*types.BookingView, error) {
	date, err := time.ParseInLocation("2006-01-02", req.Date, time.Local)
	if err != nil || date.Before(time.Now().AddDate(0, 0, -1)) {
		return nil, common.ErrBookingInvalid
	}
	slotOK := false
	for _, s := range bookingSlots {
		if s == req.Slot {
			slotOK = true
			break
		}
	}
	if !slotOK {
		return nil, common.ErrBookingInvalid
	}
	serviceID, _ := strconv.ParseInt(req.ServiceID, 10, 64)
	var svcItem model.ServiceItem
	if err := sc.DB.Where("id = ? AND status = 1", serviceID).First(&svcItem).Error; err != nil {
		return nil, common.ErrBookingInvalid
	}
	storeID, _ := strconv.ParseInt(req.StoreID, 10, 64)
	var store model.Store
	storeName := "平台服务"
	if err := sc.DB.First(&store, storeID).Error; err == nil {
		storeName = store.Name
	}
	booking := model.ServiceBooking{
		ID: common.NewID(), BookingNo: common.NewBizNo("BK"), MemberID: memberID,
		ServiceID: serviceID, ServiceName: svcItem.Name, StoreID: storeID,
		BookingDate: date, TimeSlot: req.Slot,
		Contact: strings.TrimSpace(req.Contact), Phone: strings.TrimSpace(req.Phone),
		PetName: strings.TrimSpace(req.PetName), Price: svcItem.Price,
		Status: model.BookingPending, VerifyCode: fmt.Sprintf("%06d", time.Now().UnixNano()%1000000),
	}
	if err := sc.DB.Create(&booking).Error; err != nil {
		return nil, err
	}
	return &types.BookingView{
		BookingNo: booking.BookingNo, ServiceName: booking.ServiceName, StoreName: storeName,
		Date: req.Date, Slot: req.Slot, Contact: booking.Contact, Phone: booking.Phone,
		PetName: booking.PetName, Price: booking.Price.StringFixed(2),
		Status: booking.Status, StatusText: "待到店", VerifyCode: booking.VerifyCode,
	}, nil
}

func bookingView(sc *svc.ServiceContext, b model.ServiceBooking) types.BookingView {
	storeName := ""
	if b.StoreID > 0 {
		var s model.Store
		_ = sc.DB.Select("name").First(&s, b.StoreID).Error
		if s.ID > 0 {
			storeName = s.Name
		}
	}
	text := "待到店"
	switch b.Status {
	case model.BookingDone:
		text = "已完成"
	case model.BookingCanceled:
		text = "已取消"
	}
	return types.BookingView{
		BookingNo: b.BookingNo, ServiceName: b.ServiceName, StoreName: storeName,
		Date: b.BookingDate.Format("2006-01-02"), Slot: b.TimeSlot,
		Contact: b.Contact, Phone: b.Phone, PetName: b.PetName,
		Price: b.Price.StringFixed(2), Status: b.Status, StatusText: text,
		VerifyCode: b.VerifyCode,
	}
}

// MyBookings 我的预约
func MyBookings(sc *svc.ServiceContext, memberID int64) ([]types.BookingView, error) {
	var rows []model.ServiceBooking
	if err := sc.DB.Where("member_id = ?", memberID).Order("id DESC").Limit(50).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]types.BookingView, 0, len(rows))
	for _, b := range rows {
		out = append(out, bookingView(sc, b))
	}
	return out, nil
}

// CancelBooking 取消预约
func CancelBooking(sc *svc.ServiceContext, memberID int64, bookingNo string) error {
	res := sc.DB.Model(&model.ServiceBooking{}).
		Where("booking_no = ? AND member_id = ? AND status = ?", bookingNo, memberID, model.BookingPending).
		Update("status", model.BookingCanceled)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrBookingInvalid
	}
	return nil
}

// AdminBookings 平台端预约列表
func AdminBookings(sc *svc.ServiceContext, req *types.PageReq) ([]types.BookingView, error) {
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	var rows []model.ServiceBooking
	if err := sc.DB.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]types.BookingView, 0, len(rows))
	for _, b := range rows {
		out = append(out, bookingView(sc, b))
	}
	return out, nil
}

// VerifyBooking 核销
func VerifyBooking(sc *svc.ServiceContext, bookingNo, code string) error {
	res := sc.DB.Model(&model.ServiceBooking{}).
		Where("booking_no = ? AND verify_code = ? AND status = ?", bookingNo, code, model.BookingPending).
		Update("status", model.BookingDone)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrBookingInvalid
	}
	return nil
}

// ConsumeBargain 校验并占用砍价到底价的购买资格（CAS 保证只买一次）
func ConsumeBargain(sc *svc.ServiceContext, memberID, launchID int64) (decimal.Decimal, error) {
	var l model.BargainLaunch
	if err := sc.DB.First(&l, launchID).Error; err != nil {
		return decimal.Zero, common.ErrBargainInvalid
	}
	if l.MemberID != memberID || l.Status != model.BargainReady || time.Now().After(l.ExpireAt) {
		return decimal.Zero, common.ErrBargainInvalid
	}
	res := sc.DB.Exec(
		"UPDATE bargain_launch SET status = ?, updated_at = now() WHERE id = ? AND status = ?",
		model.BargainBought, launchID, model.BargainReady)
	if res.Error != nil {
		return decimal.Zero, res.Error
	}
	if res.RowsAffected == 0 {
		return decimal.Zero, common.ErrBargainInvalid
	}
	return l.CurrentPrice, nil
}

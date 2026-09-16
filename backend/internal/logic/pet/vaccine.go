// vaccine.go 疫苗/驱虫到期提醒：每日扫描 pet_product.next_vaccine_date / next_deworm_date，
// 3 天内到期 → 给该商品最近一笔已支付订单的买家发通知（biz_key 幂等，同一到期日只发一次）。
package pet

import (
	"context"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"pet/backend/internal/logic/notify"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
)

const careWindowDays = 3

// StartVaccineReminders 启动护理提醒任务：启动即扫一次，之后每天一轮
func StartVaccineReminders(ctx context.Context, sc *svc.ServiceContext) {
	scanCareReminders(sc)
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			scanCareReminders(sc)
		}
	}
}

type careBuyer struct {
	MemberID int64
	OrderNo  string
}

func scanCareReminders(sc *svc.ServiceContext) {
	today := time.Now()
	to := today.AddDate(0, 0, careWindowDays)
	scanKind(sc, "next_vaccine_date", "vaccine:v:", "疫苗提醒", today, to)
	scanKind(sc, "next_deworm_date", "deworm:w:", "驱虫提醒", today, to)
}

func scanKind(sc *svc.ServiceContext, col, keyPrefix, title string, today, to time.Time) {
	var products []model.PetProduct
	if err := sc.DB.Select("id", "title", col).
		Where(col+" IS NOT NULL AND "+col+" BETWEEN ? AND ?", today.Format("2006-01-02"), to.Format("2006-01-02")).
		Find(&products).Error; err != nil {
		logx.Errorf("care: 扫描 %s 失败: %v", col, err)
		return
	}
	for _, p := range products {
		var due *time.Time
		if col == "next_vaccine_date" {
			due = p.NextVaccineDate
		} else {
			due = p.NextDewormDate
		}
		if due == nil {
			continue
		}
		var buyers []careBuyer
		if err := sc.DB.Model(&model.OrderItem{}).
			Select("orders.member_id AS member_id", "orders.order_no AS order_no").
			Joins("JOIN orders ON orders.id = order_item.order_id").
			Where("order_item.product_id = ? AND orders.status IN ?", p.ID,
				[]int{model.OrderPaid, model.OrderCompleted}).
			Order("orders.id DESC").Limit(1).Scan(&buyers).Error; err != nil || len(buyers) == 0 {
			continue
		}
		dateStr := due.Format("01月02日")
		bizKey := keyPrefix + strconv.FormatInt(p.ID, 10) + ":" + due.Format("2006-01-02")
		if err := notify.Enqueue(sc, buyers[0].MemberID, model.NotifySceneCare,
			bizKey, title, p.Title+" "+dateStr+" 到期", buyers[0].OrderNo); err != nil {
			logx.Errorf("care: 入队失败 bizKey=%s: %v", bizKey, err)
		}
	}
}

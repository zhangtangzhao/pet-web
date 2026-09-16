// export.go 平台端数据导出：订单 / 会员 / 积分流水 → CSV（UTF-8 BOM，Excel 直开），单次上限 2 万行。
package manage

import (
	"time"

	"pet/backend/internal/logic/growth"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
)

const exportRowLimit = 20000

func csvTime(t time.Time) string { return t.In(time.Local).Format("2006-01-02 15:04:05") }
func csvTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return csvTime(*t)
}

// ExportOrders 订单导出
func ExportOrders(sc *svc.ServiceContext) ([][]string, error) {
	var rows []model.Order
	if err := sc.DB.Order("id DESC").Limit(exportRowLimit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := [][]string{{"订单号", "会员ID", "商品金额", "优惠金额", "等级优惠", "服务费", "运费", "实付金额", "状态", "支付时间", "创建时间"}}
	for _, o := range rows {
		out = append(out, []string{
			o.OrderNo,
			strconvI64(o.MemberID),
			o.TotalAmount.StringFixed(2),
			o.DiscountAmount.StringFixed(2),
			o.LevelDiscount.StringFixed(2),
			o.ServiceFee.StringFixed(2),
			o.ShipFee.StringFixed(2),
			o.PayAmount.StringFixed(2),
			model.OrderStatusText(o.Status),
			csvTimePtr(o.PaidAt),
			csvTime(o.CreatedAt),
		})
	}
	return out, nil
}

// ExportMembers 会员导出
func ExportMembers(sc *svc.ServiceContext) ([][]string, error) {
	var rows []model.Member
	if err := sc.DB.Order("id DESC").Limit(exportRowLimit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := [][]string{{"会员ID", "昵称", "手机号", "积分", "成长值", "等级", "邀请码", "邀请人ID", "状态", "注册时间"}}
	for _, m := range rows {
		out = append(out, []string{
			strconvI64(m.ID),
			m.Nickname,
			m.Phone,
			strconvI64(m.Points),
			strconvI64(m.GrowthValue),
			growth.LevelName(growth.LevelOf(m.GrowthValue)),
			m.InviteCode,
			strconvI64(m.InvitedBy),
			map[int]string{0: "停用", 1: "正常"}[m.Status],
			csvTime(m.CreatedAt),
		})
	}
	return out, nil
}

// ExportPoints 积分流水导出
func ExportPoints(sc *svc.ServiceContext) ([][]string, error) {
	var rows []model.PointsLog
	if err := sc.DB.Order("id DESC").Limit(exportRowLimit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := [][]string{{"流水ID", "会员ID", "变动", "变动后余额", "原因", "关联单号", "时间"}}
	for _, p := range rows {
		out = append(out, []string{
			strconvI64(p.ID),
			strconvI64(p.MemberID),
			strconvI64(p.Change),
			strconvI64(p.BalanceAfter),
			p.Reason,
			p.Ref,
			csvTime(p.CreatedAt),
		})
	}
	return out, nil
}

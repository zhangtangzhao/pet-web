// tasks.go 任务中心：每日任务 + 成长任务，声明式条件校验 + 幂等发放。
package growth

import (
	"context"
	"time"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

type taskDef struct {
	Key    string
	Title  string
	Desc   string
	Reward int
	Type   string // daily / growth
}

var taskDefs = []taskDef{
	{"view3", "逛一逛", "今日浏览 3 个不同商品", 10, "daily"},
	{"share", "好物分享", "今日分享 1 次商品", 5, "daily"},
	{"post", "晒单互动", "今日发布 1 条晒单", 15, "daily"},
	{"profile", "完善档案", "填写生日等资料", 20, "growth"},
	{"first_order", "首单礼", "完成首笔订单", 50, "growth"},
}

// Tasks 任务列表（含今日/历史完成态）
func Tasks(sc *svc.ServiceContext, memberID int64) (*types.TaskListResp, error) {
	today := time.Now()
	var done []model.MemberTask
	if err := sc.DB.Where("member_id = ? AND (task_date = ? OR task_key IN ?)",
		memberID, today.Format("2006-01-02"), []string{"profile", "first_order"}).
		Find(&done).Error; err != nil {
		return nil, err
	}
	doneSet := map[string]bool{}
	for _, d := range done {
		doneSet[d.TaskKey] = true
	}
	list := make([]types.TaskView, 0, len(taskDefs))
	for _, t := range taskDefs {
		list = append(list, types.TaskView{
			Key: t.Key, Title: t.Title, Desc: t.Desc, Reward: t.Reward, Type: t.Type,
			Claimed: doneSet[t.Key],
		})
	}
	resp := &types.TaskListResp{List: list}
	// 浏览任务条件：今日浏览的不同商品数
	return resp, nil
}

func int64String(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	var digits []byte
	if neg {
		v = -v
	}
	for v > 0 {
		digits = append([]byte{byte('0' + v%10)}, digits...)
		v /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}

// ClaimTask 领取任务奖励：校验条件 + 幂等入表 + 发积分
func ClaimTask(sc *svc.ServiceContext, memberID int64, key string) error {
	var def *taskDef
	for i := range taskDefs {
		if taskDefs[i].Key == key {
			def = &taskDefs[i]
			break
		}
	}
	if def == nil {
		return common.ErrTaskInvalid
	}
	today := time.Now()
	date := today
	if def.Type == "growth" {
		date = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	}
	var cnt int64
	sc.DB.Model(&model.MemberTask{}).
		Where("member_id = ? AND task_key = ? AND task_date = ?", memberID, key, date.Format("2006-01-02")).
		Count(&cnt)
	if cnt > 0 {
		return common.ErrTaskInvalid
	}
	// 条件校验
	switch key {
	case "view3":
		n, _ := sc.Rdb.LLen(context.Background(), "rl:view:"+int64String(memberID)).Result()
		if n < 3 {
			return common.NewErr(400, 42103, "今日还没逛够 3 个商品哦")
		}
	case "share":
		// 分享行为客户端上报，直接发放
	case "post":
		var n int64
		sc.DB.Model(&model.CommunityPost{}).
			Where("member_id = ? AND created_at >= ?", memberID, today.Format("2006-01-02")).Count(&n)
		if n == 0 {
			return common.NewErr(400, 42103, "今天还没发晒单哦")
		}
	case "profile":
		var m model.Member
		sc.DB.Select("birthday").First(&m, memberID)
		if m.Birthday == nil {
			return common.NewErr(400, 42103, "请先在我的页面完善生日信息")
		}
	case "first_order":
		var n int64
		sc.DB.Model(&model.Order{}).Where("member_id = ? AND status >= ?", memberID, model.OrderPaid).Count(&n)
		if n == 0 {
			return common.NewErr(400, 42103, "完成首笔订单后可领取")
		}
	}
	res := sc.DB.Create(&model.MemberTask{
		ID: common.NewID(), MemberID: memberID, TaskKey: key,
		RewardPoints: def.Reward, TaskDate: date,
	})
	if res.Error != nil {
		return common.ErrTaskInvalid // 唯一键冲突 = 已领
	}
	_, _ = Credit(sc, memberID, int64(def.Reward), "task_"+key, today.Format("2006-01-02"))
	return nil
}

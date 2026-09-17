// account.go 账号注销（合规）：申请 → 冷却期（可撤销）→ 冷却期结束由 closer 匿名化。
package auth

import (
	"time"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// 购买协议版本与正文（修改正文必须升版本，历史订单按签署时版本留档）
const AgreementVersion = "v1"

const AgreementContent = `一、健康告知：宠物为活体，健康状态以交付时检疫证明与视频确认为准；买方应在收货后 24 小时内完成验宠并留存视频证据，逾期视为认可宠物现状。
二、运输免责：因航空/铁路/专车托运等第三方运输导致的非人为延误，平台与卖家协助跟进但免责于运输方责任范围。
三、售后边界：非重大疾病（普通感冒、轻微应激等）不构成退换依据；重大传染病以具备资质的宠物医院出具的检测报告为准。
四、注销说明：账号注销设有冷静期，冷静期后个人身份信息将被匿名化处理，交易记录依法律法规要求保留。`

// AgreementCurrent 当前电子购买协议（版本 + 文案；文案与 mobile 静态页保持一致）
func AgreementCurrent() *types.AgreementResp {
	return &types.AgreementResp{
		Version: AgreementVersion,
		Title:   "宠物活体购买协议与健康告知",
		Content: AgreementContent,
	}
}

// DeleteStatus 注销申请状态
func DeleteStatus(sc *svc.ServiceContext, memberID int64) (*types.AccountDeleteStatusResp, error) {
	var m model.Member
	if err := sc.DB.Select("delete_requested_at", "delete_cooldown_until").First(&m, memberID).Error; err != nil {
		return nil, err
	}
	resp := &types.AccountDeleteStatusResp{CooldownDays: sc.Config.Account.DeletionCooldownDays}
	if m.DeleteRequestedAt != nil {
		resp.Requested = true
		if m.DeleteCooldownUntil != nil {
			resp.CooldownUntil = m.DeleteCooldownUntil.Format(time.RFC3339)
		}
	}
	return resp, nil
}

// DeleteRequest 申请注销：进入冷静期，期间禁交易；可随时撤销
func DeleteRequest(sc *svc.ServiceContext, memberID int64) error {
	days := sc.Config.Account.DeletionCooldownDays
	if days <= 0 {
		days = 7
	}
	until := time.Now().AddDate(0, 0, days)
	res := sc.DB.Model(&model.Member{}).
		Where("id = ? AND delete_requested_at IS NULL", memberID).
		Updates(map[string]any{
			"delete_requested_at":   time.Now(),
			"delete_cooldown_until": until,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.NewErr(400, 42001, "已提交过注销申请，等待冷静期结束")
	}
	return nil
}

// DeleteCancel 冷静期内撤销注销
func DeleteCancel(sc *svc.ServiceContext, memberID int64) error {
	res := sc.DB.Model(&model.Member{}).
		Where("id = ? AND delete_requested_at IS NOT NULL", memberID).
		Updates(map[string]any{"delete_requested_at": nil, "delete_cooldown_until": nil})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.NewErr(400, 40001, "当前没有进行中的注销申请")
	}
	return nil
}

// AnonymizeDeleted 冷却期结束的账号匿名化（保留订单交易记录，清除个人身份信息）
func AnonymizeDeleted(sc *svc.ServiceContext) int64 {
	var ids []int64
	if err := sc.DB.Model(&model.Member{}).
		Where("delete_requested_at IS NOT NULL AND delete_cooldown_until IS NOT NULL AND delete_cooldown_until < now()").
		Limit(100).Pluck("id", &ids).Error; err != nil {
		return 0
	}
	for _, id := range ids {
		_ = sc.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("DELETE FROM wechat_auth WHERE member_id = ?", id).Error; err != nil {
				return err
			}
			res := tx.Exec(`
				UPDATE member SET
					nickname = '已注销用户', phone = '', avatar = '', gender = 0,
					points = 0, growth_value = 0, level_reached = 0,
					invite_code = '', invited_by = 0, blacklist = 0,
					birthday = NULL, free_ship_cards = 0, last_login_at = NULL,
					delete_requested_at = NULL, delete_cooldown_until = NULL,
					updated_at = now()
				WHERE id = ?`, id)
			return res.Error
		})
	}
	return int64(len(ids))
}

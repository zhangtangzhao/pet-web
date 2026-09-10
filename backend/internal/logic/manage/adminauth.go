package manage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

var captchaStore = base64Captcha.DefaultMemStore

// GenCaptcha 生成图形验证码（base64 图片）
func GenCaptcha() (*types.CaptchaResp, error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	c := base64Captcha.NewCaptcha(driver, captchaStore)
	id, b64, _, err := c.Generate()
	if err != nil {
		return nil, err
	}
	return &types.CaptchaResp{CaptchaID: id, ImageB64: b64}, nil
}

func verifyCaptcha(id, code string) bool {
	return captchaStore.Verify(id, code, true)
}

const adminRefreshKeyFmt = "refresh:admin:%d:%s"

func adminRandomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// IssueAdminTokens 签发平台端双 token（独立 secret + 独立 Redis 命名空间）
func IssueAdminTokens(sc *svc.ServiceContext, a *model.AdminUser) (*types.AdminLoginResp, error) {
	cfg := sc.Config.Auth
	access, err := common.GenToken(cfg.AdminAccessSecret, time.Duration(cfg.AdminAccessExpire)*time.Second, a.ID, common.TokenTypeAdmin, a.Role)
	if err != nil {
		return nil, err
	}
	raw := adminRandomToken()
	refreshTTL := time.Duration(sc.Config.RefreshTokenExpireDays) * 24 * time.Hour
	if err := sc.Rdb.Set(context.Background(), fmt.Sprintf(adminRefreshKeyFmt, a.ID, raw), "1", refreshTTL).Err(); err != nil {
		return nil, err
	}
	return &types.AdminLoginResp{
		AccessToken:  access,
		RefreshToken: fmt.Sprintf("A%d.%s", a.ID, raw),
		ExpiresIn:    cfg.AdminAccessExpire,
		Admin:        AdminView(a),
	}, nil
}

// AdminView 管理员信息视图
func AdminView(a *model.AdminUser) types.AdminInfo {
	return types.AdminInfo{
		ID:       strconv.FormatInt(a.ID, 10),
		Username: a.Username,
		Nickname: a.Nickname,
		Role:     a.Role,
	}
}

// AdminLogin 管理员登录（bcrypt 校验，图形验证码可选）
func AdminLogin(sc *svc.ServiceContext, req *types.AdminLoginReq) (*types.AdminLoginResp, error) {
	if req.Username == "" || req.Password == "" {
		return nil, common.ErrParam
	}
	if req.CaptchaID != "" && !verifyCaptcha(req.CaptchaID, req.CaptchaCode) {
		return nil, common.ErrCaptcha
	}
	var a model.AdminUser
	if err := sc.DB.Where("username = ?", req.Username).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrAdminLogin
		}
		return nil, err
	}
	if a.Status != 1 {
		return nil, common.ErrForbidden
	}
	if bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(req.Password)) != nil {
		return nil, common.ErrAdminLogin
	}
	now := time.Now()
	_ = sc.DB.Model(&model.AdminUser{}).Where("id = ?", a.ID).Update("last_login_at", &now).Error
	return IssueAdminTokens(sc, &a)
}

// AdminRefresh 平台端刷新 token（"A{uid}.{raw}" 双段格式，GetDel 一次性轮换）
func AdminRefresh(sc *svc.ServiceContext, refreshToken string) (*types.AdminLoginResp, error) {
	if !strings.HasPrefix(refreshToken, "A") {
		return nil, common.ErrUnauthorized
	}
	var adminID int64
	var raw string
	if n, err := fmt.Sscanf(strings.TrimPrefix(refreshToken, "A"), "%d.%s", &adminID, &raw); err != nil || n != 2 {
		return nil, common.ErrUnauthorized
	}
	key := fmt.Sprintf(adminRefreshKeyFmt, adminID, raw)
	val, err := sc.Rdb.GetDel(context.Background(), key).Result()
	if err != nil || val != "1" {
		return nil, common.ErrTokenExpired
	}
	var a model.AdminUser
	if err := sc.DB.First(&a, adminID).Error; err != nil {
		return nil, common.ErrTokenExpired
	}
	if a.Status != 1 {
		return nil, common.ErrForbidden
	}
	return IssueAdminTokens(sc, &a)
}

// AdminLogout 吊销平台端 refresh_token
func AdminLogout(sc *svc.ServiceContext, adminID int64, refreshToken string) {
	if refreshToken == "" {
		return
	}
	parts := strings.SplitN(refreshToken, ".", 2)
	if len(parts) != 2 || parts[1] == "" {
		return
	}
	sc.Rdb.Del(context.Background(), fmt.Sprintf(adminRefreshKeyFmt, adminID, parts[1]))
}

// ChangeAdminPassword 修改密码（原密码校验 + bcrypt 重哈希）
func ChangeAdminPassword(sc *svc.ServiceContext, adminID int64, req *types.ChangePasswordReq) error {
	if len(req.NewPassword) < 8 {
		return common.ErrPasswordLength
	}
	var a model.AdminUser
	if err := sc.DB.First(&a, adminID).Error; err != nil {
		return common.ErrNotFound
	}
	if bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(req.OldPassword)) != nil {
		return common.ErrOldPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return sc.DB.Model(&model.AdminUser{}).Where("id = ?", adminID).
		Update("password_hash", string(hash)).Error
}

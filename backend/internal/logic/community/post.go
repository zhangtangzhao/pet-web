// Package community 晒单广场：发布（敏感词过滤 + 人工审核）、浏览、点赞、删除。
package community

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/notify"
	"pet/backend/internal/logic/risk"
	"pet/backend/internal/logic/sensitive"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

const (
	maxContent = 500
	maxImages  = 9
)

func postView(sc *svc.ServiceContext, p model.CommunityPost, memberID int64, nick, avatar string) types.PostView {
	images := []string{}
	_ = json.Unmarshal([]byte(p.Images), &images)
	liked := false
	if memberID > 0 {
		var cnt int64
		sc.DB.Model(&model.CommunityPostLike{}).
			Where("post_id = ? AND member_id = ?", p.ID, memberID).Count(&cnt)
		liked = cnt > 0
	}
	return types.PostView{
		ID: strconv.FormatInt(p.ID, 10), MemberID: strconv.FormatInt(p.MemberID, 10),
		Nickname: nick, Avatar: avatar,
		Content: p.Content, Images: images,
		LikeCount: p.LikeCount, Liked: liked,
		Status: p.Status, CreatedAt: p.CreatedAt.Format(time.RFC3339),
	}
}

// List 广场列表（仅显示中，游标 id DESC）
func List(sc *svc.ServiceContext, memberID int64, req *types.PostListReq) (*types.PostListResp, error) {
	limit := req.Limit
	if limit < 1 || limit > 50 {
		limit = 10
	}
	q := sc.DB.Table("community_post p").
		Select("p.*, m.nickname AS nickname, m.avatar AS avatar").
		Joins("LEFT JOIN member m ON m.id = p.member_id").
		Where("p.status = ?", model.PostShown)
	if req.Cursor != "" {
		q = q.Where("p.id < ?", req.Cursor)
	}
	var rows []struct {
		model.CommunityPost
		Nickname string
		Avatar   string
	}
	if err := q.Order("p.id DESC").Limit(limit + 1).Scan(&rows).Error; err != nil {
		return nil, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	list := make([]types.PostView, 0, len(rows))
	for _, r := range rows {
		list = append(list, postView(sc, r.CommunityPost, memberID, r.Nickname, r.Avatar))
	}
	return &types.PostListResp{List: list, HasMore: hasMore}, nil
}

// Create 发布晒单：敏感词过滤 + 风控（联系方式拦截）→ 待审核
func Create(sc *svc.ServiceContext, memberID int64, req *types.PostCreateReq) error {
	_ = req
	if err := risk.CheckReviewContent(sc, memberID, strings.TrimSpace(req.Content)); err != nil {
		return err
	}
	content := sensitive.Filter(sc.DB, strings.TrimSpace(req.Content))
	if content == "" && len(req.Images) == 0 {
		return common.NewErr(400, 40001, "说点什么或晒张图吧")
	}
	if utf8.RuneCountInString(content) > maxContent || len(req.Images) > maxImages {
		return common.ErrParam
	}
	images := []string{}
	for _, u := range req.Images {
		if u != "" && len(u) <= 512 {
			images = append(images, u)
		}
	}
	imgJSON, _ := json.Marshal(images)
	return sc.DB.Create(&model.CommunityPost{
		ID:       common.NewID(),
		MemberID: memberID,

		Content:  content,
		Images:   string(imgJSON),
		Status:   model.PostPending,
	}).Error
}

// Delete 删除（本人）
func Delete(sc *svc.ServiceContext, memberID int64, postID int64) error {
	res := sc.DB.Where("id = ? AND member_id = ?", postID, memberID).Delete(&model.CommunityPost{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// LikeToggle 点赞 / 取消（唯一约束幂等）
func LikeToggle(sc *svc.ServiceContext, memberID, postID int64) (bool, error) {
	var post model.CommunityPost
	if err := sc.DB.Where("id = ? AND status = ?", postID, model.PostShown).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, common.ErrNotFound
		}
		return false, err
	}
	var like model.CommunityPostLike
	err := sc.DB.Where("post_id = ? AND member_id = ?", postID, memberID).First(&like).Error
	if err == nil {
		if err := sc.DB.Delete(&like).Error; err != nil {
			return false, err
		}
		sc.DB.Model(&model.CommunityPost{}).Where("id = ?", postID).
			UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
		return false, nil
	}
	if err := sc.DB.Create(&model.CommunityPostLike{
		ID: common.NewID(), PostID: postID, MemberID: memberID,
	}).Error; err != nil {
		return false, err
	}
	sc.DB.Model(&model.CommunityPost{}).Where("id = ?", postID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1"))
	return true, nil
}

// MyList 我的动态（全部状态）
func MyList(sc *svc.ServiceContext, memberID int64) ([]types.PostView, error) {
	var posts []model.CommunityPost
	if err := sc.DB.Where("member_id = ?", memberID).Order("id DESC").Limit(50).Find(&posts).Error; err != nil {
		return nil, err
	}
	list := make([]types.PostView, 0, len(posts))
	for _, p := range posts {
		list = append(list, postView(sc, p, memberID, "", ""))
	}
	return list, nil
}

// ─────────────────────────── 平台端 ───────────────────────────

func AdminList(sc *svc.ServiceContext, status int) ([]types.PostView, error) {
	q := sc.DB.Table("community_post p").
		Select("p.*, m.nickname AS nickname, m.avatar AS avatar").
		Joins("LEFT JOIN member m ON m.id = p.member_id")
	if status > 0 {
		q = q.Where("p.status = ?", status)
	}
	var rows []struct {
		model.CommunityPost
		Nickname string
		Avatar   string
	}
	if err := q.Order("p.id DESC").Limit(100).Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.PostView, 0, len(rows))
	for _, r := range rows {
		list = append(list, postView(sc, r.CommunityPost, 0, r.Nickname, r.Avatar))
	}
	return list, nil
}

func AdminSetStatus(sc *svc.ServiceContext, postID int64, status int) error {
	if status != model.PostShown && status != model.PostHidden {
		return common.ErrParam
	}
	var post model.CommunityPost
	if err := sc.DB.First(&post, postID).Error; err != nil {
		return common.ErrNotFound
	}
	if err := sc.DB.Model(&model.CommunityPost{}).Where("id = ?", postID).
		Update("status", status).Error; err != nil {
		return err
	}
	if post.Status == model.PostPending {
		title, content := "动态审核通过", "你的晒单已通过审核，已发布到广场"
		if status == model.PostHidden {
			title, content = "动态未通过审核", "你发布的动态未通过审核，如有疑问请联系客服"
		}
		key := "postaudit:" + strconv.FormatInt(postID, 10) + ":" + strconv.Itoa(status)
		_ = notify.Enqueue(sc, post.MemberID, model.NotifySceneContent, key, title, content, "")
	}
	return nil
}

func AdminDelete(sc *svc.ServiceContext, postID int64) error {
	res := sc.DB.Delete(&model.CommunityPost{}, postID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}



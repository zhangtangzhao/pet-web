// recommend.go 智能选宠：按用户条件（预算/家庭/生活习惯）从在售宠物推荐 3 只。
// 配置了大模型走 LLM 结构化推荐（解析失败/不足 3 个自动降级或补位）；
// 未配置走本地规则打分（关键词解析 + 档案字段匹配），零配置可用。
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zeromicro/go-zero/core/logx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/pet"
	"pet/backend/internal/model"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

const (
	recTopN      = 3  // 推荐数量
	catalogLimit = 200 // 送入 LLM 的目录上限（按销量取头部）
)

type recCandidate struct {
	p         model.PetProduct
	breedName string
	catName   string
}

type recScored struct {
	cand   recCandidate
	score  int
	reason string
}

// PetAIRecommend 智能选宠推荐入口
func PetAIRecommend(sc *svc.ServiceContext, memberID int64, req *types.AIRecommendReq) (*types.AIRecommendResp, error) {
	requirement := strings.TrimSpace(req.Requirement)
	if requirement == "" || utf8.RuneCountInString(requirement) > 500 {
		return nil, common.ErrParam
	}
	catalog := loadCatalog(sc)
	if len(catalog) == 0 {
		return &types.AIRecommendResp{Items: []types.RecommendItem{}, Source: "rule"}, nil
	}

	if sc.Config.AIEnabled() {
		if err := checkRecQuota(sc, memberID); err != nil {
			return nil, err
		}
		if items, ok := aiRecommend(sc, catalog, requirement); ok {
			if len(items) < recTopN { // AI 结果不足 3 个 → 规则分候选补位
				items = fillFromRules(items, catalog, requirement)
				sort.SliceStable(items, func(i, j int) bool { return items[i].Score > items[j].Score })
			}
			return &types.AIRecommendResp{Items: items, Source: "ai"}, nil
		}
		logx.Infof("recommend: LLM 推荐失败，降级规则打分 member=%d", memberID)
	}
	items := make([]types.RecommendItem, 0, recTopN)
	for _, rs := range ruleRecommend(catalog, requirement, recTopN) {
		items = append(items, toItem(rs))
	}
	return &types.AIRecommendResp{Items: items, Source: "rule"}, nil
}

// ─────────────────────────── 在售目录 ───────────────────────────

func loadCatalog(sc *svc.ServiceContext) []recCandidate {
	var products []model.PetProduct
	if err := sc.DB.Where("status = ?", model.ProductOnSale).
		Order("sales DESC, id DESC").Limit(catalogLimit).Find(&products).Error; err != nil {
		logx.Errorf("recommend: 加载在售目录失败: %v", err)
		return nil
	}
	breedIDs := make([]int64, 0, len(products))
	catIDs := make([]int64, 0, len(products))
	for _, p := range products {
		breedIDs = append(breedIDs, p.BreedID)
		catIDs = append(catIDs, p.CategoryID)
	}
	breedName := map[int64]string{}
	catName := map[int64]string{}
	if len(breedIDs) > 0 {
		var breeds []model.Breed
		if err := sc.DB.Where("id IN ?", breedIDs).Find(&breeds).Error; err == nil {
			for _, b := range breeds {
				breedName[b.ID] = b.Name
			}
		}
	}
	if len(catIDs) > 0 {
		var cats []model.Category
		if err := sc.DB.Where("id IN ?", catIDs).Find(&cats).Error; err == nil {
			for _, c := range cats {
				catName[c.ID] = c.Name
			}
		}
	}
	catalog := make([]recCandidate, 0, len(products))
	for _, p := range products {
		catalog = append(catalog, recCandidate{p: p, breedName: breedName[p.BreedID], catName: catName[p.CategoryID]})
	}
	return catalog
}

// catalogText 紧凑目录行（性格截 12 字防 prompt 膨胀）
func catalogText(catalog []recCandidate) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("【在售宠物目录，共 %d 只（按销量排序，可能仅部分在售）】\n", len(catalog)))
	b.WriteString("格式：#编号 品种|价格元|性别|月龄|体型|性格\n")
	for _, c := range catalog {
		age := pet.AgeText(c.p.BirthDate)
		if age == "" {
			age = "未知"
		}
		personality := c.p.Personality
		if utf8.RuneCountInString(personality) > 12 {
			personality = string([]rune(personality)[:12])
		}
		if strings.TrimSpace(personality) == "" {
			personality = "无"
		}
		breed := orDefault(c.breedName, orDefault(c.catName, "未知"))
		b.WriteString(fmt.Sprintf("#%d %s|%s元|%s|%s|%s|%s\n",
			c.p.ID, breed, c.p.Price.StringFixed(0), pet.GenderText(c.p.PetGender), age,
			orDefault(c.p.BodyType, "未知"), personality))
	}
	return b.String()
}

func orDefault(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

// ─────────────────────────── LLM 路径 ───────────────────────────

func aiRecommend(sc *svc.ServiceContext, catalog []recCandidate, requirement string) ([]types.RecommendItem, bool) {
	system := "你是宠物交易平台的智能选宠顾问。根据用户条件从在售宠物目录中挑选最合适的 3 只。" +
		"只输出 JSON 数组，格式：[{\"productId\":目录编号,\"score\":0到100的整数推荐分,\"reason\":\"不超过60字的中文推荐理由，说明为什么适合该用户\"}]。" +
		"productId 必须取自目录中的编号；score 按匹配程度打分并严格降序；不要输出 JSON 以外的任何文字。"
	user := catalogText(catalog) + "\n【用户条件】\n" + requirement

	answer, err := sc.AIChatWithTemperature(context.Background(),
		[]svc.AIMessage{{Role: "system", Content: system}, {Role: "user", Content: user}}, 0.3)
	if err != nil {
		logx.Errorf("recommend: LLM 调用失败: %v", err)
		return nil, false
	}

	valid := make(map[int64]recCandidate, len(catalog))
	for _, c := range catalog {
		valid[c.p.ID] = c
	}
	var parsed []struct {
		ProductID json.Number `json:"productId"`
		Score     int         `json:"score"`
		Reason    string      `json:"reason"`
	}
	if err := json.Unmarshal(extractJSON(answer), &parsed); err != nil {
		logx.Errorf("recommend: LLM 输出解析失败: %v", err)
		return nil, false
	}
	seen := map[int64]bool{}
	items := make([]types.RecommendItem, 0, recTopN)
	for _, it := range parsed {
		id, err := it.ProductID.Int64()
		cand, inCatalog := valid[id]
		if err != nil || !inCatalog || seen[id] { // 目录外/重复编号丢弃
			continue
		}
		seen[id] = true
		score := it.Score
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		items = append(items, types.RecommendItem{
			ProductCard: cardOf(cand),
			Score:       score,
			Reason:      truncateRunes(strings.TrimSpace(it.Reason), 60),
		})
		if len(items) == recTopN {
			break
		}
	}
	if len(items) == 0 {
		return nil, false
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Score > items[j].Score })
	return items, true
}

// extractJSON 截取首个 '[' 到末个 ']'（容忍 ```json 围栏与前后杂文本）
func extractJSON(s string) []byte {
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start < 0 || end <= start {
		return []byte("[]")
	}
	return []byte(s[start : end+1])
}

// fillFromRules 规则分候选补位至 3 个（不与已有项重复）
func fillFromRules(items []types.RecommendItem, catalog []recCandidate, requirement string) []types.RecommendItem {
	exist := map[string]bool{}
	for _, it := range items {
		exist[it.ID] = true
	}
	for _, rs := range ruleRecommend(catalog, requirement, catalogLimit) {
		if len(items) == recTopN {
			break
		}
		id := strconv.FormatInt(rs.cand.p.ID, 10)
		if exist[id] {
			continue
		}
		items = append(items, toItem(rs))
	}
	return items
}

// ─────────────────────────── 规则打分路径 ───────────────────────────

type recSignals struct {
	budgetMax float64
	budgetSet bool
	species   string // 猫 / 狗（空=不限）
	apartment bool   // 公寓/小户型
	family    bool   // 有小孩/老人
	busy      bool   // 上班族/常出差
	novice    bool   // 新手
	quiet     bool   // 喜欢安静
	active    bool   // 喜欢活泼
}

var (
	budgetNumRe = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(万|千|[kK])?`)
	budgetPreRe = regexp.MustCompile(`(?:预算|不超过|低于|最多)\s*(\d+(?:\.\d+)?)\s*(万|千|[kK])?`)
	budgetSufRe = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(万|千|[kK])?\s*(以内|以下)`)
)

// parseBudget 解析预算上限：优先锚定（预算/不超过…/…以内/以下），否则取首个数字
func parseBudget(s string) (float64, bool) {
	mult := func(m []string) float64 {
		v, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0
		}
		switch m[2] {
		case "万":
			v *= 10000
		case "千", "k", "K":
			v *= 1000
		}
		return v
	}
	if m := budgetPreRe.FindStringSubmatch(s); m != nil {
		return mult(m), true
	}
	if m := budgetSufRe.FindStringSubmatch(s); m != nil {
		return mult(m), true
	}
	if m := budgetNumRe.FindStringSubmatch(s); m != nil {
		return mult(m), true
	}
	return 0, false
}

func parseSignals(s string) recSignals {
	sig := recSignals{}
	if max, ok := parseBudget(s); ok && max > 0 {
		sig.budgetMax, sig.budgetSet = max, true
	}
	if strings.Contains(s, "猫") && !strings.Contains(s, "狗") {
		sig.species = "猫"
	} else if strings.Contains(s, "狗") && !strings.Contains(s, "猫") {
		sig.species = "狗"
	}
	if strings.Contains(s, "公寓") || strings.Contains(s, "小户型") {
		sig.apartment = true
	}
	if strings.Contains(s, "小孩") || strings.Contains(s, "老人") || strings.Contains(s, "孩子") {
		sig.family = true
	}
	if strings.Contains(s, "上班") || strings.Contains(s, "出差") || strings.Contains(s, "没时间") || strings.Contains(s, "时间少") {
		sig.busy = true
	}
	if strings.Contains(s, "新手") || strings.Contains(s, "第一次") || strings.Contains(s, "第一次养") {
		sig.novice = true
	}
	if strings.Contains(s, "安静") {
		sig.quiet = true
	}
	if strings.Contains(s, "活泼") || strings.Contains(s, "互动") {
		sig.active = true
	}
	return sig
}

// ruleRecommend 规则打分排序取前 need 个
func ruleRecommend(catalog []recCandidate, requirement string, need int) []recScored {
	sig := parseSignals(requirement)

	scored := make([]recScored, 0, len(catalog))
	var overBudget []recScored
	for _, c := range catalog {
		rs := scoreCandidate(c, sig)
		if sig.budgetSet && c.p.Price.InexactFloat64() > sig.budgetMax {
			rs.score -= 40 // 超预算重罚
			overBudget = append(overBudget, rs)
			continue
		}
		scored = append(scored, rs)
	}
	if len(scored) < need { // 剔除后不足 → 保留超预算候选（仍带罚分）
		scored = append(scored, overBudget...)
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		if scored[i].cand.p.Sales != scored[j].cand.p.Sales {
			return scored[i].cand.p.Sales > scored[j].cand.p.Sales
		}
		return scored[i].cand.p.ID > scored[j].cand.p.ID
	})
	if len(scored) > need {
		scored = scored[:need]
	}
	return scored
}

func scoreCandidate(c recCandidate, sig recSignals) recScored {
	score := 60
	var reasons []string

	traits := c.p.Personality
	speciesText := c.catName + c.breedName

	if sig.species != "" {
		if strings.Contains(speciesText, sig.species) {
			score += 20
			reasons = append(reasons, "品种符合你的偏好")
		} else {
			score -= 15
		}
	}
	if sig.budgetSet {
		if c.p.Price.InexactFloat64() <= sig.budgetMax {
			score += 20
			reasons = append(reasons, "价格在预算内")
		}
	}
	if sig.apartment && (strings.Contains(c.p.BodyType, "小型") || strings.Contains(speciesText, "猫")) {
		score += 8
		reasons = append(reasons, "适合公寓饲养")
	}
	if sig.family && (strings.Contains(traits, "温顺") || strings.Contains(traits, "亲人")) {
		score += 8
		reasons = append(reasons, "性格温顺亲人，适合有孩子老人的家庭")
	}
	if sig.busy && (strings.Contains(traits, "独立") || strings.Contains(traits, "安静")) {
		score += 8
		reasons = append(reasons, "性格独立，适合上班族")
	}
	if sig.novice && (strings.Contains(traits, "温顺") || strings.Contains(traits, "好养") || c.p.VaccineDesc != "") {
		score += 8
		reasons = append(reasons, "好养活，适合新手")
	}
	if sig.quiet && strings.Contains(traits, "安静") {
		score += 8
		reasons = append(reasons, "性格安静")
	}
	if sig.active && (strings.Contains(traits, "活泼") || strings.Contains(traits, "粘人")) {
		score += 8
		reasons = append(reasons, "性格活泼粘人，互动性强")
	}
	if strings.Contains(c.p.VaccineDesc, "齐") || strings.Contains(c.p.VaccineDesc, "完") {
		score += 4
	}
	if c.p.Sales > 0 {
		score += int(math.Min(math.Log10(float64(c.p.Sales+1))*4, 6))
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "整体条件与你的需求匹配")
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return recScored{cand: c, score: score, reason: truncateRunes(strings.Join(reasons, " · "), 60)}
}

// ─────────────────────────── 公共工具 ───────────────────────────

func cardOf(c recCandidate) types.ProductCard {
	return types.ProductCard{
		ID:            strconv.FormatInt(c.p.ID, 10),
		Title:         c.p.Title,
		MainImage:     c.p.MainImage,
		Price:         c.p.Price.StringFixed(2),
		OriginalPrice: c.p.OriginalPrice.StringFixed(2),
		BreedName:     c.breedName,
		PetGender:     c.p.PetGender,
		FavoriteCount: c.p.FavoriteCount,
		Sales:         c.p.Sales,
		Status:        c.p.Status,
	}
}

func toItem(rs recScored) types.RecommendItem {
	return types.RecommendItem{ProductCard: cardOf(rs.cand), Score: rs.score, Reason: rs.reason}
}

func truncateRunes(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	return string([]rune(s)[:limit])
}

// checkRecQuota 推荐频控（仅 LLM 路径调用；值 ≤0 回落 AI 问宠的频控配置）
func checkRecQuota(sc *svc.ServiceContext, memberID int64) error {
	cfg := sc.Config.AI
	interval := cfg.RecommendIntervalSeconds
	if interval <= 0 {
		interval = cfg.AskIntervalSeconds
	}
	daily := cfg.RecommendDailyLimit
	if daily <= 0 {
		daily = cfg.DailyLimit
	}
	ctx := context.Background()
	ok, err := sc.Rdb.SetNX(ctx, fmt.Sprintf("ai:rec:interval:%d", memberID),
		1, time.Duration(interval)*time.Second).Result()
	if err != nil {
		return err
	}
	if !ok {
		return common.ErrAIRecFrequency
	}
	dayKey := fmt.Sprintf("ai:rec:day:%d:%s", memberID, time.Now().Format("20060102"))
	cnt, err := sc.Rdb.Incr(ctx, dayKey).Result()
	if err != nil {
		return err
	}
	if cnt == 1 {
		sc.Rdb.Expire(ctx, dayKey, 24*time.Hour)
	}
	if cnt > int64(daily) {
		return common.ErrAIRecDaily
	}
	return nil
}

package service

import (
	"strings"
	"time"
	"unicode/utf8"
	"yuanju/internal/model"
)

type ArticleQualityInput struct {
	Candidate ArticleCollectionCandidate
	Keyword   string
	Rank      int
	Config    model.ArticleQualityConfig
	Now       time.Time
}

type ArticleQualityResult struct {
	Score             int
	Reasons           []model.ArticleQualityReason
	Skip              bool
	SkipReason        string
	AutoPublished     bool
	AutoPublishReason string
}

func ScoreArticleCandidate(input ArticleQualityInput) ArticleQualityResult {
	cfg := normalizeQualityScoringConfig(input.Config)
	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	candidate := input.Candidate
	reasons := make([]model.ArticleQualityReason, 0, 8)
	source := strings.TrimSpace(candidate.SourceName)
	if sourceMatches(source, cfg.SourceBlacklist) {
		reasons = append(reasons, qualityReason("source_blacklist", -100, "来源命中黑名单"))
		return ArticleQualityResult{Score: 0, Reasons: reasons, Skip: true, SkipReason: "来源命中黑名单"}
	}

	score := 20
	reasons = append(reasons, qualityReason("base", 20, "基础可用性"))

	rankPoints := rankQualityPoints(input.Rank)
	score += rankPoints
	reasons = append(reasons, qualityReason("rank", rankPoints, rankQualityMessage(input.Rank)))

	recencyPoints := recencyQualityPoints(candidate.PublishedAtSource, now)
	score += recencyPoints
	reasons = append(reasons, qualityReason("recency", recencyPoints, recencyQualityMessage(candidate.PublishedAtSource, now)))

	bodyPoints, bodyMessage := bodyQualityPoints(candidate)
	score += bodyPoints
	reasons = append(reasons, qualityReason("body", bodyPoints, bodyMessage))

	relevancePoints := relevanceQualityPoints(candidate, input.Keyword, cfg.BonusKeywords)
	score += relevancePoints
	reasons = append(reasons, qualityReason("relevance", relevancePoints, "关键词与标题/摘要/正文匹配"))

	if sourceMatches(source, cfg.PreferredSources) {
		score += 10
		reasons = append(reasons, qualityReason("preferred_source", 10, "来源命中优先公众号"))
	}

	usabilityPoints, usabilityMessage := usabilityQualityPoints(candidate)
	score += usabilityPoints
	reasons = append(reasons, qualityReason("usability", usabilityPoints, usabilityMessage))

	score = clampScore(score)
	if cfg.Enabled && !cfg.AllowWithoutBody && strings.TrimSpace(candidate.BodyContent) == "" {
		return ArticleQualityResult{Score: score, Reasons: reasons, Skip: true, SkipReason: "未获取正文，且质量配置要求必须有正文"}
	}
	if cfg.Enabled && score < cfg.MinQualityScore {
		return ArticleQualityResult{Score: score, Reasons: reasons, Skip: true, SkipReason: "质量分低于阈值"}
	}
	return ArticleQualityResult{Score: score, Reasons: reasons}
}

func normalizeQualityScoringConfig(cfg model.ArticleQualityConfig) model.ArticleQualityConfig {
	if cfg.MinQualityScore < 0 {
		cfg.MinQualityScore = 0
	}
	if cfg.MinQualityScore > 100 {
		cfg.MinQualityScore = 100
	}
	cfg.BonusKeywords = compactStringList(cfg.BonusKeywords)
	cfg.SourceBlacklist = compactStringList(cfg.SourceBlacklist)
	cfg.PreferredSources = compactStringList(cfg.PreferredSources)
	return cfg
}

func rankQualityPoints(rank int) int {
	switch {
	case rank <= 0:
		return 8
	case rank == 1:
		return 20
	case rank <= 3:
		return 16
	case rank <= 5:
		return 12
	case rank <= 10:
		return 8
	default:
		return 4
	}
}

func rankQualityMessage(rank int) string {
	if rank <= 0 {
		return "搜索排序未知"
	}
	return "搜索排序第 " + intToString(rank) + " 位"
}

func recencyQualityPoints(publishedAt *time.Time, now time.Time) int {
	if publishedAt == nil {
		return 4
	}
	age := now.Sub(*publishedAt)
	switch {
	case age <= 7*24*time.Hour:
		return 15
	case age <= 30*24*time.Hour:
		return 12
	case age <= 90*24*time.Hour:
		return 8
	default:
		return 3
	}
}

func recencyQualityMessage(publishedAt *time.Time, now time.Time) string {
	if publishedAt == nil {
		return "发布时间未知"
	}
	if now.Sub(*publishedAt) <= 30*24*time.Hour {
		return "发布时间较新"
	}
	return "发布时间较早"
}

func bodyQualityPoints(candidate ArticleCollectionCandidate) (int, string) {
	bodyLength := utf8.RuneCountInString(strings.TrimSpace(candidate.BodyContent))
	if bodyLength >= 800 {
		return 20, "正文完整度较高"
	}
	if bodyLength >= 300 {
		return 14, "正文可用但偏短"
	}
	if bodyLength > 0 {
		return 8, "正文过短"
	}
	switch candidate.BodyFetchStatus {
	case model.ArticleBodyFetchStatusWechatVideoPage, model.ArticleBodyFetchStatusWechatAntispider, model.ArticleBodyFetchStatusSogouVerifyRequired:
		return 0, "正文受反爬或富媒体限制"
	default:
		return 4, "未获取正文，仅保留搜索摘要"
	}
}

func relevanceQualityPoints(candidate ArticleCollectionCandidate, keyword string, bonusKeywords []string) int {
	text := strings.ToLower(candidate.Title + " " + candidate.SearchSnippet + " " + candidate.BodyContent)
	score := 0
	kw := strings.ToLower(strings.TrimSpace(keyword))
	if kw != "" && strings.Contains(text, kw) {
		score += 8
	}
	if strings.TrimSpace(candidate.Title) != "" && kw != "" && strings.Contains(strings.ToLower(candidate.Title), kw) {
		score += 4
	}
	for _, bonus := range bonusKeywords {
		if strings.Contains(text, strings.ToLower(bonus)) {
			score += 2
		}
	}
	if score > 14 {
		score = 14
	}
	return score
}

func usabilityQualityPoints(candidate ArticleCollectionCandidate) (int, string) {
	title := strings.TrimSpace(candidate.Title)
	if title == "" || strings.TrimSpace(candidate.OriginalURL) == "" {
		return -20, "标题或链接缺失"
	}
	if strings.Contains(strings.ToLower(title), "视频") || candidate.BodyFetchStatus == model.ArticleBodyFetchStatusWechatVideoPage {
		return -6, "视频/富媒体文章不利于仿写"
	}
	return 8, "标题和链接可用"
}

func sourceMatches(source string, rules []string) bool {
	if strings.TrimSpace(source) == "" {
		return false
	}
	source = strings.ToLower(source)
	for _, rule := range rules {
		rule = strings.ToLower(strings.TrimSpace(rule))
		if rule != "" && strings.Contains(source, rule) {
			return true
		}
	}
	return false
}

func qualityReason(reasonType string, points int, message string) model.ArticleQualityReason {
	return model.ArticleQualityReason{Type: reasonType, Points: points, Message: message}
}

func compactStringList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func clampScore(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func intToString(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	n := value
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

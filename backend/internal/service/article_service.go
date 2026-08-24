package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
	"yuanju/internal/model"
)

type ArticleAnalysisPromptInput struct {
	Article      model.Article
	CategoryName string
	TagNames     []string
}

var droppedArticleURLParams = map[string]bool{
	"chksm":          true,
	"from":           true,
	"isappinstalled": true,
	"scene":          true,
	"version":        true,
	"platform":       true,
	"clicktime":      true,
}

func NormalizeArticleOriginalURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("original_url 不能为空")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("original_url 无效: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("original_url 必须包含协议和域名")
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	parsed.RawQuery = normalizeArticleQuery(parsed.Query())
	return parsed.String(), nil
}

func ArticleOriginalURLHash(raw string) string {
	normalized, err := NormalizeArticleOriginalURL(raw)
	if err != nil {
		normalized = strings.TrimSpace(raw)
	}
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

func ValidateArticleStatusTransition(from, to model.ArticleStatus) error {
	if from == to {
		return nil
	}
	allowed := map[model.ArticleStatus]map[model.ArticleStatus]bool{
		model.ArticleStatusCandidate: {
			model.ArticleStatusPublished: true,
			model.ArticleStatusRejected:  true,
			model.ArticleStatusDeleted:   true,
		},
		model.ArticleStatusPublished: {
			model.ArticleStatusTakenDown: true,
			model.ArticleStatusDeleted:   true,
		},
		model.ArticleStatusRejected: {
			model.ArticleStatusPublished: true,
			model.ArticleStatusDeleted:   true,
		},
		model.ArticleStatusTakenDown: {
			model.ArticleStatusPublished: true,
			model.ArticleStatusDeleted:   true,
		},
	}
	if allowed[from][to] {
		return nil
	}
	return fmt.Errorf("非法文章状态流转: %s -> %s", from, to)
}

func BuildArticleAnalysisPrompt(input ArticleAnalysisPromptInput) string {
	article := input.Article
	var b strings.Builder
	b.WriteString("你是缘聚资讯模块的内容分析助手。请仅基于下列已采集正文、公开搜索信息与后台分类标签，生成阅读辅助和仿写拆解，禁止补写未提供的原文细节。\n\n")
	b.WriteString("文章公开信息：\n")
	writePromptLine(&b, "标题", article.Title)
	writePromptLine(&b, "来源公众号", article.SourceName)
	if article.PublishedAtSource != nil {
		writePromptLine(&b, "来源发布时间", article.PublishedAtSource.In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04"))
	}
	writePromptLine(&b, "公开搜索摘要", article.SearchSnippet)
	writePromptLine(&b, "平台摘要", article.Summary)
	writePromptLine(&b, "原文链接", article.OriginalURL)
	if article.FullTextAuthorized {
		writePromptLine(&b, "已采集正文", truncateRunes(article.BodyContent, 6000))
	}
	writePromptLine(&b, "后台分类", input.CategoryName)
	if len(input.TagNames) > 0 {
		tags := append([]string(nil), input.TagNames...)
		sort.Strings(tags)
		writePromptLine(&b, "后台标签", strings.Join(tags, "、"))
	}
	b.WriteString("\n输出 JSON，字段包括 one_sentence_summary、key_points、target_readers、related_topics、suggested_tags、title_pattern、opening_style、structure_outline、expression_style、rewrite_angles。")
	return b.String()
}

func truncateRunes(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "\n……"
}

func normalizeArticleQuery(values url.Values) string {
	next := url.Values{}
	for key, vals := range values {
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "utm_") || droppedArticleURLParams[lower] {
			continue
		}
		for _, v := range vals {
			next.Add(key, v)
		}
	}
	return next.Encode()
}

func writePromptLine(b *strings.Builder, label, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	b.WriteString("- ")
	b.WriteString(label)
	b.WriteString("：")
	b.WriteString(value)
	b.WriteString("\n")
}

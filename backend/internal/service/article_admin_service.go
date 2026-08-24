package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"yuanju/configs"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/pkg/crypto"
)

type ArticleBatchActionInput struct {
	IDs                   []string
	Action                string
	AdminID               string
	Note                  string
	AllowPublishWithoutAI bool
}

type ArticleBatchActionResult struct {
	Updated int      `json:"updated"`
	Skipped []string `json:"skipped"`
}

type ArticleAIBatchResult struct {
	Succeeded int      `json:"succeeded"`
	Failed    []string `json:"failed"`
}

func SaveArticleBody(articleID, bodyContent string) (*model.Article, error) {
	bodyContent = strings.TrimSpace(bodyContent)
	if bodyContent == "" {
		return nil, fmt.Errorf("正文内容不能为空")
	}
	article, err := repository.UpdateArticleBody(articleID, bodyContent, "")
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, sql.ErrNoRows
	}
	return article, nil
}

func FetchAndSaveArticleBody(ctx context.Context, articleID, rawURL string) (*model.Article, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("原文链接不能为空")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("原文链接格式无效")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("原文链接仅支持 http 或 https")
	}
	provider := NewSogouWeChatProvider("", &http.Client{Timeout: 10 * time.Second})
	content, finalURL, status, reason, err := provider.fetchArticleBody(ctx, rawURL)
	if err != nil {
		return nil, fmt.Errorf("正文抓取失败: %w", err)
	}
	if strings.TrimSpace(content) == "" {
		if reason == "" {
			reason = string(status)
		}
		return nil, fmt.Errorf("未能从该链接解析正文：%s", reason)
	}
	if strings.TrimSpace(finalURL) == "" {
		finalURL = rawURL
	}
	article, err := repository.UpdateArticleBody(articleID, content, finalURL)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, sql.ErrNoRows
	}
	return article, nil
}

func BatchUpdateArticles(input ArticleBatchActionInput) (ArticleBatchActionResult, error) {
	var result ArticleBatchActionResult
	target, err := articleStatusForAction(input.Action)
	if err != nil {
		return result, err
	}
	for _, id := range input.IDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		article, err := repository.GetArticleByID(id)
		if err != nil {
			return result, err
		}
		if article == nil {
			result.Skipped = append(result.Skipped, id)
			continue
		}
		if target == model.ArticleStatusPublished && article.AIAnalysis == nil && !input.AllowPublishWithoutAI {
			result.Skipped = append(result.Skipped, id)
			continue
		}
		if err := ValidateArticleStatusTransition(article.Status, target); err != nil {
			result.Skipped = append(result.Skipped, id)
			continue
		}
		updated, err := repository.UpdateArticleStatusWithAudit(id, input.AdminID, target, input.Note)
		if err != nil {
			return result, err
		}
		if updated != nil {
			result.Updated++
		}
	}
	return result, nil
}

func GetPublishedArticleDetail(articleID string) (*model.Article, error) {
	article, err := repository.GetPublishedArticleByID(articleID)
	if err != nil || article == nil {
		return article, err
	}
	if err := repository.IncrementArticleViewCount(articleID); err != nil {
		return nil, err
	}
	article.ViewCount++
	return article, nil
}

func GenerateArticleAIAnalysis(articleID string) (*model.ArticleAIAnalysis, error) {
	article, err := repository.GetArticleByID(articleID)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, sql.ErrNoRows
	}
	promptCfg, err := repository.GetArticleAIPrompt()
	if err != nil {
		return nil, err
	}
	provider, err := repository.GetActiveArticleAIProvider()
	if err != nil {
		return nil, err
	}
	if provider == nil {
		err := fmt.Errorf("未配置启用的资讯 AI Provider")
		_ = repository.RecordArticleAIAnalysisFailure(articleID, err.Error())
		return nil, err
	}
	apiKey, err := crypto.Decrypt(provider.APIKeyEncrypted, configs.AppConfig.AdminEncryptionKey)
	if err != nil {
		_ = repository.RecordArticleAIAnalysisFailure(articleID, "资讯 AI Provider API Key 解密失败")
		return nil, err
	}
	systemPrompt := "你是缘聚资讯模块的内容分析助手，只能基于已提供的正文、公开搜索信息和后台标签生成摘要和仿写拆解，不得编造未提供的原文细节。"
	if promptCfg != nil && strings.TrimSpace(promptCfg.Content) != "" {
		systemPrompt = promptCfg.Content
	}
	userPrompt := BuildArticleAnalysisPrompt(ArticleAnalysisPromptInput{Article: *article})
	baseURL := strings.TrimSuffix(strings.TrimSuffix(provider.BaseURL, "/v1"), "/")
	content, _, err := callOpenAICompatibleWithLog(baseURL+"/v1/chat/completions", apiKey, provider.Model, systemPrompt, userPrompt)
	if err != nil {
		_ = repository.RecordArticleAIAnalysisFailure(articleID, err.Error())
		return nil, err
	}
	var analysis model.ArticleAIAnalysis
	if err := json.Unmarshal([]byte(stripJSONFence(content)), &analysis); err != nil {
		_ = repository.RecordArticleAIAnalysisFailure(articleID, "AI 返回 JSON 解析失败: "+err.Error())
		return nil, err
	}
	if err := repository.SaveArticleAIAnalysis(articleID, analysis); err != nil {
		return nil, err
	}
	return &analysis, nil
}

func BatchGenerateArticleAIAnalysis(ids []string) ArticleAIBatchResult {
	var result ArticleAIBatchResult
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, err := GenerateArticleAIAnalysis(id); err != nil {
			result.Failed = append(result.Failed, id)
			continue
		}
		result.Succeeded++
	}
	return result
}

func articleStatusForAction(action string) (model.ArticleStatus, error) {
	switch action {
	case "publish":
		return model.ArticleStatusPublished, nil
	case "reject":
		return model.ArticleStatusRejected, nil
	case "take_down", "takedown":
		return model.ArticleStatusTakenDown, nil
	case "delete":
		return model.ArticleStatusDeleted, nil
	default:
		return "", fmt.Errorf("未知文章批量动作: %s", action)
	}
}

func stripJSONFence(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}

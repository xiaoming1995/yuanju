package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode"
	"yuanju/configs"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/internal/service"
	"yuanju/pkg/crypto"

	"github.com/gin-gonic/gin"
)

const articleModuleClosedMessage = "资讯模块暂未开放"

func GetArticleModuleSettings(c *gin.Context) {
	enabled, err := repository.IsArticleModuleEnabled()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取资讯模块设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"module_enabled": enabled})
}

func AdminGetArticleModuleSettings(c *gin.Context) {
	enabled, err := repository.IsArticleModuleEnabled()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取资讯模块设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"module_enabled": enabled})
}

func AdminUpdateArticleModuleSettings(c *gin.Context) {
	var req struct {
		ModuleEnabled *bool `json:"module_enabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ModuleEnabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module_enabled 必填"})
		return
	}
	if err := repository.SetBoolSetting(repository.SettingArticlesModuleEnabled, *req.ModuleEnabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存资讯模块设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"module_enabled": *req.ModuleEnabled})
}

func ListArticles(c *gin.Context) {
	if !ensureArticleModuleEnabled(c) {
		return
	}
	page, limit, offset := pagination(c)
	articles, total, err := repository.ListArticles(repository.ArticleListFilter{
		Status:          "published",
		Query:           c.Query("q"),
		CategoryID:      c.Query("category"),
		TagID:           c.Query("tag"),
		Sort:            c.DefaultQuery("sort", "latest"),
		MinQualityScore: queryInt(c, "min_quality_score"),
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资讯失败"})
		return
	}
	clearArticleBodies(articles)
	categories, _ := repository.ListArticleCategories(true)
	tags, _ := repository.ListArticleTags(true)
	c.JSON(http.StatusOK, gin.H{"articles": articles, "total": total, "page": page, "limit": limit, "categories": categories, "tags": tags})
}

func GetArticleDetail(c *gin.Context) {
	if !ensureArticleModuleEnabled(c) {
		return
	}
	article, err := service.GetPublishedArticleDetail(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取资讯失败"})
		return
	}
	if article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "资讯不存在或已下架"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"article": article})
}

func TrackArticleOriginalClick(c *gin.Context) {
	if !ensureArticleModuleEnabled(c) {
		return
	}
	userID := c.GetString("user_id")
	article, err := repository.GetPublishedArticleByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取资讯失败"})
		return
	}
	if article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "资讯不存在或已下架"})
		return
	}
	if err := repository.RecordArticleOriginalClick(article.ID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "记录点击失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"original_url": article.OriginalURL})
}

func ensureArticleModuleEnabled(c *gin.Context) bool {
	enabled, err := repository.IsArticleModuleEnabled()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取资讯模块设置失败"})
		return false
	}
	if !enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": articleModuleClosedMessage, "code": "article_module_closed"})
		return false
	}
	return true
}

func ensureArticleModuleEnabledForAdminOperation(c *gin.Context) bool {
	enabled, err := repository.IsArticleModuleEnabled()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取资讯模块设置失败"})
		return false
	}
	if !enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "资讯模块已关闭，请先打开模块总开关", "code": "article_module_closed"})
		return false
	}
	return true
}

func AdminListArticles(c *gin.Context) {
	page, limit, offset := pagination(c)
	articles, total, err := repository.ListArticles(repository.ArticleListFilter{
		Status:          c.Query("status"),
		Query:           c.Query("q"),
		CategoryID:      c.Query("category"),
		TagID:           c.Query("tag"),
		Sort:            c.DefaultQuery("sort", "latest"),
		MinQualityScore: queryInt(c, "min_quality_score"),
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资讯失败"})
		return
	}
	clearArticleBodies(articles)
	c.JSON(http.StatusOK, gin.H{"articles": articles, "total": total, "page": page, "limit": limit})
}

func AdminGetArticleDetail(c *gin.Context) {
	article, err := repository.GetArticleByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取资讯失败"})
		return
	}
	if article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "资讯不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"article": article})
}

func AdminUpdateArticleBody(c *gin.Context) {
	var req struct {
		BodyContent string `json:"body_content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	article, err := service.SaveArticleBody(c.Param("id"), req.BodyContent)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "资讯不存在"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"article": article})
}

func AdminFetchArticleBody(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	article, err := service.FetchAndSaveArticleBody(c.Request.Context(), c.Param("id"), req.URL)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "资讯不存在"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"article": article})
}

func AdminArticleBatchAction(c *gin.Context) {
	var req struct {
		IDs                   []string `json:"ids"`
		Action                string   `json:"action"`
		Note                  string   `json:"note"`
		AllowPublishWithoutAI bool     `json:"allow_publish_without_ai"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	adminID := c.GetString("admin_id")
	result, err := service.BatchUpdateArticles(service.ArticleBatchActionInput{
		IDs:                   req.IDs,
		Action:                req.Action,
		AdminID:               adminID,
		Note:                  req.Note,
		AllowPublishWithoutAI: req.AllowPublishWithoutAI,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}

func AdminListArticleCategories(c *gin.Context) {
	items, err := repository.ListArticleCategories(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询分类失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": items})
}

func AdminCreateArticleCategory(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Slug      string `json:"slug"`
		SortOrder int    `json:"sort_order"`
		Active    *bool  `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	item, err := repository.CreateArticleCategory(req.Name, ensureSlug(req.Slug, req.Name), req.SortOrder, active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建分类失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"category": item})
}

func AdminUpdateArticleCategory(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Slug      string `json:"slug"`
		SortOrder int    `json:"sort_order"`
		Active    bool   `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := repository.UpdateArticleCategory(c.Param("id"), req.Name, ensureSlug(req.Slug, req.Name), req.SortOrder, req.Active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新分类失败"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "分类不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"category": item})
}

func AdminListArticleTags(c *gin.Context) {
	items, err := repository.ListArticleTags(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询标签失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": items})
}

func AdminCreateArticleTag(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Slug   string `json:"slug"`
		Active *bool  `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	item, err := repository.CreateArticleTag(req.Name, ensureSlug(req.Slug, req.Name), active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建标签失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"tag": item})
}

func AdminUpdateArticleTag(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Slug   string `json:"slug"`
		Active bool   `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := repository.UpdateArticleTag(c.Param("id"), req.Name, ensureSlug(req.Slug, req.Name), req.Active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新标签失败"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "标签不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tag": item})
}

func AdminListArticleKeywords(c *gin.Context) {
	items, err := repository.ListArticleKeywords(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询关键词失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"keywords": items})
}

func AdminCreateArticleKeyword(c *gin.Context) {
	var req struct {
		Keyword string `json:"keyword" binding:"required"`
		Active  *bool  `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	item, err := repository.CreateArticleKeyword(req.Keyword, active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建关键词失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"keyword": item})
}

func AdminUpdateArticleKeyword(c *gin.Context) {
	var req struct {
		Keyword string `json:"keyword" binding:"required"`
		Active  bool   `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := repository.UpdateArticleKeyword(c.Param("id"), req.Keyword, req.Active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新关键词失败"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "关键词不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"keyword": item})
}

func AdminTriggerArticleCollection(c *gin.Context) {
	if !ensureArticleModuleEnabledForAdminOperation(c) {
		return
	}
	cfg, _ := repository.GetArticleCollectionConfig()
	maxResults := 20
	if cfg != nil && cfg.MaxResultsPerRun > 0 {
		maxResults = cfg.MaxResultsPerRun
	}
	task, err := service.RunManualArticleCollectionTask(c.Request.Context(), articleCollectionProviderFromConfig(cfg), maxResults)
	if err != nil {
		if errors.Is(err, service.ErrArticleModuleDisabled) {
			c.JSON(http.StatusForbidden, gin.H{"error": "资讯模块已关闭，请先打开模块总开关", "code": "article_module_closed"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"task": task, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": task})
}

func AdminListArticleCollectionTasks(c *gin.Context) {
	_, limit, offset := pagination(c)
	tasks, err := repository.ListArticleCollectionTasks(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询采集任务失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func AdminListArticleCollectionTaskItems(c *gin.Context) {
	_, limit, offset := pagination(c)
	items, err := repository.ListArticleCollectionTaskItems(c.Param("id"), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询采集明细失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func AdminRetryArticleCollectionTask(c *gin.Context) {
	if !ensureArticleModuleEnabledForAdminOperation(c) {
		return
	}
	cfg, _ := repository.GetArticleCollectionConfig()
	maxResults := 20
	if cfg != nil && cfg.MaxResultsPerRun > 0 {
		maxResults = cfg.MaxResultsPerRun
	}
	task, err := service.RetryArticleCollectionTask(c.Request.Context(), c.Param("id"), articleCollectionProviderFromConfig(cfg), maxResults)
	if err != nil {
		if errors.Is(err, service.ErrArticleModuleDisabled) {
			c.JSON(http.StatusForbidden, gin.H{"error": "资讯模块已关闭，请先打开模块总开关", "code": "article_module_closed"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"task": task, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": task})
}

func AdminGetArticleCollectionConfig(c *gin.Context) {
	cfg, err := repository.GetArticleCollectionConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询采集配置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

func AdminUpdateArticleCollectionConfig(c *gin.Context) {
	var req struct {
		Enabled                    bool   `json:"enabled"`
		Frequency                  string `json:"frequency"`
		ScheduleInterval           int    `json:"schedule_interval_minutes"`
		MaxResultsPerRun           int    `json:"max_results_per_run"`
		SearchPageMin              int    `json:"search_page_min"`
		SearchPageMax              int    `json:"search_page_max"`
		AutoPublishEnabled         bool   `json:"auto_publish_enabled"`
		AutoPublishMinQualityScore *int   `json:"auto_publish_min_quality_score"`
		AutoPublishRequiresBody    *bool  `json:"auto_publish_requires_body"`
		AutoPublishMaxPerRun       *int   `json:"auto_publish_max_per_run"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Frequency == "" {
		req.Frequency = "daily"
	}
	if req.ScheduleInterval <= 0 {
		if req.Frequency == "weekly" {
			req.ScheduleInterval = 10080
		} else {
			req.ScheduleInterval = 1440
		}
	}
	if req.MaxResultsPerRun <= 0 {
		req.MaxResultsPerRun = 20
	}
	if req.SearchPageMin <= 0 {
		req.SearchPageMin = 1
	}
	if req.SearchPageMax <= 0 {
		req.SearchPageMax = 5
	}
	autoPublishMinQualityScore := 90
	if req.AutoPublishMinQualityScore != nil {
		autoPublishMinQualityScore = *req.AutoPublishMinQualityScore
	}
	autoPublishRequiresBody := true
	if req.AutoPublishRequiresBody != nil {
		autoPublishRequiresBody = *req.AutoPublishRequiresBody
	}
	autoPublishMaxPerRun := 3
	if req.AutoPublishMaxPerRun != nil {
		autoPublishMaxPerRun = *req.AutoPublishMaxPerRun
	}
	cfg, err := repository.UpdateArticleCollectionConfig(model.ArticleCollectionConfig{
		Enabled:                    req.Enabled,
		Frequency:                  req.Frequency,
		ScheduleInterval:           req.ScheduleInterval,
		MaxResultsPerRun:           req.MaxResultsPerRun,
		SearchPageMin:              req.SearchPageMin,
		SearchPageMax:              req.SearchPageMax,
		AutoPublishEnabled:         req.AutoPublishEnabled,
		AutoPublishMinQualityScore: autoPublishMinQualityScore,
		AutoPublishRequiresBody:    autoPublishRequiresBody,
		AutoPublishMaxPerRun:       autoPublishMaxPerRun,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新采集配置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

func AdminGetArticleQualityConfig(c *gin.Context) {
	cfg, err := repository.GetArticleQualityConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询质量配置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

func AdminUpdateArticleQualityConfig(c *gin.Context) {
	var req struct {
		Enabled               bool     `json:"quality_filter_enabled"`
		MinQualityScore       int      `json:"min_quality_score"`
		AllowWithoutBody      bool     `json:"allow_without_body"`
		BonusKeywords         []string `json:"bonus_keywords"`
		SourceBlacklist       []string `json:"source_blacklist"`
		PreferredSources      []string `json:"preferred_sources"`
		AIQualityCheckEnabled bool     `json:"ai_quality_check_enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg, err := repository.UpdateArticleQualityConfig(model.ArticleQualityConfig{
		Enabled:               req.Enabled,
		MinQualityScore:       req.MinQualityScore,
		AllowWithoutBody:      req.AllowWithoutBody,
		BonusKeywords:         req.BonusKeywords,
		SourceBlacklist:       req.SourceBlacklist,
		PreferredSources:      req.PreferredSources,
		AIQualityCheckEnabled: req.AIQualityCheckEnabled,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新质量配置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

func articleCollectionProviderFromConfig(cfg *model.ArticleCollectionConfig) service.ArticleSearchProvider {
	provider := service.NewSogouWeChatProvider("", nil)
	if cfg == nil {
		return provider
	}
	return provider.WithSearchPageRange(cfg.SearchPageMin, cfg.SearchPageMax)
}

func AdminGetArticleAIConfig(c *gin.Context) {
	provider, _ := repository.GetActiveArticleAIProvider()
	prompt, _ := repository.GetArticleAIPrompt()
	if provider != nil {
		provider.APIKeyMasked = provider.APIKeyPreview
	}
	c.JSON(http.StatusOK, gin.H{"provider": provider, "prompt": prompt})
}

func AdminUpdateArticleAIConfig(c *gin.Context) {
	var req struct {
		PromptContent     string `json:"prompt_content"`
		PromptDescription string `json:"prompt_description"`
		ProviderName      string `json:"provider_name"`
		ProviderType      string `json:"provider_type"`
		BaseURL           string `json:"base_url"`
		Model             string `json:"model"`
		APIKey            string `json:"api_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var prompt any
	var provider any
	var err error
	if strings.TrimSpace(req.PromptContent) != "" {
		prompt, err = repository.UpdateArticleAIPrompt(req.PromptContent, req.PromptDescription)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新 Prompt 失败"})
			return
		}
	}
	if req.ProviderName != "" && req.BaseURL != "" && req.Model != "" && req.APIKey != "" {
		encrypted, encErr := crypto.Encrypt(req.APIKey, configs.AppConfig.AdminEncryptionKey)
		if encErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Key 加密失败"})
			return
		}
		provider, err = repository.CreateArticleAIProvider(req.ProviderName, req.ProviderType, req.BaseURL, req.Model, encrypted, crypto.MaskPlainKey(req.APIKey), true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新 Provider 失败"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"provider": provider, "prompt": prompt})
}

func AdminGenerateArticleAIAnalysis(c *gin.Context) {
	analysis, err := service.GenerateArticleAIAnalysis(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "生成资讯 AI 分析失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"analysis": analysis})
}

func AdminBatchGenerateArticleAIAnalysis(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result := service.BatchGenerateArticleAIAnalysis(req.IDs)
	c.JSON(http.StatusOK, gin.H{"result": result})
}

func pagination(c *gin.Context) (page, limit, offset int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return page, limit, (page - 1) * limit
}

func queryInt(c *gin.Context, key string) int {
	value, _ := strconv.Atoi(c.Query(key))
	return value
}

func clearArticleBodies(articles []model.Article) {
	for i := range articles {
		articles[i].BodyContent = ""
	}
}

func ensureSlug(slug, fallback string) string {
	slug = strings.TrimSpace(slug)
	if slug != "" {
		return slug
	}
	var b strings.Builder
	for _, r := range strings.ToLower(fallback) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if b.Len() > 0 {
			b.WriteByte('-')
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "article-taxonomy"
	}
	return result
}

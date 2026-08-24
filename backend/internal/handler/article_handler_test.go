package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"yuanju/configs"
	"yuanju/internal/middleware"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/pkg/database"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func withArticleHandlerDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	pg, err := tcpg.Run(ctx,
		"postgres:16-alpine",
		tcpg.WithDatabase("yuanju_article_handler_test"),
		tcpg.WithUsername("test"),
		tcpg.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	origDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = origDB })

	configs.AppConfig.JWTSecret = "test-user-secret"
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	if _, err := database.Migrate(database.ModeApply); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func articleHandlerRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/articles/settings", GetArticleModuleSettings)
	articles := r.Group("/api/articles", middleware.Auth())
	articles.GET("", ListArticles)
	articles.GET("/:id", GetArticleDetail)
	articles.POST("/:id/original-click", TrackArticleOriginalClick)
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.GET("/articles/module-settings", AdminGetArticleModuleSettings)
	admin.PUT("/articles/module-settings", AdminUpdateArticleModuleSettings)
	admin.POST("/articles/collect", AdminTriggerArticleCollection)
	admin.POST("/articles/collection-tasks/:id/retry", AdminRetryArticleCollectionTask)
	admin.POST("/articles/batch-action", AdminArticleBatchAction)
	return r
}

func articleHandlerRouterWithContext(userID, adminID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	articles := r.Group("/api/articles", func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	r.GET("/api/articles/settings", GetArticleModuleSettings)
	articles.GET("", ListArticles)
	articles.GET("/:id", GetArticleDetail)
	articles.POST("/:id/original-click", TrackArticleOriginalClick)
	admin := r.Group("/api/admin", func(c *gin.Context) {
		c.Set("admin_id", adminID)
		c.Next()
	})
	admin.GET("/articles/module-settings", AdminGetArticleModuleSettings)
	admin.PUT("/articles/module-settings", AdminUpdateArticleModuleSettings)
	admin.POST("/articles/collect", AdminTriggerArticleCollection)
	admin.POST("/articles/batch-action", AdminArticleBatchAction)
	admin.PUT("/articles/:id/body", AdminUpdateArticleBody)
	admin.POST("/articles/:id/fetch-body", AdminFetchArticleBody)
	admin.GET("/articles/collection-tasks/:id/items", AdminListArticleCollectionTaskItems)
	admin.POST("/articles/collection-tasks/:id/retry", AdminRetryArticleCollectionTask)
	admin.GET("/articles/collection-config", AdminGetArticleCollectionConfig)
	admin.PUT("/articles/collection-config", AdminUpdateArticleCollectionConfig)
	admin.GET("/articles/quality-config", AdminGetArticleQualityConfig)
	admin.PUT("/articles/quality-config", AdminUpdateArticleQualityConfig)
	return r
}

func TestArticleListRequiresUserAuth(t *testing.T) {
	withArticleHandlerDB(t)
	r := articleHandlerRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestArticleModuleSettingsDefaultAndAdminUpdate(t *testing.T) {
	db := withArticleHandlerDB(t)
	adminID := insertArticleHandlerAdmin(t, db)
	r := articleHandlerRouterWithContext("", adminID)

	if _, err := db.Exec(`DELETE FROM system_settings WHERE key=$1`, repository.SettingArticlesModuleEnabled); err != nil {
		t.Fatalf("delete setting: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/articles/settings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("public settings status=%d body=%s", w.Code, w.Body.String())
	}
	var publicRes struct {
		ModuleEnabled bool `json:"module_enabled"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &publicRes); err != nil {
		t.Fatalf("decode public settings: %v", err)
	}
	if !publicRes.ModuleEnabled {
		t.Fatal("missing article module setting should default enabled")
	}

	req = httptest.NewRequest(http.MethodPut, "/api/admin/articles/module-settings", strings.NewReader(`{"module_enabled":false}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("admin update settings status=%d body=%s", w.Code, w.Body.String())
	}

	enabled, err := repository.GetBoolSetting(repository.SettingArticlesModuleEnabled, true)
	if err != nil {
		t.Fatalf("read saved setting: %v", err)
	}
	if enabled {
		t.Fatal("article module setting should be disabled after admin update")
	}
}

func TestArticleModuleSettingsRequireAdminAuth(t *testing.T) {
	withArticleHandlerDB(t)
	r := articleHandlerRouter()

	req := httptest.NewRequest(http.MethodPut, "/api/admin/articles/module-settings", strings.NewReader(`{"module_enabled":false}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("non-admin update status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestArticleModuleDisabledBlocksAdminCollectionActions(t *testing.T) {
	db := withArticleHandlerDB(t)
	adminID := insertArticleHandlerAdmin(t, db)
	r := articleHandlerRouterWithContext("", adminID)
	if err := repository.SetBoolSetting(repository.SettingArticlesModuleEnabled, false); err != nil {
		t.Fatalf("disable article module: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/articles/collect", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("manual collect status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "article_module_closed") {
		t.Fatalf("manual collect should return module closed code: %s", w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/articles/collection-tasks/task-id/retry", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("retry collect status=%d body=%s", w.Code, w.Body.String())
	}

	var taskCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM article_collection_tasks`).Scan(&taskCount); err != nil {
		t.Fatalf("count collection tasks: %v", err)
	}
	if taskCount != 0 {
		t.Fatalf("collection task count=%d, want 0", taskCount)
	}
}

func TestArticleDetailPublishedOnlyAndIncrementsView(t *testing.T) {
	db := withArticleHandlerDB(t)
	userID := insertArticleHandlerUser(t, db)
	article, _, err := repository.InsertArticleCandidate(repository.ArticleCandidateInput{
		Title:            "发布文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=900&idx=1&sn=def",
		CanonicalURLHash: "handler-detail-published",
	})
	if err != nil {
		t.Fatalf("insert article: %v", err)
	}
	if _, err := db.Exec(`UPDATE articles SET status='published' WHERE id=$1`, article.ID); err != nil {
		t.Fatalf("publish fixture: %v", err)
	}
	hidden, _, err := repository.InsertArticleCandidate(repository.ArticleCandidateInput{
		Title:            "隐藏文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=901&idx=1&sn=def",
		CanonicalURLHash: "handler-detail-hidden",
	})
	if err != nil {
		t.Fatalf("insert hidden: %v", err)
	}

	r := articleHandlerRouterWithContext(userID, "")
	req := httptest.NewRequest(http.MethodGet, "/api/articles/"+article.ID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("published detail status=%d body=%s", w.Code, w.Body.String())
	}
	var views int
	if err := db.QueryRow(`SELECT view_count FROM articles WHERE id=$1`, article.ID).Scan(&views); err != nil {
		t.Fatalf("read views: %v", err)
	}
	if views != 1 {
		t.Fatalf("view_count=%d, want 1", views)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/articles/"+hidden.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("hidden detail status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestArticleModuleDisabledBlocksUserAccessWithoutMutations(t *testing.T) {
	db := withArticleHandlerDB(t)
	userID := insertArticleHandlerUser(t, db)
	article, _, err := repository.InsertArticleCandidate(repository.ArticleCandidateInput{
		Title:            "发布文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=910&idx=1&sn=def",
		CanonicalURLHash: "handler-module-disabled",
	})
	if err != nil {
		t.Fatalf("insert article: %v", err)
	}
	if _, err := db.Exec(`UPDATE articles SET status='published' WHERE id=$1`, article.ID); err != nil {
		t.Fatalf("publish fixture: %v", err)
	}
	if err := repository.SetBoolSetting(repository.SettingArticlesModuleEnabled, false); err != nil {
		t.Fatalf("disable article module: %v", err)
	}

	r := articleHandlerRouterWithContext(userID, "")
	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/articles"},
		{http.MethodGet, "/api/articles/" + article.ID},
		{http.MethodPost, "/api/articles/" + article.ID + "/original-click"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("%s %s status=%d body=%s", tc.method, tc.path, w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "article_module_closed") {
			t.Fatalf("%s %s missing closed code: %s", tc.method, tc.path, w.Body.String())
		}
	}

	var views, clicks int
	if err := db.QueryRow(`SELECT view_count, original_click_count FROM articles WHERE id=$1`, article.ID).Scan(&views, &clicks); err != nil {
		t.Fatalf("read counters: %v", err)
	}
	if views != 0 || clicks != 0 {
		t.Fatalf("disabled module mutated counters: views=%d clicks=%d", views, clicks)
	}
}

func TestArticleOriginalClickRequiresPublishedArticle(t *testing.T) {
	db := withArticleHandlerDB(t)
	userID := insertArticleHandlerUser(t, db)
	article, _, err := repository.InsertArticleCandidate(repository.ArticleCandidateInput{
		Title:            "发布文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=902&idx=1&sn=def",
		CanonicalURLHash: "handler-click-published",
	})
	if err != nil {
		t.Fatalf("insert article: %v", err)
	}
	if _, err := db.Exec(`UPDATE articles SET status='published' WHERE id=$1`, article.ID); err != nil {
		t.Fatalf("publish fixture: %v", err)
	}
	r := articleHandlerRouterWithContext(userID, "")
	req := httptest.NewRequest(http.MethodPost, "/api/articles/"+article.ID+"/original-click", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("click status=%d body=%s", w.Code, w.Body.String())
	}
	var clicks int
	if err := db.QueryRow(`SELECT original_click_count FROM articles WHERE id=$1`, article.ID).Scan(&clicks); err != nil {
		t.Fatalf("read clicks: %v", err)
	}
	if clicks != 1 {
		t.Fatalf("original_click_count=%d, want 1", clicks)
	}
}

func TestAdminBatchPublishRequiresExplicitWithoutAIConfirmation(t *testing.T) {
	db := withArticleHandlerDB(t)
	adminID := insertArticleHandlerAdmin(t, db)
	article, _, err := repository.InsertArticleCandidate(repository.ArticleCandidateInput{
		Title:            "待发布文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=903&idx=1&sn=def",
		CanonicalURLHash: "handler-admin-publish",
	})
	if err != nil {
		t.Fatalf("insert article: %v", err)
	}
	r := articleHandlerRouterWithContext("", adminID)

	body, _ := json.Marshal(map[string]any{"ids": []string{article.ID}, "action": "publish"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/articles/batch-action", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("batch without confirmation status=%d body=%s", w.Code, w.Body.String())
	}
	if got := readArticleStatus(t, db, article.ID); got != model.ArticleStatusCandidate {
		t.Fatalf("status=%s, want candidate without confirmation", got)
	}

	body, _ = json.Marshal(map[string]any{"ids": []string{article.ID}, "action": "publish", "allow_publish_without_ai": true})
	req = httptest.NewRequest(http.MethodPost, "/api/admin/articles/batch-action", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("batch with confirmation status=%d body=%s", w.Code, w.Body.String())
	}
	if got := readArticleStatus(t, db, article.ID); got != model.ArticleStatusPublished {
		t.Fatalf("status=%s, want published with confirmation", got)
	}
}

func TestAdminUpdateArticleBodyStoresManualContent(t *testing.T) {
	db := withArticleHandlerDB(t)
	adminID := insertArticleHandlerAdmin(t, db)
	article, _, err := repository.InsertArticleCandidate(repository.ArticleCandidateInput{
		Title:            "待补正文文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://weixin.sogou.com/link?url=abc",
		CanonicalURLHash: "handler-admin-update-body",
	})
	if err != nil {
		t.Fatalf("insert article: %v", err)
	}
	r := articleHandlerRouterWithContext("", adminID)

	body, _ := json.Marshal(map[string]any{"body_content": "第一段正文\n第二段正文"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/articles/"+article.ID+"/body", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update body status=%d body=%s", w.Code, w.Body.String())
	}
	var stored string
	var authorized bool
	if err := db.QueryRow(`SELECT body_content, full_text_authorized FROM articles WHERE id=$1`, article.ID).Scan(&stored, &authorized); err != nil {
		t.Fatalf("read body: %v", err)
	}
	if stored != "第一段正文\n第二段正文" || !authorized {
		t.Fatalf("stored body=%q authorized=%v", stored, authorized)
	}
}

func TestAdminFetchArticleBodyStoresParsedContentAndURL(t *testing.T) {
	db := withArticleHandlerDB(t)
	adminID := insertArticleHandlerAdmin(t, db)
	article, _, err := repository.InsertArticleCandidate(repository.ArticleCandidateInput{
		Title:            "待抓正文文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://weixin.sogou.com/link?url=def",
		CanonicalURLHash: "handler-admin-fetch-body",
	})
	if err != nil {
		t.Fatalf("insert article: %v", err)
	}
	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><h1 id="activity-name">标题</h1><div id="js_content"><p>直接链接正文第一段</p><p>直接链接正文第二段</p></div></body></html>`))
	}))
	t.Cleanup(page.Close)
	r := articleHandlerRouterWithContext("", adminID)

	body, _ := json.Marshal(map[string]any{"url": page.URL + "/article"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/articles/"+article.ID+"/fetch-body", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("fetch body status=%d body=%s", w.Code, w.Body.String())
	}
	var stored, originalURL string
	var authorized bool
	if err := db.QueryRow(`SELECT body_content, original_url, full_text_authorized FROM articles WHERE id=$1`, article.ID).Scan(&stored, &originalURL, &authorized); err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(stored, "直接链接正文第一段") || !strings.Contains(stored, "直接链接正文第二段") || !authorized {
		t.Fatalf("stored body=%q authorized=%v", stored, authorized)
	}
	if originalURL != page.URL+"/article" {
		t.Fatalf("original_url=%q, want fetched url", originalURL)
	}
}

func TestAdminListArticleCollectionTaskItemsReturnsCollectedArticleData(t *testing.T) {
	db := withArticleHandlerDB(t)
	adminID := insertArticleHandlerAdmin(t, db)
	article, _, err := repository.InsertArticleCandidate(repository.ArticleCandidateInput{
		Title:            "任务明细文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=906&idx=1&sn=def",
		CanonicalURLHash: "handler-task-items",
		QualityScore:     86,
		QualityReasons:   []model.ArticleQualityReason{{Type: "rank", Points: 20, Message: "搜索排序靠前"}},
		BodyFetchStatus:  model.ArticleBodyFetchStatusSucceeded,
	})
	if err != nil {
		t.Fatalf("insert article: %v", err)
	}
	task, err := repository.CreateArticleCollectionTask("manual", 1)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := repository.AddArticleCollectionTaskItemWithQuality(repository.ArticleCollectionTaskItemInput{
		TaskID:            task.ID,
		ArticleID:         article.ID,
		OriginalURL:       article.OriginalURL,
		Keyword:           "八字",
		Status:            model.CollectionItemStatusInserted,
		QualityScore:      86,
		QualityReasons:    []model.ArticleQualityReason{{Type: "rank", Points: 20, Message: "搜索排序靠前"}},
		AutoPublished:     true,
		AutoPublishReason: "定时采集自动发布：质量分 86 >= 阈值 80，正文已获取",
	}); err != nil {
		t.Fatalf("add task item: %v", err)
	}

	r := articleHandlerRouterWithContext("", adminID)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/articles/collection-tasks/"+task.ID+"/items", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("task items status=%d body=%s", w.Code, w.Body.String())
	}
	var res struct {
		Items []model.ArticleCollectionTaskItem `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(res.Items) != 1 {
		t.Fatalf("items=%d, want 1", len(res.Items))
	}
	got := res.Items[0]
	if got.ArticleTitle != "任务明细文章" || got.SourceName != "命理参考" || got.Keyword != "八字" || got.Status != model.CollectionItemStatusInserted {
		t.Fatalf("unexpected task item: %+v", got)
	}
	if got.QualityScore != 86 || len(got.QualityReasons) != 1 || got.QualityReasons[0].Message != "搜索排序靠前" {
		t.Fatalf("quality metadata not returned: %+v", got)
	}
	if !got.AutoPublished || !strings.Contains(got.AutoPublishReason, "自动发布") {
		t.Fatalf("auto publish metadata not returned: %+v", got)
	}
}

func TestAdminArticleCollectionConfigAutoPublishUpdate(t *testing.T) {
	db := withArticleHandlerDB(t)
	adminID := insertArticleHandlerAdmin(t, db)
	r := articleHandlerRouterWithContext("", adminID)

	body := `{"enabled":true,"frequency":"daily","schedule_interval_minutes":30,"max_results_per_run":12,"search_page_min":2,"search_page_max":4,"auto_publish_enabled":true,"auto_publish_min_quality_score":88,"auto_publish_requires_body":true,"auto_publish_max_per_run":2}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/articles/collection-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update collection config status=%d body=%s", w.Code, w.Body.String())
	}
	var updateRes struct {
		Config model.ArticleCollectionConfig `json:"config"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &updateRes); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if !updateRes.Config.AutoPublishEnabled || updateRes.Config.AutoPublishMinQualityScore != 88 || !updateRes.Config.AutoPublishRequiresBody || updateRes.Config.AutoPublishMaxPerRun != 2 {
		t.Fatalf("unexpected saved auto publish config: %+v", updateRes.Config)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/admin/articles/collection-config", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get collection config status=%d body=%s", w.Code, w.Body.String())
	}
	var getRes struct {
		Config model.ArticleCollectionConfig `json:"config"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &getRes); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if getRes.Config.AutoPublishMinQualityScore != 88 || getRes.Config.AutoPublishMaxPerRun != 2 {
		t.Fatalf("unexpected fetched auto publish config: %+v", getRes.Config)
	}
}

func TestAdminArticleQualityConfigUpdate(t *testing.T) {
	db := withArticleHandlerDB(t)
	adminID := insertArticleHandlerAdmin(t, db)
	r := articleHandlerRouterWithContext("", adminID)

	body := `{"quality_filter_enabled":true,"min_quality_score":75,"allow_without_body":false,"bonus_keywords":["财运"],"source_blacklist":["低质号"],"preferred_sources":["命理参考"],"ai_quality_check_enabled":false}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/articles/quality-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update quality config status=%d body=%s", w.Code, w.Body.String())
	}
	var updateRes struct {
		Config model.ArticleQualityConfig `json:"config"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &updateRes); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if !updateRes.Config.Enabled || updateRes.Config.MinQualityScore != 75 || updateRes.Config.AllowWithoutBody {
		t.Fatalf("unexpected saved quality config: %+v", updateRes.Config)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/admin/articles/quality-config", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get quality config status=%d body=%s", w.Code, w.Body.String())
	}
	var getRes struct {
		Config model.ArticleQualityConfig `json:"config"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &getRes); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if getRes.Config.MinQualityScore != 75 || len(getRes.Config.PreferredSources) != 1 || getRes.Config.PreferredSources[0] != "命理参考" {
		t.Fatalf("unexpected fetched quality config: %+v", getRes.Config)
	}
}

func TestAdminArticleTaxonomyAndKeywordCRUD(t *testing.T) {
	withArticleHandlerDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/categories", AdminCreateArticleCategory)
	r.PUT("/categories/:id", AdminUpdateArticleCategory)
	r.POST("/tags", AdminCreateArticleTag)
	r.PUT("/tags/:id", AdminUpdateArticleTag)
	r.POST("/keywords", AdminCreateArticleKeyword)
	r.PUT("/keywords/:id", AdminUpdateArticleKeyword)

	categoryID := postAndReadID(t, r, "/categories", `{"name":"八字入门"}`, "category")
	putJSON(t, r, "/categories/"+categoryID, `{"name":"八字基础","active":false}`)
	tagID := postAndReadID(t, r, "/tags", `{"name":"格局"}`, "tag")
	putJSON(t, r, "/tags/"+tagID, `{"name":"格局法","active":false}`)
	keywordID := postAndReadID(t, r, "/keywords", `{"keyword":"八字格局"}`, "keyword")
	putJSON(t, r, "/keywords/"+keywordID, `{"keyword":"八字格局案例","active":false}`)
}

func insertArticleHandlerUser(t *testing.T, db *sql.DB) string {
	t.Helper()
	var id string
	if err := db.QueryRow(`INSERT INTO users (email, password_hash) VALUES ('article-handler-user@example.com', 'hash') RETURNING id`).Scan(&id); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func insertArticleHandlerAdmin(t *testing.T, db *sql.DB) string {
	t.Helper()
	var id string
	if err := db.QueryRow(`INSERT INTO admins (email, password_hash, name) VALUES ('article-handler-admin@example.com', 'hash', 'Admin') RETURNING id`).Scan(&id); err != nil {
		t.Fatalf("insert admin: %v", err)
	}
	return id
}

func readArticleStatus(t *testing.T, db *sql.DB, articleID string) model.ArticleStatus {
	t.Helper()
	var status model.ArticleStatus
	if err := db.QueryRow(`SELECT status FROM articles WHERE id=$1`, articleID).Scan(&status); err != nil {
		t.Fatalf("read status: %v", err)
	}
	return status
}

func postAndReadID(t *testing.T, r *gin.Engine, path, body, key string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST %s status=%d body=%s", path, w.Code, w.Body.String())
	}
	var payload map[string]map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	id, _ := payload[key]["id"].(string)
	if id == "" {
		t.Fatalf("response missing %s.id: %s", key, w.Body.String())
	}
	return id
}

func putJSON(t *testing.T, r *gin.Engine, path, body string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("PUT %s status=%d body=%s", path, w.Code, w.Body.String())
	}
}

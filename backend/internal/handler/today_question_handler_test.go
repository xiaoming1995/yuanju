package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
	"yuanju/configs"
	"yuanju/internal/middleware"
	"yuanju/internal/repository"
	"yuanju/internal/service"
	"yuanju/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func mockTodayQuestionOutput(ctx *service.TodayQuestionContext) string {
	support, caution := []string{}, []string{}
	for _, evidence := range ctx.Evidence {
		if (evidence.Direction == "candidate" || evidence.Direction == "favorable") && len(support) < 4 {
			support = append(support, evidence.ID)
		}
		if (evidence.Direction == "review" || evidence.Direction == "adverse") && len(caution) < 4 {
			caution = append(caution, evidence.ID)
		}
	}
	level := "参考有限"
	if len(support) > 0 && len(caution) > 0 {
		level = "信号交错"
	} else if len(support) > 0 {
		level = "偏有利"
	} else if len(caution) > 0 {
		level = "需留意"
	}
	value, _ := json.Marshal(map[string]any{"level": level, "summary": "命盘整体方向仅作间接参考，不能据此判断这件事的实际结果。", "support_ids": support, "caution_ids": caution, "advice": []string{"列清今天要做的具体步骤", "核对时间和实际限制条件"}})
	return string(value)
}

func TestTodayQuestionRejectsInvalidAndUnauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/user/today-question", func(c *gin.Context) { c.Set("user_id", "test-user"); c.Next() }, PostTodayQuestion)
	for _, body := range []string{`{"question":"短"}`, `{"question":"今天应该买入股票吗"}`, `{}`} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/user/today-question", strings.NewReader(body)))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("invalid request got %d: %s", w.Code, w.Body.String())
		}
	}
	protected := gin.New()
	protected.POST("/api/user/today-question", middleware.Auth(), PostTodayQuestion)
	w := httptest.NewRecorder()
	protected.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/user/today-question", strings.NewReader(`{"question":"今天下午有一场面试"}`)))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated request got %d", w.Code)
	}
}

func TestTodayQuestionBoundUserTodayWithMockProvider(t *testing.T) {
	_ = withArticleHandlerDB(t)
	if err := repository.SetTodayQuestionDailyLimit(10); err != nil {
		t.Fatal(err)
	}
	oldConfig := configs.AppConfig
	oldNow := todayQuestionNow
	t.Cleanup(func() { configs.AppConfig = oldConfig; todayQuestionNow = oldNow })
	todayQuestionNow = func() time.Time { return time.Date(2026, 11, 5, 8, 0, 0, 0, time.UTC) }
	var owner, other string
	for _, id := range []*string{&owner, &other} {
		if err := database.DB.QueryRow(`INSERT INTO users (email,password_hash) VALUES (gen_random_uuid()::text || '@example.com','hash') RETURNING id`).Scan(id); err != nil {
			t.Fatal(err)
		}
	}
	ownerAPI := gin.New()
	ownerAPI.POST("/api/user/personal-chart", func(c *gin.Context) { c.Set("user_id", owner); c.Next() }, SavePersonalChart)
	ownerAPI.POST("/api/user/today-question", func(c *gin.Context) { c.Set("user_id", owner); c.Next() }, PostTodayQuestion)
	otherAPI := gin.New()
	otherAPI.POST("/api/user/today-question", func(c *gin.Context) { c.Set("user_id", other); c.Next() }, PostTodayQuestion)
	setup := httptest.NewRecorder()
	ownerAPI.ServeHTTP(setup, httptest.NewRequest(http.MethodPost, "/api/user/personal-chart", strings.NewReader(`{"year":1990,"month":1,"day":1,"hour":12,"gender":"male"}`)))
	if setup.Code != http.StatusOK {
		t.Fatalf("setup: %d %s", setup.Code, setup.Body.String())
	}
	var saved struct {
		Chart struct {
			ID string `json:"id"`
		} `json:"chart"`
	}
	if err := json.Unmarshal(setup.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	chart, err := repository.GetPersonalChart(owner)
	if err != nil || chart == nil {
		t.Fatalf("bound chart: %v", err)
	}
	result, err := service.LoadOrCalculateResult(chart)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := service.BuildTodayQuestionContext("2026-11-05", chart.ID, "今天下午二面，担心临场表达", result)
	if err != nil {
		t.Fatal(err)
	}
	support, caution := []string{}, []string{}
	for _, e := range ctx.Evidence {
		if (e.Direction == "candidate" || e.Direction == "favorable") && len(support) < 4 {
			support = append(support, e.ID)
		}
		if (e.Direction == "review" || e.Direction == "adverse") && len(caution) < 4 {
			caution = append(caution, e.ID)
		}
	}
	if len(support)+len(caution) == 0 {
		t.Fatal("test chart lacks directional evidence")
	}
	level := "信号交错"
	if len(caution) == 0 {
		level = "偏有利"
	} else if len(support) == 0 {
		level = "需留意"
	}
	content, _ := json.Marshal(map[string]any{"level": level, "summary": "当前有具体的面试命理参考，仍应认真准备表达。", "support_ids": support, "caution_ids": caution, "advice": []string{"整理项目案例和职责", "核对岗位要求及面试时间"}})
	providerFail := false
	providerInvalid := false
	providerCalls := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerCalls++
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if providerFail {
			http.Error(w, "model unavailable", http.StatusBadGateway)
			return
		}
		responseContent := string(content)
		if providerInvalid {
			responseContent = "{}"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": responseContent}, "finish_reason": "stop"}}, "usage": map[string]int{"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}})
	}))
	defer provider.Close()
	configs.AppConfig.DeepSeekAPIKey = "test"
	configs.AppConfig.OpenAIAPIKey = ""
	configs.AppConfig.AIBaseURL = provider.URL
	configs.AppConfig.AIPromptLog = false
	requestBody := `{"question":"今天下午二面，担心临场表达","date":"2030-01-01","chart_id":"foreign","category":"travel"}`
	otherW := httptest.NewRecorder()
	otherAPI.ServeHTTP(otherW, httptest.NewRequest(http.MethodPost, "/api/user/today-question", strings.NewReader(requestBody)))
	if otherW.Code != http.StatusConflict {
		t.Fatalf("unbound user got %d", otherW.Code)
	}
	w := httptest.NewRecorder()
	ownerAPI.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/user/today-question", strings.NewReader(requestBody)))
	if w.Code != http.StatusOK {
		t.Fatalf("bound request: %d %s", w.Code, w.Body.String())
	}
	var body struct {
		Data service.TodayQuestionAnswer `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.Date != "2026-11-05" || body.Data.ChartID != saved.Chart.ID || body.Data.Category != "interview" || body.Data.Level != level || len(body.Data.Layers) < 2 {
		t.Fatalf("wrong bound result: %+v", body.Data)
	}
	adminAPI := gin.New()
	admin := adminAPI.Group("/api/admin", middleware.AdminAuth())
	admin.GET("/today-questions", AdminListTodayQuestions)
	admin.GET("/today-questions/:id", AdminGetTodayQuestion)
	unauthorized := httptest.NewRecorder()
	adminAPI.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/admin/today-questions", nil))
	if unauthorized.Code != http.StatusUnauthorized || strings.Contains(unauthorized.Body.String(), "二面") {
		t.Fatalf("admin route exposed private data: %d %s", unauthorized.Code, unauthorized.Body.String())
	}
	adminToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "yuanju-admin", "sub": "test-admin", "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(configs.AppConfig.AdminJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	adminGet := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("Authorization", "Bearer "+adminToken)
		response := httptest.NewRecorder()
		adminAPI.ServeHTTP(response, request)
		return response
	}
	for _, path := range []string{"/api/admin/today-questions?page=0", "/api/admin/today-questions?from=bad", "/api/admin/today-questions?status=unknown", "/api/admin/today-questions/not-a-uuid"} {
		if response := adminGet(path); response.Code != http.StatusBadRequest {
			t.Fatalf("invalid filter %s got %d", path, response.Code)
		}
	}
	list := adminGet("/api/admin/today-questions?status=success&from=2026-11-05&to=2026-11-05&user=" + owner + "&page=1&limit=1")
	if list.Code != http.StatusOK || strings.Contains(list.Body.String(), "ai_user_prompt") || strings.Contains(list.Body.String(), "context_data") {
		t.Fatalf("list status/privacy: %d %s", list.Code, list.Body.String())
	}
	var listed struct {
		Data struct {
			Total int                              `json:"total"`
			Items []repository.TodayQuestionRecord `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil || listed.Data.Total != 1 || len(listed.Data.Items) != 1 {
		t.Fatalf("filtered list: %+v %v", listed, err)
	}
	detail := adminGet("/api/admin/today-questions/" + listed.Data.Items[0].ID)
	var detailed struct {
		Data repository.TodayQuestionRecord `json:"data"`
	}
	if err := json.Unmarshal(detail.Body.Bytes(), &detailed); err != nil || detail.Code != http.StatusOK ||
		detailed.Data.Question != "今天下午二面，担心临场表达" || detailed.Data.ChartID != chart.ID ||
		detailed.Data.TotalTokens != 30 || detailed.Data.PromptTokens != 10 || detailed.Data.CompletionTokens != 20 ||
		len(detailed.Data.ContextData) == 0 || len(detailed.Data.AnswerData) == 0 ||
		!strings.Contains(detailed.Data.AIUserPrompt, "以下 JSON") || !strings.Contains(detailed.Data.AIOutput, "support_ids") {
		t.Fatalf("missing request/AI detail: %+v, %v", detailed.Data, err)
	}
	var usageID string
	if err := database.DB.QueryRow(`SELECT token_usage_id FROM today_question_requests WHERE id=$1`, listed.Data.Items[0].ID).Scan(&usageID); err != nil {
		t.Fatalf("today question not linked to shared token usage: %v", err)
	}
	usageInput, usageOutput, err := repository.GetTokenUsageContent(usageID)
	if err != nil || !strings.Contains(usageInput, "以下 JSON") || !strings.Contains(usageOutput, "support_ids") {
		t.Fatalf("existing token detail lost today question content: %v", err)
	}
	providerInvalid = true
	invalidOutput := httptest.NewRecorder()
	ownerAPI.ServeHTTP(invalidOutput, httptest.NewRequest(http.MethodPost, "/api/user/today-question", strings.NewReader(requestBody)))
	if invalidOutput.Code != http.StatusBadGateway {
		t.Fatalf("invalid model output got %d", invalidOutput.Code)
	}
	invalidList := adminGet("/api/admin/today-questions?status=error&user=" + owner)
	var invalidRows struct {
		Data struct {
			Items []repository.TodayQuestionRecord `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(invalidList.Body.Bytes(), &invalidRows); err != nil || len(invalidRows.Data.Items) != 1 {
		t.Fatalf("invalid output not listed: %s %v", invalidList.Body.String(), err)
	}
	invalidDetail := adminGet("/api/admin/today-questions/" + invalidRows.Data.Items[0].ID)
	var invalidRecord struct {
		Data repository.TodayQuestionRecord `json:"data"`
	}
	if err := json.Unmarshal(invalidDetail.Body.Bytes(), &invalidRecord); err != nil || invalidRecord.Data.AIOutput != "{}" || invalidRecord.Data.TotalTokens != 30 || len(invalidRecord.Data.AnswerData) != 0 {
		t.Fatalf("invalid output/usage not retained: %+v %v", invalidRecord.Data, err)
	}
	providerInvalid = false
	providerFail = true
	failed := httptest.NewRecorder()
	ownerAPI.ServeHTTP(failed, httptest.NewRequest(http.MethodPost, "/api/user/today-question", strings.NewReader(requestBody)))
	if failed.Code != http.StatusBadGateway || strings.Contains(failed.Body.String(), "model unavailable") {
		t.Fatalf("provider failure leaked or misreported: %d %s", failed.Code, failed.Body.String())
	}
	providerFail = false
	generalContext, err := service.BuildTodayQuestionContext("2026-11-05", chart.ID, "今天能做成这件事吗", result)
	if err != nil {
		t.Fatal(err)
	}
	content = []byte(mockTodayQuestionOutput(generalContext))
	beforeGeneralCalls := providerCalls
	general := httptest.NewRecorder()
	ownerAPI.ServeHTTP(general, httptest.NewRequest(http.MethodPost, "/api/user/today-question", strings.NewReader(`{"question":"今天能做成这件事吗"}`)))
	if general.Code != http.StatusOK || strings.Contains(general.Body.String(), "依据不足") || providerCalls != beforeGeneralCalls+1 {
		t.Fatalf("general response: %d %s", general.Code, general.Body.String())
	}
	var generalPromptLength, generalTokens int
	if err := database.DB.QueryRow(`SELECT length(COALESCE(t.input_content,q.ai_user_prompt)), COALESCE(t.total_tokens,0)
		FROM today_question_requests q LEFT JOIN token_usage_logs t ON t.id=q.token_usage_id
		WHERE q.user_id=$1 AND q.question='今天能做成这件事吗' AND q.status='success'
		ORDER BY q.created_at DESC LIMIT 1`, owner).Scan(&generalPromptLength, &generalTokens); err != nil || generalPromptLength == 0 || generalTokens != 30 {
		t.Fatalf("general request skipped AI or token audit: prompt=%d tokens=%d err=%v", generalPromptLength, generalTokens, err)
	}
	content = []byte(`{"level":"偏有利","summary":"今天这件事有明确的个人支持依据，可以顺利完成。","support_ids":[],"caution_ids":[],"advice":["列清今天要做的具体步骤","核对时间和实际限制条件"]}`)
	unsupported := httptest.NewRecorder()
	ownerAPI.ServeHTTP(unsupported, httptest.NewRequest(http.MethodPost, "/api/user/today-question", strings.NewReader(`{"question":"今天能做成这件事吗"}`)))
	if unsupported.Code != http.StatusBadGateway {
		t.Fatalf("unsupported favorable answer accepted: %d %s", unsupported.Code, unsupported.Body.String())
	}
	var refunded bool
	if err := database.DB.QueryRow(`SELECT NOT quota_charged FROM today_question_requests
		WHERE user_id=$1 AND question='今天能做成这件事吗' AND status='error'
		ORDER BY created_at DESC LIMIT 1`, owner).Scan(&refunded); err != nil || !refunded {
		t.Fatalf("invalid no-evidence answer did not refund quota: %v %v", refunded, err)
	}
	for status, expected := range map[string]string{"error": "AI 调用失败", "success": "今天能做成这件事吗"} {
		response := adminGet("/api/admin/today-questions?status=" + status + "&user=" + owner)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), expected) {
			t.Fatalf("%s request not recorded: %d %s", status, response.Code, response.Body.String())
		}
	}
	otherID, _, err := repository.ReserveTodayQuestionRequest(other, chart.ID, "2026-11-05", "general", "今天还有其他安排吗", []byte(`{"evidence":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.CompleteTodayQuestionRequest(otherID, repository.TodayQuestionCompletion{Status: "insufficient"}); err != nil {
		t.Fatal(err)
	}
	ownerList := adminGet("/api/admin/today-questions?user=" + owner)
	if strings.Contains(ownerList.Body.String(), "今天还有其他安排吗") {
		t.Fatal("user filter included another user's request")
	}
	pageTwo := adminGet("/api/admin/today-questions?page=2&limit=1")
	if pageTwo.Code != http.StatusOK || strings.Contains(pageTwo.Body.String(), "ai_user_prompt") {
		t.Fatalf("pagination/privacy failed: %d %s", pageTwo.Code, pageTwo.Body.String())
	}
	var usageRows, linkedRows, duplicateSnapshot int
	if err := database.DB.QueryRow(`SELECT count(*) FROM token_usage_logs WHERE call_type='today_question'`).Scan(&usageRows); err != nil {
		t.Fatal(err)
	}
	if err := database.DB.QueryRow(`SELECT count(*) FILTER (WHERE token_usage_id IS NOT NULL), count(*) FILTER (WHERE total_tokens<>0) FROM today_question_requests WHERE user_id=$1`, owner).Scan(&linkedRows, &duplicateSnapshot); err != nil {
		t.Fatal(err)
	}
	if usageRows != 4 || linkedRows != 4 || duplicateSnapshot != 0 {
		t.Fatalf("shared token usage not single-sourced: usage=%d linked=%d duplicate=%d", usageRows, linkedRows, duplicateSnapshot)
	}
	if _, err := database.DB.Exec(`UPDATE token_usage_logs SET created_at='2001-01-02' WHERE id=$1`, usageID); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.RollupClosedMonthsAndDelete(); err != nil {
		t.Fatalf("monthly token rollup failed: %v", err)
	}
	archived := adminGet("/api/admin/today-questions/" + listed.Data.Items[0].ID)
	var archivedRecord struct {
		Data repository.TodayQuestionRecord `json:"data"`
	}
	if err := json.Unmarshal(archived.Body.Bytes(), &archivedRecord); err != nil || archivedRecord.Data.TotalTokens != 30 ||
		!strings.Contains(archivedRecord.Data.AIUserPrompt, "以下 JSON") || !strings.Contains(archivedRecord.Data.AIOutput, "support_ids") {
		t.Fatalf("monthly rollup lost question audit: %+v %v", archivedRecord.Data, err)
	}
}

func TestTodayQuestionDailyQuotaAndSettings(t *testing.T) {
	_ = withArticleHandlerDB(t)
	oldConfig := configs.AppConfig
	t.Cleanup(func() { configs.AppConfig = oldConfig })
	providerOutput := ""
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": providerOutput}, "finish_reason": "stop"}}, "usage": map[string]int{"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}})
	}))
	defer provider.Close()
	configs.AppConfig.DeepSeekAPIKey = "test"
	configs.AppConfig.OpenAIAPIKey = ""
	configs.AppConfig.AIBaseURL = provider.URL
	oldNow := todayQuestionNow
	t.Cleanup(func() { todayQuestionNow = oldNow })
	todayQuestionNow = func() time.Time { return time.Date(2026, 11, 5, 15, 59, 0, 0, time.UTC) }
	var userID string
	if err := database.DB.QueryRow(`INSERT INTO users (email,password_hash) VALUES (gen_random_uuid()::text || '@example.com','hash') RETURNING id`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	userAPI := gin.New()
	userAPI.POST("/api/user/personal-chart", func(c *gin.Context) { c.Set("user_id", userID); c.Next() }, SavePersonalChart)
	userAPI.POST("/api/user/today-question", func(c *gin.Context) { c.Set("user_id", userID); c.Next() }, PostTodayQuestion)
	setup := httptest.NewRecorder()
	userAPI.ServeHTTP(setup, httptest.NewRequest(http.MethodPost, "/api/user/personal-chart", strings.NewReader(`{"year":1990,"month":1,"day":1,"hour":12,"gender":"male"}`)))
	if setup.Code != http.StatusOK {
		t.Fatalf("setup: %d %s", setup.Code, setup.Body.String())
	}
	var chartID string
	if err := database.DB.QueryRow(`SELECT chart_id FROM user_personal_charts WHERE user_id=$1`, userID).Scan(&chartID); err != nil {
		t.Fatal(err)
	}
	bound, err := repository.GetPersonalChart(userID)
	if err != nil || bound == nil {
		t.Fatalf("bound chart: %v", err)
	}
	calculated, err := service.LoadOrCalculateResult(bound)
	if err != nil {
		t.Fatal(err)
	}
	setOutput := func(date string) {
		t.Helper()
		ctx, err := service.BuildTodayQuestionContext(date, chartID, "今天能做成这件事吗", calculated)
		if err != nil {
			t.Fatal(err)
		}
		providerOutput = mockTodayQuestionOutput(ctx)
	}
	setOutput("2026-11-05")
	if _, err := database.DB.Exec(`INSERT INTO today_question_requests
		(user_id,chart_id,question_date,category,question,context_data,status)
		VALUES ($1,$2,'2026-11-05','general','旧请求','{}','legacy')`, userID, chartID); err != nil {
		t.Fatal(err)
	}
	question := `{"question":"今天能做成这件事吗"}`
	post := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		userAPI.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/user/today-question", strings.NewReader(question)))
		return w
	}
	if first := post(); first.Code != http.StatusOK || strings.Contains(first.Body.String(), "依据不足") {
		t.Fatalf("first request: %d %s", first.Code, first.Body.String())
	}
	if second := post(); second.Code != http.StatusTooManyRequests || !strings.Contains(second.Body.String(), "明天 0 点") {
		t.Fatalf("second request: %d %s", second.Code, second.Body.String())
	}
	var charged int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM today_question_requests WHERE user_id=$1 AND question_date='2026-11-05' AND quota_charged`, userID).Scan(&charged); err != nil || charged != 1 {
		t.Fatalf("general answer must count once: %d %v", charged, err)
	}
	todayQuestionNow = func() time.Time { return time.Date(2026, 11, 5, 16, 1, 0, 0, time.UTC) }
	setOutput("2026-11-06")
	if nextDay := post(); nextDay.Code != http.StatusOK {
		t.Fatalf("next Beijing day: %d %s", nextDay.Code, nextDay.Body.String())
	}

	adminAPI := gin.New()
	admin := adminAPI.Group("/api/admin", middleware.AdminAuth())
	admin.GET("/today-questions/settings", AdminGetTodayQuestionSettings)
	admin.PUT("/today-questions/settings", AdminUpdateTodayQuestionSettings)
	requestSetting := func(method, body, token string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/api/admin/today-questions/settings", strings.NewReader(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		adminAPI.ServeHTTP(w, req)
		return w
	}
	if denied := requestSetting(http.MethodPut, `{"daily_limit":2}`, ""); denied.Code != http.StatusUnauthorized {
		t.Fatalf("public setting update: %d", denied.Code)
	}
	if denied := requestSetting(http.MethodGet, "", ""); denied.Code != http.StatusUnauthorized {
		t.Fatalf("public setting read: %d", denied.Code)
	}
	adminToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "yuanju-admin", "sub": "test-admin", "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(configs.AppConfig.AdminJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	if current := requestSetting(http.MethodGet, "", adminToken); current.Code != http.StatusOK || !strings.Contains(current.Body.String(), `"daily_limit":1`) {
		t.Fatalf("default setting: %d %s", current.Code, current.Body.String())
	}
	for _, body := range []string{`{"daily_limit":0}`, `{"daily_limit":101}`, `{"daily_limit":1.5}`} {
		if invalid := requestSetting(http.MethodPut, body, adminToken); invalid.Code != http.StatusBadRequest {
			t.Fatalf("invalid setting %s: %d", body, invalid.Code)
		}
	}
	if updated := requestSetting(http.MethodPut, `{"daily_limit":2}`, adminToken); updated.Code != http.StatusOK || !strings.Contains(updated.Body.String(), `"daily_limit":2`) {
		t.Fatalf("update setting: %d %s", updated.Code, updated.Body.String())
	}
	if more := post(); more.Code != http.StatusOK {
		t.Fatalf("second request under N=2: %d %s", more.Code, more.Body.String())
	}
	if over := post(); over.Code != http.StatusTooManyRequests {
		t.Fatalf("third request under N=2: %d %s", over.Code, over.Body.String())
	}
	if err := repository.SetTodayQuestionDailyLimit(1); err != nil {
		t.Fatal(err)
	}
	chart, err := repository.GetPersonalChart(userID)
	if err != nil || chart == nil {
		t.Fatalf("bound chart: %v", err)
	}
	reserve := func(date string) (string, error) {
		id, _, err := repository.ReserveTodayQuestionRequest(userID, chart.ID, date, "general", "今天还有别的安排吗", []byte(`{"evidence":[]}`))
		return id, err
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := reserve("2026-11-07")
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	accepted, denied := 0, 0
	for err := range results {
		if err == nil {
			accepted++
		} else if errors.Is(err, repository.ErrTodayQuestionDailyLimit) {
			denied++
		} else {
			t.Fatalf("concurrent reserve: %v", err)
		}
	}
	if accepted != 1 || denied != 1 {
		t.Fatalf("concurrent quota: accepted=%d denied=%d", accepted, denied)
	}
	failedID, err := reserve("2026-11-08")
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.CompleteTodayQuestionRequest(failedID, repository.TodayQuestionCompletion{Status: "error", ErrorMsg: "mock AI failed"}); err != nil {
		t.Fatal(err)
	}
	if _, err := reserve("2026-11-08"); err != nil {
		t.Fatalf("technical failure should refund: %v", err)
	}
	staleID, err := reserve("2026-11-09")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`UPDATE today_question_requests SET created_at=NOW()-INTERVAL '11 minutes' WHERE id=$1`, staleID); err != nil {
		t.Fatal(err)
	}
	if _, err := reserve("2026-11-09"); err != nil {
		t.Fatalf("stale processing should release quota: %v", err)
	}
}

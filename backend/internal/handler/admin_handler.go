package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"yuanju/configs"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/internal/service"
	"yuanju/pkg/crypto"
	"yuanju/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ====== Admin Auth ======

func AdminRegister(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Name     string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing, _ := repository.GetAdminByEmail(req.Email)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱已注册"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	admin, err := repository.CreateAdmin(req.Email, string(hash), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	token, _ := generateAdminToken(admin)
	c.JSON(http.StatusCreated, gin.H{"token": token, "admin": admin})
}

func AdminLogin(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	admin, err := repository.GetAdminByEmail(req.Email)
	if err != nil || admin == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
		return
	}
	token, _ := generateAdminToken(admin)
	c.JSON(http.StatusOK, gin.H{"token": token, "admin": admin})
}

func generateAdminToken(admin *model.Admin) (string, error) {
	claims := jwt.MapClaims{
		"sub":   admin.ID,
		"email": admin.Email,
		"iss":   "yuanju-admin",
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(configs.AppConfig.AdminJWTSecret))
}

// ====== Admin LLM Providers ======

func AdminListProviders(c *gin.Context) {
	providers, err := repository.ListLLMProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	for i := range providers {
		providers[i].APIKeyMasked = providers[i].APIKeyPreview
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers, "predefined": model.PredefinedProviders})
}

func AdminCreateProvider(c *gin.Context) {
	var req struct {
		Name            string  `json:"name" binding:"required"`
		Type            string  `json:"type" binding:"required"`
		BaseURL         string  `json:"base_url" binding:"required"`
		Model           string  `json:"model" binding:"required"`
		APIKey          string  `json:"api_key" binding:"required"`
		ThinkingEnabled bool    `json:"thinking_enabled"`
		InputPriceCny   float64 `json:"input_price_cny"`
		OutputPriceCny  float64 `json:"output_price_cny"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.InputPriceCny == 0 {
		req.InputPriceCny = 1.0
	}
	if req.OutputPriceCny == 0 {
		req.OutputPriceCny = 2.0
	}
	preview := crypto.MaskPlainKey(req.APIKey)
	encrypted, err := crypto.Encrypt(req.APIKey, configs.AppConfig.AdminEncryptionKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Key 加密失败"})
		return
	}
	p, err := repository.CreateLLMProvider(req.Name, req.Type, req.BaseURL, req.Model, encrypted, preview, req.ThinkingEnabled, req.InputPriceCny, req.OutputPriceCny)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败: " + err.Error()})
		return
	}
	p.APIKeyMasked = preview
	c.JSON(http.StatusCreated, gin.H{"provider": p})
}

func AdminUpdateProvider(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name            string  `json:"name"`
		BaseURL         string  `json:"base_url"`
		Model           string  `json:"model"`
		APIKey          string  `json:"api_key"` // 可选，不传则不更新
		ThinkingEnabled bool    `json:"thinking_enabled"`
		InputPriceCny   float64 `json:"input_price_cny"`
		OutputPriceCny  float64 `json:"output_price_cny"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 如果未传新 key，保留旧的（需要先查）
	providers, _ := repository.ListLLMProviders()
	var oldEncrypted string
	for _, p := range providers {
		if p.ID == id {
			oldEncrypted = p.APIKeyEncrypted
			break
		}
	}
	encrypted := oldEncrypted
	preview := ""
	if req.APIKey != "" {
		var err error
		preview = crypto.MaskPlainKey(req.APIKey)
		encrypted, err = crypto.Encrypt(req.APIKey, configs.AppConfig.AdminEncryptionKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Key 加密失败"})
			return
		}
	}
	if err := repository.UpdateLLMProvider(id, req.Name, req.BaseURL, req.Model, encrypted, preview, req.ThinkingEnabled, req.InputPriceCny, req.OutputPriceCny); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已更新"})
}

func AdminActivateProvider(c *gin.Context) {
	id := c.Param("id")
	if !repository.LLMProviderExists(id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider 不存在"})
		return
	}
	if err := repository.ActivateLLMProvider(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "切换失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已激活"})
}

func AdminDeleteProvider(c *gin.Context) {
	id := c.Param("id")
	deleted, err := repository.DeleteLLMProvider(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	if !deleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先切换到其他 Provider 再删除，或 Provider 不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// AdminTestProvider POST /api/admin/llm-providers/:id/test
func AdminTestProvider(c *gin.Context) {
	id := c.Param("id")
	providers, _ := repository.ListLLMProviders()
	var target *model.LLMProvider
	for i := range providers {
		if providers[i].ID == id {
			p := providers[i]
			target = &p
			break
		}
	}
	if target == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider 不存在"})
		return
	}
	apiKey, err := crypto.Decrypt(target.APIKeyEncrypted, configs.AppConfig.AdminEncryptionKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解密失败"})
		return
	}
	baseURL := strings.TrimSuffix(strings.TrimSuffix(target.BaseURL, "/v1"), "/")
	start := time.Now()
	testErr := service.TestProviderConnection(baseURL+"/v1/chat/completions", apiKey, target.Model)
	elapsed := int(time.Since(start).Milliseconds())
	if testErr != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": testErr.Error(), "latency_ms": elapsed})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "latency_ms": elapsed})
}

// ====== Admin Stats ======

func AdminGetStats(c *gin.Context) {
	stats := gin.H{}

	var totalUsers, todayUsers int
	database.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&totalUsers)
	database.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE`).Scan(&todayUsers)

	var totalCharts, todayCharts int
	database.DB.QueryRow(`SELECT COUNT(*) FROM bazi_charts`).Scan(&totalCharts)
	database.DB.QueryRow(`SELECT COUNT(*) FROM bazi_charts WHERE created_at >= CURRENT_DATE`).Scan(&todayCharts)

	var totalAI, todayAI int
	database.DB.QueryRow(`SELECT COUNT(*) FROM ai_requests_log`).Scan(&totalAI)
	database.DB.QueryRow(`SELECT COUNT(*) FROM ai_requests_log WHERE created_at >= CURRENT_DATE`).Scan(&todayAI)

	stats["total_users"] = totalUsers
	stats["today_users"] = todayUsers
	stats["total_charts"] = totalCharts
	stats["today_charts"] = todayCharts
	stats["total_ai_requests"] = totalAI
	stats["today_ai_requests"] = todayAI

	c.JSON(http.StatusOK, stats)
}

func AdminGetAIStats(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT p.name, COUNT(*) as total,
		       SUM(CASE WHEN l.status='success' THEN 1 ELSE 0 END) as success_count
		FROM ai_requests_log l
		LEFT JOIN llm_providers p ON l.provider_id = p.id
		GROUP BY p.name
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	type ProviderStat struct {
		Provider     string  `json:"provider"`
		Total        int     `json:"total"`
		SuccessCount int     `json:"success_count"`
		SuccessRate  float64 `json:"success_rate"`
	}
	var result []ProviderStat
	for rows.Next() {
		var s ProviderStat
		rows.Scan(&s.Provider, &s.Total, &s.SuccessCount)
		if s.Total > 0 {
			s.SuccessRate = float64(s.SuccessCount) / float64(s.Total) * 100
		}
		result = append(result, s)
	}
	c.JSON(http.StatusOK, gin.H{"by_provider": result})
}

func AdminGetUsers(c *gin.Context) {
	q := c.Query("q")
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	pageSize := 20
	offset := (page - 1) * pageSize

	query := `
		SELECT u.id, u.email, u.nickname, COALESCE(u.source, 'self_registered'), u.created_at, u.disabled_at,
		       COUNT(b.id) as chart_count,
		       (SELECT COUNT(*) FROM compatibility_readings cr WHERE cr.user_id = u.id) as compat_count
		FROM users u
		LEFT JOIN bazi_charts b ON b.user_id = u.id
		WHERE ($1 = '' OR u.email ILIKE '%' || $1 || '%')
		GROUP BY u.id, u.email, u.nickname, u.source, u.created_at, u.disabled_at
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := database.DB.Query(query, q, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	type UserRow struct {
		ID          string     `json:"id"`
		Email       string     `json:"email"`
		Nickname    string     `json:"nickname"`
		Source      string     `json:"source"`
		CreatedAt   string     `json:"created_at"`
		DisabledAt  *time.Time `json:"disabled_at"`
		ChartCount  int        `json:"chart_count"`
		CompatCount int        `json:"compat_count"`
	}
	var users []UserRow
	for rows.Next() {
		var u UserRow
		rows.Scan(&u.ID, &u.Email, &u.Nickname, &u.Source, &u.CreatedAt, &u.DisabledAt, &u.ChartCount, &u.CompatCount)
		users = append(users, u)
	}

	var total int
	database.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE $1 = '' OR email ILIKE '%' || $1 || '%'`, q).Scan(&total)

	c.JSON(http.StatusOK, gin.H{"users": users, "total": total})
}

func AdminCreateUser(c *gin.Context) {
	var req service.AdminCreateUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := service.CreateUserByAdmin(req)
	if err != nil {
		if err.Error() == "该邮箱已被注册" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// AdminSetUserDisabled 禁用/解禁用户
func AdminSetUserDisabled(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Disabled *bool `json:"disabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Disabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "disabled 必填"})
		return
	}
	if err := repository.SetUserDisabled(id, *req.Disabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"disabled": *req.Disabled})
}

func AdminGetRegistrationSetting(c *gin.Context) {
	enabled, err := repository.GetBoolSetting(repository.SettingRegistrationEnabled, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取注册设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"registration_enabled": enabled})
}

func AdminUpdateRegistrationSetting(c *gin.Context) {
	var req struct {
		RegistrationEnabled *bool `json:"registration_enabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RegistrationEnabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "registration_enabled 必填"})
		return
	}
	if err := repository.SetBoolSetting(repository.SettingRegistrationEnabled, *req.RegistrationEnabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存注册设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"registration_enabled": *req.RegistrationEnabled})
}

// AdminListAILogs 分页查询 AI 调用日志明细
func AdminListAILogs(c *gin.Context) {
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil || page < 1 {
			page = 1
		}
	}
	statusFilter := c.Query("status") // "success" | "error" | "" (all)

	logs, total, err := repository.ListAIRequestLogs(page, pageSize, statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}
	if logs == nil {
		logs = []model.AIRequestLog{}
	}
	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AdminGetAILogsSummary 近 7 天 AI 调用统计摘要
func AdminGetAILogsSummary(c *gin.Context) {
	stats, err := repository.GetAILogsSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}
	if stats == nil {
		stats = []repository.AILogDayStat{}
	}
	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// AdminListCharts 获取全量用户的起盘日记流水
func AdminListCharts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	charts, total, err := repository.ListBaziCharts(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排盘历史失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  charts,
		"total": total,
		"page":  page,
	})
}

// ====== Admin Report Cache ======

// AdminClearAllReports 清空所有 AI 报告缓存（强制下次重新生成）
func AdminClearAllReports(c *gin.Context) {
	affected, err := repository.DeleteAllReports()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "已清除所有报告缓存",
		"deleted": affected,
	})
}

// AdminClearReportByChart 清除指定命盘的 AI 报告缓存
func AdminClearReportByChart(c *gin.Context) {
	chartID := c.Param("chart_id")
	if chartID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 chart_id 参数"})
		return
	}
	affected, err := repository.DeleteReportByChartID(chartID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "已清除该命盘报告缓存",
		"deleted": affected,
	})
}

// AdminListLiunianReports 获取某命盘下所有的流年批断记录
func AdminListLiunianReports(c *gin.Context) {
	chartID := c.Param("chart_id")
	if chartID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 chart_id 参数"})
		return
	}
	reports, err := repository.GetLiunianReportsByChartID(chartID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取流年记录失败: " + err.Error()})
		return
	}
	if reports == nil {
		reports = []model.AILiunianReport{}
	}
	c.JSON(http.StatusOK, gin.H{"data": reports})
}

// AdminDeleteLiunianReport 单独删除某一流年报告缓存
func AdminDeleteLiunianReport(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 id 参数"})
		return
	}
	if err := repository.DeleteLiunianReportByID(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已清除该流年记录缓存"})
}

// AdminListCompatReadings 后台全量合盘明细（分页，只读）
func AdminListCompatReadings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := repository.AdminListCompatibilityReadings(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取合盘明细失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total, "page": page})
}

// AdminResetUserPassword 后台重置指定用户密码
func AdminResetUserPassword(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码至少需要8位"})
		return
	}
	if err := service.ResetUserPassword(id, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置密码失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已重置"})
}

// AdminGetCompatReadingDetail 后台合盘详情（只读）
func AdminGetCompatReadingDetail(c *gin.Context) {
	id := c.Param("id")
	detail, err := repository.GetCompatibilityDetail(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取合盘详情失败"})
		return
	}
	if detail == nil || detail.Reading == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "合盘记录不存在"})
		return
	}
	email, _ := repository.GetCompatibilityReadingUserEmail(id)
	c.JSON(http.StatusOK, gin.H{"data": detail, "user_email": email})
}

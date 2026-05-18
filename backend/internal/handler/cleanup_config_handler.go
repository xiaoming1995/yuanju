package handler

import (
	"net/http"
	"strconv"
	"yuanju/internal/repository"
	"yuanju/internal/service"

	"github.com/gin-gonic/gin"
)

// CleanupConfigResponse 是 admin 后台读取的清理配置（已 clamp 到合法区间）。
type CleanupConfigResponse struct {
	Enabled       bool `json:"enabled"`
	RetentionDays int  `json:"retention_days"`
	RunHour       int  `json:"run_hour"`
}

// AdminGetCleanupConfig GET /api/admin/cleanup-config
// 直读 algo_config 中 cleanup_* 键，缺失则走默认值。
func AdminGetCleanupConfig(c *gin.Context) {
	cfg, err := service.GetCleanupConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, CleanupConfigResponse{
		Enabled:       cfg.Enabled,
		RetentionDays: cfg.RetentionDays,
		RunHour:       cfg.RunHour,
	})
}

// AdminUpdateCleanupConfig PUT /api/admin/cleanup-config
// body: {enabled, retention_days, run_hour}。三个字段都要传；clamp 在写入前完成。
func AdminUpdateCleanupConfig(c *gin.Context) {
	var body struct {
		Enabled       *bool `json:"enabled"`
		RetentionDays *int  `json:"retention_days"`
		RunHour       *int  `json:"run_hour"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Enabled == nil || body.RetentionDays == nil || body.RunHour == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "enabled / retention_days / run_hour 都不能省"})
		return
	}

	retention := clamp(*body.RetentionDays, 1, 3650)
	runHour := clamp(*body.RunHour, 0, 23)

	enabledStr := "false"
	if *body.Enabled {
		enabledStr = "true"
	}

	if err := repository.UpsertAlgoConfig("cleanup_enabled", enabledStr, "是否启用自动数据清理任务"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := repository.UpsertAlgoConfig("cleanup_retention_days", strconv.Itoa(retention), "AI 缓存表与请求日志的保留天数，超期自动删除"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := repository.UpsertAlgoConfig("cleanup_run_hour", strconv.Itoa(runHour), "每日清理任务执行时刻（24h 制小时，0-23）"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CleanupConfigResponse{
		Enabled:       *body.Enabled,
		RetentionDays: retention,
		RunHour:       runHour,
	})
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

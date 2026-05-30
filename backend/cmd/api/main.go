package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	"yuanju/configs"
	"yuanju/internal/handler"
	"yuanju/internal/middleware"
	"yuanju/internal/repository"
	"yuanju/internal/service"
	"yuanju/pkg/database"
	"yuanju/pkg/prompt"
	"yuanju/pkg/seed"

	"github.com/gin-gonic/gin"
)

func main() {
	cleanupOnce := flag.Bool("cleanup-once", false, "运行一次清理任务后退出，不启动 HTTP server")
	migrateDryRun := flag.Bool("migrate-dry-run", false, "打印 pending migration 后退出，不动 DB")
	migrateApply := flag.Bool("migrate-apply", false, "强制跑一次 migration 后退出，不启动 HTTP server")
	flag.Parse()

	// 加载配置
	configs.Load()

	// 连接数据库
	database.Connect()

	// CLI 分支：迁移工具命令在 Migrate 之前处理，避免重复迁移
	if *migrateDryRun {
		rep, err := database.Migrate(database.ModeDryRun)
		if err != nil {
			log.Fatalf("dry-run 失败: %v", err)
		}
		fmt.Println(string(database.MarshalMigrationReport(rep)))
		os.Exit(0)
	}
	if *migrateApply {
		rep, err := database.Migrate(database.ModeApply)
		fmt.Println(string(database.MarshalMigrationReport(rep)))
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 默认启动路径：跑 ModeStartup 迁移（0001 fatal、0002+ warn-only）
	if _, err := database.Migrate(database.ModeStartup); err != nil {
		log.Printf("[migrate] startup unexpected error: %v", err)
	}

	// Prompt 注册表对齐：把代码侧 Canonical 写入 ai_prompts 表（未自定义行 sync 到当前版本）。
	if err := prompt.SyncCanonical(database.DB); err != nil {
		log.Printf("[prompt-sync] startup error: %v", err)
	}

	// 确保 logo 上传目录存在
	if err := os.MkdirAll(filepath.Join(configs.AppConfig.UploadDir, "brand-logos"), 0755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}

	// 种子数据：将 .env 中已有的 API Key 写入 llm_providers
	seed.SeedLLMProviders()
	seed.SeedLLMPrices()
	seed.SeedCostAlertThresholds()

	// 加载算法配置（含调候用神 seed）
	if err := service.LoadAlgoConfig(); err != nil {
		log.Printf("算法配置加载失败（使用默认值）: %v", err)
	}

	// 构造 cleanup 服务（在判断 --cleanup-once 之前，因为两种模式都用到）
	cleanupSvc := service.NewCleanupService(makeCleanupDeps())

	if *cleanupOnce {
		rep := cleanupSvc.RunOnce(context.Background())
		fmt.Println(string(service.MarshalRunReport(rep)))
		os.Exit(0)
	}

	// 起 cleanup scheduler（后台 goroutine）
	schedCtx, cancelSched := context.WithCancel(context.Background())
	go cleanupSvc.StartScheduler(schedCtx)

	// Cost alert scheduler — 5 min ticker, emits structured JSON log on threshold breach
	costAlertScheduler := service.NewCostAlertScheduler()
	go costAlertScheduler.StartScheduler(schedCtx)
	defer cancelSched()

	// SIGTERM/SIGINT 触发 cancelSched 让 scheduler 干净退出
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh
		log.Println("收到退出信号，取消 cleanup scheduler")
		cancelSched()
	}()

	// 初始化路由
	r := gin.Default()

	// Cap multipart in-memory buffer well below Gin's 32 MiB default.
	// The upload handler enforces 2 MiB per file; this is belt-and-suspenders
	// against clients omitting Content-Length (chunked encoding).
	r.MaxMultipartMemory = 4 << 20 // 4 MiB

	// 跨域中间件
	r.Use(middleware.CORS())

	// Public read-only mount: ONLY the brand-logos subdirectory is exposed.
	// Do NOT broaden this to UploadDir root — future features may put private
	// files under UploadDir/<other-subdir> and they must not be served publicly.
	r.Static("/static/uploads/brand-logos", filepath.Join(configs.AppConfig.UploadDir, "brand-logos"))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "缘聚命理 API"})
	})

	// API 路由组
	api := r.Group("/api")
	{
		// 普通用户认证路由
		auth := api.Group("/auth")
		{
			auth.GET("/registration-settings", handler.RegistrationSettings)
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
			auth.GET("/me", middleware.Auth(), handler.Me)
		}

		// 八字路由
		bazi := api.Group("/bazi")
		{
			bazi.POST("/calculate", middleware.OptionalAuth(), handler.Calculate)
			bazi.POST("/report/:chart_id", middleware.Auth(), handler.GenerateReport)
			bazi.POST("/report-stream/:chart_id", middleware.Auth(), handler.GenerateReportStream)
			bazi.POST("/polished-report/:chart_id", middleware.Auth(), handler.GenerateAndSavePolishedReport)
			bazi.GET("/polished-report/:chart_id", middleware.Auth(), handler.GetPolishedReport)
			bazi.POST("/liunian-report/:chart_id", middleware.Auth(), handler.GenerateLiunianReport)
			bazi.POST("/past-events-stream/:chart_id", middleware.Auth(), handler.HandlePastEventsStream)
			// 思路 E：即时年份 + 流式大运总结
			bazi.POST("/past-events/years/:chart_id", middleware.Auth(), handler.HandlePastEventsYears)
			bazi.POST("/past-events/dayun-summary-stream/:chart_id", middleware.Auth(), handler.HandleDayunSummariesStream)
			bazi.GET("/history", middleware.Auth(), handler.GetHistory)
			bazi.GET("/history/:id", middleware.Auth(), handler.GetHistoryDetail)
			bazi.PATCH("/history/:id/display-name", middleware.Auth(), handler.UpdateHistoryDisplayName)
			bazi.POST("/liu-yue", handler.HandleLiuYue) // 流月查询（无需登录）
		}

		compatibility := api.Group("/compatibility", middleware.Auth())
		{
			compatibility.POST("/readings", handler.CreateCompatibilityReading)
			compatibility.GET("/readings", handler.GetCompatibilityHistory)
			compatibility.GET("/readings/:id", handler.GetCompatibilityDetail)
			compatibility.POST("/readings/:id/report", handler.GenerateCompatibilityReport)
		}

		user := api.Group("/user", middleware.Auth())
		{
			user.GET("/profile", handler.GetUserProfile)
			user.GET("/export-brand", handler.RequireUserID, handler.GetExportBrand)
			user.PUT("/export-brand", handler.RequireUserID, handler.UpdateExportBrand)
			user.POST("/export-brand/logo", handler.RequireUserID, handler.UploadExportBrandLogo)
			user.DELETE("/export-brand/logo", handler.RequireUserID, handler.DeleteExportBrandLogo)
		}

		// 神煞注解（公开）
		api.GET("/shensha/annotations", handler.GetShenshaAnnotations)

		// Admin 路由组（独立鉴权）
		admin := api.Group("/admin")
		{
			// Admin 认证（无需 Token）
			admin.POST("/auth/register", handler.AdminRegister)
			admin.POST("/auth/login", handler.AdminLogin)

			// 需要 Admin Token 的路由
			adminAuth := admin.Group("", middleware.AdminAuth())
			{
				// LLM Provider 管理
				adminAuth.GET("/llm-providers", handler.AdminListProviders)
				adminAuth.POST("/llm-providers", handler.AdminCreateProvider)
				adminAuth.PUT("/llm-providers/:id", handler.AdminUpdateProvider)
				adminAuth.PUT("/llm-providers/:id/activate", handler.AdminActivateProvider)
				adminAuth.POST("/llm-providers/:id/test", handler.AdminTestProvider)
				adminAuth.DELETE("/llm-providers/:id", handler.AdminDeleteProvider)

				// 统计
				adminAuth.GET("/stats", handler.AdminGetStats)
				adminAuth.GET("/stats/ai", handler.AdminGetAIStats)

				// 用户与数据流水管理
				adminAuth.GET("/users", handler.AdminGetUsers)
				adminAuth.POST("/users", handler.AdminCreateUser)
				adminAuth.GET("/settings/registration", handler.AdminGetRegistrationSetting)
				adminAuth.PUT("/settings/registration", handler.AdminUpdateRegistrationSetting)
				adminAuth.GET("/charts", handler.AdminListCharts)
				adminAuth.GET("/compatibility/readings", handler.AdminListCompatReadings)
				adminAuth.GET("/compatibility/readings/:id", handler.AdminGetCompatReadingDetail)
				adminAuth.GET("/charts/:chart_id/liunian", handler.AdminListLiunianReports)

				// AI 调用日志
				adminAuth.GET("/ai-logs", handler.AdminListAILogs)
				adminAuth.GET("/ai-logs/summary", handler.AdminGetAILogsSummary)

				// Token 用量统计
				adminAuth.GET("/token-usage/summary", handler.AdminGetTokenUsageSummary)
				adminAuth.GET("/token-usage/detail", handler.AdminGetTokenUsageDetail)
				adminAuth.GET("/token-usage/content/:id", handler.AdminGetTokenUsageContent)
				adminAuth.GET("/token-usage/budget-status", handler.AdminGetBudgetStatus)

				// 报告缓存管理
				adminAuth.DELETE("/reports/cache", handler.AdminClearAllReports)
				adminAuth.DELETE("/reports/cache/:chart_id", handler.AdminClearReportByChart)
				adminAuth.DELETE("/liunian/:id", handler.AdminDeleteLiunianReport)

				// 名人录管理
				adminAuth.GET("/celebrities", handler.AdminListCelebrities)
				adminAuth.POST("/celebrities", handler.AdminCreateCelebrity)
				adminAuth.PUT("/celebrities/:id", handler.AdminUpdateCelebrity)
				adminAuth.DELETE("/celebrities/:id", handler.AdminDeleteCelebrity)
				adminAuth.POST("/celebrities/ai-generate", handler.AdminAIGenerateCelebrities)

				// Prompt 管理
				adminAuth.GET("/prompts", handler.GetPrompts)
				adminAuth.PUT("/prompts/:module", handler.UpdatePrompt)
				adminAuth.POST("/prompts/:module/reset", handler.ResetPromptToCanonical)

				// 算法参数管理
				adminAuth.GET("/algo-config", handler.AdminGetAlgoConfig)
				adminAuth.PUT("/algo-config/:key", handler.AdminUpdateAlgoConfig)
				adminAuth.POST("/algo-config/reload", handler.AdminReloadAlgoConfig)

				// 数据清理任务配置
				adminAuth.GET("/cleanup-config", handler.AdminGetCleanupConfig)
				adminAuth.PUT("/cleanup-config", handler.AdminUpdateCleanupConfig)

				// 调候用神规则管理
				adminAuth.GET("/algo-tiaohou", handler.AdminGetAlgoTiaohou)
				adminAuth.PUT("/algo-tiaohou/:day_gan/:month_zhi", handler.AdminUpdateAlgoTiaohou)
				adminAuth.DELETE("/algo-tiaohou/:day_gan/:month_zhi", handler.AdminDeleteAlgoTiaohou)

				// 神煞注解管理（Admin）
				adminAuth.PUT("/shensha-annotations/:name", handler.AdminUpdateShenshaAnnotation)
			}
		}
	}

	port := configs.AppConfig.Port
	log.Printf("🚀 缘聚命理服务启动，端口：%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// makeCleanupDeps 把 repository 层的具体函数 wrap 成 service.CleanupDeps，
// 同时把 repository.RollupReport 适配成 service.RollupReport（结构等价）。
func makeCleanupDeps() service.CleanupDeps {
	wrapTime := func(f func(time.Time) (int64, error)) service.TableCleaner {
		return func(_ context.Context, cutoff time.Time) (int64, error) {
			return f(cutoff)
		}
	}
	return service.CleanupDeps{
		AIReports:     wrapTime(repository.DeleteAIReportsOlderThan),
		Polished:      wrapTime(repository.DeletePolishedReportsOlderThan),
		Liunian:       wrapTime(repository.DeleteLiunianReportsOlderThan),
		PastEvents:    wrapTime(repository.DeletePastEventsOlderThan),
		DayunSummary:  wrapTime(repository.DeleteDayunSummariesOlderThan),
		CompatReports: wrapTime(repository.DeleteAICompatibilityReportsOlderThan),
		RequestLogs:   wrapTime(repository.DeleteRequestLogsOlderThan),
		TokenRollup: func(_ context.Context) (service.RollupReport, error) {
			r, err := repository.RollupClosedMonthsAndDelete()
			return service.RollupReport{
				MonthsAggregated:      r.MonthsAggregated,
				RowsInsertedOrUpdated: r.RowsInsertedOrUpdated,
				SourceRowsDeleted:     r.SourceRowsDeleted,
			}, err
		},
	}
}

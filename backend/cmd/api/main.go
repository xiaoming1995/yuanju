package main

import (
	"log"
	"yuanju/configs"
	"yuanju/internal/handler"
	"yuanju/internal/middleware"
	"yuanju/internal/service"
	"yuanju/pkg/database"
	"yuanju/pkg/seed"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	configs.Load()

	// 连接数据库
	database.Connect()
	database.Migrate()

	// 种子数据：将 .env 中已有的 API Key 写入 llm_providers
	seed.SeedLLMProviders()

	// 加载算法配置（含调候用神 seed）
	if err := service.LoadAlgoConfig(); err != nil {
		log.Printf("算法配置加载失败（使用默认值）: %v", err)
	}

	// 初始化路由
	r := gin.Default()

	// 跨域中间件
	r.Use(middleware.CORS())

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
			bazi.POST("/liunian-report/:chart_id", middleware.Auth(), handler.GenerateLiunianReport)
			bazi.POST("/past-events-stream/:chart_id", middleware.Auth(), handler.HandlePastEventsStream)
			// 思路 E：即时年份 + 流式大运总结
			bazi.POST("/past-events/years/:chart_id", middleware.Auth(), handler.HandlePastEventsYears)
			bazi.POST("/past-events/dayun-summary-stream/:chart_id", middleware.Auth(), handler.HandleDayunSummariesStream)
			bazi.GET("/history", middleware.Auth(), handler.GetHistory)
			bazi.GET("/history/:id", middleware.Auth(), handler.GetHistoryDetail)
			bazi.POST("/liu-yue", handler.HandleLiuYue) // 流月查询（无需登录）
		}

		compatibility := api.Group("/compatibility", middleware.Auth())
		{
			compatibility.POST("/readings", handler.CreateCompatibilityReading)
			compatibility.GET("/readings", handler.GetCompatibilityHistory)
			compatibility.GET("/readings/:id", handler.GetCompatibilityDetail)
			compatibility.POST("/readings/:id/report", handler.GenerateCompatibilityReport)
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
				adminAuth.GET("/charts", handler.AdminListCharts)
				adminAuth.GET("/charts/:chart_id/liunian", handler.AdminListLiunianReports)

				// AI 调用日志
				adminAuth.GET("/ai-logs", handler.AdminListAILogs)
				adminAuth.GET("/ai-logs/summary", handler.AdminGetAILogsSummary)

				// Token 用量统计
				adminAuth.GET("/token-usage/summary", handler.AdminGetTokenUsageSummary)
				adminAuth.GET("/token-usage/detail", handler.AdminGetTokenUsageDetail)

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

				// 算法参数管理
				adminAuth.GET("/algo-config", handler.AdminGetAlgoConfig)
				adminAuth.PUT("/algo-config/:key", handler.AdminUpdateAlgoConfig)
				adminAuth.POST("/algo-config/reload", handler.AdminReloadAlgoConfig)

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

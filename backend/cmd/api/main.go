package main

import (
	"log"
	"yuanju/configs"
	"yuanju/internal/handler"
	"yuanju/internal/middleware"
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
			bazi.GET("/history", middleware.Auth(), handler.GetHistory)
			bazi.GET("/history/:id", middleware.Auth(), handler.GetHistoryDetail)
		}

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
				adminAuth.DELETE("/llm-providers/:id", handler.AdminDeleteProvider)

				// 统计
				adminAuth.GET("/stats", handler.AdminGetStats)
				adminAuth.GET("/stats/ai", handler.AdminGetAIStats)

				// 用户与数据流水管理
				adminAuth.GET("/users", handler.AdminGetUsers)
				adminAuth.GET("/charts", handler.AdminListCharts)

				// AI 调用日志
				adminAuth.GET("/ai-logs", handler.AdminListAILogs)
				adminAuth.GET("/ai-logs/summary", handler.AdminGetAILogsSummary)

				// 报告缓存管理
				adminAuth.DELETE("/reports/cache", handler.AdminClearAllReports)
				adminAuth.DELETE("/reports/cache/:chart_id", handler.AdminClearReportByChart)

				// 名人录管理
				adminAuth.GET("/celebrities", handler.AdminListCelebrities)
				adminAuth.POST("/celebrities", handler.AdminCreateCelebrity)
				adminAuth.PUT("/celebrities/:id", handler.AdminUpdateCelebrity)
				adminAuth.DELETE("/celebrities/:id", handler.AdminDeleteCelebrity)
				adminAuth.POST("/celebrities/ai-generate", handler.AdminAIGenerateCelebrities)
			}
		}
	}

	port := configs.AppConfig.Port
	log.Printf("🚀 缘聚命理服务启动，端口：%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

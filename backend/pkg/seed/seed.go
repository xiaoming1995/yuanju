package seed

import (
	"log"
	"yuanju/configs"
	"yuanju/pkg/crypto"
	"yuanju/pkg/database"
)

// SeedLLMProviders 将 .env 中已有的 API Key 写入 llm_providers（若表为空）
// 任务 1.5：.env 配置迁移到数据库
func SeedLLMProviders() {
	var count int
	database.DB.QueryRow(`SELECT COUNT(*) FROM llm_providers`).Scan(&count)
	if count > 0 {
		return // 已有数据，跳过
	}

	seeded := false

	if configs.AppConfig.DeepSeekAPIKey != "" {
		encrypted, err := crypto.Encrypt(configs.AppConfig.DeepSeekAPIKey, configs.AppConfig.AdminEncryptionKey)
		if err == nil {
			_, err = database.DB.Exec(
				`INSERT INTO llm_providers (name, type, base_url, model, api_key_encrypted, active)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				"DeepSeek", "deepseek",
				configs.AppConfig.AIBaseURL,
				"deepseek-chat",
				encrypted,
				true, // 第一个设为激活
			)
			if err == nil {
				log.Println("✅ 种子数据：DeepSeek Provider 已写入数据库并激活")
				seeded = true
			}
		}
	}

	if configs.AppConfig.OpenAIAPIKey != "" {
		encrypted, err := crypto.Encrypt(configs.AppConfig.OpenAIAPIKey, configs.AppConfig.AdminEncryptionKey)
		if err == nil {
			_, err = database.DB.Exec(
				`INSERT INTO llm_providers (name, type, base_url, model, api_key_encrypted, active)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				"OpenAI", "openai",
				"https://api.openai.com",
				"gpt-4o-mini",
				encrypted,
				!seeded, // 仅当没有 DeepSeek 时才激活
			)
			if err == nil {
				log.Println("✅ 种子数据：OpenAI Provider 已写入数据库")
			}
		}
	}
}

// SeedLLMPrices 将默认模型定价写入 algo_config（ON CONFLICT DO NOTHING，不覆盖 Admin 已改的值）
func SeedLLMPrices() {
	prices := []struct {
		key, value, description string
	}{
		{"llm_price_flash_input", "0.27", "deepseek-v4-flash 输入单价（CNY/百万tokens）"},
		{"llm_price_flash_output", "1.10", "deepseek-v4-flash 输出单价（CNY/百万tokens）"},
		{"llm_price_pro_input", "4.00", "deepseek-v4-pro 输入单价（CNY/百万tokens）"},
		{"llm_price_pro_output", "16.00", "deepseek-v4-pro 输出单价（CNY/百万tokens）"},
		{"llm_price_default_input", "1.00", "未知模型 fallback 输入单价（CNY/百万tokens）"},
		{"llm_price_default_output", "2.00", "未知模型 fallback 输出单价（CNY/百万tokens）"},
	}
	for _, p := range prices {
		if _, err := database.DB.Exec(
			`INSERT INTO algo_config (key, value, description)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (key) DO NOTHING`,
			p.key, p.value, p.description,
		); err != nil {
			log.Printf("[seed] LLM 定价 seed 失败: key=%s err=%v", p.key, err)
		}
	}
	log.Println("✅ 种子数据：LLM 定价配置已写入 algo_config（ON CONFLICT DO NOTHING）")
}

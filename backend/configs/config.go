package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT (普通用户)
	JWTSecret string

	// Admin
	AdminJWTSecret     string
	AdminEncryptionKey string

	// AI
	DeepSeekAPIKey string
	OpenAIAPIKey   string
	AIBaseURL      string
}

var AppConfig Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用环境变量")
	}

	AppConfig = Config{
		Port:               getEnv("PORT", "9002"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://yuanju:yuanju123@localhost:5432/yuanju?sslmode=disable"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:          getEnv("JWT_SECRET", "yuanju-secret-key-change-in-production"),
		AdminJWTSecret:     getEnv("ADMIN_JWT_SECRET", "yuanju-admin-secret-change-in-production"),
		AdminEncryptionKey: getEnv("ADMIN_ENCRYPTION_KEY", "yuanju-enc-key-32bytespadding!!"),
		DeepSeekAPIKey:     getEnv("DEEPSEEK_API_KEY", ""),
		OpenAIAPIKey:       getEnv("OPENAI_API_KEY", ""),
		AIBaseURL:          getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

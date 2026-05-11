package service

import (
	"strconv"
	"strings"
	"yuanju/internal/repository"
)

// matchModelTier 将模型名映射到定价档位：pro / flash / default
func matchModelTier(modelName string) string {
	lower := strings.ToLower(modelName)
	if strings.Contains(lower, "v4-pro") || strings.Contains(lower, "reasoner") {
		return "pro"
	}
	if strings.Contains(lower, "v4-flash") || strings.Contains(lower, "chat") {
		return "flash"
	}
	return "default"
}

// GetModelPrice 返回指定模型的 (inputPriceCNY, outputPriceCNY)，单位：CNY / 百万 tokens。
// 从 algo_config 表实时读取；读取失败时使用硬编码兜底值。
func GetModelPrice(modelName string) (inputPrice, outputPrice float64) {
	rows, err := repository.GetAllAlgoConfig()
	if err != nil {
		return 1.0, 2.0
	}
	prices := make(map[string]string, 6)
	for _, r := range rows {
		if strings.HasPrefix(r.Key, "llm_price_") {
			prices[r.Key] = r.Value
		}
	}
	tier := matchModelTier(modelName)
	inputPrice = parsePrice(prices["llm_price_"+tier+"_input"], 1.0)
	outputPrice = parsePrice(prices["llm_price_"+tier+"_output"], 2.0)
	return inputPrice, outputPrice
}

// CalcCost 根据 token 数量和模型名返回预估费用（CNY）。
func CalcCost(modelName string, promptTokens, completionTokens int) float64 {
	inputPrice, outputPrice := GetModelPrice(modelName)
	return float64(promptTokens)/1_000_000*inputPrice +
		float64(completionTokens)/1_000_000*outputPrice
}

func parsePrice(s string, fallback float64) float64 {
	if s == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return fallback
	}
	return v
}

package service

import (
	"database/sql"
	"yuanju/internal/repository"
)

// GetModelPrice 返回指定模型的 (inputPriceCNY, outputPriceCNY)，单位：CNY / 百万 tokens。
// 从 llm_providers 表按 model 字段精确查询；找不到时 fallback 到默认值。
func GetModelPrice(modelName string) (inputPrice, outputPrice float64) {
	in, out, err := repository.GetPriceByModel(modelName)
	if err != nil && err != sql.ErrNoRows {
		return 1.0, 2.0
	}
	if err == sql.ErrNoRows || (in == 0 && out == 0) {
		return 1.0, 2.0
	}
	return in, out
}

// CalcCost 根据 token 数量和模型名返回预估费用（CNY）。
func CalcCost(modelName string, promptTokens, completionTokens int) float64 {
	inputPrice, outputPrice := GetModelPrice(modelName)
	return float64(promptTokens)/1_000_000*inputPrice +
		float64(completionTokens)/1_000_000*outputPrice
}

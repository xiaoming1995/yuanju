package bazi

import "sync"

// AlgoConfig 算法层全局配置，由 internal/service 在启动时注入
type AlgoConfig struct {
	// 大运吉凶判定参数
	JiHanMin    int     // 极寒：寒性元素（水+金）最低数量，默认 4
	JiReMin     int     // 极热：暖性元素（火+木）最低数量，默认 4
	ShenQiangPct float64 // 身强判定阈值（生助比例%），默认 40.0

	// 调候用神覆盖（DB 优先，key = "日干_月支"，如 "甲_子"）
	TiaohouOverrides map[string]TiaohouRule
}

// 默认值常量（DB 无配置时的 fallback）
const (
	DefaultJiHanMin     = 4
	DefaultJiReMin      = 4
	DefaultShenQiangPct = 40.0
)

var (
	globalAlgoConfig AlgoConfig
	algoConfigMu     sync.RWMutex
)

func init() {
	// 初始化为默认值
	globalAlgoConfig = AlgoConfig{
		JiHanMin:         DefaultJiHanMin,
		JiReMin:          DefaultJiReMin,
		ShenQiangPct:     DefaultShenQiangPct,
		TiaohouOverrides: map[string]TiaohouRule{},
	}
}

// SetAlgoConfig 由 service 层在启动/热重载时调用，注入从 DB 读取的配置
func SetAlgoConfig(cfg AlgoConfig) {
	algoConfigMu.Lock()
	defer algoConfigMu.Unlock()
	globalAlgoConfig = cfg
}

// GetAlgoConfig 算法层内部读取当前配置（快照，线程安全）
func GetAlgoConfig() AlgoConfig {
	algoConfigMu.RLock()
	defer algoConfigMu.RUnlock()
	return globalAlgoConfig
}

package service

import (
	"log"
	"strconv"
	"strings"
	"yuanju/internal/repository"
	"yuanju/pkg/bazi"
)

// LoadAlgoConfig 从数据库加载算法配置，注入 bazi 层缓存。
// 若 algo_tiaohou 表为空，自动从硬编码字典 seed。
func LoadAlgoConfig() error {
	// 1. 读取算法参数
	rows, err := repository.GetAllAlgoConfig()
	if err != nil {
		return err
	}

	jiHanMin := bazi.DefaultJiHanMin
	jiReMin := bazi.DefaultJiReMin
	shenQiangPct := bazi.DefaultShenQiangPct

	for _, r := range rows {
		switch r.Key {
		case "jixiong_jiHan_min":
			if v, err := strconv.Atoi(r.Value); err == nil {
				jiHanMin = v
			}
		case "jixiong_jiRe_min":
			if v, err := strconv.Atoi(r.Value); err == nil {
				jiReMin = v
			}
		case "jixiong_shenQiang_pct":
			if v, err := strconv.ParseFloat(r.Value, 64); err == nil {
				shenQiangPct = v
			}
		}
	}

	// 2. 若调候表为空，自动 seed 硬编码字典
	count, err := repository.CountAlgoTiaohou()
	if err != nil {
		return err
	}
	if count == 0 {
		log.Println("[algo_config] algo_tiaohou 表为空，开始 seed 默认调候规则...")
		dict := bazi.GetDefaultTiaohouDict()
		for key, rule := range dict {
			parts := strings.SplitN(key, "_", 2)
			if len(parts) != 2 {
				continue
			}
			xiElements := strings.Join(rule.Yongshen, ",")
			if seedErr := repository.UpsertAlgoTiaohou(parts[0], parts[1], xiElements, rule.Text); seedErr != nil {
				log.Printf("[algo_config] seed tiaohou %s 失败: %v", key, seedErr)
			}
		}
		log.Printf("[algo_config] seed 完成，共 %d 条", len(dict))
	}

	// 3. 读取调候覆盖规则
	tiaohouRows, err := repository.GetAllAlgoTiaohou("")
	if err != nil {
		return err
	}
	overrides := make(map[string]bazi.TiaohouRule, len(tiaohouRows))
	for _, r := range tiaohouRows {
		yongshen := []string{}
		if r.XiElements != "" {
			yongshen = strings.Split(r.XiElements, ",")
		}
		overrides[r.DayGan+"_"+r.MonthZhi] = bazi.TiaohouRule{
			Yongshen: yongshen,
			Text:     r.Text,
		}
	}

	// 4. 注入 bazi 层
	bazi.SetAlgoConfig(bazi.AlgoConfig{
		JiHanMin:         jiHanMin,
		JiReMin:          jiReMin,
		ShenQiangPct:     shenQiangPct,
		TiaohouOverrides: overrides,
	})

	log.Printf("[algo_config] 加载完成: jiHanMin=%d jiReMin=%d shenQiangPct=%.1f tiaohou覆盖=%d条",
		jiHanMin, jiReMin, shenQiangPct, len(overrides))
	return nil
}

// ReloadAlgoConfig 热重载算法配置（与 LoadAlgoConfig 等效，供 API 调用）
func ReloadAlgoConfig() error {
	return LoadAlgoConfig()
}

// CleanupConfig 是清理任务的运行时配置（每次 tick 从 algo_config 表实时读取）。
type CleanupConfig struct {
	Enabled       bool
	RetentionDays int // 已 clamp 到 [1, 3650]
	RunHour       int // 已 clamp 到 [0, 23]
}

// GetCleanupConfig 实时读取 algo_config 中清理相关键。
// 缺失的键走默认值；异常值 clamp 到合法区间。
func GetCleanupConfig() (CleanupConfig, error) {
	cfg := CleanupConfig{Enabled: true, RetentionDays: 90, RunHour: 3}

	rows, err := repository.GetAllAlgoConfig()
	if err != nil {
		return cfg, err
	}
	for _, r := range rows {
		switch r.Key {
		case "cleanup_enabled":
			cfg.Enabled = r.Value == "true" || r.Value == "1"
		case "cleanup_retention_days":
			if v, err := strconv.Atoi(r.Value); err == nil {
				cfg.RetentionDays = clampInt(v, 1, 3650)
			}
		case "cleanup_run_hour":
			if v, err := strconv.Atoi(r.Value); err == nil {
				cfg.RunHour = clampInt(v, 0, 23)
			}
		}
	}
	return cfg, nil
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

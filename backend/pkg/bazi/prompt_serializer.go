package bazi

import (
	"encoding/json"
	"regexp"
)

// CompressYearSignalsForPrompt 序列化 YearSignals slice 用于 AI prompt。
//
// 应用三项 token-saving 转换（依据 optimize-past-events-token-cost 提案档位 2）：
//
//  1. ten_god_power 仅保留 AI 推理需要的字段
//     drop: group (英文码), plain_title, plain_text, score
//     keep: dominant, group_label (中文), strength, polarity, reason
//
//  2. 去掉 year_in_dayun / dayun_phase / dayun_phase_level
//     这些字段能从年份在数组中的位置推出，AI 不需要重复输入
//
//  3. 修剪 signal.evidence 的括号注脚
//     如 "（月柱宫位，权重次之）"、"（本年有重煞，此信号仅作参考）"
//     这些 AI 用作背景但不参与判断
//
// 持久化（ai_dayun_summaries.years）、SSE 推送、前端展示 三处的 JSON 完全不受影响 —
// 本函数仅产出用于 AI prompt 的精简 JSON。
func CompressYearSignalsForPrompt(years []YearSignals) ([]byte, error) {
	compressed := make([]compactYear, len(years))
	for i, y := range years {
		compressed[i] = compactYear{
			Year:        y.Year,
			Age:         y.Age,
			GanZhi:      y.GanZhi,
			DayunGanZhi: y.DayunGanZhi,
			TenGodPower: compactTGPower{
				Dominant:   y.TenGodPower.Dominant,
				GroupLabel: y.TenGodPower.GroupLabel,
				Strength:   y.TenGodPower.Strength,
				Polarity:   y.TenGodPower.Polarity,
				Reason:     y.TenGodPower.Reason,
			},
			Signals: compactSignals(y.Signals),
		}
	}
	return json.Marshal(compressed)
}

type compactYear struct {
	Year        int             `json:"year"`
	Age         int             `json:"age"`
	GanZhi      string          `json:"gan_zhi"`
	DayunGanZhi string          `json:"dayun_gan_zhi"`
	TenGodPower compactTGPower  `json:"ten_god_power,omitempty"`
	Signals     []compactSignal `json:"signals"`
}

type compactTGPower struct {
	Dominant   string `json:"dominant"`
	GroupLabel string `json:"group_label"`
	Strength   string `json:"strength"`
	Polarity   string `json:"polarity"`
	Reason     string `json:"reason,omitempty"`
}

type compactSignal struct {
	Type     string `json:"type"`
	Evidence string `json:"evidence"`
	Polarity string `json:"polarity,omitempty"`
	Source   string `json:"source,omitempty"`
}

// evidenceParenRe 匹配全角和半角括号注脚（长度 ≤ 30 字符）。
// 不匹配嵌套括号；不匹配跨行括号。
var evidenceParenRe = regexp.MustCompile(`（[^）]{1,30}）|\([^)]{1,30}\)`)

// stripEvidenceParenthetical 去掉 evidence 字符串中的括号注脚
func stripEvidenceParenthetical(s string) string {
	return evidenceParenRe.ReplaceAllString(s, "")
}

func compactSignals(sigs []EventSignal) []compactSignal {
	out := make([]compactSignal, len(sigs))
	for i, s := range sigs {
		out[i] = compactSignal{
			Type:     s.Type,
			Evidence: stripEvidenceParenthetical(s.Evidence),
			Polarity: s.Polarity,
			Source:   s.Source,
		}
	}
	return out
}

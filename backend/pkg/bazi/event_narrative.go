package bazi

import (
	"strings"
)

// RenderYearNarrative 根据 EventSignal 列表拼接 2-3 句中文叙述
// 不调用 AI，毫秒级返回。规则：
//  1. 用神基底色作为开篇定调
//  2. 按 polarity 顺序提取关键 evidence（凶 > 吉 > 中性）
//  3. 中性年份给出"运势平稳"类描述
//
// signals 顺序约定与 GetYearEventSignals 输出一致
func RenderYearNarrative(ys YearSignals) string {
	if len(ys.Signals) == 0 {
		return "本年命理信号较弱，运势相对平稳，无明显重大变动。"
	}

	var baseline *EventSignal
	var jiSigs, xiongSigs, neutralSigs []EventSignal
	for i := range ys.Signals {
		s := ys.Signals[i]
		if s.Type == "用神基底" {
			b := s
			baseline = &b
			continue
		}
		switch s.Polarity {
		case PolarityJi:
			jiSigs = append(jiSigs, s)
		case PolarityXiong:
			xiongSigs = append(xiongSigs, s)
		default:
			neutralSigs = append(neutralSigs, s)
		}
	}

	var parts []string

	// 开篇：用神基底定调
	if baseline != nil {
		switch baseline.Polarity {
		case PolarityJi:
			parts = append(parts, ys.GanZhi+"年透用神，全年定调偏吉。")
		case PolarityXiong:
			parts = append(parts, ys.GanZhi+"年透忌神，全年定调偏凶。")
		default:
			parts = append(parts, ys.GanZhi+"年五行中和，全年基调平稳。")
		}
	} else {
		parts = append(parts, ys.GanZhi+"年命局动象如下。")
	}

	// 中段：关键 evidence（最多 2-3 条，按 polarity 优先级）
	keyEvs := pickKeyEvidence(xiongSigs, jiSigs, neutralSigs, 3)
	if len(keyEvs) > 0 {
		parts = append(parts, strings.Join(keyEvs, "；")+"。")
	}

	// 收尾：综合提示
	if len(xiongSigs) >= 2 {
		parts = append(parts, "凶象较多，宜守不宜攻，谨慎应对。")
	} else if len(jiSigs) >= 2 {
		parts = append(parts, "吉象交辉，可主动把握机遇。")
	} else if len(xiongSigs) > 0 && len(jiSigs) > 0 {
		parts = append(parts, "吉凶交杂，得失之间需留心权衡。")
	}

	return strings.Join(parts, "")
}

// pickKeyEvidence 按 polarity 优先级（凶 > 吉 > 中性）挑选最多 max 条 evidence
// 同时优先非神煞类（神煞较泛，事件类信号更具体）
func pickKeyEvidence(xiong, ji, neutral []EventSignal, max int) []string {
	var pool []EventSignal
	pool = append(pool, prioritizeNonShensha(xiong)...)
	pool = append(pool, prioritizeNonShensha(ji)...)
	pool = append(pool, prioritizeNonShensha(neutral)...)

	out := make([]string, 0, max)
	seenEv := map[string]bool{}
	for _, s := range pool {
		if len(out) >= max {
			break
		}
		// 简化 evidence：去掉重复修饰
		ev := s.Evidence
		if seenEv[ev] {
			continue
		}
		seenEv[ev] = true
		out = append(out, ev)
	}
	return out
}

// prioritizeNonShensha 把非神煞信号放前面，神煞放后
func prioritizeNonShensha(sigs []EventSignal) []EventSignal {
	var primary, secondary []EventSignal
	for _, s := range sigs {
		if s.Source == SourceShensha {
			secondary = append(secondary, s)
		} else {
			primary = append(primary, s)
		}
	}
	return append(primary, secondary...)
}

// ExtractYearSignalTypes 提取 signal types（用于前端徽标显示），去除"用神基底"等内部 type
func ExtractYearSignalTypes(ys YearSignals) []string {
	hide := map[string]bool{"用神基底": true, "综合变动": false}
	out := []string{}
	seen := map[string]bool{}
	for _, s := range ys.Signals {
		if hide[s.Type] {
			continue
		}
		if seen[s.Type] {
			continue
		}
		seen[s.Type] = true
		out = append(out, s.Type)
	}
	return out
}

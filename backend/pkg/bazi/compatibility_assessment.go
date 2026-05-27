package bazi

// classifyRelationshipTypeV3 按 design §5.1 优先级链匹配关系类型。
// 短路求值：第一个 true 即返回。
func classifyRelationshipTypeV3(total int, s CompatibilityDimensionScores) string {
	switch {
	case total >= 80:
		return "高契合型"
	case s.Zodiac == 50 && s.DayPillar >= 5:
		return "亲密层稳固型"
	case s.Zodiac == 50:
		return "属相吸引型"
	case s.DayPillar >= 5 || s.EightChars >= 14:
		return "亲密外围支撑型"
	default:
		return "合盘无加成"
	}
}

type decisionAdviceV3 struct {
	Recommendation string
	Verdict        string
	Confidence     string
	Conditions     []string
	DoNext         []string
	Avoid          []string
}

// buildDecisionAdviceV3 按 design §5.2 三档生成建议。hitsCount = 4 模块中分>0 的个数。
func buildDecisionAdviceV3(total, hitsCount int) decisionAdviceV3 {
	var rec, verdict string
	switch {
	case total >= 80:
		rec = "continue"
		verdict = "适合继续推进"
	case total >= 60:
		rec = "observe"
		verdict = "建议谨慎观察"
	default:
		rec = "caution"
		verdict = "不宜过早重投入"
	}
	var confidence string
	switch {
	case hitsCount >= 3:
		confidence = "high"
	case hitsCount >= 1:
		confidence = "medium"
	default:
		confidence = "low"
	}
	conditions, doNext, avoid := decisionAdviceTextsV3(rec)
	return decisionAdviceV3{
		Recommendation: rec, Verdict: verdict, Confidence: confidence,
		Conditions: conditions, DoNext: doNext, Avoid: avoid,
	}
}

// decisionAdviceTextsV3 三档 conditions/do_next/avoid 文案模板（设计 §5.2 末段）。
func decisionAdviceTextsV3(recommendation string) (conditions, doNext, avoid []string) {
	switch recommendation {
	case "continue":
		return []string{
				"维持现有沟通节奏与现实安排",
				"在关键决策上保持双方同步",
			},
			[]string{
				"把长期承接的关键议题（住、责任分工）逐项落地",
				"用具体行为而非情绪强度判断关系稳定性",
			},
			[]string{
				"误以为 4 模块全命中就免维护，关系仍需经营",
				"用合盘结果替代日常沟通的具体内容",
			}
	case "observe":
		return []string{
				"在一到两个月内验证沟通节奏是否稳定",
				"把容易争执的话题具体化处理",
			},
			[]string{
				"先观察冲突后双方修复能力",
				"把短期吸引点和长期承接点分开评估",
			},
			[]string{
				"在关系规则未稳定前过早绑定重大决定",
				"用单一模块的结果（如属相相合）替代全局判断",
			}
	default:
		return []string{
				"先稳定个人节奏，再考虑重投入",
				"避免在缺少支点的阶段做长期承诺",
			},
			[]string{
				"用 1–3 件具体生活议题观察对方现实承接能力",
				"建立可暂停的关系边界",
			},
			[]string{
				"用『感觉』替代『结构证据』推动关系升级",
				"忽略合盘提示的弱支点强行投入",
			}
	}
}

// buildDurationAssessmentV3 按 design §5.3 三窗口阈值生成评估。
func buildDurationAssessmentV3(s CompatibilityDimensionScores) CompatibilityDurationAssessment {
	short := shortWindowLevel(s)
	mid := midWindowLevel(s)
	long := longWindowLevel(s)
	return CompatibilityDurationAssessment{
		OverallBand: durationBandV3(long),
		Windows: CompatibilityDurationWindows{
			ThreeMonths:  CompatibilityDurationWindow{Level: short},
			OneYear:      CompatibilityDurationWindow{Level: mid},
			TwoYearsPlus: CompatibilityDurationWindow{Level: long},
		},
		Summary: durationSummaryV3(short, long),
		Reasons: nil, // 由调用方注入 evidence reasons
	}
}

func shortWindowLevel(s CompatibilityDimensionScores) CompatibilityDurationLevel {
	switch {
	case s.Zodiac == 50 && s.Nayin == 20:
		return CompatibilityDurationHigh
	case s.Zodiac == 50 || s.Nayin == 20:
		return CompatibilityDurationMedium
	default:
		return CompatibilityDurationLow
	}
}

func midWindowLevel(s CompatibilityDimensionScores) CompatibilityDurationLevel {
	switch {
	case s.DayPillar == 10 || (s.DayPillar == 5 && s.Zodiac == 50):
		return CompatibilityDurationHigh
	case s.DayPillar >= 5:
		return CompatibilityDurationMedium
	default:
		return CompatibilityDurationLow
	}
}

func longWindowLevel(s CompatibilityDimensionScores) CompatibilityDurationLevel {
	switch {
	case s.EightChars >= 14 && s.DayPillar >= 5:
		return CompatibilityDurationHigh
	case s.EightChars >= 7 || s.DayPillar == 10:
		return CompatibilityDurationMedium
	default:
		return CompatibilityDurationLow
	}
}

func durationBandV3(longLevel CompatibilityDurationLevel) string {
	switch longLevel {
	case CompatibilityDurationHigh:
		return "long_term"
	case CompatibilityDurationMedium:
		return "medium_term"
	default:
		return "short_term"
	}
}

func durationSummaryV3(short, long CompatibilityDurationLevel) string {
	switch {
	case short == CompatibilityDurationHigh && long == CompatibilityDurationHigh:
		return "属相与纳音支撑短期吸引，日柱与八字承接长期稳定，关系发展通道顺畅。"
	case short == CompatibilityDurationHigh && long == CompatibilityDurationLow:
		return "短期靠近感强，但长期承接薄弱——关系更像「先热后难」型，需要把短期热度引导到现实安排。"
	case long == CompatibilityDurationHigh:
		return "短期需要时间培养亲近感，长期承接稳——关系更适合慢热经营。"
	default:
		return "这段关系的维持性更依赖阶段中的现实磨合而非最初的命盘指向。"
	}
}

// durationReasonsFromEvidence 把前 3 条命中 evidence 转为 Reason 字符串。
func durationReasonsFromEvidence(evidences []CompatibilityEvidence) []string {
	out := make([]string, 0, 3)
	for _, ev := range evidences {
		if ev.Polarity != "positive" {
			continue
		}
		out = append(out, ev.Title+"："+ev.Detail)
		if len(out) >= 3 {
			break
		}
	}
	if len(out) == 0 {
		out = append(out, "当前盘面未触发任何加分模块，建议结合现实相处判断维持性。")
	}
	return out
}

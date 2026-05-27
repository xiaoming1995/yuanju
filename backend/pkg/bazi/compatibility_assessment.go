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

// buildStageRisksV3 生成 3 个窗口的风险描述。文案按 (窗口×level) 9 套模板。
func buildStageRisksV3(duration CompatibilityDurationAssessment, evidences []CompatibilityEvidence) []CompatibilityStageRisk {
	zodiacKeys := evidenceKeysByDimension(evidences, "zodiac", "nayin")
	dayKeys := evidenceKeysByDimension(evidences, "day_pillar")
	eightKeys := evidenceKeysByDimension(evidences, "eight_chars")
	return []CompatibilityStageRisk{
		stageRiskV3("three_months", duration.Windows.ThreeMonths.Level, zodiacKeys),
		stageRiskV3("one_year", duration.Windows.OneYear.Level, dayKeys),
		stageRiskV3("two_years_plus", duration.Windows.TwoYearsPlus.Level, eightKeys),
	}
}

func stageRiskV3(window string, level CompatibilityDurationLevel, evidenceKeys []string) CompatibilityStageRisk {
	main, trigger, advice := stageRiskTextV3(window, level)
	return CompatibilityStageRisk{
		Window:       window,
		RiskLevel:    string(level),
		MainRisk:     main,
		Trigger:      trigger,
		Advice:       advice,
		EvidenceKeys: evidenceKeys,
	}
}

func stageRiskTextV3(window string, level CompatibilityDurationLevel) (main, trigger, advice string) {
	switch window {
	case "three_months":
		switch level {
		case CompatibilityDurationHigh:
			return "靠近感强但节奏需要校准", "对方推进速度与你不同步时", "保持轻量频繁互动，不急于规则化关系。"
		case CompatibilityDurationMedium:
			return "短期吸引点有限", "缺乏话题或场景持续输入时", "刻意制造共同体验，避免单方追逐。"
		default:
			return "短期吸引基础薄弱", "热度退去后缺少留存点", "用现实生活节奏检验是否值得继续投入。"
		}
	case "one_year":
		switch level {
		case CompatibilityDurationHigh:
			return "亲密层稳固但仍需经营", "生活节奏被外部压力打乱时", "建立稳定的冲突修复机制。"
		case CompatibilityDurationMedium:
			return "亲密层有支撑但易波动", "情绪强度替代具体沟通时", "把分歧拆成具体事项，不情绪化判断关系本身。"
		default:
			return "亲密层缺乏天然契合", "对方亲密表达与你期待错位时", "先观察互相调整的意愿，再做长期承诺。"
		}
	default:
		switch level {
		case CompatibilityDurationHigh:
			return "长期稳定基础好", "责任分工与资源投入需要落地时", "建立可持续的责任分工与共同计划。"
		case CompatibilityDurationMedium:
			return "长期承接强度中等", "现实压力（住、家庭、收入）进入关系时", "用阶段性目标替代『未来无限期』式承诺。"
		default:
			return "长期承接薄弱", "需要共同处理重大现实议题时", "在做长期承诺前重新评估关系结构。"
		}
	}
}

func evidenceKeysByDimension(evidences []CompatibilityEvidence, dims ...string) []string {
	set := map[string]bool{}
	for _, d := range dims {
		set[d] = true
	}
	out := make([]string, 0, 2)
	for _, ev := range evidences {
		if set[ev.Dimension] && ev.EvidenceKey != "" {
			out = append(out, ev.EvidenceKey)
		}
	}
	return out
}

// buildRelationshipStrategyV3 按 recommendation 三档切换 12 句策略模板（4 句 × 3 档）。
func buildRelationshipStrategyV3(recommendation string) CompatibilityRelationshipStrategy {
	switch recommendation {
	case "continue":
		return CompatibilityRelationshipStrategy{
			Communication: "重要议题用明确约定替代情绪试探。",
			Conflict:      "冲突先暂停升级，再回到具体事件与责任分工。",
			Reality:       "长期计划拆成可验证的小步骤，逐项落地。",
			Boundary:      "保持双方个人节奏，避免过早形成单向依赖。",
		}
	case "observe":
		return CompatibilityRelationshipStrategy{
			Communication: "重要话题做到事先沟通规则，再讨论内容。",
			Conflict:      "争执后给彼此 24 小时冷却，再回到事实层处理。",
			Reality:       "用 1–2 个生活议题（出行、家庭联系）观察现实承接能力。",
			Boundary:      "在关系规则未稳定前，避免重大物质或时间投入。",
		}
	default:
		return CompatibilityRelationshipStrategy{
			Communication: "用具体行为而非情绪强度作为判断锚点。",
			Conflict:      "冲突后先评估是否值得继续修复，再决定行动。",
			Reality:       "把共同决策的频率与强度降到最低，先稳定个人节奏。",
			Boundary:      "明确可暂停 / 可退出的关系边界，避免被动滑入承诺。",
		}
	}
}

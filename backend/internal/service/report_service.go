package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"text/template"
	"time"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/pkg/bazi"
)

// BaziReportInput AI报告生成输入
type BaziReportInput struct {
	ChartID string
	Chart   *model.BaziChart
	Result  *bazi.BaziResult
}

// buildLifeStageHint 按"该段大运 age<18 的年份占比"生成 prompt 提示词
// youngCount=0 → 空（成人期不附加）
// youngCount=totalCount → 全段读书期
// 0 < youngCount < totalCount → 跨界期（前 N 年读书、后 M 年入社会）
func buildLifeStageHint(youngCount, totalCount int) string {
	if totalCount <= 0 || youngCount <= 0 {
		return ""
	}
	if youngCount == totalCount {
		return "本段大运全部年份处于读书期，请以学业、性格塑造、同窗关系为主轴撰写 summary。"
	}
	adultCount := totalCount - youngCount
	return fmt.Sprintf("本段大运跨越读书期与成人期（前 %d 年读书、后 %d 年入社会），summary 请分两段叙述：先讲读书期学业 / 性格，再讲成人期事业 / 婚恋。", youngCount, adultCount)
}

// LoadOrCalculateResult 加载或懒计算命盘的 BaziResult 快照
//
// 优先从 bazi_charts.result_json 反序列化（无 lunar-go 调用，毫秒级）；
// 老命盘 result_json 为空时，调一次 bazi.Calculate 并回写 DB；
// 之后所有下游（流年、大运、报告）都拿到一致的"原子化"原局，避免算法版本漂移。
func LoadOrCalculateResult(chart *model.BaziChart) (*bazi.BaziResult, error) {
	if chart == nil {
		return nil, fmt.Errorf("chart 为 nil")
	}
	raw, err := repository.GetChartResultJSON(chart.ID)
	if err != nil {
		return nil, fmt.Errorf("读取命盘快照失败: %v", err)
	}
	if len(raw) > 0 {
		var cached bazi.BaziResult
		if err := json.Unmarshal(raw, &cached); err == nil {
			bazi.EnsureTenGodRelation(&cached)
			return &cached, nil
		}
		// 反序列化失败：日志告警后回退到重新计算（防止脏数据卡死）
		log.Printf("[LoadOrCalculateResult] result_json 反序列化失败 chart_id=%s: %v；回退至 Calculate", chart.ID, err)
	}
	// 懒加载：调一次 lunar-go 算出 BaziResult，写回 DB
	result := bazi.Calculate(chart.BirthYear, chart.BirthMonth, chart.BirthDay, chart.BirthHour,
		chart.Gender, false, 0, chart.CalendarType, chart.IsLeapMonth)
	if marshalled, mErr := json.Marshal(result); mErr == nil {
		if sErr := repository.SaveChartResultJSON(chart.ID, marshalled); sErr != nil {
			log.Printf("[LoadOrCalculateResult] 写回 result_json 失败 chart_id=%s: %v", chart.ID, sErr)
		}
	}
	return result, nil
}

// formatYongshenInfo 将 BaziResult 的 yongshen 字段格式化为 prompt 可读的文案
// 优先使用 t0 调候命中信息（含具体天干），fallback 至旧版"用神/忌神"五行格式
func formatYongshenInfo(result *bazi.BaziResult) string {
	if result == nil {
		return ""
	}
	switch result.YongshenStatus {
	case bazi.YongshenStatusTiaohouHit:
		if len(result.YongshenGans) > 0 {
			return fmt.Sprintf("调候用神：%s（%s）／ 忌神：%s",
				strings.Join(result.YongshenGans, "、"), result.Yongshen, result.Jishen)
		}
	case bazi.YongshenStatusTiaohouMissFallback:
		miss := strings.Join(result.YongshenMissing, "、")
		if miss == "" {
			return fmt.Sprintf("用神（扶抑回退）：%s ／ 忌神：%s", result.Yongshen, result.Jishen)
		}
		return fmt.Sprintf("用神（扶抑回退，调候缺：%s）：%s ／ 忌神：%s",
			miss, result.Yongshen, result.Jishen)
	}
	if result.Yongshen != "" || result.Jishen != "" {
		return fmt.Sprintf("用神：%s ／ 忌神：%s", result.Yongshen, result.Jishen)
	}
	if result.Tiaohou != nil && len(result.Tiaohou.Expected) > 0 {
		return fmt.Sprintf("调候喜神：%s（综合用神信息缺失，以调候为参考）",
			strings.Join(result.Tiaohou.Expected, "、"))
	}
	return ""
}

// ===== 结构化报告 JSON 类型 =====

type reportAnalysis struct {
	Logic   string `json:"logic"`
	Summary string `json:"summary"`
}

type reportChapter struct {
	Title  string `json:"title"`
	Brief  string `json:"brief"`
	Detail string `json:"detail"`
}

type structuredReport struct {
	Yongshen string          `json:"yongshen"`
	Jishen   string          `json:"jishen"`
	Analysis reportAnalysis  `json:"analysis"`
	Chapters []reportChapter `json:"chapters"`
}

// buildBaziPrompt 构建八字报告 Prompt
// v2 优化：现代解读风格、月令格局推断CoT（子平真诠）、调候用神注入（穷通宝鉴）、
//
//	章节数据锚点、大运时间定位（注入当前年份）
func buildBaziPrompt(r *bazi.BaziResult) string {
	currentYear := time.Now().Year()

	joinOrNone := func(ss []string) string {
		if len(ss) == 0 {
			return "无"
		}
		return strings.Join(ss, "、")
	}

	firstShiShen := func(ss []string) string {
		if len(ss) == 0 {
			return "—"
		}
		return ss[0]
	}

	hideGanStr := func(hg []string) string {
		if len(hg) == 0 {
			return "无"
		}
		return strings.TrimSpace(strings.Join(hg, " "))
	}

	genderText := "女"
	if r.Gender == "male" {
		genderText = "男"
	}

	// ===大运序列===
	dayunStr := ""
	if r.StartYunSolar != "" {
		dayunStr += fmt.Sprintf("起运：%s\n", r.StartYunSolar)
	}
	for i, dy := range r.Dayun {
		if i >= 10 {
			break
		}
		shenshaStr := ""
		if len(dy.ShenSha) > 0 {
			shenshaStr = " 神煞=[" + strings.Join(dy.ShenSha, ",") + "]"
		}
		dayunStr += fmt.Sprintf("(%d) %d岁~%d岁（%d年起）：[%s%s] 干十神=%s 支十神=%s 长生=%s%s\n",
			i+1, dy.StartAge, dy.StartAge+9, dy.StartYear,
			dy.Gan, dy.Zhi,
			dy.GanShiShen, dy.ZhiShiShen, dy.DiShi,
			shenshaStr,
		)
	}
	if dayunStr == "" {
		dayunStr = "（暂无大运数据）\n"
	}

	// ===引擎五行统计初步参考===
	yongshenHint := ""
	if r.Yongshen != "" {
		yongshenHint = fmt.Sprintf(
			"\n[引擎五行统计初步参考]\n"+
				"引擎基于五行比例初步判断：用神=[%s]，忌神=[%s]\n"+
				"（此为统计参考值，请结合格局与调候综合确认）\n",
			r.Yongshen, r.Jishen,
		)
	}

	// ===调候用神===
	tiaohouStr := ""
	if r.Tiaohou != nil {
		// 构建透出/藏干状态描述
		touDesc := "无"
		if len(r.Tiaohou.Tou) > 0 {
			touDesc = strings.Join(r.Tiaohou.Tou, "、")
		}
		cangDesc := "无"
		if len(r.Tiaohou.Cang) > 0 {
			cangDesc = strings.Join(r.Tiaohou.Cang, "、")
		}
		expectedDesc := "无"
		if len(r.Tiaohou.Expected) > 0 {
			expectedDesc = strings.Join(r.Tiaohou.Expected, "、")
		}

		// 判断调候用神满足程度
		satisfyNote := ""
		touCount := len(r.Tiaohou.Tou)
		cangCount := len(r.Tiaohou.Cang)
		expectedCount := len(r.Tiaohou.Expected)
		if touCount > 0 && touCount >= expectedCount {
			satisfyNote = "→ 调候用神透干齐全，寒暖燥湿均衡，命局完整度高。"
		} else if touCount > 0 {
			satisfyNote = fmt.Sprintf("→ 调候用神部分透干（%d/%d），有一定调候基础。", touCount, expectedCount)
		} else if cangCount > 0 {
			satisfyNote = "→ 调候用神仅藏于地支，力量偏弱，需行运引出方能发挥。"
		} else {
			satisfyNote = "→ 调候用神完全缺失，寒暖失衡，命局存在明显短板，需大运补足。"
		}

		tiaohouStr = fmt.Sprintf(
			"\n[调候用神-穷通宝鉴精算]\n"+
				"日主[%s]生于[%s月]，调候理论指出：%s\n"+
				"理论调候用神：%s\n"+
				"本命局透干：%s\n"+
				"本命局藏干：%s\n"+
				"%s\n",
			r.DayGan, r.MonthZhi, r.Tiaohou.Text,
			expectedDesc,
			touDesc,
			cangDesc,
			satisfyNote,
		)
	}

	// ===系统定格结果===
	minggeStr := ""
	hasSystemMingge := strings.TrimSpace(r.MingGe) != ""
	if hasSystemMingge {
		minggeDesc := strings.TrimSpace(r.MingGeDesc)
		if minggeDesc == "" {
			minggeDesc = "该格局由后端命格引擎按月令透干优先法自动判定。"
		}
		minggeStr = fmt.Sprintf(
			"\n[系统定格结果]\n"+
				"本命局已由后端命格引擎完成定格，以下格局结论为唯一有效主格，请严格以此为准，不得重新改判格名。\n"+
				"主格：%s\n"+
				"格局说明：%s\n"+
				"解释边界：\n"+
				"1. 你可以判断此格局在原局中属于成格、破格、有救、偏弱，或受杂气干扰。\n"+
				"2. 你可以说明该格局与调候、用神、神煞之间的关系。\n"+
				"3. 你可以写“兼带某某倾向”或“局中亦见某某气象”，但不得把主格改写成其他格局。\n"+
				"4. 若你认为传统流派存在其他取格可能，只能作为补充说明，不得覆盖主格结论。\n",
			r.MingGe,
			minggeDesc,
		)
	}

	dayunYear := fmt.Sprintf("%d", currentYear)

	var stepOnePrompt string
	var stepTwoPrompt string
	if hasSystemMingge {
		stepOnePrompt = fmt.Sprintf("[第一步：三模块综合解释（在心中完成，不要在报告中输出计算过程）]\n"+
			"⚠️ 必须完整执行以下四步，不可合并或跳过任何模块。\n\n"+
			"a. 【调候判断 — 主导依据】\n"+
			"   月支=%s%s，主气十神=%s\n"+
			"   读取[调候用神-穷通宝鉴精算]区块，判断命局寒暖燥湿是否平衡；\n"+
			"   判断调候用神是否透干、是否得地、是否仍需大运补足；\n"+
			"   若调候明显失衡，需在最终分析中明确指出“调候优先”。\n\n"+
			"b. 【格局解释 — 以系统定格为准】\n"+
			"   主格固定为[系统定格结果]中的格局，不得改判；\n"+
			"   你只需判断该格局在原局中是否成格、是否受破、是否有救、是否偏弱；\n"+
			"   结合月令、透干、生扶、制化关系，解释此格为何成立、为何受损、为何仍可用或不可尽信；\n"+
			"   若局中同时存在其它明显结构，必须明确写出“兼带某某倾向”或“局中亦见某某气象”一句，再继续解释主次关系；\n"+
			"   格局模块用于解释主格，不再重新决定格局名称。\n\n"+
			"c. 【神煞修饰 — 辅助依据】\n"+
			"   扫描命盘全部神煞（天乙贵人/驿马/羊刃/桃花/华盖等）；\n"+
			"   判断关键神煞如何修饰人格特征、事件色彩与运势表现；\n"+
			"   神煞不得推翻系统定格结果，也不得单独决定最终喜忌。\n\n"+
			"d. 【综合结论】\n"+
			"   先确认系统主格的成立程度，再结合调候与全局五行，判断最终喜用神与忌神；\n"+
			"   若格局倾向与调候方向不一致，必须明确说明主次，例如：“虽以某格立局，但原局寒湿偏重，仍当先取火土调候为先。”\n\n",
			r.MonthGan, r.MonthZhi, firstShiShen(r.MonthZhiShiShen),
		)
		stepTwoPrompt = "[第二步：生成命局分析总览 analysis.logic]\n" +
			"写一段整体分析（500-800字），专业命理分析风格，如实批断，不偏不倚，并严格遵循以下结构：\n" +
			"- 开头必须先写系统主格，使用“此命以【X格】立局”或同义表达；\n" +
			"- 紧接着说明该格局在原局中属于成格、破格、有救、偏弱中的哪一种；\n" +
			"- 然后说明该格局与调候、最终喜用神之间的关系；\n" +
			"- 首段直接说明命盘的核心喜用神与忌神。阐述依据时，需通过自然语言融合调候、格局与神煞的结论。绝对禁止出现“百分比”、“加权”、“权重”、“65%”等机械化算分词汇。\n" +
			"- 若模块存在冲突（如格局倾向与调候方向不一致），请用专业口吻解释主次取舍（例如：“虽以正官格立局，但命局偏寒，首当以火土暖局为要”）。\n" +
			"- 若局中兼象明显，必须显式写出“兼带某某倾向”或“局中亦见某某气象”，但不得覆盖系统主格。\n" +
			"- 接续说明日干强弱依据（月令得令、失令状态）\n" +
			"- 提取1~2个最亮眼的特征或神煞进行个性点评\n" +
			"- 若命局存在明显缺陷（如格局遭破、用神缺失、忌神过重），必须如实指出，不可回避或美化。\n\n" +
			"[格局一致性约束]\n" +
			"最终输出中，凡涉及格局主名的表述，必须与[系统定格结果]一致。\n" +
			"若存在其它结构，只能写为“兼带某某倾向”或“局中亦见某某气象”，不得覆盖系统主格。\n\n"
	} else {
		stepOnePrompt = fmt.Sprintf("[第一步：三模块加权推断（在心中完成，不要在报告中输出计算过程）]\n"+
			"⚠️ 必须完整执行以下四步，不可合并或跳过任何模块。\n\n"+
			"a. 【调候用神评分 — 权重65票】\n"+
			"   月支=%s%s，主气十神=%s\n"+
			"   读取[调候用神-穷通宝鉴精算]区块的精算数据和大运征兆；\n"+
			"   对每个五行/天干判断方向：喜→该五行 +65票，忌→该五行 -65票。\n\n"+
			"b. 【格局评分 — 权重25票】\n"+
			"   严格按 System Prompt 中的【格局判断规则】公式执行：\n"+
			"   ①查月支藏干权重表，按权重顺序检查是否透出天干；\n"+
			"   ②有透干者以透干十神定格，无透干者以月令本气定格（弱格）；\n"+
			"   ③按知识库「格局高低」判断成格/破格；\n"+
			"   ④按知识库「用神取法」确定格局喜用神；\n"+
			"   对格局喜用神五行 +25票，忌神五行 -25票。\n\n"+
			"c. 【神煞综合评分 — 权重10票】\n"+
			"   扫描命盘全部神煞（天乙贵人/驿马/羊刃/桃花/华盖等）；\n"+
			"   判断每个神煞对各五行的影响方向：利→ +10票，不利→ -10票，中性→ 0票。\n\n"+
			"d. 【加权合并】\n"+
			"   对每个五行汇总 a+b+c 的总票数；\n"+
			"   总分为正 → 喜用神；总分为负 → 忌神；\n"+
			"   若模块间方向矛盾，以票数高者为准，标注「[模块]与调候存在出入，以调候优先」。\n\n",
			r.MonthGan, r.MonthZhi, firstShiShen(r.MonthZhiShiShen),
		)
		stepTwoPrompt = "[第二步：生成命局分析总览 analysis.logic]\n" +
			"写一段整体分析（500-800字），专业命理分析风格，如实批断，不偏不倚：\n" +
			"- 首段直接说明命盘的核心喜用神与忌神。阐述依据时，需通过自然语言融合调候、格局与神煞的结论。绝对禁止出现“百分比”、“加权”、“权重”、“65%”等机械化算分词汇。\n" +
			"- 若模块存在冲突（如格局用神与调候用神不一致），请用专业口吻解释主次取舍（例如：“虽因格局喜水木，但命局偏寒，首当以火土暖局为要”）。\n" +
			"- 接续说明日干强弱依据（月令得令、失令状态）\n" +
			"- 提取1~2个最亮眼的特征或神煞进行个性点评\n" +
			"- 若命局存在明显缺陷（如用神缺失、忌神过重、格局遭破），必须如实指出，不可回避或美化。\n\n"
	}

	prompt := "以下命盘数据已由算法精确计算，请基于精算结果进行深度命理解读。\n\n" +
		fmt.Sprintf("[八字命盘]\n"+
			"年柱：%s%s（%s%s）| 藏干：%s\n"+
			"月柱：%s%s（%s%s）| 藏干：%s\n"+
			"日柱：%s%s（%s%s）| 藏干：%s <- 日干代表本人\n"+
			"时柱：%s%s（%s%s）| 藏干：%s\n\n"+
			"纳音：年柱[%s] 月柱[%s] 日柱[%s] 时柱[%s]\n\n"+
			"五行分布：木%d个 火%d个 土%d个 金%d个 水%d个\n\n"+
			"性别：%s\n\n",
			r.YearGan, r.YearZhi, r.YearGanWuxing, r.YearZhiWuxing, hideGanStr(r.YearHideGan),
			r.MonthGan, r.MonthZhi, r.MonthGanWuxing, r.MonthZhiWuxing, hideGanStr(r.MonthHideGan),
			r.DayGan, r.DayZhi, r.DayGanWuxing, r.DayZhiWuxing, hideGanStr(r.DayHideGan),
			r.HourGan, r.HourZhi, r.HourGanWuxing, r.HourZhiWuxing, hideGanStr(r.HourHideGan),
			r.YearNaYin, r.MonthNaYin, r.DayNaYin, r.HourNaYin,
			r.Wuxing.Mu,
			r.Wuxing.Huo,
			r.Wuxing.Tu,
			r.Wuxing.Jin,
			r.Wuxing.Shui,
			genderText,
		) +
		fmt.Sprintf("[十神关系-算法精算]\n"+
			"天干：年干%s=[%s] | 月干%s=[%s] | 日干%s=[日主] | 时干%s=[%s]\n"+
			"地支主气：年支%s=[%s] | 月支%s=[%s] | 日支%s=[%s] | 时支%s=[%s]\n\n",
			r.YearGan, r.YearGanShiShen,
			r.MonthGan, r.MonthGanShiShen,
			r.DayGan,
			r.HourGan, r.HourGanShiShen,
			r.YearZhi, firstShiShen(r.YearZhiShiShen),
			r.MonthZhi, firstShiShen(r.MonthZhiShiShen),
			r.DayZhi, firstShiShen(r.DayZhiShiShen),
			r.HourZhi, firstShiShen(r.HourZhiShiShen),
		) +
		fmt.Sprintf("[十二长生]\n年柱[%s] | 月柱[%s] | 日柱[%s] | 时柱[%s]\n\n"+
			"[旬空-空亡]\n年柱[%s] | 月柱[%s] | 日柱[%s] | 时柱[%s]\n\n"+
			"[神煞]\n年柱：%s | 月柱：%s | 日柱：%s | 时柱：%s\n\n"+
			"[大运序列]\n%s",
			r.YearDiShi, r.MonthDiShi, r.DayDiShi, r.HourDiShi,
			r.YearXunKong, r.MonthXunKong, r.DayXunKong, r.HourXunKong,
			joinOrNone(r.YearShenSha), joinOrNone(r.MonthShenSha),
			joinOrNone(r.DayShenSha), joinOrNone(r.HourShenSha),
			dayunStr,
		) +
		yongshenHint +
		tiaohouStr +
		minggeStr +
		"\n" +
		stepOnePrompt +
		stepTwoPrompt +
		"[第三步：六章节报告指引]\n" +
		"各个章节在进行深度分析时，请务必参考以下依据（每部分不仅要写精简版结论，还要写专业版拆解）：\n" +
		"- 精简版：每章约80-120字，结论先行，用普通用户能快速理解的话说明核心判断。\n" +
		"- 专业版：每章约220-350字，必须自然包含结论、命理依据、现实表现、建议四类信息；不要用机械小标题堆砌，要写成连贯、可阅读的专业解读。\n" +
		"- 术语出现后必须紧跟白话解释。遇到印星、官杀、食伤、财星、用神、忌神、调候、格局等术语时，要说明它在现实中对应学习证书、事业压力、表达才华、钱财资源、助力阻力、气候平衡或人生结构等含义。\n" +
		"- 【全章节通用语言风格】所有章节（性格/感情/事业/健康/大运）均严禁出现「百分比」「%」「加权」「权重」「占比 N%」「N/8」「打分」「评分 N 分」等机械算分词汇；五行强弱必须用「旺/相/休/囚/死」「过旺/偏旺/平衡/偏弱/缺」等命理传统术语表达。\n" +
		"- 【性格特质】参考日干天性、日支十二长生、关键神煞等，既指出优势也需指出潜在短板。\n" +
		"- 【感情运势】男看财女看官，结合配偶宫（日支）与桃花红鸾等神煞，分析婚姻稳定性。\n" +
		"- 【事业财运】看官杀与食伤透出情况、财星旺衰、天乙与驿马等，指明职业方向与瓶颈。\n" +
		"- 【健康提示】看全局五行最旺与最弱项对应的脏腑（木肝胆/火心脑/土脾胃/金肺肠/水肾膀），给出养生建议。\n" +
		"- 【大运走势】当前年份=" + dayunYear + "年，结合给出的大运序列推算所处步次。请重点解读上一步大运（回顾）、当前大运（现状）和下一步大运（展望）。\n\n" +
		"[第四步：严格输出 Markdown 格式]\n" +
		"请必须使用以下严格的 Markdown 标题结构来输出最终结果。必须包含以下所有二级标题，并且标题名称必须一字不差，不要输出额外的多余信息或开头语：\n\n" +
		"## 【喜用神】\n" +
		"仅回答金、木、水、火、土中的一种或两种组合\n\n" +
		"## 【忌神】\n" +
		"仅回答金、木、水、火、土中的一种或两种组合\n\n" +
		"## 【命理摘要】\n" +
		"一句话命局核心特质（30字以内）\n\n" +
		"## 【命局分析总览】\n" +
		"第二步的命局分析总览全文\n\n" +
		"## 【性格特质-精简版】\n\n## 【性格特质-专业版】\n\n" +
		"## 【感情运势-精简版】\n\n## 【感情运势-专业版】\n\n" +
		"## 【事业财运-精简版】\n\n## 【事业财运-专业版】\n\n" +
		"## 【健康提示-精简版】\n\n## 【健康提示-专业版】\n\n" +
		"## 【大运走势-精简版】\n\n## 【大运走势-专业版】\n"

	return prompt
}

// GenerateAIReport 生成 AI 报告（始终调 LLM，生成后覆盖存库）
func GenerateAIReport(chartID string, result *bazi.BaziResult, userID *string) (*model.AIReport, error) {
	// 构建 Prompt 并调用 AI
	prompt := buildBaziPrompt(result)
	rawContent, modelName, providerID, durationMs, usage, aiErr := callAIWithSystem(prompt)

	// 记录调用日志（无论成功失败）
	status, errMsg := "success", ""
	if aiErr != nil {
		status, errMsg = "error", aiErr.Error()
	}
	repository.CreateAIRequestLog(chartID, providerID, modelName, durationMs, status, errMsg)
	log.Printf("[TokenUsage-DEBUG] report aiErr=%v usage={prompt=%d completion=%d total=%d} userID=%v",
		aiErr, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, userID)
	if aiErr == nil {
		go func() {
			if logErr := repository.CreateTokenUsageLog(userID, &chartID, "report", modelName, providerID,
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens,
				prompt, rawContent); logErr != nil {
				log.Printf("[TokenUsage] 写入失败: %v", logErr)
			}
		}()
	}

	if aiErr != nil {
		return nil, aiErr
	}

	// ===== 解析 AI 返回内容 =====

	// 清理 Markdown 代码块标记
	cleanJSON := strings.TrimSpace(rawContent)
	if strings.HasPrefix(cleanJSON, "```json") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
		cleanJSON = strings.TrimSuffix(strings.TrimSpace(cleanJSON), "```")
	} else if strings.HasPrefix(cleanJSON, "```") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```")
		cleanJSON = strings.TrimSuffix(strings.TrimSpace(cleanJSON), "```")
	}

	// 提取 {} 之间的内容
	firstBrace := strings.Index(cleanJSON, "{")
	lastBrace := strings.LastIndex(cleanJSON, "}")
	if firstBrace != -1 && lastBrace != -1 && lastBrace > firstBrace {
		cleanJSON = cleanJSON[firstBrace : lastBrace+1]
	}

	parsed, briefContent, contentStructured := parseAIReportContent(rawContent, cleanJSON)
	if briefContent == "" {
		fmt.Printf("[AI Report] 解析失败，使用原始内容。原始返回（前200字）：%s\n", rawContent[:min(200, len(rawContent))])
		briefContent = rawContent
	}

	// 回写喜忌到 chart
	if chartID != "" && parsed != nil && (parsed.Yongshen != "" || parsed.Jishen != "") {
		repository.UpdateChartYongshenJishen(chartID, parsed.Yongshen, parsed.Jishen)
		result.Yongshen = parsed.Yongshen
		result.Jishen = parsed.Jishen
	}

	// 追加免责声明到 content 字段
	content := briefContent + "\n\n---\n*本报告由 AI 辅助生成，内容仅供参考，不构成任何决策建议。*"

	// 保存报告（同时写入 content 和 content_structured）
	report, err := repository.CreateReport(chartID, content, modelName, contentStructured)
	if err != nil {
		return nil, err
	}
	return report, nil
}

// GenerateAIReportStream 流式生成 AI 报告
func GenerateAIReportStream(chartID string, result *bazi.BaziResult, userID *string, onData func(string) error, onThinking func() error) error {
	t0 := time.Now()

	// 构建 Prompt
	prompt := buildBaziPrompt(result)
	log.Printf("[ReportStream T+%dms] Prompt 构建完成 (长度=%d 字符)", time.Since(t0).Milliseconds(), len(prompt))

	rawContent, modelName, providerID, durationMs, usage, aiErr := StreamAIWithSystem(prompt, onData, onThinking)
	log.Printf("[ReportStream T+%dms] AI 流式调用结束, model=%s, duration=%dms, err=%v", time.Since(t0).Milliseconds(), modelName, durationMs, aiErr)

	status, errMsg := "success", ""
	if aiErr != nil {
		status, errMsg = "error", aiErr.Error()
	}
	repository.CreateAIRequestLog(chartID, providerID, modelName, durationMs, status, errMsg)
	log.Printf("[TokenUsage-DEBUG] aiErr=%v usage={prompt=%d completion=%d total=%d} userID=%v",
		aiErr, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, userID)
	if aiErr == nil {
		go func() {
			if logErr := repository.CreateTokenUsageLog(userID, &chartID, "report_stream", modelName, providerID,
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens,
				prompt, rawContent); logErr != nil {
				log.Printf("[TokenUsage] 写入失败: %v", logErr)
			}
		}()
	}

	if aiErr != nil {
		return aiErr
	}

	// 流结束后，异步或同步将 rawContent 处理成 JSON 持久化，供分享和旧接口使用
	parsed, briefContent, contentStructured := parseAIReportContent(rawContent, "")
	if parsed != nil && contentStructured != nil {
		if chartID != "" && (parsed.Yongshen != "" || parsed.Jishen != "") {
			repository.UpdateChartYongshenJishen(chartID, parsed.Yongshen, parsed.Jishen)
		}
	} else {
		fmt.Println("[Stream] Markdown 解析为结构化 JSON 失败。尝试兜底策略")
	}

	if briefContent == "" {
		briefContent = rawContent
	}
	content := briefContent + "\n\n---\n*本报告由 AI 辅助生成，内容仅供参考，不构成任何决策建议。*"
	_, _ = repository.CreateReport(chartID, content, modelName, contentStructured)

	return nil
}

func extractSection(md, header string) string {
	idx := strings.Index(md, "## 【"+header+"】")
	if idx == -1 {
		return ""
	}
	sub := md[idx+len("## 【"+header+"】"):]
	nextIdx := strings.Index(sub, "## 【")
	if nextIdx != -1 {
		sub = sub[:nextIdx]
	}
	return strings.TrimSpace(sub)
}

// ParseMarkdownToStructured 将 AI 生成的 Markdown 解析回 JSON 结构
func ParseMarkdownToStructured(md string) (*structuredReport, string) {
	var parsed structuredReport
	parsed.Yongshen = extractSection(md, "喜用神")
	parsed.Jishen = extractSection(md, "忌神")
	parsed.Analysis.Summary = extractSection(md, "命理摘要")
	parsed.Analysis.Logic = extractSection(md, "命局分析总览")

	chapters := []string{"性格特质", "感情运势", "事业财运", "健康提示", "大运走势"}
	for _, title := range chapters {
		brief := extractSection(md, title+"-精简版")
		detail := extractSection(md, title+"-专业版")
		if brief != "" || detail != "" {
			parsed.Chapters = append(parsed.Chapters, reportChapter{
				Title:  title,
				Brief:  brief,
				Detail: detail,
			})
		}
	}

	// 拼装备用版 Markdown
	var parts []string
	if parsed.Analysis.Summary != "" {
		parts = append(parts, "【命局概要】\n"+parsed.Analysis.Summary)
	}
	for _, ch := range parsed.Chapters {
		if ch.Brief != "" {
			parts = append(parts, "【"+ch.Title+"】\n"+ch.Brief)
		} else if ch.Detail != "" {
			parts = append(parts, "【"+ch.Title+"】\n"+ch.Detail)
		}
	}

	briefText := strings.Join(parts, "\n\n")

	return &parsed, briefText
}

func parseAIReportContent(rawContent, cleanJSON string) (*structuredReport, string, *json.RawMessage) {
	if parsed, brief := ParseMarkdownToStructured(rawContent); parsed != nil && len(parsed.Chapters) > 0 && parsed.Analysis.Logic != "" {
		raw, _ := json.Marshal(parsed)
		rawMsg := json.RawMessage(raw)
		return parsed, brief, &rawMsg
	}

	var parsed structuredReport
	trimmedJSON := strings.TrimSpace(cleanJSON)
	if trimmedJSON == "" {
		return &parsed, "", nil
	}

	if errParse := json.Unmarshal([]byte(trimmedJSON), &parsed); errParse != nil {
		fixedJSON := fixJSONStrings(trimmedJSON)
		if errFix := json.Unmarshal([]byte(fixedJSON), &parsed); errFix == nil {
			trimmedJSON = fixedJSON
		}
	}
	if len(parsed.Chapters) > 0 && parsed.Analysis.Logic != "" {
		_, brief := ParseMarkdownToStructured(rawContent)
		raw, _ := json.Marshal(parsed)
		rawMsg := json.RawMessage(raw)
		return &parsed, brief, &rawMsg
	}

	type legacyResult struct {
		Yongshen string `json:"yongshen"`
		Jishen   string `json:"jishen"`
		Report   string `json:"report"`
	}
	var legacy legacyResult
	briefContent := ""
	if errLegacy := json.Unmarshal([]byte(trimmedJSON), &legacy); errLegacy == nil && legacy.Report != "" {
		parsed.Yongshen = legacy.Yongshen
		parsed.Jishen = legacy.Jishen
		briefContent = legacy.Report
	} else {
		importRegexp := regexp.MustCompile(`"yongshen"\s*:\s*"([^"]+)"`)
		matchY := importRegexp.FindStringSubmatch(trimmedJSON)
		if len(matchY) > 1 {
			parsed.Yongshen = matchY[1]
		}
		jishenRegexp := regexp.MustCompile(`"jishen"\s*:\s*"([^"]+)"`)
		matchJ := jishenRegexp.FindStringSubmatch(trimmedJSON)
		if len(matchJ) > 1 {
			parsed.Jishen = matchJ[1]
		}
		reportRegexp := regexp.MustCompile(`(?s)"report"\s*:\s*"(.*?)"\s*}`)
		matchR := reportRegexp.FindStringSubmatch(trimmedJSON)
		if len(matchR) > 1 && strings.TrimSpace(matchR[1]) != "" {
			briefContent = strings.TrimSpace(matchR[1])
		}
	}

	return &parsed, briefContent, nil
}

// fixJSONStrings 用状态机扫描 JSON，将字符串字段内的真实控制字符（\n \r \t）
// 转义为合法的 JSON 转义序列，修复 AI 输出中未转义换行导致的解析失败。
func fixJSONStrings(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	inString := false
	escaped := false
	for _, c := range s {
		if escaped {
			buf.WriteRune(c)
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			buf.WriteRune(c)
			continue
		}
		if c == '"' {
			inString = !inString
			buf.WriteRune(c)
			continue
		}
		// 仅在字符串内替换控制字符
		if inString {
			switch c {
			case '\n':
				buf.WriteString(`\n`)
			case '\r':
				buf.WriteString(`\r`)
			case '\t':
				buf.WriteString(`\t`)
			default:
				buf.WriteRune(c)
			}
			continue
		}
		buf.WriteRune(c)
	}
	return buf.String()
}

// GenerateLiunianReport 生成流年运势分析（始终调 LLM，生成后 upsert 存库）
func GenerateLiunianReport(chartID string, targetYear int, userID *string) (*model.AILiunianReport, error) {
	// 1. 读取排盘
	chart, err := repository.GetChartByID(chartID)
	if err != nil || chart == nil {
		return nil, fmt.Errorf("无此排盘记录")
	}

	// 读取原局分析文本（如果生成过原局报告）
	natalLogic := "用户大运起伏与原局特质正在展开。"
	natalReport, _ := repository.GetReportByChartID(chartID)
	if natalReport != nil && natalReport.ContentStructured != nil {
		var parsed structuredReport
		if err := json.Unmarshal(*natalReport.ContentStructured, &parsed); err == nil && parsed.Analysis.Logic != "" {
			natalLogic = parsed.Analysis.Logic
		}
	}

	result, err := LoadOrCalculateResult(chart)
	if err != nil {
		return nil, err
	}

	var currentDayun string
	var currentDayunGSS string
	var currentDayunZSS string
	var lnGanZhi string
	var lnGanShiShen string
	var lnZhiShiShen string

	// 从历遍中抓取流年
	for _, dy := range result.Dayun {
		for _, ln := range dy.LiuNian {
			if ln.Year == targetYear {
				currentDayun = dy.Gan + dy.Zhi
				currentDayunGSS = dy.GanShiShen
				currentDayunZSS = dy.ZhiShiShen
				lnGanZhi = ln.GanZhi
				lnGanShiShen = ln.GanShiShen
				lnZhiShiShen = ln.ZhiShiShen
				break
			}
		}
		if lnGanZhi != "" {
			break
		}
	}

	// 构造模板数据
	tplData := model.LiunianTemplateData{
		NatalAnalysisLogic:     natalLogic,
		CurrentDayunGanZhi:     currentDayun,
		CurrentDayunGanShiShen: currentDayunGSS,
		CurrentDayunZhiShiShen: currentDayunZSS,
		TargetYear:             targetYear,
		TargetYearGanZhi:       lnGanZhi,
		TargetYearGanShiShen:   lnGanShiShen,
		TargetYearZhiShiShen:   lnZhiShiShen,
	}

	// 3. 读取 Prompt 模板
	promptConfig, err := repository.GetPromptByModule("liunian")
	if err != nil || promptConfig == nil {
		return nil, fmt.Errorf("未找到系统预设的流年Prompt")
	}

	tmpl, err := template.New("liunian").Parse(promptConfig.Content)
	if err != nil {
		return nil, fmt.Errorf("后台Prompt模板语法错误: %v", err)
	}

	var parsedPrompt bytes.Buffer
	if err := tmpl.Execute(&parsedPrompt, tplData); err != nil {
		return nil, fmt.Errorf("拼接Prompt上下文失败: %v", err)
	}

	// 4. 调用 AI（使用知识库增强的 System Prompt）
	rawContent, modelName, providerID, durationMs, usage, aiErr := callAIWithSystem(parsedPrompt.String())
	status, errMsg := "success", ""
	if aiErr != nil {
		status, errMsg = "error", aiErr.Error()
		repository.CreateAIRequestLog(chartID, providerID, modelName, durationMs, status, errMsg)
		return nil, aiErr
	}
	repository.CreateAIRequestLog(chartID, providerID, modelName, durationMs, status, errMsg)
	go func() {
		if logErr := repository.CreateTokenUsageLog(userID, &chartID, "liunian", modelName, providerID,
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens,
			parsedPrompt.String(), rawContent); logErr != nil {
			log.Printf("[TokenUsage] 写入失败: %v", logErr)
		}
	}()

	// 解析 JSON
	cleanJSON := strings.TrimSpace(rawContent)
	if strings.HasPrefix(cleanJSON, "```json") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
		cleanJSON = strings.TrimSuffix(strings.TrimSpace(cleanJSON), "```")
	} else if strings.HasPrefix(cleanJSON, "```") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```")
		cleanJSON = strings.TrimSuffix(strings.TrimSpace(cleanJSON), "```")
	}
	firstBrace := strings.Index(cleanJSON, "{")
	lastBrace := strings.LastIndex(cleanJSON, "}")
	if firstBrace != -1 && lastBrace != -1 && lastBrace > firstBrace {
		cleanJSON = cleanJSON[firstBrace : lastBrace+1]
	}

	// 由于是简单的 map结构，直接检查是否可用
	var reportData map[string]interface{}
	if err := json.Unmarshal([]byte(cleanJSON), &reportData); err != nil {
		// 尝试修复换行符
		fixedJSON := fixJSONStrings(cleanJSON)
		if errFix := json.Unmarshal([]byte(fixedJSON), &reportData); errFix != nil {
			return nil, fmt.Errorf("解析AI流年输出失败: %v", errFix)
		}
		cleanJSON = fixedJSON
	}

	rawMsg := json.RawMessage(cleanJSON)

	// 5. 存入数据库
	return repository.CreateLiunianReport(chartID, targetYear, currentDayun, &rawMsg, modelName)
}

// GeneratePastEventsStream 过往年份事件推算 SSE 流式生成（始终调 LLM，生成后 upsert 存库）
func GeneratePastEventsStream(chartID string, userID *string, onData func(string) error, onThinking func() error) error {
	// 1. 读取排盘
	chart, err := repository.GetChartByID(chartID)
	if err != nil || chart == nil {
		return fmt.Errorf("未找到排盘记录")
	}

	result, err := LoadOrCalculateResult(chart)
	if err != nil {
		return err
	}

	// 3. 算法扫描所有过往年份信号
	currentYear := time.Now().Year()
	// 从起运年龄开始（起运前无大运加持，信号意义有限）
	minAge := 0
	if len(result.Dayun) > 0 {
		minAge = result.Dayun[0].StartAge
	}
	yearSignals := bazi.GetAllYearSignals(result, chart.Gender, currentYear, minAge)

	yearsJSON, err := json.Marshal(yearSignals)
	if err != nil {
		return fmt.Errorf("序列化年份信号失败: %v", err)
	}

	// 4. 原局概要
	natalSummary := fmt.Sprintf(
		"年柱%s%s（%s/%s） 月柱%s%s（%s/%s） 日柱%s%s（%s/%s） 时柱%s%s（%s/%s）",
		result.YearGan, result.YearZhi, result.YearGanShiShen, result.YearZhiShiShen,
		result.MonthGan, result.MonthZhi, result.MonthGanShiShen, result.MonthZhiShiShen,
		result.DayGan, result.DayZhi, "日元", result.DayZhiShiShen,
		result.HourGan, result.HourZhi, result.HourGanShiShen, result.HourZhiShiShen,
	)

	// 5. 读取 Prompt 模板
	promptConfig, err := repository.GetPromptByModule("past_events")
	if err != nil || promptConfig == nil {
		return fmt.Errorf("未找到 past_events Prompt 配置")
	}

	genderLabel := "男"
	if chart.Gender == "female" {
		genderLabel = "女"
	}

	// 构建大运列表描述，供 AI 撰写大运整体总结
	var dayunLines []string
	for _, dy := range result.Dayun {
		dayunLines = append(dayunLines, fmt.Sprintf("大运 %s%s %d-%d岁（%d-%d年） [%s/%s]",
			dy.Gan, dy.Zhi, dy.StartAge, dy.StartAge+9,
			dy.StartYear, dy.EndYear,
			dy.GanShiShen, dy.ZhiShiShen,
		))
	}

	// 用神/忌神信息（t0 调候优先，fallback 至扶抑）
	yongshenInfo := formatYongshenInfo(result)

	// 大运合化标签（汇总所有合化大运）
	dayunHuahe := strings.Join(bazi.CollectDayunHuaheLines(result), "\n")

	// 加权身强弱评分明细
	level, score, detail := bazi.GetStrengthDetail(result)
	levelMap := map[string]string{
		"vstrong": "极强", "strong": "强", "neutral": "中和", "weak": "弱", "vweak": "极弱",
	}
	strengthDetail := fmt.Sprintf("%s(评分%d): %s", levelMap[level], score, detail)

	// 原局格局描述（暂无格局判定算法，留空字符串；后续 change 补足）
	gejuSummary := ""

	tplData := model.PastEventsTemplateData{
		Gender:         genderLabel,
		DayGan:         result.DayGan,
		NatalSummary:   natalSummary,
		YearsData:      string(yearsJSON),
		DayunList:      strings.Join(dayunLines, "\n"),
		YongshenInfo:   yongshenInfo,
		GejuSummary:    gejuSummary,
		DayunHuahe:     dayunHuahe,
		StrengthDetail: strengthDetail,
	}

	tmpl, err := template.New("past_events").Parse(promptConfig.Content)
	if err != nil {
		return fmt.Errorf("past_events Prompt 模板语法错误: %v", err)
	}

	var parsedPrompt bytes.Buffer
	if err := tmpl.Execute(&parsedPrompt, tplData); err != nil {
		return fmt.Errorf("Prompt模板渲染失败: %v", err)
	}

	// 6. SSE 流式 AI 调用（过往事件推算固定关闭推理思考模式，避免 Qwen3 等推理模型 3 分钟+思考阶段）
	rawContent, modelName, providerID, durationMs, usage, aiErr := StreamAIWithSystemNoThink(parsedPrompt.String(), onData, onThinking)
	status, errMsg := "success", ""
	if aiErr != nil {
		status, errMsg = "error", aiErr.Error()
	}
	repository.CreateAIRequestLog(chartID, providerID, modelName, durationMs, status, errMsg)
	if aiErr == nil {
		go func() {
			if logErr := repository.CreateTokenUsageLog(userID, &chartID, "dayun", modelName, providerID,
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens,
				parsedPrompt.String(), rawContent); logErr != nil {
				log.Printf("[TokenUsage] 写入失败: %v", logErr)
			}
		}()
	}
	if aiErr != nil {
		return aiErr
	}

	// 7. 解析并存库
	cleanJSON := strings.TrimSpace(rawContent)
	if strings.HasPrefix(cleanJSON, "```json") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
		cleanJSON = strings.TrimSuffix(strings.TrimSpace(cleanJSON), "```")
	} else if strings.HasPrefix(cleanJSON, "```") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```")
		cleanJSON = strings.TrimSuffix(strings.TrimSpace(cleanJSON), "```")
	}
	if i := strings.Index(cleanJSON, "{"); i > 0 {
		cleanJSON = cleanJSON[i:]
	}
	if i := strings.LastIndex(cleanJSON, "}"); i >= 0 && i < len(cleanJSON)-1 {
		cleanJSON = cleanJSON[:i+1]
	}

	rawMsg := json.RawMessage(cleanJSON)
	_, _ = repository.CreatePastEvents(chartID, &rawMsg, modelName)

	return nil
}

// ─── 思路 E：模板化年份 + 分段大运 AI ─────────────────────────────────────────

// PastEventsYearItem 单个年份的算法+模板叙述
type PastEventsYearItem struct {
	Year            int                     `json:"year"`
	Age             int                     `json:"age"`
	GanZhi          string                  `json:"gan_zhi"`
	DayunGanZhi     string                  `json:"dayun_gan_zhi"`
	DayunIndex      int                     `json:"dayun_index"`
	YearInDayun     int                     `json:"year_in_dayun,omitempty"`
	DayunPhase      string                  `json:"dayun_phase,omitempty"`
	TenGodPower     bazi.TenGodPowerProfile `json:"ten_god_power,omitempty"`
	Signals         []string                `json:"signals"`
	Narrative       string                  `json:"narrative"`
	EvidenceSummary []string                `json:"evidence_summary,omitempty"`
}

// PastEventsYearsResponse 一次性返回所有年份的算法叙述（无 AI）
type PastEventsYearsResponse struct {
	Years     []PastEventsYearItem `json:"years"`
	DayunMeta []DayunMetaItem      `json:"dayun_meta"`
	Generated string               `json:"generated_by"`
}

type DayunMetaItem struct {
	Index       int                     `json:"index"`
	GanZhi      string                  `json:"gan_zhi"`
	StartAge    int                     `json:"start_age"`
	EndAge      int                     `json:"end_age"`
	StartYr     int                     `json:"start_year"`
	EndYr       int                     `json:"end_year"`
	TenGodPower bazi.TenGodPowerProfile `json:"ten_god_power,omitempty"`
}

// GeneratePastEventsYears 根据算法 + 模板生成所有年份叙述（毫秒级，无 AI）
func GeneratePastEventsYears(chartID string) (*PastEventsYearsResponse, error) {
	chart, err := repository.GetChartByID(chartID)
	if err != nil || chart == nil {
		return nil, fmt.Errorf("未找到排盘记录")
	}

	result, err := LoadOrCalculateResult(chart)
	if err != nil {
		return nil, err
	}

	currentYear := time.Now().Year()
	minAge := 0
	if len(result.Dayun) > 0 {
		minAge = result.Dayun[0].StartAge
	}
	yearSignals := bazi.GetAllYearSignals(result, chart.Gender, currentYear, minAge)

	// 构造大运索引（dayunGanzhi → index）
	dyIndex := map[string]int{}
	dayunMeta := make([]DayunMetaItem, 0, len(result.Dayun))
	for _, dy := range result.Dayun {
		gz := dy.Gan + dy.Zhi
		dyIndex[gz] = dy.Index
		dayunMeta = append(dayunMeta, DayunMetaItem{
			Index:       dy.Index,
			GanZhi:      gz,
			StartAge:    dy.StartAge,
			EndAge:      dy.StartAge + 9,
			StartYr:     dy.StartYear,
			EndYr:       dy.EndYear,
			TenGodPower: bazi.BuildDayunTenGodPower(result, dy),
		})
	}

	years := make([]PastEventsYearItem, 0, len(yearSignals))
	for _, ys := range yearSignals {
		years = append(years, PastEventsYearItem{
			Year:            ys.Year,
			Age:             ys.Age,
			GanZhi:          ys.GanZhi,
			DayunGanZhi:     ys.DayunGanZhi,
			DayunIndex:      dyIndex[ys.DayunGanZhi],
			YearInDayun:     ys.YearInDayun,
			DayunPhase:      ys.DayunPhase,
			TenGodPower:     ys.TenGodPower,
			Signals:         bazi.ExtractYearSignalTypes(ys),
			Narrative:       bazi.RenderYearNarrative(ys),
			EvidenceSummary: bazi.RenderEvidenceSummary(ys),
		})
	}

	return &PastEventsYearsResponse{
		Years:     years,
		DayunMeta: dayunMeta,
		Generated: "algo-template",
	}, nil
}

// DayunSummaryStreamItem SSE 流式推送的单段大运 summary
type DayunSummaryStreamItem struct {
	DayunIndex int      `json:"dayun_index"`
	GanZhi     string   `json:"gan_zhi"`
	Themes     []string `json:"themes"`
	Summary    string   `json:"summary"`
	Cached     bool     `json:"cached"`
	Error      string   `json:"error,omitempty"`
}

func cachedDayunSummaryToStreamItem(cached *model.AIDayunSummary, fallbackGanZhi string) (DayunSummaryStreamItem, bool) {
	if cached == nil || strings.TrimSpace(cached.Summary) == "" {
		return DayunSummaryStreamItem{}, false
	}

	var themes []string
	if cached.Themes != nil && len(*cached.Themes) > 0 {
		if err := json.Unmarshal(*cached.Themes, &themes); err != nil {
			return DayunSummaryStreamItem{}, false
		}
	}

	gz := strings.TrimSpace(cached.DayunGanZhi)
	if gz == "" {
		gz = fallbackGanZhi
	}

	return DayunSummaryStreamItem{
		DayunIndex: cached.DayunIndex,
		GanZhi:     gz,
		Themes:     themes,
		Summary:    cached.Summary,
		Cached:     true,
	}, true
}

// GenerateDayunSummariesStream 按大运分段调 AI 生成 themes + summary，每段独立缓存与推送
func GenerateDayunSummariesStream(chartID string, userID *string, onItem func(item DayunSummaryStreamItem) error) error {
	chart, err := repository.GetChartByID(chartID)
	if err != nil || chart == nil {
		return fmt.Errorf("未找到排盘记录")
	}

	result, err := LoadOrCalculateResult(chart)
	if err != nil {
		return err
	}

	// 用神信息 / 身强弱明细 / 大运合化 标签缓存
	yongshenInfo := formatYongshenInfo(result)
	strengthLevel, strengthScore, strengthDtl := bazi.GetStrengthDetail(result)
	levelMap := map[string]string{"vstrong": "极强", "strong": "强", "neutral": "中和", "weak": "弱", "vweak": "极弱"}
	strengthDetail := fmt.Sprintf("%s(评分%d): %s", levelMap[strengthLevel], strengthScore, strengthDtl)

	natalSummary := fmt.Sprintf(
		"年柱%s%s 月柱%s%s 日柱%s%s 时柱%s%s",
		result.YearGan, result.YearZhi,
		result.MonthGan, result.MonthZhi,
		result.DayGan, result.DayZhi,
		result.HourGan, result.HourZhi,
	)
	genderLabel := "男"
	if chart.Gender == "female" {
		genderLabel = "女"
	}

	// 大运合化 map（gz → 描述）
	huaheMap := bazi.CollectDayunHuaheMap(result)

	// Prompt 模板（首次启动时已 seed，未 seed 时降级为内置）
	promptTpl := `你是一位资深八字命理师。请只为下列单段大运撰写整体总结，禁止逐年罗列。

命主：性别{{.Gender}} / 日干{{.DayGan}}
原局：{{.NatalSummary}}
{{if .YongshenInfo}}用忌神：{{.YongshenInfo}}{{end}}
{{if .StrengthDetail}}身强弱：{{.StrengthDetail}}{{end}}

当前大运：{{.DayunInfo}}
{{if .HuaheTag}}合化：{{.HuaheTag}}{{end}}

本段大运 10 年的算法信号摘要（JSON，每年含 type/evidence/polarity/source/year_in_dayun/dayun_phase/dayun_phase_level；dayun_phase=gan 表示前5年天干主事，zhi 表示后5年地支主事）：
{{.YearsData}}
{{if .LifeStageHint}}
人生阶段提示：{{.LifeStageHint}}{{end}}

输出要求：
1. themes：2-4 个主题词（如"事业↑""感情动荡""贵人扶持"；读书期可用"学业突破""同窗情谊""叛逆"）
2. summary：80-120 字，综合评述这 10 年整体走势、关键转折、注意事项；若前5年与后5年信号明显不同，要点出早段/后段气质差异
3. 严格输出以下 JSON，不要 Markdown 围栏：
{"themes":["主题1","主题2"],"summary":"..."}`

	tmpl, terr := template.New("dayun_summary").Parse(promptTpl)
	if terr != nil {
		return fmt.Errorf("dayun_summary prompt 解析失败: %v", terr)
	}

	for _, dy := range result.Dayun {
		gz := dy.Gan + dy.Zhi
		dayunPower := bazi.BuildDayunTenGodPower(result, dy)

		cached, cacheErr := repository.GetDayunSummary(chartID, dy.Index)
		if cacheErr == nil {
			if item, ok := cachedDayunSummaryToStreamItem(cached, gz); ok {
				_ = onItem(item)
				continue
			}
		} else {
			log.Printf("[GenerateDayunSummariesStream] 读取大运总结缓存失败 chart=%s dayun=%d: %v", chartID, dy.Index, cacheErr)
		}

		// 取该段大运的 10 年信号
		var dySignals []bazi.YearSignals
		for i, ln := range dy.LiuNian {
			if ln.Age < 1 {
				continue
			}
			lnRunes := []rune(ln.GanZhi)
			if len(lnRunes) < 2 {
				continue
			}
			ctx, _ := bazi.NewYearSignalContextForDayunIndex(i, dy.JinBuHuan)
			sigs := bazi.GetYearEventSignalsWithContext(result, string(lnRunes[0]), string(lnRunes[1]), gz, chart.Gender, ln.Age, ctx)
			tenGodPower := bazi.BuildYearTenGodPower(result, dy, ln, ctx, dayunPower)
			dySignals = append(dySignals, bazi.YearSignals{
				Year:            ln.Year,
				Age:             ln.Age,
				GanZhi:          ln.GanZhi,
				DayunGanZhi:     gz,
				YearInDayun:     ctx.YearInDayun,
				DayunPhase:      ctx.DayunPhase,
				DayunPhaseLevel: ctx.DayunPhaseLevel,
				TenGodPower:     tenGodPower,
				Signals:         sigs,
			})
		}
		dySigsJSON, _ := json.Marshal(dySignals)

		dayunInfo := fmt.Sprintf("%s %d-%d岁（%d-%d年）[%s/%s]",
			gz, dy.StartAge, dy.StartAge+9, dy.StartYear, dy.EndYear, dy.GanShiShen, dy.ZhiShiShen)
		if dayunPower.PlainTitle != "" {
			dayunInfo += "；主导力量：" + dayunPower.PlainTitle + "（" + strings.TrimSuffix(dayunPower.PlainText, "。") + "）"
		}

		// 计算 youngRatio：本段大运中 age<18 的年份占比（起运前年份不计）
		youngCount, totalCount := 0, 0
		for _, ln := range dy.LiuNian {
			if ln.Age < 1 {
				continue
			}
			totalCount++
			if ln.Age < bazi.YoungAgeCutoff {
				youngCount++
			}
		}
		lifeStageHint := buildLifeStageHint(youngCount, totalCount)

		tplData := model.DayunSummaryTemplateData{
			Gender:         genderLabel,
			DayGan:         result.DayGan,
			NatalSummary:   natalSummary,
			YongshenInfo:   yongshenInfo,
			StrengthDetail: strengthDetail,
			DayunInfo:      dayunInfo,
			HuaheTag:       huaheMap[gz],
			YearsData:      string(dySigsJSON),
			LifeStageHint:  lifeStageHint,
		}
		var pbuf bytes.Buffer
		if err := tmpl.Execute(&pbuf, tplData); err != nil {
			_ = onItem(DayunSummaryStreamItem{DayunIndex: dy.Index, GanZhi: gz, Error: "prompt 渲染失败"})
			continue
		}

		// 4. 调 AI（非推理模式）
		var collect strings.Builder
		_, modelName, providerID, durationMs, usage, aiErr := StreamAIWithSystemNoThink(pbuf.String(), func(chunk string) error {
			collect.WriteString(chunk)
			return nil
		}, nil)
		status, errMsg := "success", ""
		if aiErr != nil {
			status, errMsg = "error", aiErr.Error()
		}
		repository.CreateAIRequestLog(chartID, providerID, modelName, durationMs, status, errMsg)
		if aiErr == nil {
			promptStr := pbuf.String()
			outputStr := collect.String()
			go func(u TokenUsage, mn, pid string) {
				if logErr := repository.CreateTokenUsageLog(userID, &chartID, "dayun", mn, pid,
					u.PromptTokens, u.CompletionTokens, u.TotalTokens, u.ReasoningTokens, u.CacheHitTokens, u.CacheMissTokens,
					promptStr, outputStr); logErr != nil {
					log.Printf("[TokenUsage] 写入失败: %v", logErr)
				}
			}(usage, modelName, providerID)
		}
		if aiErr != nil {
			_ = onItem(DayunSummaryStreamItem{DayunIndex: dy.Index, GanZhi: gz, Error: aiErr.Error()})
			continue
		}

		// 5. 解析 JSON
		raw := strings.TrimSpace(collect.String())
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSuffix(raw, "```")
		raw = strings.TrimSpace(raw)
		if i := strings.Index(raw, "{"); i > 0 {
			raw = raw[i:]
		}
		if i := strings.LastIndex(raw, "}"); i >= 0 && i < len(raw)-1 {
			raw = raw[:i+1]
		}
		var parsed struct {
			Themes  []string `json:"themes"`
			Summary string   `json:"summary"`
		}
		if jerr := json.Unmarshal([]byte(raw), &parsed); jerr != nil {
			_ = onItem(DayunSummaryStreamItem{DayunIndex: dy.Index, GanZhi: gz, Error: "解析 AI JSON 失败"})
			continue
		}

		// 6. 写缓存
		themesJSON, _ := json.Marshal(parsed.Themes)
		themesRaw := json.RawMessage(themesJSON)
		_ = repository.UpsertDayunSummary(chartID, dy.Index, gz, &themesRaw, parsed.Summary, modelName)

		// 7. 推送
		_ = onItem(DayunSummaryStreamItem{
			DayunIndex: dy.Index,
			GanZhi:     gz,
			Themes:     parsed.Themes,
			Summary:    parsed.Summary,
			Cached:     false,
		})
	}
	return nil
}

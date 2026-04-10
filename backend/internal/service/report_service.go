package service

import (
	"bytes"
	"encoding/json"
	"fmt"
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
func buildBaziPrompt(r *bazi.BaziResult, celebs []model.CelebrityRecord) string {
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

	// ===名人参考库===
	celebStr := ""
	if len(celebs) > 0 {
		celebStr = "\n[名人参考库]\n"
		for _, c := range celebs {
			celebStr += fmt.Sprintf("- %s（%s）: 特征=[%s]\n", c.Name, c.Career, c.Traits)
		}
	}

	dayunYear := fmt.Sprintf("%d", currentYear)

	prompt := "以下命盘数据已由算法精确计算，请基于精算结果进行深度命理解读。\n\n" +
		fmt.Sprintf("[八字命盘]\n"+
			"年柱：%s%s（%s%s）| 藏干：%s\n"+
			"月柱：%s%s（%s%s）| 藏干：%s\n"+
			"日柱：%s%s（%s%s）| 藏干：%s <- 日干代表本人\n"+
			"时柱：%s%s（%s%s）| 藏干：%s\n\n"+
			"纳音：年柱[%s] 月柱[%s] 日柱[%s] 时柱[%s]\n\n"+
			"五行分布：木%d个(%.0f%%) 火%d个(%.0f%%) 土%d个(%.0f%%) 金%d个(%.0f%%) 水%d个(%.0f%%)\n\n"+
			"性别：%s\n\n",
			r.YearGan, r.YearZhi, r.YearGanWuxing, r.YearZhiWuxing, hideGanStr(r.YearHideGan),
			r.MonthGan, r.MonthZhi, r.MonthGanWuxing, r.MonthZhiWuxing, hideGanStr(r.MonthHideGan),
			r.DayGan, r.DayZhi, r.DayGanWuxing, r.DayZhiWuxing, hideGanStr(r.DayHideGan),
			r.HourGan, r.HourZhi, r.HourGanWuxing, r.HourZhiWuxing, hideGanStr(r.HourHideGan),
			r.YearNaYin, r.MonthNaYin, r.DayNaYin, r.HourNaYin,
			r.Wuxing.Mu, r.Wuxing.MuPct,
			r.Wuxing.Huo, r.Wuxing.HuoPct,
			r.Wuxing.Tu, r.Wuxing.TuPct,
			r.Wuxing.Jin, r.Wuxing.JinPct,
			r.Wuxing.Shui, r.Wuxing.ShuiPct,
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
		celebStr +
		"\n" +
		fmt.Sprintf("[第一步：三模块加权推断（在心中完成，不要在报告中输出计算过程）]\n"+
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
		) +
		"[第二步：生成命局分析总览 analysis.logic]\n" +
		"写一段整体分析（300-500字），专业命理分析风格，如实批断，不偏不倚：\n" +
		"- 首段直接说明命盘的核心喜用神与忌神。阐述依据时，需通过自然语言融合调候、格局与神煞的结论。绝对禁止出现“百分比”、“加权”、“权重”、“65%”等机械化算分词汇。\n" +
		"- 若模块存在冲突（如格局用神与调候用神不一致），请用专业口吻解释主次取舍（例如：“虽因格局喜水木，但命局偏寒，首当以火土暖局为要”）。\n" +
		"- 接续说明日干强弱依据（月令得令、失令状态）\n" +
		"- 提取1~2个最亮眼的特征或神煞进行个性点评\n" +
		"- 若命局存在明显缺陷（如用神缺失、忌神过重、格局遭破），必须如实指出，不可回避或美化。\n\n" +
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
		"## 【大运走势-精简版】\n\n## 【大运走势-专业版】\n\n" +
		"## 【命理分身-精简版】\n一句话提炼相似名人核心特质\n\n## 【命理分身-专业版】\n相似度剖析与寄语\n"

	return prompt
}

// GenerateAIReport 生成 AI 报告（带缓存）
func GenerateAIReport(chartID string, result *bazi.BaziResult) (*model.AIReport, error) {
	// 检查缓存
	cached, err := repository.GetReportByChartID(chartID)
	if err == nil && cached != nil {
		return cached, nil
	}

	// 构建 Prompt 并调用 AI
	celebs, _ := repository.ListCelebrities(true)
	prompt := buildBaziPrompt(result, celebs)
	rawContent, modelName, providerID, durationMs, aiErr := callAIWithSystem(prompt)

	// 记录调用日志（无论成功失败）
	status, errMsg := "success", ""
	if aiErr != nil {
		status, errMsg = "error", aiErr.Error()
	}
	repository.CreateAIRequestLog(chartID, providerID, modelName, durationMs, status, errMsg)

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

	var parsed structuredReport
	var contentStructured *json.RawMessage
	briefContent := ""

	// 尝试解析新结构化格式（第一次：原始内容）
	trimmedJSON := strings.TrimSpace(cleanJSON)
	if errParse := json.Unmarshal([]byte(trimmedJSON), &parsed); errParse != nil {
		// 第一次失败：用状态机修复字符串内的非法控制字符（如真实换行）
		fixedJSON := fixJSONStrings(trimmedJSON)
		if errFix := json.Unmarshal([]byte(fixedJSON), &parsed); errFix == nil {
			cleanJSON = fixedJSON // 用修复后的版本继续
		}
		// 无论修复是否成功，继续走下面的条件判断
	}
	if len(parsed.Chapters) > 0 && parsed.Analysis.Logic != "" {
		// 新格式解析成功：拼接 brief 作为兜底 content
		parts := []string{}
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
		briefContent = strings.Join(parts, "\n\n")

		// 写入结构化 JSON
		raw, _ := json.Marshal(parsed)
		rawMsg := json.RawMessage(raw)
		contentStructured = &rawMsg
	} else {
		// 降级：尝试旧格式（{yongshen, jishen, report}）
		type legacyResult struct {
			Yongshen string `json:"yongshen"`
			Jishen   string `json:"jishen"`
			Report   string `json:"report"`
		}
		var legacy legacyResult
		if errLegacy := json.Unmarshal([]byte(strings.TrimSpace(cleanJSON)), &legacy); errLegacy == nil && legacy.Report != "" {
			parsed.Yongshen = legacy.Yongshen
			parsed.Jishen = legacy.Jishen
			briefContent = legacy.Report
		} else {
			// 正则兜底提取
			importRegexp := regexp.MustCompile(`"yongshen"\s*:\s*"([^"]+)"`)
			matchY := importRegexp.FindStringSubmatch(cleanJSON)
			if len(matchY) > 1 {
				parsed.Yongshen = matchY[1]
			}
			jishenRegexp := regexp.MustCompile(`"jishen"\s*:\s*"([^"]+)"`)
			matchJ := jishenRegexp.FindStringSubmatch(cleanJSON)
			if len(matchJ) > 1 {
				parsed.Jishen = matchJ[1]
			}
			reportRegexp := regexp.MustCompile(`(?s)"report"\s*:\s*"(.*?)"\s*}`)
			matchR := reportRegexp.FindStringSubmatch(cleanJSON)
			if len(matchR) > 1 && strings.TrimSpace(matchR[1]) != "" {
				briefContent = strings.TrimSpace(matchR[1])
			}
		}

		// 全部失败：原始文本作为报告
		if briefContent == "" {
			fmt.Printf("[AI Report] 解析失败，使用原始内容。原始返回（前200字）：%s\n", rawContent[:min(200, len(rawContent))])
			briefContent = rawContent
		}
		// 降级路径：content_structured 置 nil
		contentStructured = nil
	}

	// 回写喜忌到 chart
	if chartID != "" && (parsed.Yongshen != "" || parsed.Jishen != "") {
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
func GenerateAIReportStream(chartID string, result *bazi.BaziResult, onData func(string) error) error {
	// 检查缓存
	cached, err := repository.GetReportByChartID(chartID)
	if err == nil && cached != nil {
		// 缓存命中：直接把全量内容模拟成一个数据块推下去
		_ = onData(cached.Content)
		return nil
	}

	// 构建 Prompt
	celebs, _ := repository.ListCelebrities(true)
	prompt := buildBaziPrompt(result, celebs)

	rawContent, modelName, providerID, durationMs, aiErr := StreamAIWithSystem(prompt, onData)
	
	status, errMsg := "success", ""
	if aiErr != nil {
		status, errMsg = "error", aiErr.Error()
	}
	repository.CreateAIRequestLog(chartID, providerID, modelName, durationMs, status, errMsg)

	if aiErr != nil {
		return aiErr
	}

	// 流结束后，异步或同步将 rawContent 处理成 JSON 持久化，供分享和旧接口使用
	parsed, briefContent := ParseMarkdownToStructured(rawContent)
	var contentStructured *json.RawMessage
	if parsed != nil && len(parsed.Chapters) > 0 {
		raw, _ := json.Marshal(parsed)
		rawMsg := json.RawMessage(raw)
		contentStructured = &rawMsg

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

	chapters := []string{"性格特质", "感情运势", "事业财运", "健康提示", "大运走势", "命理分身"}
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

// GenerateLiunianReport 生成流年运势分析
func GenerateLiunianReport(chartID string, targetYear int) (*model.AILiunianReport, error) {
	// 1. 检查缓存
	cached, err := repository.GetLiunianReport(chartID, targetYear)
	if err == nil && cached != nil {
		return cached, nil
	}

	// 2. 读取排盘
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

	result := bazi.Calculate(chart.BirthYear, chart.BirthMonth, chart.BirthDay, chart.BirthHour, chart.Gender, false, 0, chart.CalendarType, chart.IsLeapMonth)

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
	rawContent, modelName, providerID, durationMs, aiErr := callAIWithSystem(parsedPrompt.String())
	status, errMsg := "success", ""
	if aiErr != nil {
		status, errMsg = "error", aiErr.Error()
		repository.CreateAIRequestLog(chartID, providerID, modelName, durationMs, status, errMsg)
		return nil, aiErr
	}
	repository.CreateAIRequestLog(chartID, providerID, modelName, durationMs, status, errMsg)

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

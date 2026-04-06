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
		dayunStr += fmt.Sprintf("(%d) %d岁~%d岁（%d年起）：[%s%s] 干十神=%s 支十神=%s 长生=%s\n",
			i+1, dy.StartAge, dy.StartAge+9, dy.StartYear,
			dy.Gan, dy.Zhi,
			dy.GanShiShen, dy.ZhiShiShen, dy.DiShi,
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
		tiaohouStr = fmt.Sprintf(
			"\n[调候用神-穷通宝鉴]\n"+
				"日主[%s]生于[%s月]，调候理论指出：%s\n",
			r.DayGan, r.MonthZhi, r.Tiaohou.Text,
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
		fmt.Sprintf("[第一步：在心中完成以下推断，不要在报告中输出此步骤]\n"+
			"a. 月令格局判断（子平真诠）：\n"+
			"   月支=%s%s，主气十神=%s；\n"+
			"   判断月令主气十神是否透出天干（成格/破格）；\n"+
			"   明确格局名称（正官格/七杀格/食神格/伤官格/正财格/偏财格/正印格/偏印格/建禄格/羊刃格等）；\n"+
			"   确定格局顺用或逆用方向，得出格局用神。\n"+
			"b. 调候整合（穷通宝鉴）：\n"+
			"   结合[调候用神]区块数据，综合格局用神，确认最终喜用神与忌神。\n"+
			"   格局用神与调候用神一致则更有力；有出入时以格局用神为主，调候为辅。\n"+
			"c. 挑选1~2个最关键神煞，点出其对性格或运势的影响。\n\n",
			r.MonthGan, r.MonthZhi, firstShiShen(r.MonthZhiShiShen),
		) +
		"[第二步：生成命局分析总览 analysis.logic]\n" +
		"写一段整体分析（300-500字），现代叙事风格：\n" +
		"- 开门见山说格局定性（如：这是一个正印格命局...）\n" +
		"- 简述日主强弱依据（月令得令/失令）\n" +
		"- 说明最终用神/忌神及简要推导\n" +
		"- 点评1~2个关键神煞的个性影响\n" +
		"- 全程口语化、有温度，有专业感\n\n" +
		"[第三步：生成六章节报告]\n" +
		"每章两个版本：\n" +
		"- brief：约100字，通俗直接，结论先行\n" +
		"- detail：约350字，有命盘数据支撑，不堆砌术语，用「也就是说」等方式解释专业概念\n\n" +
		"章节与分析锚点：\n\n" +
		"【性格特质】\n" +
		"参考：日主五行天性特质、日支十二长生（内在潜质）、比劫/印绶十神力量（自我表达与驱动力）、关键神煞（天乙贵人/文昌等）。\n" +
		"分析：性格优势、内在驱动力、潜在短板。\n\n" +
		"【感情运势】\n" +
		"参考：男命看财星（正财/偏财）位置与旺衰；女命看官杀（正官/七杀）位置与力量；日支星运（感情宫）；桃花/红鸾/天喜等感情神煞。\n" +
		"分析：感情缘分特质、伴侣类型、婚姻稳定性与注意事项。\n\n" +
		"【事业财运】\n" +
		"参考：官杀（事业星）与食伤（才能星）天干透出情况；财星旺衰；天乙贵人/文昌/驿马等神煞。\n" +
		"分析：适合职业方向、事业发展节奏、财运特质。\n\n" +
		"【健康提示】\n" +
		"参考：五行最旺（需泄耗）与最弱（需补益）的脏腑对应（木肝胆/火心肠/土脾胃/金肺肠/水肾膀胱）；旬空地支；凶煞影响。\n" +
		"分析：体质倾向、易发健康问题、养生建议。\n\n" +
		"【大运走势】\n" +
		"参考：当前年份=" + dayunYear + "年，结合起运时间和大运序列，推算当前所处大运步次并明确说明；重点解读当前大运和下一步大运对事业感情的影响。\n" +
		"分析：人生节奏、当前运势处境、近期方向与建议。\n\n" +
		"【命理分身】\n" +
		"参考：日主五行特质+格局名称+关键神煞，与名人参考库匹配，选最相近的一位名人。\n" +
		"分析：相似之处剖析（侧重命理特质相通而非完全相同）、名人启示、一句有温度的结尾寄语。\n\n" +
		"[第四步：输出JSON（必须严格遵守，不输出任何其他内容）]\n" +
		"❗JSON格式规范：所有字符串字段中的换行必须用 \\n 表示，禁止使用真实换行符，否则JSON解析会失败。\n" +
		"{\n" +
		"  \"yongshen\": \"最终确认的喜用神五行（如：木火）\",\n" +
		"  \"jishen\": \"最终确认的忌神五行（如：金水）\",\n" +
		"  \"analysis\": {\n" +
		"    \"logic\": \"第二步命局分析总览全文（换行用\\n）\",\n" +
		"    \"summary\": \"一句话命局核心特质（30字以内）\"\n" +
		"  },\n" +
		"  \"chapters\": [\n" +
		"    {\"title\": \"性格特质\", \"brief\": \"...\", \"detail\": \"...\"},\n" +
		"    {\"title\": \"感情运势\", \"brief\": \"...\", \"detail\": \"...\"},\n" +
		"    {\"title\": \"事业财运\", \"brief\": \"...\", \"detail\": \"...\"},\n" +
		"    {\"title\": \"健康提示\", \"brief\": \"...\", \"detail\": \"...\"},\n" +
		"    {\"title\": \"大运走势\", \"brief\": \"...\", \"detail\": \"...\"},\n" +
		"    {\"title\": \"命理分身\", \"brief\": \"一句话提炼相似名人特质\", \"detail\": \"相似度剖析与寄语...\"}\n" +
		"  ]\n" +
		"}"

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

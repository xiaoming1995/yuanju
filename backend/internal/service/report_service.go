package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
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

// buildBaziPrompt 构建八字报告 Prompt（增强版：推理链条可见，精简/专业双输出，附加名人匹配）
func buildBaziPrompt(r *bazi.BaziResult, celebs []model.CelebrityRecord) string {
	// 辅助：将 []string 以顿号连接，空时返回"无"
	joinOrNone := func(ss []string) string {
		if len(ss) == 0 {
			return "无"
		}
		parts := make([]string, len(ss))
		copy(parts, ss)
		return strings.Join(parts, "、")
	}

	// 辅助：取地支主气十神（第一个元素）
	firstShiShen := func(ss []string) string {
		if len(ss) == 0 {
			return "—"
		}
		return ss[0]
	}

	// 辅助：藏干格式化
	hideGanStr := func(hg []string) string {
		if len(hg) == 0 {
			return "无"
		}
		return strings.TrimSpace(strings.Join(hg, " "))
	}

	// 性别文本
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
		dayunStr += fmt.Sprintf("（%d）%d岁～%d岁（%d年起）：[%s%s] 干十神=%s 支十神=%s 长生=%s\n",
			i+1, dy.StartAge, dy.StartAge+9, dy.StartYear,
			dy.Gan, dy.Zhi,
			dy.GanShiShen, dy.ZhiShiShen, dy.DiShi,
		)
	}
	if dayunStr == "" {
		dayunStr = "（暂无大运数据）\n"
	}

	// ===引擎初步推算===（可选，非空时才注入）
	yongshenHint := ""
	if r.Yongshen != "" {
		yongshenHint = fmt.Sprintf(
			"\n===引擎初步推算===\n引擎基于五行统计初步判断：用神=[%s]，忌神=[%s]\n（此为算法参考值，请结合十神全局格局确认或微调）\n",
			r.Yongshen, r.Jishen,
		)
	}

	// ===名人参考库===
	celebStr := ""
	if len(celebs) > 0 {
		celebStr = "\n===名人参考库===\n"
		for _, c := range celebs {
			celebStr += fmt.Sprintf("- %s（%s）: 特征=[%s]\n", c.Name, c.Career, c.Traits)
		}
	}

	prompt := fmt.Sprintf(
		`你是一位精通八字命理的专业命理师。以下命盘数据已由算法精确计算，请基于这些精算结果进行深度命理解读。

===八字命盘===
年柱：%s%s（%s%s）| 藏干：%s
月柱：%s%s（%s%s）| 藏干：%s
日柱：%s%s（%s%s）| 藏干：%s ← 日干代表本人
时柱：%s%s（%s%s）| 藏干：%s

纳音：年柱「%s」月柱「%s」日柱「%s」时柱「%s」

五行分布：
木：%d个（%.0f%%）火：%d个（%.0f%%）土：%d个（%.0f%%）金：%d个（%.0f%%）水：%d个（%.0f%%）

性别：%s

===十神关系（算法精算）===
天干十神：年干%s=[%s] | 月干%s=[%s] | 日干%s=[日主] | 时干%s=[%s]
地支主气十神：年支%s=[%s] | 月支%s=[%s] | 日支%s=[%s] | 时支%s=[%s]

===十二长生（日主对各柱星运）===
年柱：[%s] | 月柱：[%s] | 日柱：[%s] | 时柱：[%s]

===旬空（空亡）===
年柱：[%s] | 月柱：[%s] | 日柱：[%s] | 时柱：[%s]

===神煞===
年柱：%s | 月柱：%s | 日柱：%s | 时柱：%s

===大运序列===
%s%s%s
===第一步：综合精算数据整合判断（请在心中完成，不要在报告中输出此步骤）===
基于以上算法精算数据，在心中完成以下专业整合：
1. 月令考察：月支=[%s%s]，主气十神=[%s]，结合日主星运=[%s]，综合评估日主得令/失令、得地/失地状况；
2. 用神确认：参考引擎初步推算，结合天干十神全局格局，确认或微调最终用神/忌神；
3. 神煞特质：归纳此命局中最关键的1~3个神煞（如有），点出其对性格和运势的影响。

===第二步：生成命局分析总览（analysis.logic）===
写一段整体推理总览（300-600字），展示你的完整分析思路：
- 先写专业术语版推导（如：「日主甲木，月令戌土为财星，月令失令...」）
- 紧跟一句通俗白话解释（如：「简单说就是...」）
- 涵盖：日主强弱判断 + 用神忌神推理依据 + 格局定性 + 关键神煞点评

===第三步：生成六章节报告===
请按以下六个章节撰写报告，每章需提供两个版本：
- brief（约100字精简摘要，通俗易懂）
- detail（约350字详细分析，每个结论先写术语依据，再跟一句白话解释）

章节：
【性格特质】分析性格特点、内在天赋和潜在挑战，结合日主五行、月令十神特性展开，可与神煞特质呼应。
【感情运势】分析感情缘分、婚姻状况、伴侣特质，以及感情方面的注意事项。
【事业财运】分析适合的职业方向、事业发展路径、财运状况和投资建议。
【健康提示】根据五行强弱和藏干分析需要注意的健康领域，提供养生建议。
【大运走势】结合起运年龄和各步大运干支十神，解读人生各阶段的整体运势节奏；重点分析当前大运和近1~2步大运。
【命理分身】分析命理相似名人。若上下文中提供了“名人参考库”，请挑选一位五行或日主特征最相似的名人，分析相似之处，并给出寄语。若未匹配到合适名人请自行推演一位，但优先使用参考库。

===第四步：输出格式（非常重要）===
你必须且只能以合法的 JSON 格式输出最终结果，不要输出任何额外的解释或说话头。结构必须严格如下：
{
  "yongshen": "在此填入你推断的用神（喜用神）五行汉字，如：木火",
  "jishen": "在此填入你推断的忌神五行汉字，如：金水",
  "analysis": {
    "logic": "在此填入第二步的命局分析总览全文（内容中的换行请使用 \n 转义符）",
    "summary": "在此填入一句话命局核心特质（30字以内）"
  },
  "chapters": [
    {
      "title": "性格特质",
      "brief": "在此填入性格特质精简摘要（约100字）",
      "detail": "在此填入性格特质详细分析（约350字，含术语依据+白话解释，换行用 \n）"
    },
    {
      "title": "感情运势",
      "brief": "...",
      "detail": "..."
    },
    {
      "title": "事业财运",
      "brief": "...",
      "detail": "..."
    },
    {
      "title": "健康提示",
      "brief": "...",
      "detail": "..."
    },
    {
      "title": "大运走势",
      "brief": "...",
      "detail": "..."
    },
    {
      "title": "命理分身",
      "brief": "一句话提炼相似名人之特质",
      "detail": "详细的相似度剖析与名人寄语..."
    }
  ]
}`,
		// 八字命盘
		r.YearGan, r.YearZhi, r.YearGanWuxing, r.YearZhiWuxing, hideGanStr(r.YearHideGan),
		r.MonthGan, r.MonthZhi, r.MonthGanWuxing, r.MonthZhiWuxing, hideGanStr(r.MonthHideGan),
		r.DayGan, r.DayZhi, r.DayGanWuxing, r.DayZhiWuxing, hideGanStr(r.DayHideGan),
		r.HourGan, r.HourZhi, r.HourGanWuxing, r.HourZhiWuxing, hideGanStr(r.HourHideGan),
		// 纳音
		r.YearNaYin, r.MonthNaYin, r.DayNaYin, r.HourNaYin,
		// 五行分布
		r.Wuxing.Mu, r.Wuxing.MuPct,
		r.Wuxing.Huo, r.Wuxing.HuoPct,
		r.Wuxing.Tu, r.Wuxing.TuPct,
		r.Wuxing.Jin, r.Wuxing.JinPct,
		r.Wuxing.Shui, r.Wuxing.ShuiPct,
		// 性别
		genderText,
		// 天干十神
		r.YearGan, r.YearGanShiShen,
		r.MonthGan, r.MonthGanShiShen,
		r.DayGan,
		r.HourGan, r.HourGanShiShen,
		// 地支主气十神
		r.YearZhi, firstShiShen(r.YearZhiShiShen),
		r.MonthZhi, firstShiShen(r.MonthZhiShiShen),
		r.DayZhi, firstShiShen(r.DayZhiShiShen),
		r.HourZhi, firstShiShen(r.HourZhiShiShen),
		// 十二长生
		r.YearDiShi, r.MonthDiShi, r.DayDiShi, r.HourDiShi,
		// 旬空
		r.YearXunKong, r.MonthXunKong, r.DayXunKong, r.HourXunKong,
		// 神煞
		joinOrNone(r.YearShenSha), joinOrNone(r.MonthShenSha),
		joinOrNone(r.DayShenSha), joinOrNone(r.HourShenSha),
		// 大运序列
		dayunStr,
		// 引擎初步推算（可能为空）
		yongshenHint,
		// 名人库（可能为空）
		celebStr,
		// 第一步月令参数
		r.MonthGan, r.MonthZhi, firstShiShen(r.MonthZhiShiShen), r.MonthDiShi,
	)

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
	rawContent, modelName, providerID, durationMs, aiErr := callAI(prompt)

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

	// 尝试解析新结构化格式
	if errParse := json.Unmarshal([]byte(strings.TrimSpace(cleanJSON)), &parsed); errParse == nil &&
		len(parsed.Chapters) > 0 && parsed.Analysis.Logic != "" {
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

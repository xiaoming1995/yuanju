package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/pkg/bazi"
)

const (
	polishMinChars = 20
	polishMaxChars = 300
)

// validatePolishSituation 校验用户输入的当前情况字符串。
// 长度按 unicode rune 计数（中文也算 1 字符），范围 [20, 300]。
func validatePolishSituation(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return fmt.Errorf("当前情况描述不能为空")
	}
	n := utf8.RuneCountInString(trimmed)
	if n < polishMinChars {
		return fmt.Errorf("当前情况描述至少 %d 字（当前 %d 字）", polishMinChars, n)
	}
	if n > polishMaxChars {
		return fmt.Errorf("当前情况描述最多 %d 字（当前 %d 字）", polishMaxChars, n)
	}
	return nil
}

// buildPolishPrompt 构建润色 prompt：
// 原版 5 章 markdown + 命理数据 + 用户输入 → 让 LLM 逐章重写。
func buildPolishPrompt(originalReport *model.AIReport, result *bazi.BaziResult, userSituation string) string {
	originalContent := ""
	if originalReport != nil {
		originalContent = originalReport.Content
	}

	yongshen := result.Yongshen
	jishen := result.Jishen
	mingGe := result.MingGe
	mingGeDesc := result.MingGeDesc

	genderText := "女"
	if result.Gender == "male" {
		genderText = "男"
	}

	return strings.Join([]string{
		"你是一位资深命理师傅。下面给你一份已经生成好的命理解读「原版」、一份命理算法的精算数据、以及用户当下的具体处境。",
		"你的任务：基于原版结论，结合用户处境，重写一份「润色版」 —— 内容贴近用户当下，口吻像师傅当面讲给他听，避免报告体。",
		"",
		"[原版命理解读 · 严格保留命局结论]",
		originalContent,
		"",
		"[命理算法精算数据]",
		fmt.Sprintf("出生：%d年%d月%d日%d时 · %s命",
			result.BirthYear, result.BirthMonth, result.BirthDay, result.BirthHour, genderText),
		fmt.Sprintf("用神：%s ／ 忌神：%s", strDefault(yongshen, "—"), strDefault(jishen, "—")),
		fmt.Sprintf("命格：%s（%s）", strDefault(mingGe, "—"), strDefault(mingGeDesc, "—")),
		"",
		"[用户当下情况 · 用户自述]",
		strings.TrimSpace(userSituation),
		"",
		"[改写要求]",
		"1. 保持原版的 5 章结构（性格特质 / 感情运势 / 事业财运 / 健康提示 / 大运走势），章节顺序不变。",
		"2. 每章 200-300 字。第二人称「你」叙述，师傅口吻，温润不冷漠。",
		"3. 每章首段必须引出用户情况里相关的事，再带回命理依据说明为什么。",
		"4. 不可改变原版的命局结论：用神 / 忌神 / 命格 / 十神等核心判断与原版一致。",
		"5. 严禁出现 Markdown 加粗 `**`、斜体 `*`、子标题 `###` 等格式符号。",
		"6. 严禁出现「百分比 / 加权 / 权重 / 由此可见 / 综上所述 / 通过分析可得 / 显然 / 总体而言」等公文体词汇。",
		"7. 五行强弱用「旺/相/休/囚/死」「过旺/偏旺/平衡/偏弱/缺」等命理术语，不用数字。",
		"8. 段落自然换行（章内多段之间空一行），不用列表 / bullet。",
		"",
		"[输出格式 · 严格遵守]",
		"必须按以下顺序、一字不差使用 Markdown 二级标题；每章标题独占一行；除标题外不要其它格式标记：",
		"",
		"## 【性格特质】",
		"...（200-300 字润色版正文，至少 2 段，段间空行）...",
		"",
		"## 【感情运势】",
		"...",
		"",
		"## 【事业财运】",
		"...",
		"",
		"## 【健康提示】",
		"...",
		"",
		"## 【大运走势】",
		"...",
	}, "\n")
}

// strDefault 返回 s 或 fallback（s 为空时）。
func strDefault(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

// PolishReport 执行润色：load 原版 → 构建 prompt → 调 LLM → 解析 → upsert
//
// 不写 user_id 参数 — token 记账只在原版调用时执行；润色版仅记 token 字段到本表。
func PolishReport(chartID string, result *bazi.BaziResult, userSituation string) (*model.PolishedReport, error) {
	// 1. 校验输入
	if err := validatePolishSituation(userSituation); err != nil {
		return nil, err
	}

	// 2. load 原版（必须存在）
	original, err := repository.GetReportByChartID(chartID)
	if err != nil {
		return nil, fmt.Errorf("读取原版报告失败: %w", err)
	}
	if original == nil {
		return nil, fmt.Errorf("请先生成原版命理解读，再尝试润色")
	}

	// 3. 构建 prompt
	prompt := buildPolishPrompt(original, result, strings.TrimSpace(userSituation))

	// 4. 调 LLM（非流式，复用原版 ai_client）
	log.Printf("[Polish] 开始润色 chart_id=%s situation_len=%d", chartID, utf8.RuneCountInString(userSituation))
	rawContent, modelName, _, durationMs, usage, aiErr := callAIWithSystem(prompt)
	if aiErr != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", aiErr)
	}
	log.Printf("[Polish] LLM 返回 chart_id=%s duration_ms=%d tokens=%d/%d/%d",
		chartID, durationMs, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens)

	// 5. 解析 Markdown → 5 章 JSON
	cleanContent := strings.TrimSpace(rawContent)
	parsed, _ := ParseMarkdownToStructured(cleanContent)
	var contentStructured *json.RawMessage
	if parsed != nil && len(parsed.Chapters) > 0 {
		raw, _ := json.Marshal(parsed)
		rawMsg := json.RawMessage(raw)
		contentStructured = &rawMsg
	}

	// 6. UPSERT 到 ai_polished_reports
	report, err := repository.UpsertPolishedReport(
		chartID, strings.TrimSpace(userSituation), cleanContent, modelName,
		contentStructured,
		usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens,
	)
	if err != nil {
		return nil, fmt.Errorf("保存润色版失败: %w", err)
	}

	return report, nil
}

package service

import (
	"fmt"
	"strings"

	"yuanju/pkg/bazi"
)

// validatedKeywords 是 narrative 中出现就必须能在算法 evidence 里追溯到的字符串。
// 选取标准：
//
//	(1) 命理学有具体结构含义的术语（用神位/忌神位/喜神位）
//	(2) 命理学有特定干支事件含义的术语（伏吟/反吟/大运合化/三会/三合/受冲/受刑/双重命中/力度倍增）
//	(3) 神煞名（每一个神煞临运 都对应 event_signals.go 的具体生成路径）
//
// 不在此列的术语（如十神名「食神/伤官」、单纯干支「甲乙丙丁」、宫位「日柱/月柱」）
// AI 可以自由发挥而不被卡，因为它们可从 BaziResult 直接推导。
var validatedKeywords = []string{
	// 位
	"用神位", "忌神位", "喜神位",
	// 强变动
	"伏吟", "反吟", "大运合化", "三会", "三合",
	// 硬事件标记
	"受冲", "受刑", "双重命中", "力度倍增",
	// 神煞（与 event_signals.go::shenshaTable 对齐）
	"驿马", "桃花", "华盖", "白虎", "丧门", "吊客", "灾煞", "流霞",
	"天医", "天喜", "天乙", "天德", "月德", "文昌", "太极", "福星",
	"红艳", "孤辰", "寡宿", "羊刃", "亡神", "劫煞", "披麻", "咸池",
	"勾绞", "国印",
}

// ExtractEvidenceKeywords 在单年 signals 集合上扫描 validatedKeywords，
// 返回该年 evidence 已识别到的命理术语清单（按 validatedKeywords 顺序去重）。
//
// 用途：把"白名单"以 allowed_keywords 字段注入 prompt，让 AI 不必从 evidence
// 长文里猜测哪个词是"算法已认证"的，直接对照清单写 narrative。
// 缓解 AI 凭训练语料先验编造神煞/伏吟等关键词的问题。
func ExtractEvidenceKeywords(signals []bazi.EventSignal) []string {
	if len(signals) == 0 {
		return nil
	}
	var sb strings.Builder
	for _, s := range signals {
		sb.WriteString(s.Evidence)
		sb.WriteString("\n")
	}
	text := sb.String()
	out := make([]string, 0, 4)
	for _, kw := range validatedKeywords {
		if strings.Contains(text, kw) {
			out = append(out, kw)
		}
	}
	return out
}

// ValidateYearNarrative 校验单年 narrative 引用的命理术语是否能在该年算法
// signals 的 Evidence 中追溯到。
//
// 返回 (true, "") 表示通过；返回 (false, reason) 表示某个关键词无源可溯，
// reason 包含具体哪个词没匹配上。
//
// 设计意图：拦截 AI 自信地说错（编造神煞/编造用神位事件）的最常见路径。
// 不做语义级校验（"宜防健康" 是否合理）—— 那是 AI 的判断权限范围。
//
// 空 narrative（"" — AI 决定不写）总是通过。
func ValidateYearNarrative(narrative string, signals []bazi.EventSignal) (bool, string) {
	if narrative == "" {
		return true, ""
	}
	// 拼一份 evidence 全文，逐关键词查一次 substring 包含。
	var evidenceBuf strings.Builder
	for _, s := range signals {
		evidenceBuf.WriteString(s.Evidence)
		evidenceBuf.WriteString("\n")
	}
	allEvidence := evidenceBuf.String()
	for _, kw := range validatedKeywords {
		if strings.Contains(narrative, kw) && !strings.Contains(allEvidence, kw) {
			return false, fmt.Sprintf("narrative 出现 %q 但算法 evidence 无对应来源", kw)
		}
	}
	return true, ""
}

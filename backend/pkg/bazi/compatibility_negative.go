package bazi

import "fmt"

// detectNegativeSignals 同位比较两人四柱（年-年/月-月/日-日/时-时），
// 检出地支冲/刑/害与天干相克，产出 polarity="negative" 的 evidence。
// 与正向评分口径一致（同名柱比较），不影响任何分数。日柱权重最高。
func detectNegativeSignals(a, b *BaziResult) []CompatibilityEvidence {
	if a == nil || b == nil {
		return nil
	}
	pillars := []negPillar{
		{name: "day", label: "日柱", dimension: "day_pillar", weight: 10,
			ganA: a.DayGan, zhiA: a.DayZhi, ganB: b.DayGan, zhiB: b.DayZhi},
		{name: "month", label: "月柱", dimension: "eight_chars", weight: 6,
			ganA: a.MonthGan, zhiA: a.MonthZhi, ganB: b.MonthGan, zhiB: b.MonthZhi},
		{name: "year", label: "年柱", dimension: "zodiac", weight: 5,
			ganA: a.YearGan, zhiA: a.YearZhi, ganB: b.YearGan, zhiB: b.YearZhi},
		{name: "hour", label: "时柱", dimension: "eight_chars", weight: 4,
			ganA: a.HourGan, zhiA: a.HourZhi, ganB: b.HourGan, zhiB: b.HourZhi},
	}
	out := make([]CompatibilityEvidence, 0, 4)
	for _, p := range pillars {
		out = append(out, pillarNegatives(p)...)
	}
	return out
}

// negPillar 描述一柱的同位比较输入与落位元数据。
type negPillar struct {
	name      string // year/month/day/hour（用于 evidence_key）
	label     string // 年柱/月柱/日柱/时柱（用于文案）
	dimension string // 落入报告分节：year→zodiac，month/hour→eight_chars，day→day_pillar
	weight    int
	ganA, zhiA, ganB, zhiB string
}

func pillarNegatives(p negPillar) []CompatibilityEvidence {
	var out []CompatibilityEvidence
	if branchChong(p.zhiA, p.zhiB) {
		out = append(out, negEvidence(p, "chong", "地支相冲", fmt.Sprintf(
			"%s地支 %s 与 %s 相冲——这是两股直接对撞的力量，落到关系里就是这块容易顶牛、各执一端，需要主动让一步才不至于僵住。",
			p.label, p.zhiA, p.zhiB)))
	}
	if branchXing(p.zhiA, p.zhiB) {
		out = append(out, negEvidence(p, "xing", "地支相刑", fmt.Sprintf(
			"%s地支 %s 与 %s 相刑——刑主纠缠、暗耗，容易反复在同一件事上磨人，要警惕翻旧账式的内耗。",
			p.label, p.zhiA, p.zhiB)))
	}
	if branchHai(p.zhiA, p.zhiB) {
		out = append(out, negEvidence(p, "hai", "地支相害", fmt.Sprintf(
			"%s地支 %s 与 %s 相害（穿）——害主暗里别扭、好心办坏事，容易因误解积小怨，要把话说开别憋着。",
			p.label, p.zhiA, p.zhiB)))
	}
	if ke, actor, target := ganKe(p.ganA, p.ganB); ke {
		out = append(out, negEvidence(p, "gan_ke", "天干相克", fmt.Sprintf(
			"%s天干 %s 克 %s（%s克%s）——一方在气势上压住另一方，相处容易一强一弱、一个主导一个迁就，时间长了被压的一方会憋屈。",
			p.label, actor, target, wxPinyin2CN[ganWuxing[actor]], wxPinyin2CN[ganWuxing[target]])))
	}
	return out
}

func negEvidence(p negPillar, kind, typeLabel, detail string) CompatibilityEvidence {
	return CompatibilityEvidence{
		EvidenceKey: "neg_" + p.name + "_" + kind,
		Dimension:   p.dimension,
		Type:        p.label + typeLabel,
		Polarity:    "negative",
		Source:      p.dimension,
		Title:       p.label + typeLabel,
		Detail:      detail,
		Weight:      p.weight,
	}
}

// branchChong 判定两地支是否相冲（六冲）。
func branchChong(x, y string) bool {
	if x == "" || y == "" {
		return false
	}
	return sixChong[x] == y
}

// branchXing 判定两地支是否相刑：异支查 sixXing 双向；同支查自刑。
func branchXing(x, y string) bool {
	if x == "" || y == "" {
		return false
	}
	if x == y {
		return selfXing[x]
	}
	return sixXing[x] == y || sixXing[y] == x
}

// branchHai 判定两地支是否相害（穿）。
func branchHai(x, y string) bool {
	if x == "" || y == "" {
		return false
	}
	return sixHai[x] == y
}

// ganKe 判定两天干是否五行相克，返回 (是否相克, 克方, 被克方)。
// 涵盖天干相冲（庚冲甲等本质即金克木）。
func ganKe(x, y string) (bool, string, string) {
	if x == "" || y == "" {
		return false, "", ""
	}
	wx, wy := ganWuxing[x], ganWuxing[y]
	if wx == "" || wy == "" {
		return false, "", ""
	}
	if wxKe[wx] == wy {
		return true, x, y
	}
	if wxKe[wy] == wx {
		return true, y, x
	}
	return false, "", ""
}

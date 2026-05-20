package bazi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStripEvidenceParenthetical_FullWidth(t *testing.T) {
	in := "用神受冲（月柱宫位，权重次之）"
	want := "用神受冲"
	if got := stripEvidenceParenthetical(in); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestStripEvidenceParenthetical_HalfWidth(t *testing.T) {
	in := "white tiger临命 (本年有重煞，此信号仅作参考)"
	want := "white tiger临命 "
	if got := stripEvidenceParenthetical(in); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestStripEvidenceParenthetical_MultipleParens(t *testing.T) {
	in := "流年甲冲原局戊月（用神位），用神受冲，应期凶（月柱宫位，权重次之）"
	got := stripEvidenceParenthetical(in)
	if strings.Contains(got, "（") || strings.Contains(got, "）") {
		t.Errorf("expected no parens left, got %q", got)
	}
}

func TestStripEvidenceParenthetical_NoParens(t *testing.T) {
	in := "用神受冲，应期凶"
	if got := stripEvidenceParenthetical(in); got != in {
		t.Errorf("expected unchanged, got %q", got)
	}
}

func TestStripEvidenceParenthetical_EmptyString(t *testing.T) {
	if got := stripEvidenceParenthetical(""); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestCompressYearSignalsForPrompt_DropsRedundantFields(t *testing.T) {
	years := []YearSignals{
		{
			Year:            2020,
			Age:             25,
			GanZhi:          "庚子",
			DayunGanZhi:     "壬辰",
			YearInDayun:     5,
			DayunPhase:      "gan",
			DayunPhaseLevel: "凶",
			TenGodPower: TenGodPowerProfile{
				Dominant:   "正官",
				Group:      "official",
				GroupLabel: "官杀",
				Strength:   "strong",
				Polarity:   "pressure",
				PlainTitle: "责任明显",
				PlainText:  "事业、规则、地位变动会更明显。",
				Score:      6,
				Reason:     "流年天干为正官",
			},
			Signals: []EventSignal{
				{Type: "事业", Evidence: "庚透干为正官（事业之星）", Polarity: "吉", Source: "柱位互动"},
			},
		},
	}
	compressed, err := CompressYearSignalsForPrompt(years)
	if err != nil {
		t.Fatalf("compress error: %v", err)
	}
	out := string(compressed)

	// 不应出现的字段
	mustNotContain := []string{"year_in_dayun", "dayun_phase", "plain_title", "plain_text", "score", `"group"`}
	for _, kw := range mustNotContain {
		if strings.Contains(out, kw) {
			t.Errorf("compressed should NOT contain %q; got: %s", kw, out)
		}
	}

	// 应保留的字段
	mustContain := []string{"year", "age", "gan_zhi", "dayun_gan_zhi", "ten_god_power", "dominant", "group_label", "strength", "polarity", "signals", "type", "evidence"}
	for _, kw := range mustContain {
		if !strings.Contains(out, kw) {
			t.Errorf("compressed should contain %q; got: %s", kw, out)
		}
	}
}

func TestCompressYearSignalsForPrompt_StripsEvidenceParens(t *testing.T) {
	years := []YearSignals{
		{
			Year:    2020,
			Age:     25,
			GanZhi:  "庚子",
			Signals: []EventSignal{
				{Type: "综合变动", Evidence: "庚冲原局甲月（用神位），用神受冲（月柱宫位，权重次之）", Polarity: "凶"},
			},
		},
	}
	compressed, _ := CompressYearSignalsForPrompt(years)
	out := string(compressed)
	if strings.Contains(out, "（月柱宫位") || strings.Contains(out, "权重次之") {
		t.Errorf("parenthetical should be stripped; got: %s", out)
	}
	// 主体内容保留
	if !strings.Contains(out, "用神受冲") {
		t.Errorf("main evidence should survive; got: %s", out)
	}
}

func TestCompressYearSignalsForPrompt_RoughSizeSaving(t *testing.T) {
	years := []YearSignals{
		{
			Year:            2020,
			Age:             25,
			GanZhi:          "庚子",
			DayunGanZhi:     "壬辰",
			YearInDayun:     5,
			DayunPhase:      "gan",
			DayunPhaseLevel: "凶",
			TenGodPower: TenGodPowerProfile{
				Dominant:   "正官",
				Group:      "official",
				GroupLabel: "官杀",
				Strength:   "strong",
				Polarity:   "pressure",
				PlainTitle: "责任明显",
				PlainText:  "事业、规则、地位变动会更明显，需稳步推进。",
				Score:      6,
				Reason:     "流年天干为正官",
			},
			Signals: []EventSignal{
				{Type: "事业", Evidence: "庚透干为正官（事业之星，社会地位提升）", Polarity: "吉"},
				{Type: "健康", Evidence: "羊刃临运，宜防开刀血光（本年有重煞，此信号仅作参考）", Polarity: "凶"},
				{Type: "综合变动", Evidence: "流年庚冲原局甲月（用神位，权重次之）", Polarity: "凶"},
			},
		},
	}
	rawSize := func() int {
		raw, _ := json.Marshal(years)
		return len(raw)
	}()
	compressed, _ := CompressYearSignalsForPrompt(years)
	compressedSize := len(compressed)

	// 压缩后字节数应明显小于原 JSON（目标 ~20%+ 节省）
	saving := float64(rawSize-compressedSize) / float64(rawSize)
	if saving < 0.15 {
		t.Errorf("expected at least 15%% size reduction, got %.1f%% (raw=%d compressed=%d)", saving*100, rawSize, compressedSize)
	}
	t.Logf("compression saving: %.1f%% (raw=%d → compressed=%d)", saving*100, rawSize, compressedSize)
}

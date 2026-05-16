package service

import (
	"strings"
	"testing"

	"yuanju/internal/model"
	"yuanju/pkg/bazi"
)

func TestValidatePolishSituation(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"空字符串", "", true},
		{"全空白", "   \n  ", true},
		{"中文 19 字", strings.Repeat("一", 19), true},
		{"中文 20 字", strings.Repeat("一", 20), false},
		{"中文 300 字", strings.Repeat("一", 300), false},
		{"中文 301 字", strings.Repeat("一", 301), true},
		{"混合 25 字", "今年考虑跳槽，目前 work 太烧脑，想换条路走走看", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validatePolishSituation(c.input)
			if (err != nil) != c.wantErr {
				t.Fatalf("want err=%v, got %v", c.wantErr, err)
			}
		})
	}
}

func TestBuildPolishPrompt_IncludesAllParts(t *testing.T) {
	original := &model.AIReport{
		Content: "## 【性格特质】\n原版性格章...\n## 【感情运势】\n原版感情章...",
	}
	result := &bazi.BaziResult{
		BirthYear: 1995, BirthMonth: 5, BirthDay: 12, BirthHour: 10,
		Gender: "male", Yongshen: "火、土", Jishen: "金", MingGe: "食神生财格", MingGeDesc: "月令偏财得用",
	}
	prompt := buildPolishPrompt(original, result, "  今年考虑跳槽，跟对象有点摩擦，想看看大方向  ")

	mustContain := []string{
		"原版命理解读",
		"原版性格章",
		"火、土",
		"金",
		"食神生财格",
		"今年考虑跳槽",
		"## 【性格特质-专业版】",
		"## 【大运走势-专业版】",
		"师傅口吻",
		"不可改变原版的命局结论",
	}
	for _, s := range mustContain {
		if !strings.Contains(prompt, s) {
			t.Errorf("prompt 缺少关键字: %q", s)
		}
	}
}

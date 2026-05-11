package service

import "testing"

func Test_matchModelTier(t *testing.T) {
	cases := []struct {
		model string
		want  string
	}{
		{"deepseek-v4-pro", "pro"},
		{"deepseek-v4-pro-20250101", "pro"},
		{"deepseek-reasoner", "pro"},
		{"deepseek-v4-flash", "flash"},
		{"deepseek-chat", "flash"},
		{"DEEPSEEK-V4-FLASH", "flash"},
		{"gpt-4o", "default"},
		{"qwen3-32b", "default"},
		{"", "default"},
	}
	for _, c := range cases {
		got := matchModelTier(c.model)
		if got != c.want {
			t.Errorf("matchModelTier(%q) = %q, want %q", c.model, got, c.want)
		}
	}
}

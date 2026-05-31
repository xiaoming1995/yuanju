package prompt

import "testing"

func TestDriftStatus(t *testing.T) {
	def := MustGet("compatibility") // 已注册模块，取当前出厂版

	cases := []struct {
		name          string
		module        string
		content       string
		canonicalHash string
		want          string
	}{
		{
			name:          "aligned: 内容与分支点都等于当前出厂版",
			module:        "compatibility",
			content:       def.Content,
			canonicalHash: def.Hash,
			want:          DriftAligned,
		},
		{
			name:          "customized: 基于当前出厂版但内容被改",
			module:        "compatibility",
			content:       "管理员改过的文案",
			canonicalHash: def.Hash,
			want:          DriftCustomized,
		},
		{
			name:          "outdated: 分支点落后于当前出厂版",
			module:        "compatibility",
			content:       "随便什么旧内容",
			canonicalHash: "old-hash-from-v2",
			want:          DriftOutdated,
		},
		{
			name:          "outdated: 分支点为空(历史遗留行)也算落后",
			module:        "compatibility",
			content:       def.Content,
			canonicalHash: "",
			want:          DriftOutdated,
		},
		{
			name:          "unregistered: 代码无此 canonical",
			module:        "no_such_module",
			content:       "x",
			canonicalHash: "y",
			want:          DriftUnregistered,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DriftStatus(tc.module, tc.content, tc.canonicalHash)
			if got != tc.want {
				t.Errorf("DriftStatus(%q,...) = %q, want %q", tc.module, got, tc.want)
			}
		})
	}
}

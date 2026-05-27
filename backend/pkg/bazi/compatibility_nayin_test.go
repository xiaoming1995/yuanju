package bazi

import "testing"

func TestNayinElement_AllSixtyGanzhi(t *testing.T) {
	cases := []struct {
		ganzhi string
		want   string
	}{
		{"甲子", "金"}, {"乙丑", "金"}, // 海中金
		{"丙寅", "火"}, {"丁卯", "火"}, // 炉中火
		{"戊辰", "木"}, {"己巳", "木"}, // 大林木
		{"庚午", "土"}, {"辛未", "土"}, // 路旁土
		{"壬申", "金"}, {"癸酉", "金"}, // 剑锋金
		{"甲戌", "火"}, {"乙亥", "火"}, // 山头火
		{"丙子", "水"}, {"丁丑", "水"}, // 涧下水
		{"戊寅", "土"}, {"己卯", "土"}, // 城头土
		{"庚辰", "金"}, {"辛巳", "金"}, // 白蜡金
		{"壬午", "木"}, {"癸未", "木"}, // 杨柳木
		{"甲申", "水"}, {"乙酉", "水"}, // 泉中水
		{"丙戌", "土"}, {"丁亥", "土"}, // 屋上土
		{"戊子", "火"}, {"己丑", "火"}, // 霹雳火
		{"庚寅", "木"}, {"辛卯", "木"}, // 松柏木
		{"壬辰", "水"}, {"癸巳", "水"}, // 长流水
		{"甲午", "金"}, {"乙未", "金"}, // 沙中金
		{"丙申", "火"}, {"丁酉", "火"}, // 山下火
		{"戊戌", "木"}, {"己亥", "木"}, // 平地木
		{"庚子", "土"}, {"辛丑", "土"}, // 壁上土
		{"壬寅", "金"}, {"癸卯", "金"}, // 金箔金
		{"甲辰", "火"}, {"乙巳", "火"}, // 覆灯火
		{"丙午", "水"}, {"丁未", "水"}, // 天河水
		{"戊申", "土"}, {"己酉", "土"}, // 大驿土
		{"庚戌", "金"}, {"辛亥", "金"}, // 钗钏金
		{"壬子", "木"}, {"癸丑", "木"}, // 桑柘木
		{"甲寅", "水"}, {"乙卯", "水"}, // 大溪水
		{"丙辰", "土"}, {"丁巳", "土"}, // 沙中土
		{"戊午", "火"}, {"己未", "火"}, // 天上火
		{"庚申", "木"}, {"辛酉", "木"}, // 石榴木
		{"壬戌", "水"}, {"癸亥", "水"}, // 大海水
	}
	for _, tc := range cases {
		got := nayinElement(tc.ganzhi)
		if got != tc.want {
			t.Errorf("nayinElement(%q) = %q, want %q", tc.ganzhi, got, tc.want)
		}
	}
}

func TestNayinElement_Unknown_ReturnsEmpty(t *testing.T) {
	if got := nayinElement(""); got != "" {
		t.Errorf("empty input: got %q, want \"\"", got)
	}
	if got := nayinElement("XX"); got != "" {
		t.Errorf("unknown ganzhi: got %q, want \"\"", got)
	}
}

func TestNayinRelation_Cases(t *testing.T) {
	cases := []struct {
		a, b string
		want string
	}{
		{"金", "水", "sheng"},
		{"水", "金", "sheng"},
		{"火", "土", "sheng"},
		{"金", "金", "same"},
		{"金", "木", "ke"},
		{"木", "金", "ke"},
		{"", "金", ""},
		{"金", "", ""},
	}
	for _, tc := range cases {
		if got := nayinRelation(tc.a, tc.b); got != tc.want {
			t.Errorf("nayinRelation(%q,%q) = %q, want %q", tc.a, tc.b, got, tc.want)
		}
	}
}

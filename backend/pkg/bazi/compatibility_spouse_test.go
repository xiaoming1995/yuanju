package bazi

import (
	"reflect"
	"testing"
)

func TestDetectSpouseStarSignal(t *testing.T) {
	cases := []struct {
		name string
		in   *BaziResult
		want SpouseStarSignal
	}{
		{
			name: "男命：正财透月干、偏财藏年支，日支藏干非财",
			in: &BaziResult{
				Gender: "male", DayGan: "甲",
				MonthGanShiShen: "正财",
				YearZhiShiShen:  []string{"偏财"},
				DayZhiShiShen:   []string{"伤官"},
			},
			want: SpouseStarSignal{
				Available: true, Present: true, Category: "财星",
				StarNames:              []string{"正财", "偏财"},
				Positions:              []string{"月干(透)", "年支(藏)"},
				Visible:                true,
				InSpousePalace:         false,
				DayBranchHiddenShiShen: []string{"伤官"},
			},
		},
		{
			name: "男命：财星坐日支（入夫妻宫，仅藏不透）",
			in: &BaziResult{
				Gender: "male", DayGan: "甲",
				DayZhiShiShen: []string{"正财"},
			},
			want: SpouseStarSignal{
				Available: true, Present: true, Category: "财星",
				StarNames:              []string{"正财"},
				Positions:              []string{"日支(藏)"},
				Visible:                false,
				InSpousePalace:         true,
				DayBranchHiddenShiShen: []string{"正财"},
			},
		},
		{
			name: "女命：正官透时干",
			in: &BaziResult{
				Gender: "female", DayGan: "甲",
				HourGanShiShen: "正官",
				DayZhiShiShen:  []string{"比肩"},
			},
			want: SpouseStarSignal{
				Available: true, Present: true, Category: "官杀",
				StarNames:              []string{"正官"},
				Positions:              []string{"时干(透)"},
				Visible:                true,
				InSpousePalace:         false,
				DayBranchHiddenShiShen: []string{"比肩"},
			},
		},
		{
			name: "男命：配偶星不现（命中无财）",
			in: &BaziResult{
				Gender: "male", DayGan: "甲",
				YearGanShiShen: "比肩",
				DayZhiShiShen:  []string{"劫财"},
			},
			want: SpouseStarSignal{
				Available: true, Present: false, Category: "财星",
				DayBranchHiddenShiShen: []string{"劫财"},
			},
		},
		{
			name: "性别缺失 → 不可用",
			in:   &BaziResult{Gender: "", DayGan: "甲"},
			want: SpouseStarSignal{Available: false},
		},
		{
			name: "缺日柱 → 不可用",
			in:   &BaziResult{Gender: "male", DayGan: ""},
			want: SpouseStarSignal{},
		},
		{
			name: "nil → 不可用",
			in:   nil,
			want: SpouseStarSignal{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectSpouseStarSignal(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("DetectSpouseStarSignal()\n got=%+v\nwant=%+v", got, tc.want)
			}
		})
	}
}

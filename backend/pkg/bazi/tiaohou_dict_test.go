package bazi

import "testing"

// TestTiaohouDictFullCoverage 校对 120 条字典全覆盖（10 天干 × 12 月支）
func TestTiaohouDictFullCoverage(t *testing.T) {
	tianGans := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	monthZhis := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

	for _, dg := range tianGans {
		for _, mz := range monthZhis {
			got := LookupTiaohouYongshen(dg, mz)
			if len(got) == 0 {
				t.Errorf("LookupTiaohouYongshen(%q,%q) 返回空，字典覆盖不全", dg, mz)
			}
		}
	}
}

// TestLookupReturnsCopy 验证返回值是字典副本，外部修改不影响下次查询
func TestLookupReturnsCopy(t *testing.T) {
	first := LookupTiaohouYongshen("甲", "寅")
	if len(first) == 0 {
		t.Fatalf("expected non-empty, got %v", first)
	}
	first[0] = "TAMPERED"
	second := LookupTiaohouYongshen("甲", "寅")
	if second[0] == "TAMPERED" {
		t.Errorf("LookupTiaohouYongshen 返回的不是副本，外部修改污染了原字典")
	}
}

// TestLookupUnknownReturnsNil 非法组合应返回 nil
func TestLookupUnknownReturnsNil(t *testing.T) {
	got := LookupTiaohouYongshen("X", "?")
	if got != nil {
		t.Errorf("expected nil for unknown combo, got %v", got)
	}
}

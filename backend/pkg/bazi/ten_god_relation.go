package bazi

type TenGodRelationMatrix struct {
	DayMaster     TenGodDayMaster        `json:"day_master"`
	HeavenlyStems []TenGodStemRelation   `json:"heavenly_stems"`
	HiddenStems   []TenGodHiddenStemGroup `json:"hidden_stems"`
}

type TenGodDayMaster struct {
	Gan    string `json:"gan"`
	Wuxing string `json:"wuxing"`
	Label  string `json:"label"`
}

type TenGodStemRelation struct {
	Pillar      string `json:"pillar"`
	PillarLabel string `json:"pillar_label"`
	Gan         string `json:"gan"`
	Wuxing      string `json:"wuxing"`
	TenGod      string `json:"ten_god"`
	Group       string `json:"group,omitempty"`
	GroupLabel  string `json:"group_label,omitempty"`
	Relation    string `json:"relation"`
	Summary     string `json:"summary"`
}

type TenGodHiddenStemGroup struct {
	Pillar      string                 `json:"pillar"`
	PillarLabel string                 `json:"pillar_label"`
	Branch      string                 `json:"branch"`
	Items       []TenGodHiddenStemItem `json:"items"`
}

type TenGodHiddenStemItem struct {
	Gan        string `json:"gan"`
	Wuxing     string `json:"wuxing"`
	TenGod     string `json:"ten_god"`
	Group      string `json:"group,omitempty"`
	GroupLabel string `json:"group_label,omitempty"`
	Relation   string `json:"relation"`
	Summary    string `json:"summary"`
}

func EnsureTenGodRelation(result *BaziResult) {
	if result == nil {
		return
	}
	if result.TenGodRelation != nil && result.TenGodRelation.DayMaster.Gan != "" {
		return
	}
	if result.DayGan == "" {
		return
	}
	result.TenGodRelation = BuildTenGodRelationMatrix(result)
}

func BuildTenGodRelationMatrix(result *BaziResult) *TenGodRelationMatrix {
	if result == nil || result.DayGan == "" {
		return nil
	}

	dayWx := result.DayGanWuxing
	if dayWx == "" {
		dayWx = ganWuxingCN(result.DayGan)
	}

	return &TenGodRelationMatrix{
		DayMaster: TenGodDayMaster{
			Gan:    result.DayGan,
			Wuxing: dayWx,
			Label:  result.DayGan + dayWx,
		},
		HeavenlyStems: []TenGodStemRelation{
			buildStemRelation(result.DayGan, "year", "年干", result.YearGan, result.YearGanWuxing, result.YearGanShiShen, false),
			buildStemRelation(result.DayGan, "month", "月干", result.MonthGan, result.MonthGanWuxing, result.MonthGanShiShen, false),
			buildStemRelation(result.DayGan, "day", "日干", result.DayGan, result.DayGanWuxing, result.DayGanShiShen, true),
			buildStemRelation(result.DayGan, "hour", "时干", result.HourGan, result.HourGanWuxing, result.HourGanShiShen, false),
		},
		HiddenStems: []TenGodHiddenStemGroup{
			buildHiddenStemGroup(result.DayGan, "year", "年支", result.YearZhi, result.YearHideGan),
			buildHiddenStemGroup(result.DayGan, "month", "月支", result.MonthZhi, result.MonthHideGan),
			buildHiddenStemGroup(result.DayGan, "day", "日支", result.DayZhi, result.DayHideGan),
			buildHiddenStemGroup(result.DayGan, "hour", "时支", result.HourZhi, result.HourHideGan),
		},
	}
}

func buildStemRelation(dayGan, pillar, label, gan, wuxing, tenGod string, isDayMaster bool) TenGodStemRelation {
	if wuxing == "" {
		wuxing = ganWuxingCN(gan)
	}
	if isDayMaster {
		return TenGodStemRelation{
			Pillar:      pillar,
			PillarLabel: label,
			Gan:         gan,
			Wuxing:      wuxing,
			TenGod:      "日主 / 日元",
			Relation:    "命主自身",
			Summary:     "这是命盘的参照点，其他十神都以此天干为中心推导。",
		}
	}
	if tenGod == "" {
		tenGod = GetShiShen(dayGan, gan)
	}
	group, groupLabel := tenGodGroupInfo(tenGod)
	return TenGodStemRelation{
		Pillar:      pillar,
		PillarLabel: label,
		Gan:         gan,
		Wuxing:      wuxing,
		TenGod:      tenGod,
		Group:       group,
		GroupLabel:  groupLabel,
		Relation:    tenGodRelationLabel(tenGod),
		Summary:     tenGodSummary(tenGod),
	}
}

func buildHiddenStemGroup(dayGan, pillar, label, branch string, hideGan []string) TenGodHiddenStemGroup {
	group := TenGodHiddenStemGroup{
		Pillar:      pillar,
		PillarLabel: label,
		Branch:      branch,
		Items:       []TenGodHiddenStemItem{},
	}
	for _, gan := range hideGan {
		if gan == "" {
			continue
		}
		tenGod := GetShiShen(dayGan, gan)
		if tenGod == "" {
			continue
		}
		groupName, groupLabel := tenGodGroupInfo(tenGod)
		group.Items = append(group.Items, TenGodHiddenStemItem{
			Gan:        gan,
			Wuxing:     ganWuxingCN(gan),
			TenGod:     tenGod,
			Group:      groupName,
			GroupLabel: groupLabel,
			Relation:   tenGodRelationLabel(tenGod),
			Summary:    tenGodSummary(tenGod),
		})
	}
	return group
}

func tenGodGroupInfo(tenGod string) (string, string) {
	info, ok := TenGodGroupOf(tenGod)
	if !ok {
		return "", ""
	}
	return info.Group, info.Label
}

func tenGodRelationLabel(tenGod string) string {
	switch tenGod {
	case "比肩", "劫财":
		return "同我"
	case "食神", "伤官":
		return "我生"
	case "正财", "偏财":
		return "我克"
	case "正官", "七杀":
		return "克我"
	case "正印", "偏印":
		return "生我"
	default:
		return ""
	}
}

func tenGodSummary(tenGod string) string {
	switch tenGod {
	case "比肩":
		return "自我、同类、独立意识与同辈协作。"
	case "劫财":
		return "同辈竞争、资源分配、行动冲劲与合作博弈。"
	case "食神":
		return "表达、才艺、稳定输出、口福与享受。"
	case "伤官":
		return "创意表达、突破规则、才华外放与锋芒。"
	case "正财":
		return "稳定财富、务实经营、责任感与现实积累。"
	case "偏财":
		return "机会资源、流动财富、人情经营与商业嗅觉。"
	case "正官":
		return "规则、职位、责任、名誉与秩序感。"
	case "七杀":
		return "外部压力、竞争、规则挑战与行动魄力。"
	case "正印":
		return "学习、贵人、保护、资质与正统资源。"
	case "偏印":
		return "灵感、研究、特殊资源、独特思维与保护。"
	default:
		return ""
	}
}

func ganWuxingCN(gan string) string {
	if wx, ok := ganWuxing[gan]; ok {
		return wxPinyin2CN[wx]
	}
	return ""
}

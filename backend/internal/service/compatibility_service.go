package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/pkg/bazi"
)

const compatibilityAnalysisVersion = "v1"

func CreateCompatibilityReading(userID string, selfProfile, partnerProfile model.CompatibilityBirthProfile) (*model.CompatibilityDetail, error) {
	selfProfile = normalizeCompatibilityProfile(selfProfile)
	partnerProfile = normalizeCompatibilityProfile(partnerProfile)

	selfChart := bazi.Calculate(
		selfProfile.Year, selfProfile.Month, selfProfile.Day, selfProfile.Hour,
		selfProfile.Gender, false, 0, selfProfile.CalendarType, selfProfile.IsLeapMonth,
	)
	partnerChart := bazi.Calculate(
		partnerProfile.Year, partnerProfile.Month, partnerProfile.Day, partnerProfile.Hour,
		partnerProfile.Gender, false, 0, partnerProfile.CalendarType, partnerProfile.IsLeapMonth,
	)

	analysis := bazi.AnalyzeCompatibility(selfChart, partnerChart)
	reading, err := repository.CreateCompatibilityReading(
		userID,
		string(analysis.OverallLevel),
		model.CompatibilityDimensionScores{
			Attraction:    analysis.DimensionScores.Attraction,
			Stability:     analysis.DimensionScores.Stability,
			Communication: analysis.DimensionScores.Communication,
			Practicality:  analysis.DimensionScores.Practicality,
		},
		model.CompatibilityDurationAssessment{
			OverallBand: analysis.DurationAssessment.OverallBand,
			Windows: model.CompatibilityDurationWindows{
				ThreeMonths:  model.CompatibilityDurationWindow{Level: string(analysis.DurationAssessment.Windows.ThreeMonths.Level)},
				OneYear:      model.CompatibilityDurationWindow{Level: string(analysis.DurationAssessment.Windows.OneYear.Level)},
				TwoYearsPlus: model.CompatibilityDurationWindow{Level: string(analysis.DurationAssessment.Windows.TwoYearsPlus.Level)},
			},
			Summary: analysis.DurationAssessment.Summary,
			Reasons: analysis.DurationAssessment.Reasons,
		},
		analysis.SummaryTags,
		compatibilityAnalysisVersion,
	)
	if err != nil {
		return nil, err
	}

	selfSnapshot, _ := json.Marshal(selfChart)
	partnerSnapshot, _ := json.Marshal(partnerChart)
	selfRaw := json.RawMessage(selfSnapshot)
	partnerRaw := json.RawMessage(partnerSnapshot)

	if _, err := repository.CreateCompatibilityParticipant(reading.ID, "self", "我", selfChart.ChartHash, selfProfile, &selfRaw); err != nil {
		return nil, err
	}
	if _, err := repository.CreateCompatibilityParticipant(reading.ID, "partner", "对方", partnerChart.ChartHash, partnerProfile, &partnerRaw); err != nil {
		return nil, err
	}

	for _, item := range analysis.Evidences {
		if _, err := repository.CreateCompatibilityEvidence(reading.ID, model.CompatibilityEvidence{
			Dimension: string(item.Dimension),
			Type:      item.Type,
			Polarity:  string(item.Polarity),
			Source:    item.Source,
			Title:     item.Title,
			Detail:    item.Detail,
			Weight:    item.Weight,
		}); err != nil {
			return nil, err
		}
	}

	return repository.GetCompatibilityDetail(reading.ID)
}

func GetCompatibilityDetailForUser(readingID, userID string) (*model.CompatibilityDetail, error) {
	ownerID, err := repository.GetCompatibilityReadingOwner(readingID)
	if err != nil {
		return nil, err
	}
	if ownerID == "" {
		return nil, nil
	}
	if ownerID != userID {
		return &model.CompatibilityDetail{}, fmt.Errorf("forbidden")
	}
	return repository.GetCompatibilityDetail(readingID)
}

func GetCompatibilityHistoryForUser(userID string, limit, offset int) ([]model.CompatibilityHistoryItem, error) {
	return repository.ListCompatibilityHistory(userID, limit, offset)
}

func GenerateCompatibilityReport(readingID, userID string) (*model.AICompatibilityReport, error) {
	ownerID, err := repository.GetCompatibilityReadingOwner(readingID)
	if err != nil {
		return nil, err
	}
	if ownerID == "" {
		return nil, fmt.Errorf("未找到合盘记录")
	}
	if ownerID != userID {
		return nil, fmt.Errorf("forbidden")
	}

	cached, err := repository.GetLatestCompatibilityReport(readingID)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}

	detail, err := repository.GetCompatibilityDetail(readingID)
	if err != nil {
		return nil, err
	}
	if detail == nil || detail.Reading == nil {
		return nil, fmt.Errorf("未找到合盘记录")
	}

	promptConfig, err := repository.GetPromptByModule("compatibility")
	if err != nil {
		return nil, err
	}
	tplContent := compatibilityPromptFallback()
	if promptConfig != nil && strings.TrimSpace(promptConfig.Content) != "" {
		tplContent = promptConfig.Content
	}

	tplData, err := buildCompatibilityPromptData(detail)
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("compatibility").Parse(tplContent)
	if err != nil {
		return nil, fmt.Errorf("compatibility Prompt 模板语法错误: %v", err)
	}
	var parsed bytes.Buffer
	if err := tmpl.Execute(&parsed, tplData); err != nil {
		return nil, fmt.Errorf("compatibility Prompt 渲染失败: %v", err)
	}

	rawContent, modelName, _, _, aiErr := callAIWithSystem(parsed.String())
	if aiErr != nil {
		return nil, aiErr
	}
	rawContent = strings.TrimSpace(rawContent)

	var structured model.CompatibilityStructuredReport
	var structuredRaw *json.RawMessage
	if err := json.Unmarshal([]byte(rawContent), &structured); err == nil {
		if marshalled, mErr := json.Marshal(structured); mErr == nil {
			raw := json.RawMessage(marshalled)
			structuredRaw = &raw
		}
	}

	return repository.CreateCompatibilityReport(readingID, rawContent, modelName, structuredRaw)
}

func buildCompatibilityPromptData(detail *model.CompatibilityDetail) (*model.CompatibilityPromptData, error) {
	var selfP, partnerP *model.CompatibilityParticipant
	for i := range detail.Participants {
		p := &detail.Participants[i]
		if p.Role == "self" {
			selfP = p
		} else if p.Role == "partner" {
			partnerP = p
		}
	}
	if selfP == nil || partnerP == nil {
		return nil, fmt.Errorf("合盘参与者信息不完整")
	}

	selfSummary, err := compatibilityParticipantSummary(selfP)
	if err != nil {
		return nil, err
	}
	partnerSummary, err := compatibilityParticipantSummary(partnerP)
	if err != nil {
		return nil, err
	}
	scoresJSON, _ := json.Marshal(detail.Reading.DimensionScores)
	durationJSON, _ := json.Marshal(detail.Reading.DurationAssessment)
	evidencesJSON, _ := json.Marshal(detail.Evidences)

	return &model.CompatibilityPromptData{
		SelfLabel:           selfP.DisplayName,
		PartnerLabel:        partnerP.DisplayName,
		SelfChartSummary:    selfSummary,
		PartnerChartSummary: partnerSummary,
		ScoresJSON:          string(scoresJSON),
		DurationJSON:        string(durationJSON),
		EvidencesJSON:       string(evidencesJSON),
		SummaryTags:         strings.Join(detail.Reading.SummaryTags, "、"),
	}, nil
}

func compatibilityParticipantSummary(p *model.CompatibilityParticipant) (string, error) {
	if p.ChartSnapshot == nil {
		return "", fmt.Errorf("chart_snapshot 缺失")
	}
	var result bazi.BaziResult
	if err := json.Unmarshal(*p.ChartSnapshot, &result); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"%s：%s%s·%s%s·%s%s·%s%s；日主=%s；五行=%d木/%d火/%d土/%d金/%d水；用神=%s；忌神=%s。",
		p.DisplayName,
		result.YearGan, result.YearZhi,
		result.MonthGan, result.MonthZhi,
		result.DayGan, result.DayZhi,
		result.HourGan, result.HourZhi,
		result.DayGan,
		result.Wuxing.Mu, result.Wuxing.Huo, result.Wuxing.Tu, result.Wuxing.Jin, result.Wuxing.Shui,
		result.Yongshen, result.Jishen,
	), nil
}

func normalizeCompatibilityProfile(p model.CompatibilityBirthProfile) model.CompatibilityBirthProfile {
	if p.CalendarType == "" {
		p.CalendarType = "solar"
	}
	return p
}

func compatibilityPromptFallback() string {
	return `你是一位专业、克制、直断的八字合盘分析师。请根据双方命盘摘要、四维分数和结构化证据，输出一份关于婚恋/姻缘匹配的分析。

人物标识：
- A：{{.SelfLabel}}
- B：{{.PartnerLabel}}

A 命盘摘要：
{{.SelfChartSummary}}

B 命盘摘要：
{{.PartnerChartSummary}}

四维分数（JSON）：
{{.ScoresJSON}}

缘分时长评估（JSON）：
{{.DurationJSON}}

关系摘要标签：
{{.SummaryTags}}

结构化证据（JSON）：
{{.EvidencesJSON}}

输出严格为 JSON：
{
  "summary": "总体判断",
  "dimensions": [
    { "key": "attraction", "title": "吸引力", "content": "..." },
    { "key": "stability", "title": "稳定度", "content": "..." },
    { "key": "communication", "title": "沟通协同", "content": "..." },
    { "key": "practicality", "title": "现实磨合", "content": "..." }
  ],
  "duration_assessment": {
    "overall_band": "medium_term",
    "summary": "阶段性维持判断",
    "reasons": ["...", "..."],
    "windows": {
      "three_months": { "level": "high" },
      "one_year": { "level": "medium" },
      "two_years_plus": { "level": "low" }
    }
  },
  "risks": ["...", "..."],
  "advice": "..."
}`
}

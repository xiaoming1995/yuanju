package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"text/template"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/pkg/bazi"
)

const compatibilityAnalysisVersion = "v2"

var compatibilityRelationshipStageLabels = map[string]string{
	"ambiguous":              "暧昧中",
	"dating":                 "恋爱中",
	"long_distance":          "异地中",
	"reconciliation":         "分手/复合中",
	"marriage_or_engagement": "已婚/谈婚论嫁",
	"crush":                  "单恋/暗恋",
	"general":                "一般关系判断",
}

var compatibilityPrimaryQuestionLabels = map[string]string{
	"continue_investment":      "值不值得继续投入",
	"marriage_suitability":     "对方适不适合结婚",
	"recurring_conflict":       "为什么总是反复拉扯",
	"reconciliation_potential": "复合有没有意义",
	"long_term_stability":      "长期能不能稳定",
	"relationship_strategy":    "怎么相处更顺",
	"general":                  "综合关系判断",
}

var compatibilityQuestionGuidance = map[string]string{
	"continue_investment":      "重点回答是否继续投入、下一步观察什么、投入节奏如何控制，以及短期要避免什么。",
	"marriage_suitability":     "重点回答婚姻适配、长期稳定、现实承接、冲突处理、家庭责任和谈婚前必须确认的问题。",
	"recurring_conflict":       "重点回答反复拉扯的结构原因、冲突触发点、修复条件，以及哪些互动模式要停止。",
	"reconciliation_potential": "重点回答是否建议复合、原问题是否可修复、复合后最容易重复的模式、验证信号和边界条件。",
	"long_term_stability":      "重点回答长期稳定基础、阶段性压力、责任分工和关系持续经营条件。",
	"relationship_strategy":    "重点回答沟通、冲突、现实安排、边界和投入节奏的具体经营策略。",
	"general":                  "重点回答关系是否值得继续观察、主要优势、主要风险和下一步行动。",
}

func CreateCompatibilityReading(userID string, selfProfile, partnerProfile model.CompatibilityBirthProfile, contexts ...model.CompatibilityContext) (*model.CompatibilityDetail, error) {
	selfProfile = normalizeCompatibilityProfile(selfProfile)
	partnerProfile = normalizeCompatibilityProfile(partnerProfile)
	context := model.CompatibilityContext{}
	if len(contexts) > 0 {
		context = contexts[0]
	}
	context = normalizeCompatibilityContext(context)

	selfChart := bazi.Calculate(
		selfProfile.Year, selfProfile.Month, selfProfile.Day, selfProfile.Hour,
		selfProfile.Gender, false, 0, selfProfile.CalendarType, selfProfile.IsLeapMonth,
	)
	partnerChart := bazi.Calculate(
		partnerProfile.Year, partnerProfile.Month, partnerProfile.Day, partnerProfile.Hour,
		partnerProfile.Gender, false, 0, partnerProfile.CalendarType, partnerProfile.IsLeapMonth,
	)

	analysis := bazi.AnalyzeCompatibility(selfChart, partnerChart)
	consulting := mapCompatibilityConsultingAssessment(analysis.ConsultingAssessment)
	reading, err := repository.CreateCompatibilityReading(
		userID,
		string(analysis.OverallLevel),
		model.CompatibilityDimensionScores{
			Attraction:    analysis.DimensionScores.Attraction,
			Stability:     analysis.DimensionScores.Stability,
			Communication: analysis.DimensionScores.Communication,
			Practicality:  analysis.DimensionScores.Practicality,
		},
		mapCompatibilityScoreExplanations(analysis.ScoreExplanations),
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
		consulting,
		analysis.SummaryTags,
		compatibilityAnalysisVersion,
		context,
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
			EvidenceKey:    item.EvidenceKey,
			Dimension:      string(item.Dimension),
			Type:           item.Type,
			Polarity:       string(item.Polarity),
			Source:         item.Source,
			Perspective:    item.Perspective,
			Actor:          item.Actor,
			Target:         item.Target,
			RelatedSources: item.RelatedSources,
			Title:          item.Title,
			Detail:         item.Detail,
			Weight:         item.Weight,
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
	detail, err := repository.GetCompatibilityDetail(readingID)
	if err != nil || detail == nil {
		return detail, err
	}
	if changed, err := ensureCompatibilityDurationAssessment(detail); err != nil {
		return nil, err
	} else if changed {
		if err := repository.UpdateCompatibilityDurationAssessment(readingID, detail.Reading.DurationAssessment); err != nil {
			return nil, err
		}
	}
	if changed, err := ensureCompatibilityEvidenceKeys(detail); err != nil {
		return nil, err
	} else if changed {
		if err := persistCompatibilityEvidenceKeys(detail); err != nil {
			return nil, err
		}
	}
	if changed, err := ensureCompatibilityConsultingAssessment(detail); err != nil {
		return nil, err
	} else if changed {
		if err := repository.UpdateCompatibilityConsultingAssessment(readingID, detail.Reading.ConsultingAssessment); err != nil {
			return nil, err
		}
	}
	return detail, nil
}

func mapCompatibilityConsultingAssessment(in bazi.CompatibilityConsultingAssessment) model.CompatibilityConsultingAssessment {
	return model.CompatibilityConsultingAssessment{
		RelationshipDiagnosis: model.CompatibilityRelationshipDiagnosis{
			RelationshipType: in.RelationshipDiagnosis.RelationshipType,
			Verdict:          in.RelationshipDiagnosis.Verdict,
			Summary:          in.RelationshipDiagnosis.Summary,
			TopFindings:      mapCompatibilityFindings(in.RelationshipDiagnosis.TopFindings),
		},
		DecisionAdvice: model.CompatibilityDecisionAdvice{
			Recommendation: in.DecisionAdvice.Recommendation,
			Confidence:     in.DecisionAdvice.Confidence,
			Conditions:     in.DecisionAdvice.Conditions,
			DoNext:         in.DecisionAdvice.DoNext,
			Avoid:          in.DecisionAdvice.Avoid,
		},
		StageRisks: mapCompatibilityStageRisks(in.StageRisks),
		RelationshipStrategy: model.CompatibilityRelationshipStrategy{
			Communication: in.RelationshipStrategy.Communication,
			Conflict:      in.RelationshipStrategy.Conflict,
			Reality:       in.RelationshipStrategy.Reality,
			Boundary:      in.RelationshipStrategy.Boundary,
		},
		ClaimEvidenceLinks: mapCompatibilityClaimLinks(in.ClaimEvidenceLinks),
	}
}

func mapCompatibilityFindings(in []bazi.CompatibilityFinding) []model.CompatibilityFinding {
	out := make([]model.CompatibilityFinding, 0, len(in))
	for _, item := range in {
		out = append(out, model.CompatibilityFinding{
			Text:         item.Text,
			EvidenceKeys: item.EvidenceKeys,
		})
	}
	return out
}

func mapCompatibilityStageRisks(in []bazi.CompatibilityStageRisk) []model.CompatibilityStageRisk {
	out := make([]model.CompatibilityStageRisk, 0, len(in))
	for _, item := range in {
		out = append(out, model.CompatibilityStageRisk{
			Window:       item.Window,
			RiskLevel:    item.RiskLevel,
			MainRisk:     item.MainRisk,
			Trigger:      item.Trigger,
			Advice:       item.Advice,
			EvidenceKeys: item.EvidenceKeys,
		})
	}
	return out
}

func mapCompatibilityClaimLinks(in []bazi.CompatibilityClaimEvidenceLink) []model.CompatibilityClaimEvidenceLink {
	out := make([]model.CompatibilityClaimEvidenceLink, 0, len(in))
	for _, item := range in {
		out = append(out, model.CompatibilityClaimEvidenceLink{
			ClaimID:      item.ClaimID,
			Claim:        item.Claim,
			EvidenceKeys: item.EvidenceKeys,
			Reasoning:    item.Reasoning,
			Caveat:       item.Caveat,
		})
	}
	return out
}

func mapCompatibilityScoreExplanations(in []bazi.CompatibilityScoreExplanation) []model.CompatibilityScoreExplanation {
	out := make([]model.CompatibilityScoreExplanation, 0, len(in))
	for _, item := range in {
		out = append(out, model.CompatibilityScoreExplanation{
			Dimension:            string(item.Dimension),
			PositiveFactor:       item.PositiveFactor,
			NegativeFactor:       item.NegativeFactor,
			PositiveEvidenceKeys: item.PositiveEvidenceKeys,
			NegativeEvidenceKeys: item.NegativeEvidenceKeys,
			Summary:              item.Summary,
		})
	}
	return out
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
	if changed, err := ensureCompatibilityDurationAssessment(detail); err != nil {
		return nil, err
	} else if changed {
		if err := repository.UpdateCompatibilityDurationAssessment(readingID, detail.Reading.DurationAssessment); err != nil {
			return nil, err
		}
	}
	if changed, err := ensureCompatibilityEvidenceKeys(detail); err != nil {
		return nil, err
	} else if changed {
		if err := persistCompatibilityEvidenceKeys(detail); err != nil {
			return nil, err
		}
	}
	if changed, err := ensureCompatibilityConsultingAssessment(detail); err != nil {
		return nil, err
	} else if changed {
		if err := repository.UpdateCompatibilityConsultingAssessment(readingID, detail.Reading.ConsultingAssessment); err != nil {
			return nil, err
		}
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

	rawContent, modelName, providerID, _, usage, aiErr := callAIWithSystem(parsed.String())
	if aiErr != nil {
		return nil, aiErr
	}
	compatPrompt := parsed.String()
	go func(uid string) {
		if logErr := repository.CreateTokenUsageLog(&uid, nil, "compatibility", modelName, providerID,
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens,
			compatPrompt, rawContent); logErr != nil {
			log.Printf("[TokenUsage] compatibility 写入失败: %v", logErr)
		}
	}(userID)
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

func ensureCompatibilityDurationAssessment(detail *model.CompatibilityDetail) (bool, error) {
	if detail == nil || detail.Reading == nil {
		return false, nil
	}
	if detail.Reading.DurationAssessment.OverallBand != "" {
		return false, nil
	}

	selfResult, partnerResult, err := compatibilityResultsFromDetail(detail)
	if err != nil {
		return false, err
	}

	analysis := bazi.AnalyzeCompatibility(selfResult, partnerResult)
	detail.Reading.DurationAssessment = model.CompatibilityDurationAssessment{
		OverallBand: analysis.DurationAssessment.OverallBand,
		Windows: model.CompatibilityDurationWindows{
			ThreeMonths:  model.CompatibilityDurationWindow{Level: string(analysis.DurationAssessment.Windows.ThreeMonths.Level)},
			OneYear:      model.CompatibilityDurationWindow{Level: string(analysis.DurationAssessment.Windows.OneYear.Level)},
			TwoYearsPlus: model.CompatibilityDurationWindow{Level: string(analysis.DurationAssessment.Windows.TwoYearsPlus.Level)},
		},
		Summary: analysis.DurationAssessment.Summary,
		Reasons: analysis.DurationAssessment.Reasons,
	}
	return true, nil
}

func ensureCompatibilityConsultingAssessment(detail *model.CompatibilityDetail) (bool, error) {
	if detail == nil || detail.Reading == nil {
		return false, nil
	}
	if detail.Reading.ConsultingAssessment.RelationshipDiagnosis.RelationshipType != "" {
		return false, nil
	}

	selfResult, partnerResult, err := compatibilityResultsFromDetail(detail)
	if err != nil {
		return false, err
	}
	analysis := bazi.AnalyzeCompatibility(selfResult, partnerResult)
	detail.Reading.ConsultingAssessment = mapCompatibilityConsultingAssessment(analysis.ConsultingAssessment)
	return true, nil
}

func ensureCompatibilityEvidenceKeys(detail *model.CompatibilityDetail) (bool, error) {
	if detail == nil || detail.Reading == nil || len(detail.Evidences) == 0 {
		return false, nil
	}
	needsBackfill := false
	for _, item := range detail.Evidences {
		if strings.TrimSpace(item.EvidenceKey) == "" {
			needsBackfill = true
			break
		}
	}
	if !needsBackfill {
		return false, nil
	}

	selfResult, partnerResult, err := compatibilityResultsFromDetail(detail)
	if err != nil {
		return false, err
	}
	analysis := bazi.AnalyzeCompatibility(selfResult, partnerResult)
	return applyCompatibilityEvidenceKeys(detail.Evidences, analysis.Evidences), nil
}

func applyCompatibilityEvidenceKeys(existing []model.CompatibilityEvidence, generated []bazi.CompatibilityEvidence) bool {
	used := make([]bool, len(generated))
	changed := false
	for i := range existing {
		if strings.TrimSpace(existing[i].EvidenceKey) != "" {
			continue
		}
		for j, candidate := range generated {
			if used[j] || !compatibilityEvidenceMatches(existing[i], candidate) {
				continue
			}
			existing[i].EvidenceKey = candidate.EvidenceKey
			used[j] = true
			changed = true
			break
		}
	}
	return changed
}

func compatibilityEvidenceMatches(existing model.CompatibilityEvidence, generated bazi.CompatibilityEvidence) bool {
	return existing.Dimension == string(generated.Dimension) &&
		existing.Type == generated.Type &&
		existing.Polarity == string(generated.Polarity) &&
		existing.Source == generated.Source &&
		existing.Title == generated.Title &&
		existing.Detail == generated.Detail &&
		existing.Weight == generated.Weight
}

func persistCompatibilityEvidenceKeys(detail *model.CompatibilityDetail) error {
	for _, item := range detail.Evidences {
		if item.ID == "" || strings.TrimSpace(item.EvidenceKey) == "" {
			continue
		}
		if err := repository.UpdateCompatibilityEvidenceKey(item.ID, item.EvidenceKey); err != nil {
			return err
		}
	}
	return nil
}

func compatibilityResultsFromDetail(detail *model.CompatibilityDetail) (*bazi.BaziResult, *bazi.BaziResult, error) {
	var selfResult, partnerResult *bazi.BaziResult
	for i := range detail.Participants {
		p := &detail.Participants[i]
		result, err := compatibilityParticipantResult(p)
		if err != nil {
			return nil, nil, err
		}
		if p.Role == "self" {
			selfResult = result
		} else if p.Role == "partner" {
			partnerResult = result
		}
	}
	if selfResult == nil || partnerResult == nil {
		return nil, nil, fmt.Errorf("合盘参与者信息不完整")
	}
	return selfResult, partnerResult, nil
}

func compatibilityParticipantResult(p *model.CompatibilityParticipant) (*bazi.BaziResult, error) {
	if p.ChartSnapshot != nil {
		var result bazi.BaziResult
		if err := json.Unmarshal(*p.ChartSnapshot, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}

	profile := normalizeCompatibilityProfile(p.BirthProfile)
	result := bazi.Calculate(
		profile.Year, profile.Month, profile.Day, profile.Hour,
		profile.Gender, false, 0, profile.CalendarType, profile.IsLeapMonth,
	)
	return result, nil
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
	scoreExplanationsJSON, _ := json.Marshal(detail.Reading.ScoreExplanations)
	durationJSON, _ := json.Marshal(detail.Reading.DurationAssessment)
	consultingJSON, _ := json.Marshal(detail.Reading.ConsultingAssessment)
	evidencesJSON, _ := json.Marshal(detail.Evidences)
	context := normalizeCompatibilityContext(model.CompatibilityContext{
		RelationshipStage: detail.Reading.RelationshipStage,
		PrimaryQuestion:   detail.Reading.PrimaryQuestion,
	})
	evidenceGroupsJSON, _ := json.Marshal(groupCompatibilityEvidences(detail.Evidences))

	return &model.CompatibilityPromptData{
		SelfLabel:              selfP.DisplayName,
		PartnerLabel:           partnerP.DisplayName,
		RelationshipStage:      context.RelationshipStage,
		RelationshipStageLabel: compatibilityRelationshipStageLabels[context.RelationshipStage],
		PrimaryQuestion:        context.PrimaryQuestion,
		PrimaryQuestionLabel:   compatibilityPrimaryQuestionLabels[context.PrimaryQuestion],
		QuestionGuidance:       compatibilityQuestionGuidance[context.PrimaryQuestion],
		SelfChartSummary:       selfSummary,
		PartnerChartSummary:    partnerSummary,
		ScoresJSON:             string(scoresJSON),
		ScoreExplanationsJSON:  string(scoreExplanationsJSON),
		DurationJSON:           string(durationJSON),
		ConsultingJSON:         string(consultingJSON),
		EvidencesJSON:          string(evidencesJSON),
		EvidenceGroupsJSON:     string(evidenceGroupsJSON),
		SummaryTags:            strings.Join(detail.Reading.SummaryTags, "、"),
	}, nil
}

func groupCompatibilityEvidences(evidences []model.CompatibilityEvidence) map[string][]model.CompatibilityEvidence {
	groups := map[string][]model.CompatibilityEvidence{}
	for _, item := range evidences {
		key := item.Source
		if key == "" {
			key = "unknown"
		}
		groups[key] = append(groups[key], item)
	}
	return groups
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

func normalizeCompatibilityContext(context model.CompatibilityContext) model.CompatibilityContext {
	context.RelationshipStage = strings.TrimSpace(context.RelationshipStage)
	context.PrimaryQuestion = strings.TrimSpace(context.PrimaryQuestion)
	if _, ok := compatibilityRelationshipStageLabels[context.RelationshipStage]; !ok {
		context.RelationshipStage = "general"
	}
	if _, ok := compatibilityPrimaryQuestionLabels[context.PrimaryQuestion]; !ok {
		context.PrimaryQuestion = "general"
	}
	return context
}

func compatibilityPromptFallback() string {
	return `你是一位专业、克制、直断的八字合盘分析师。请根据双方命盘摘要、四维分数、分数解释和结构化证据，输出一份关于婚恋/姻缘匹配的分析。

人物标识：
- A：{{.SelfLabel}}
- B：{{.PartnerLabel}}

用户关系背景：
- 当前关系阶段：{{.RelationshipStageLabel}}
- 用户最关心的问题：{{.PrimaryQuestionLabel}}
- 报告侧重点：{{.QuestionGuidance}}

A 命盘摘要：
{{.SelfChartSummary}}

B 命盘摘要：
{{.PartnerChartSummary}}

四维分数（JSON）：
{{.ScoresJSON}}

四维分数解释（JSON，包含每个维度的主要支撑与压力证据）：
{{.ScoreExplanationsJSON}}

缘分时长评估（JSON）：
{{.DurationJSON}}

关系摘要标签：
{{.SummaryTags}}

咨询型结构化诊断（JSON）：
{{.ConsultingJSON}}

结构化证据（JSON）：
{{.EvidencesJSON}}

按证据来源分组（JSON）：
{{.EvidenceGroupsJSON}}

证据约束：
- 所有主要判断必须引用 evidence_key。
- 可以使用 perspective/actor/target 理解方向性证据。
- 不得输出具体结婚、分手、复合、出轨、怀孕等确定事件日期。
- 若正负证据混合，必须表达条件、边界和可验证行为，不能写成绝对命运。

问题分支要求：
- 当 primary_question = reconciliation_potential：必须直接回答是否建议复合、原问题是否可修复、复合后最容易重复的模式、需要验证的信号、以及应停止尝试的边界条件。
- 当 primary_question = marriage_suitability：必须直接回答是否适合进入婚姻/谈婚，覆盖长期稳定、现实承接、冲突处理、家庭责任边界，并列出谈婚前必须确认的问题。
- 当 primary_question = continue_investment：必须直接回答是否继续投入，覆盖下一步观察点、投入节奏、短期承诺边界、以及当前最该避免的行为。
- 其他 primary_question：围绕用户问题输出同等颗粒度的判断、验证点和边界条件。

输出严格为 JSON：
{
  "summary": "总体判断，必须基于输入证据，不使用绝对断语",
  "question_focus": {
    "title": "围绕用户问题的章节标题，例如复合判断、婚姻适配判断、继续投入判断",
    "judgment": "直接回答用户最关心的问题，但必须使用条件语言",
    "key_checks": ["接下来需要观察或确认的信号"],
    "boundary_conditions": ["出现这些情况时应放缓、暂停或重新评估"]
  },
  "relationship_diagnosis": {
    "relationship_type": "短期吸引强、长期承压型",
    "verdict": "建议谨慎观察",
    "summary": "双方初期靠近感较强，但长期稳定更依赖沟通节奏和现实安排是否能对齐。",
    "top_findings": [
      {
        "text": "吸引力有明显支点，但稳定维度存在拉扯。",
        "evidence_keys": ["spouse_palace_stability_spouse_palace_chong"]
      }
    ]
  },
  "decision_advice": {
    "recommendation": "observe",
    "confidence": "medium",
    "conditions": ["先建立稳定沟通规则"],
    "do_next": ["用一到两个月观察冲突后的修复能力"],
    "avoid": ["用短期吸引感替代长期判断"]
  },
  "stage_risks": [
    {
      "window": "three_months",
      "risk_level": "medium",
      "main_risk": "热度高但节奏不一致",
      "trigger": "一方推进过快、另一方需要空间时",
      "advice": "先约定沟通频率和边界，不急于做长期承诺",
      "evidence_keys": ["day_master_communication_day_master_controlling"]
    }
  ],
  "relationship_strategy": {
    "communication": "重要议题用明确约定替代情绪试探。",
    "conflict": "争执时先暂停升级，再回到具体事件和责任分工。",
    "reality": "长期计划需要拆成可验证的小步骤。",
    "boundary": "初期保留个人节奏，避免过快形成单方依赖。"
  },
  "claim_evidence_links": [
    {
      "claim_id": "long_term_pressure",
      "claim": "长期关系需要额外经营稳定感。",
      "evidence_keys": ["spouse_palace_stability_spouse_palace_chong"],
      "reasoning": "夫妻宫冲动和现实磨合信号叠加时，关系更容易在长期安排中反复消耗。",
      "caveat": "若双方能建立清晰沟通规则，负向信号的影响会被削弱。"
    }
  ],
  "dimensions": [
    { "key": "attraction", "title": "吸引力", "content": "基于证据的维度解释" },
    { "key": "stability", "title": "稳定度", "content": "基于证据的维度解释" },
    { "key": "communication", "title": "沟通协同", "content": "基于证据的维度解释" },
    { "key": "practicality", "title": "现实磨合", "content": "基于证据的维度解释" }
  ],
  "duration_assessment": {
    "overall_band": "medium_term",
    "summary": "阶段性维持判断",
    "reasons": ["只引用输入中已有的阶段原因"],
    "windows": {
      "three_months": { "level": "high" },
      "one_year": { "level": "medium" },
      "two_years_plus": { "level": "low" }
    }
  },
  "risks": ["基于证据的风险点"],
  "advice": "综合建议"
}`
}

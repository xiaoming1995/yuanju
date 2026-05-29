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
	"yuanju/pkg/prompt"
)

const compatibilityAnalysisVersion = "v3.1"

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

func CreateCompatibilityReading(userID string, selfProfile, partnerProfile model.CompatibilityBirthProfile, context model.CompatibilityContext, displayNames model.CompatibilityDisplayNames) (*model.CompatibilityDetail, error) {
	selfProfile = normalizeCompatibilityProfile(selfProfile)
	partnerProfile = normalizeCompatibilityProfile(partnerProfile)
	context = normalizeCompatibilityContext(context)
	displayNames = normalizeCompatibilityDisplayNames(displayNames)

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
		analysis.OverallScore,
		model.CompatibilityDimensionScores{
			Zodiac:     analysis.DimensionScores.Zodiac,
			Nayin:      analysis.DimensionScores.Nayin,
			DayPillar:  analysis.DimensionScores.DayPillar,
			EightChars: analysis.DimensionScores.EightChars,
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

	if _, err := repository.CreateCompatibilityParticipant(reading.ID, "self", displayNames.Self, selfChart.ChartHash, selfProfile, &selfRaw); err != nil {
		return nil, err
	}
	if _, err := repository.CreateCompatibilityParticipant(reading.ID, "partner", displayNames.Partner, partnerChart.ChartHash, partnerProfile, &partnerRaw); err != nil {
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
	var tplContent string
	if promptConfig != nil && strings.TrimSpace(promptConfig.Content) != "" {
		tplContent = promptConfig.Content
	} else {
		// 极端 fallback：DB 没记录（SyncCanonical 失败 + 首次部署）时，用代码注册表权威版本
		tplContent = prompt.MustGet("compatibility").Content
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
	shishen := fmt.Sprintf("年%s/%s·月%s/%s·日%s/%s·时%s/%s",
		zeroDash(result.YearGanShiShen), zeroDash(strings.Join(result.YearZhiShiShen, ",")),
		zeroDash(result.MonthGanShiShen), zeroDash(strings.Join(result.MonthZhiShiShen, ",")),
		zeroDash(result.DayGanShiShen), zeroDash(strings.Join(result.DayZhiShiShen, ",")),
		zeroDash(result.HourGanShiShen), zeroDash(strings.Join(result.HourZhiShiShen, ",")),
	)
	level, _, _ := bazi.GetStrengthDetail(&result)
	strength := compatibilityStrengthLabels[level]
	if strength == "" {
		strength = "中和"
	}
	return fmt.Sprintf(
		"%s：%s%s·%s%s·%s%s·%s%s；日主=%s；五行=%d木/%d火/%d土/%d金/%d水；十神=%s；命格=%s；旺衰=%s；用神=%s；忌神=%s。",
		p.DisplayName,
		result.YearGan, result.YearZhi,
		result.MonthGan, result.MonthZhi,
		result.DayGan, result.DayZhi,
		result.HourGan, result.HourZhi,
		result.DayGan,
		result.Wuxing.Mu, result.Wuxing.Huo, result.Wuxing.Tu, result.Wuxing.Jin, result.Wuxing.Shui,
		shishen, zeroDash(result.MingGe), strength,
		result.Yongshen, result.Jishen,
	), nil
}

// compatibilityStrengthLabels 把日主旺衰等级（GetStrengthDetail）映射为中文标签。
var compatibilityStrengthLabels = map[string]string{
	"vstrong": "身旺",
	"strong":  "偏强",
	"neutral": "中和",
	"weak":    "偏弱",
	"vweak":   "身弱",
}

func zeroDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
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

func normalizeCompatibilityDisplayNames(in model.CompatibilityDisplayNames) model.CompatibilityDisplayNames {
	self := strings.TrimSpace(in.Self)
	partner := strings.TrimSpace(in.Partner)
	if self == "" {
		self = "我"
	}
	if partner == "" {
		partner = "对方"
	}
	if len([]rune(self)) > 20 {
		self = string([]rune(self)[:20])
	}
	if len([]rune(partner)) > 20 {
		partner = string([]rune(partner)[:20])
	}
	return model.CompatibilityDisplayNames{Self: self, Partner: partner}
}

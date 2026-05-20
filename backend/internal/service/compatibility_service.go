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
	evidenceGroupsJSON, _ := json.Marshal(groupCompatibilityEvidences(detail.Evidences))

	return &model.CompatibilityPromptData{
		SelfLabel:             selfP.DisplayName,
		PartnerLabel:          partnerP.DisplayName,
		SelfChartSummary:      selfSummary,
		PartnerChartSummary:   partnerSummary,
		ScoresJSON:            string(scoresJSON),
		ScoreExplanationsJSON: string(scoreExplanationsJSON),
		DurationJSON:          string(durationJSON),
		ConsultingJSON:        string(consultingJSON),
		EvidencesJSON:         string(evidencesJSON),
		EvidenceGroupsJSON:    string(evidenceGroupsJSON),
		SummaryTags:           strings.Join(detail.Reading.SummaryTags, "、"),
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

func compatibilityPromptFallback() string {
	return `你是一位专业、克制、直断的八字合盘分析师。请根据双方命盘摘要、四维分数、分数解释和结构化证据，输出一份关于婚恋/姻缘匹配的分析。

人物标识：
- A：{{.SelfLabel}}
- B：{{.PartnerLabel}}

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

输出严格为 JSON：
{
  "summary": "总体判断，必须基于输入证据，不使用绝对断语",
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

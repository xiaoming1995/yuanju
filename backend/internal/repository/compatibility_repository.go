package repository

import (
	"database/sql"
	"encoding/json"
	"time"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

func CreateCompatibilityReading(userID, overallLevel string, overallScore int, scores model.CompatibilityDimensionScores, scoreExplanations []model.CompatibilityScoreExplanation, duration model.CompatibilityDurationAssessment, consulting model.CompatibilityConsultingAssessment, summaryTags []string, analysisVersion string, context model.CompatibilityContext) (*model.CompatibilityReading, error) {
	scoresJSON, _ := json.Marshal(scores)
	scoreExplanationsJSON, _ := json.Marshal(scoreExplanations)
	durationJSON, _ := json.Marshal(duration)
	consultingJSON, _ := json.Marshal(consulting)
	tagsJSON, _ := json.Marshal(summaryTags)

	r := &model.CompatibilityReading{}
	var rawScores, rawScoreExplanations, rawDuration, rawConsulting, rawTags []byte
	err := database.DB.QueryRow(
		`INSERT INTO compatibility_readings (user_id, overall_level, overall_score, dimension_scores, score_explanations, duration_assessment, consulting_assessment, summary_tags, analysis_version, relationship_stage, primary_question)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id, user_id, relationship_stage, primary_question, overall_score, overall_level, dimension_scores, score_explanations, duration_assessment, consulting_assessment, summary_tags, analysis_version, created_at, updated_at`,
		userID, overallLevel, overallScore, scoresJSON, scoreExplanationsJSON, durationJSON, consultingJSON, tagsJSON, analysisVersion, context.RelationshipStage, context.PrimaryQuestion,
	).Scan(&r.ID, &r.UserID, &r.RelationshipStage, &r.PrimaryQuestion, &r.OverallScore, &r.OverallLevel, &rawScores, &rawScoreExplanations, &rawDuration, &rawConsulting, &rawTags, &r.AnalysisVersion, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(rawScores, &r.DimensionScores)
	_ = json.Unmarshal(rawScoreExplanations, &r.ScoreExplanations)
	_ = json.Unmarshal(rawDuration, &r.DurationAssessment)
	_ = json.Unmarshal(rawConsulting, &r.ConsultingAssessment)
	_ = json.Unmarshal(rawTags, &r.SummaryTags)
	return r, nil
}

func CreateCompatibilityParticipant(readingID, role, displayName, chartHash string, birthProfile model.CompatibilityBirthProfile, chartSnapshot *json.RawMessage) (*model.CompatibilityParticipant, error) {
	birthJSON, _ := json.Marshal(birthProfile)

	p := &model.CompatibilityParticipant{}
	var rawBirth []byte
	err := database.DB.QueryRow(
		`INSERT INTO compatibility_participants (reading_id, role, display_name, birth_profile, chart_hash, chart_snapshot)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, reading_id, role, display_name, birth_profile, chart_hash, chart_snapshot, created_at`,
		readingID, role, displayName, birthJSON, chartHash, chartSnapshot,
	).Scan(&p.ID, &p.ReadingID, &p.Role, &p.DisplayName, &rawBirth, &p.ChartHash, &p.ChartSnapshot, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(rawBirth, &p.BirthProfile)
	return p, nil
}

func CreateCompatibilityEvidence(readingID string, evidence model.CompatibilityEvidence) (*model.CompatibilityEvidence, error) {
	out := &model.CompatibilityEvidence{}
	relatedSourcesJSON, _ := json.Marshal(evidence.RelatedSources)
	var rawRelatedSources []byte
	err := database.DB.QueryRow(
		`INSERT INTO compatibility_evidences (reading_id, evidence_key, dimension, type, polarity, source, perspective, actor, target, related_sources, title, detail, weight)
		 VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), $10, $11, $12, $13)
		 RETURNING id, reading_id, evidence_key, dimension, type, polarity, source, COALESCE(perspective, ''), COALESCE(actor, ''), COALESCE(target, ''), related_sources, title, detail, weight, created_at`,
		readingID, evidence.EvidenceKey, evidence.Dimension, evidence.Type, evidence.Polarity, evidence.Source, evidence.Perspective, evidence.Actor, evidence.Target, relatedSourcesJSON, evidence.Title, evidence.Detail, evidence.Weight,
	).Scan(&out.ID, &out.ReadingID, &out.EvidenceKey, &out.Dimension, &out.Type, &out.Polarity, &out.Source, &out.Perspective, &out.Actor, &out.Target, &rawRelatedSources, &out.Title, &out.Detail, &out.Weight, &out.CreatedAt)
	_ = json.Unmarshal(rawRelatedSources, &out.RelatedSources)
	return out, err
}

func CreateCompatibilityReport(readingID, content, modelName string, contentStructured *json.RawMessage) (*model.AICompatibilityReport, error) {
	r := &model.AICompatibilityReport{}
	err := database.DB.QueryRow(
		`INSERT INTO ai_compatibility_reports (reading_id, content, content_structured, model)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, reading_id, content, content_structured, model, created_at`,
		readingID, content, contentStructured, modelName,
	).Scan(&r.ID, &r.ReadingID, &r.Content, &r.ContentStructured, &r.Model, &r.CreatedAt)
	return r, err
}

func GetLatestCompatibilityReport(readingID string) (*model.AICompatibilityReport, error) {
	r := &model.AICompatibilityReport{}
	err := database.DB.QueryRow(
		`SELECT id, reading_id, content, content_structured, model, created_at
		 FROM ai_compatibility_reports
		 WHERE reading_id = $1
		 ORDER BY created_at DESC
		 LIMIT 1`,
		readingID,
	).Scan(&r.ID, &r.ReadingID, &r.Content, &r.ContentStructured, &r.Model, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func GetCompatibilityReadingByID(readingID string) (*model.CompatibilityReading, error) {
	r := &model.CompatibilityReading{}
	var rawScores, rawScoreExplanations, rawDuration, rawConsulting, rawTags []byte
	err := database.DB.QueryRow(
		`SELECT id, user_id, relationship_stage, primary_question, overall_score, overall_level, dimension_scores, COALESCE(score_explanations, '[]'::jsonb), duration_assessment, consulting_assessment, summary_tags, analysis_version, created_at, updated_at
		 FROM compatibility_readings
		 WHERE id = $1`,
		readingID,
	).Scan(&r.ID, &r.UserID, &r.RelationshipStage, &r.PrimaryQuestion, &r.OverallScore, &r.OverallLevel, &rawScores, &rawScoreExplanations, &rawDuration, &rawConsulting, &rawTags, &r.AnalysisVersion, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(rawScores, &r.DimensionScores)
	_ = json.Unmarshal(rawScoreExplanations, &r.ScoreExplanations)
	_ = json.Unmarshal(rawDuration, &r.DurationAssessment)
	_ = json.Unmarshal(rawConsulting, &r.ConsultingAssessment)
	_ = json.Unmarshal(rawTags, &r.SummaryTags)
	return r, nil
}

// DeleteCompatibilityReading 删除合盘记录；参与者/证据/报告子表由外键
// ON DELETE CASCADE 自动清除，计费日志保留。返回删除条数。
func DeleteCompatibilityReading(readingID string) (int64, error) {
	result, err := database.DB.Exec(`DELETE FROM compatibility_readings WHERE id=$1`, readingID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func GetCompatibilityParticipants(readingID string) ([]model.CompatibilityParticipant, error) {
	rows, err := database.DB.Query(
		`SELECT id, reading_id, role, display_name, birth_profile, chart_hash, chart_snapshot, created_at
		 FROM compatibility_participants
		 WHERE reading_id = $1
		 ORDER BY CASE role WHEN 'self' THEN 0 ELSE 1 END, created_at ASC`,
		readingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.CompatibilityParticipant{}
	for rows.Next() {
		var item model.CompatibilityParticipant
		var rawBirth []byte
		if err := rows.Scan(&item.ID, &item.ReadingID, &item.Role, &item.DisplayName, &rawBirth, &item.ChartHash, &item.ChartSnapshot, &item.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(rawBirth, &item.BirthProfile)
		out = append(out, item)
	}
	return out, rows.Err()
}

func GetCompatibilityEvidences(readingID string) ([]model.CompatibilityEvidence, error) {
	rows, err := database.DB.Query(
		`SELECT id, reading_id, evidence_key, dimension, type, polarity, source, COALESCE(perspective, ''), COALESCE(actor, ''), COALESCE(target, ''), COALESCE(related_sources, '[]'::jsonb), title, detail, weight, created_at
		 FROM compatibility_evidences
		 WHERE reading_id = $1
		 ORDER BY ABS(weight) DESC, created_at ASC`,
		readingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.CompatibilityEvidence{}
	for rows.Next() {
		var item model.CompatibilityEvidence
		var rawRelatedSources []byte
		if err := rows.Scan(&item.ID, &item.ReadingID, &item.EvidenceKey, &item.Dimension, &item.Type, &item.Polarity, &item.Source, &item.Perspective, &item.Actor, &item.Target, &rawRelatedSources, &item.Title, &item.Detail, &item.Weight, &item.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(rawRelatedSources, &item.RelatedSources)
		out = append(out, item)
	}
	return out, rows.Err()
}

func GetCompatibilityDetail(readingID string) (*model.CompatibilityDetail, error) {
	reading, err := GetCompatibilityReadingByID(readingID)
	if err != nil || reading == nil {
		return nil, err
	}
	participants, err := GetCompatibilityParticipants(readingID)
	if err != nil {
		return nil, err
	}
	evidences, err := GetCompatibilityEvidences(readingID)
	if err != nil {
		return nil, err
	}
	report, err := GetLatestCompatibilityReport(readingID)
	if err != nil {
		return nil, err
	}
	return &model.CompatibilityDetail{
		Reading:      reading,
		Participants: participants,
		Evidences:    evidences,
		LatestReport: report,
	}, nil
}

func UpdateCompatibilityDurationAssessment(readingID string, duration model.CompatibilityDurationAssessment) error {
	durationJSON, _ := json.Marshal(duration)
	_, err := database.DB.Exec(
		`UPDATE compatibility_readings
		 SET duration_assessment = $2, updated_at = NOW()
		 WHERE id = $1`,
		readingID, durationJSON,
	)
	return err
}

func UpdateCompatibilityConsultingAssessment(readingID string, consulting model.CompatibilityConsultingAssessment) error {
	consultingJSON, _ := json.Marshal(consulting)
	_, err := database.DB.Exec(
		`UPDATE compatibility_readings
		 SET consulting_assessment = $2, updated_at = NOW()
		 WHERE id = $1`,
		readingID, consultingJSON,
	)
	return err
}

func UpdateCompatibilityEvidenceKey(evidenceID, evidenceKey string) error {
	_, err := database.DB.Exec(
		`UPDATE compatibility_evidences
		 SET evidence_key = $2
		 WHERE id = $1`,
		evidenceID, evidenceKey,
	)
	return err
}

func GetCompatibilityReadingOwner(readingID string) (string, error) {
	var userID string
	err := database.DB.QueryRow(
		`SELECT user_id FROM compatibility_readings WHERE id = $1`,
		readingID,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return userID, err
}

func ListCompatibilityHistory(userID string, limit, offset int) ([]model.CompatibilityHistoryItem, error) {
	rows, err := database.DB.Query(
		`SELECT id, relationship_stage, primary_question, overall_score, overall_level, dimension_scores, summary_tags, created_at
		 FROM compatibility_readings
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.CompatibilityHistoryItem{}
	for rows.Next() {
		var item model.CompatibilityHistoryItem
		var rawScores, rawTags []byte
		if err := rows.Scan(&item.ID, &item.RelationshipStage, &item.PrimaryQuestion, &item.OverallScore, &item.OverallLevel, &rawScores, &rawTags, &item.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(rawScores, &item.DimensionScores)
		_ = json.Unmarshal(rawTags, &item.SummaryTags)
		participants, err := GetCompatibilityParticipants(item.ID)
		if err != nil {
			return nil, err
		}
		for _, p := range participants {
			if p.Role == "self" {
				item.SelfName = p.DisplayName
			} else if p.Role == "partner" {
				item.PartnerName = p.DisplayName
			}
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func ListCompatibilityReportsByReadingID(readingID string) ([]model.AICompatibilityReport, error) {
	rows, err := database.DB.Query(
		`SELECT id, reading_id, content, content_structured, model, created_at
		 FROM ai_compatibility_reports
		 WHERE reading_id = $1
		 ORDER BY created_at DESC`,
		readingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.AICompatibilityReport{}
	for rows.Next() {
		var item model.AICompatibilityReport
		if err := rows.Scan(&item.ID, &item.ReadingID, &item.Content, &item.ContentStructured, &item.Model, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CompatibilityNow() time.Time {
	return time.Now()
}

// DeleteAICompatibilityReportsOlderThan 删除超期合盘 AI 报告（不动业务表）。
func DeleteAICompatibilityReportsOlderThan(cutoff time.Time) (int64, error) {
	res, err := database.DB.Exec(`DELETE FROM ai_compatibility_reports WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// AdminListCompatibilityReadings 后台全量合盘列表（分页，含创建者邮箱）
func AdminListCompatibilityReadings(page, pageSize int) ([]model.AdminCompatListItem, int, error) {
	offset := (page - 1) * pageSize

	var total int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM compatibility_readings`).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.AdminCompatListItem{}, 0, nil
	}

	rows, err := database.DB.Query(`
		SELECT r.id, u.email, r.overall_score, r.overall_level,
		       r.relationship_stage, r.primary_question, r.analysis_version, r.created_at,
		       COALESCE((SELECT display_name FROM compatibility_participants WHERE reading_id=r.id AND role='self'  LIMIT 1), '') AS self_name,
		       COALESCE((SELECT display_name FROM compatibility_participants WHERE reading_id=r.id AND role='partner' LIMIT 1), '') AS partner_name
		FROM compatibility_readings r
		LEFT JOIN users u ON r.user_id = u.id
		ORDER BY r.created_at DESC
		LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []model.AdminCompatListItem{}
	for rows.Next() {
		var it model.AdminCompatListItem
		if err := rows.Scan(&it.ID, &it.UserEmail, &it.OverallScore, &it.OverallLevel,
			&it.RelationshipStage, &it.PrimaryQuestion, &it.AnalysisVersion, &it.CreatedAt,
			&it.SelfName, &it.PartnerName); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

// GetCompatibilityReadingUserEmail 取某条合盘创建者邮箱（游客或已删用户返回空串）
func GetCompatibilityReadingUserEmail(readingID string) (string, error) {
	var email sql.NullString
	err := database.DB.QueryRow(
		`SELECT u.email FROM compatibility_readings r LEFT JOIN users u ON r.user_id=u.id WHERE r.id=$1`,
		readingID,
	).Scan(&email)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return email.String, err
}

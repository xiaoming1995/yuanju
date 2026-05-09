package repository

import (
	"database/sql"
	"encoding/json"
	"time"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

func CreateCompatibilityReading(userID, overallLevel string, scores model.CompatibilityDimensionScores, summaryTags []string, analysisVersion string) (*model.CompatibilityReading, error) {
	scoresJSON, _ := json.Marshal(scores)
	tagsJSON, _ := json.Marshal(summaryTags)

	r := &model.CompatibilityReading{}
	var rawScores, rawTags []byte
	err := database.DB.QueryRow(
		`INSERT INTO compatibility_readings (user_id, overall_level, dimension_scores, summary_tags, analysis_version)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, overall_level, dimension_scores, summary_tags, analysis_version, created_at, updated_at`,
		userID, overallLevel, scoresJSON, tagsJSON, analysisVersion,
	).Scan(&r.ID, &r.UserID, &r.OverallLevel, &rawScores, &rawTags, &r.AnalysisVersion, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(rawScores, &r.DimensionScores)
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
	err := database.DB.QueryRow(
		`INSERT INTO compatibility_evidences (reading_id, dimension, type, polarity, source, title, detail, weight)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, reading_id, dimension, type, polarity, source, title, detail, weight, created_at`,
		readingID, evidence.Dimension, evidence.Type, evidence.Polarity, evidence.Source, evidence.Title, evidence.Detail, evidence.Weight,
	).Scan(&out.ID, &out.ReadingID, &out.Dimension, &out.Type, &out.Polarity, &out.Source, &out.Title, &out.Detail, &out.Weight, &out.CreatedAt)
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
	var rawScores, rawTags []byte
	err := database.DB.QueryRow(
		`SELECT id, user_id, overall_level, dimension_scores, summary_tags, analysis_version, created_at, updated_at
		 FROM compatibility_readings
		 WHERE id = $1`,
		readingID,
	).Scan(&r.ID, &r.UserID, &r.OverallLevel, &rawScores, &rawTags, &r.AnalysisVersion, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(rawScores, &r.DimensionScores)
	_ = json.Unmarshal(rawTags, &r.SummaryTags)
	return r, nil
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
		`SELECT id, reading_id, dimension, type, polarity, source, title, detail, weight, created_at
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
		if err := rows.Scan(&item.ID, &item.ReadingID, &item.Dimension, &item.Type, &item.Polarity, &item.Source, &item.Title, &item.Detail, &item.Weight, &item.CreatedAt); err != nil {
			return nil, err
		}
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
		`SELECT id, overall_level, dimension_scores, summary_tags, created_at
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
		if err := rows.Scan(&item.ID, &item.OverallLevel, &rawScores, &rawTags, &item.CreatedAt); err != nil {
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

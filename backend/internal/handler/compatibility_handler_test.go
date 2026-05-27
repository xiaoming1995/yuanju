package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yuanju/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newCompatibilityRouter() *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Next()
	})
	r.POST("/compatibility", CreateCompatibilityReading)
	r.GET("/compatibility/history", GetCompatibilityHistory)
	r.GET("/compatibility/:id", GetCompatibilityDetail)
	r.POST("/compatibility/:id/report", GenerateCompatibilityReport)
	return r
}

func TestCreateCompatibilityReading_MissingFields(t *testing.T) {
	r := newCompatibilityRouter()

	cases := []struct {
		name     string
		body     map[string]any
		wantCode int
	}{
		{
			name:     "empty body",
			body:     map[string]any{},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing partner",
			body: map[string]any{
				"self": map[string]any{
					"year": 1990, "month": 1, "day": 1, "hour": 0, "gender": "male",
				},
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid gender",
			body: map[string]any{
				"self": map[string]any{
					"year": 1990, "month": 1, "day": 1, "hour": 0, "gender": "unknown",
				},
				"partner": map[string]any{
					"year": 1992, "month": 6, "day": 15, "hour": 12, "gender": "female",
				},
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "year out of range",
			body: map[string]any{
				"self": map[string]any{
					"year": 1800, "month": 1, "day": 1, "hour": 0, "gender": "male",
				},
				"partner": map[string]any{
					"year": 1992, "month": 6, "day": 15, "hour": 12, "gender": "female",
				},
			},
			wantCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/compatibility", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.wantCode {
				t.Errorf("expected %d, got %d: %s", tc.wantCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestCreateCompatibilityReadingRequest_DecodesRelationshipContext(t *testing.T) {
	body := []byte(`{
		"relationship_stage": "dating",
		"primary_question": "marriage_suitability",
		"self": {"year": 1990, "month": 1, "day": 1, "hour": 0, "gender": "male"},
		"partner": {"year": 1992, "month": 6, "day": 15, "hour": 12, "gender": "female"}
	}`)
	var req CreateCompatibilityReadingRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatal(err)
	}
	if req.RelationshipStage != "dating" {
		t.Fatalf("expected relationship stage to decode, got %q", req.RelationshipStage)
	}
	if req.PrimaryQuestion != "marriage_suitability" {
		t.Fatalf("expected primary question to decode, got %q", req.PrimaryQuestion)
	}
}

func TestCreateCompatibilityReadingRequest_DecodesDisplayNames(t *testing.T) {
	body := []byte(`{
		"self_display_name": "我",
		"partner_display_name": "小王",
		"self": {"year": 1990, "month": 1, "day": 1, "hour": 0, "gender": "male"},
		"partner": {"year": 1992, "month": 6, "day": 15, "hour": 12, "gender": "female"}
	}`)
	var req CreateCompatibilityReadingRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatal(err)
	}
	if req.SelfDisplayName != "我" {
		t.Fatalf("expected self display name to decode, got %q", req.SelfDisplayName)
	}
	if req.PartnerDisplayName != "小王" {
		t.Fatalf("expected partner display name to decode, got %q", req.PartnerDisplayName)
	}
}

func TestCompatibilityProfileInput_AllowsZeroHour(t *testing.T) {
	validate, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		t.Fatal("expected gin validator engine")
	}
	input := CompatibilityProfileInput{
		Year:   1990,
		Month:  1,
		Day:    1,
		Hour:   0,
		Gender: "male",
	}
	if err := validate.Struct(input); err != nil {
		t.Fatalf("expected hour 0 to be valid, got %v", err)
	}
}

func TestCompatibilityDetailJSON_IncludesConsultingShape(t *testing.T) {
	detail := model.CompatibilityDetail{
		Reading: &model.CompatibilityReading{
			RelationshipStage: "reconciliation",
			PrimaryQuestion:   "reconciliation_potential",
			ConsultingAssessment: model.CompatibilityConsultingAssessment{
				RelationshipDiagnosis: model.CompatibilityRelationshipDiagnosis{
					RelationshipType: "短期吸引强、长期承压型",
					Verdict:          "建议谨慎观察",
				},
			},
		},
		Evidences: []model.CompatibilityEvidence{
			{EvidenceKey: "spouse_palace_stability_spouse_palace_chong", Title: "夫妻宫六冲"},
		},
	}
	raw, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, `"consulting_assessment"`) {
		t.Fatalf("expected consulting assessment in json: %s", body)
	}
	if !strings.Contains(body, `"relationship_diagnosis"`) {
		t.Fatalf("expected relationship diagnosis in json: %s", body)
	}
	if !strings.Contains(body, `"evidence_key"`) {
		t.Fatalf("expected evidence key in json: %s", body)
	}
	if !strings.Contains(body, `"relationship_stage":"reconciliation"`) {
		t.Fatalf("expected relationship stage in json: %s", body)
	}
	if !strings.Contains(body, `"primary_question":"reconciliation_potential"`) {
		t.Fatalf("expected primary question in json: %s", body)
	}
}

func TestCompatibilityHistoryItemJSON_IncludesRelationshipContext(t *testing.T) {
	item := model.CompatibilityHistoryItem{
		ID:                "reading-id",
		RelationshipStage: "ambiguous",
		PrimaryQuestion:   "continue_investment",
		OverallLevel:      "medium",
	}
	raw, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, `"relationship_stage":"ambiguous"`) {
		t.Fatalf("expected relationship stage in history json: %s", body)
	}
	if !strings.Contains(body, `"primary_question":"continue_investment"`) {
		t.Fatalf("expected primary question in history json: %s", body)
	}
}

func TestCompatibilityStructuredReportJSON_IncludesQuestionFocus(t *testing.T) {
	report := model.CompatibilityStructuredReport{
		QuestionFocus: model.CompatibilityQuestionFocus{
			Title:              "复合判断",
			Judgment:           "建议先验证原问题是否可修复。",
			KeyChecks:          []string{"冲突后能否修复"},
			BoundaryConditions: []string{"不要在规则未稳定前复合"},
		},
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, `"question_focus"`) {
		t.Fatalf("expected question focus in json: %s", body)
	}
	if !strings.Contains(body, `"boundary_conditions"`) {
		t.Fatalf("expected boundary conditions in json: %s", body)
	}
}

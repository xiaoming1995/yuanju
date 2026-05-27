# Tasks

## 1. Backend algorithm core (compatibility_*.go)
- [x] 1.1 Implement nayin 60-ganzhi lookup table + tests
- [x] 1.2 Implement branchCompatible (liuhe + sanhe half) + tests
- [x] 1.3 Implement scoreZodiac / scoreNayin / scoreDayPillar / scoreEightChars + tests
- [x] 1.4 Implement evidence list + score_explanations + summary_tags generators + tests
- [x] 1.5 Implement classifyRelationshipType + decision_advice + duration + stage_risks + strategy + tests
- [x] 1.6 Rewrite AnalyzeCompatibility entry; delete 11 legacy build*Signals helpers
- [x] 1.7 Integration tests covering perfect-hit / all-miss / threshold cases

## 2. Type & Service & DB
- [x] 2.1 Rename DimensionScores fields + add OverallScore in model
- [x] 2.2 Bump compatibilityAnalysisVersion to "v3"
- [x] 2.3 Add migration 00012 (overall_score column + COMMENT)
- [x] 2.4 Extend repository CreateCompatibilityReading / SELECTs to handle overall_score
- [x] 2.5 Verify go build ./... and go test ./... pass

## 3. AI Prompt
- [x] 3.1 Rewrite canonical_compatibility.go content for 4-module language
- [x] 3.2 Update canonical_test.go / sync_test.go assertions

## 4. Frontend
- [x] 4.1 Extend api.ts types with discriminated union v1/v2/v3
- [x] 4.2 Add ScoreOverviewV3 component + CSS
- [x] 4.3 Wire CompatibilityResultPage version dispatch
- [x] 4.4 Wire CompatibilityHistoryPage version dispatch

## 5. OpenSpec
- [x] 5.1 Validate spec deltas with /opsx validation
- [ ] 5.2 Archive after implementation merges

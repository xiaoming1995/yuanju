-- +goose Up
-- +goose StatementBegin

-- Ensure legacy liunian prompts expose gender-specific spouse-star rules.
-- The service layer also prepends these rules at runtime, so this migration is
-- only to make the maintained DB prompt clearer for admins and future resets.

UPDATE ai_prompts
SET content = REPLACE(
		REPLACE(
			content,
			'{{.NatalAnalysisLogic}}

目前正行',
			'{{.NatalAnalysisLogic}}

命主信息：
- 性别：{{.GenderLabel}}
- 日主：{{.DayGan}}
- 婚恋取象规则：{{.RelationshipStarRule}}
- 性别一致性硬性约束：{{.GenderGuardRule}}

目前正行'
		),
		'请为他详细批断',
		'请为命主详细批断'
	),
	updated_at = NOW()
WHERE module = 'liunian'
  AND content NOT LIKE '%GenderGuardRule%'
  AND content LIKE '%{{.NatalAnalysisLogic}}%'
  AND content LIKE '%请为他详细批断%';

-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin

-- Existing prompt rows are maintained in the database and are not replaced by
-- the code's canonical prompt during startup. Change only the requested term.
UPDATE ai_prompts
SET content = REPLACE(content, '八字', '干支'),
    updated_at = NOW()
WHERE module IN ('compatibility', 'liunian', 'past_events')
  AND content LIKE '%八字%';

UPDATE shensha_annotations
SET description = REPLACE(description, '八字', '干支'),
    updated_at = NOW()
WHERE description LIKE '%八字%';

-- +goose StatementEnd

-- +goose Down
-- Historical wording cannot be reconstructed after a targeted text update.

package repository

import (
	"database/sql"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

// GetPromptByModule 根据模块名称获取 Prompt
func GetPromptByModule(module string) (*model.AIPrompt, error) {
	prompt := &model.AIPrompt{}
	err := database.DB.QueryRow(
		`SELECT id, module, content, description, created_at, updated_at FROM ai_prompts WHERE module = $1`,
		module,
	).Scan(&prompt.ID, &prompt.Module, &prompt.Content, &prompt.Description, &prompt.CreatedAt, &prompt.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	return prompt, err
}

// GetAllPrompts 获取所有模块的 Prompt 设置
func GetAllPrompts() ([]model.AIPrompt, error) {
	rows, err := database.DB.Query(
		`SELECT id, module, content, description, created_at, updated_at FROM ai_prompts ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []model.AIPrompt
	for rows.Next() {
		var p model.AIPrompt
		if err := rows.Scan(&p.ID, &p.Module, &p.Content, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		prompts = append(prompts, p)
	}
	return prompts, nil
}

// UpdatePrompt 更新 Prompt 内容
func UpdatePrompt(module string, content string) error {
	// Use UPSERT syntax (ON CONFLICT) in case it doesn't exist, though typically it does.
	_, err := database.DB.Exec(
		`INSERT INTO ai_prompts (module, content, description)
		 VALUES ($1, $2, '')
		 ON CONFLICT (module) DO UPDATE SET content = EXCLUDED.content, updated_at = NOW()`,
		module, content,
	)
	return err
}

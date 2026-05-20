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
		`SELECT id, module, content, description, version, is_customized, canonical_hash, created_at, updated_at FROM ai_prompts WHERE module = $1`,
		module,
	).Scan(&prompt.ID, &prompt.Module, &prompt.Content, &prompt.Description, &prompt.Version, &prompt.IsCustomized, &prompt.CanonicalHash, &prompt.CreatedAt, &prompt.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	return prompt, err
}

// GetAllPrompts 获取所有模块的 Prompt 设置
func GetAllPrompts() ([]model.AIPrompt, error) {
	rows, err := database.DB.Query(
		`SELECT id, module, content, description, version, is_customized, canonical_hash, created_at, updated_at FROM ai_prompts ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []model.AIPrompt
	for rows.Next() {
		var p model.AIPrompt
		if err := rows.Scan(&p.ID, &p.Module, &p.Content, &p.Description, &p.Version, &p.IsCustomized, &p.CanonicalHash, &p.CreatedAt, &p.UpdatedAt); err != nil {
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

// InsertCanonical 插入新的 canonical prompt 行（is_customized = false）。
func InsertCanonical(module, version, content, hash, description string) error {
	_, err := database.DB.Exec(
		`INSERT INTO ai_prompts (module, content, description, version, is_customized, canonical_hash)
		 VALUES ($1, $2, $3, $4, false, $5)`,
		module, content, description, version, hash,
	)
	return err
}

// UpdateCanonicalContent 把已存在的（未自定义）行升级到 canonical 新版本。
// 不动 is_customized 字段（调用方应已确认该行 is_customized=false）。
func UpdateCanonicalContent(module, version, content, hash string) error {
	_, err := database.DB.Exec(
		`UPDATE ai_prompts
		 SET content = $2, version = $3, canonical_hash = $4, updated_at = NOW()
		 WHERE module = $1`,
		module, content, version, hash,
	)
	return err
}

// SetCustomized 翻转某模块的 is_customized 标记。
func SetCustomized(module string, customized bool) error {
	_, err := database.DB.Exec(
		`UPDATE ai_prompts SET is_customized = $2, updated_at = NOW() WHERE module = $1`,
		module, customized,
	)
	return err
}

// ResetToCanonical 把模块强制回到 canonical 版本：content/version/hash 全覆盖，is_customized=false。
func ResetToCanonical(module, version, content, hash string) error {
	_, err := database.DB.Exec(
		`UPDATE ai_prompts
		 SET content = $2, version = $3, canonical_hash = $4, is_customized = false, updated_at = NOW()
		 WHERE module = $1`,
		module, content, version, hash,
	)
	return err
}

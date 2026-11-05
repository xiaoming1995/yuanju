package prompt

import (
	"database/sql"
	"log"

	"yuanju/internal/model"
	"yuanju/internal/repository"
)

// syncStore is the repository surface SyncCanonical needs.
// It is satisfied by realStore (production) and fakeStore (tests).
type syncStore interface {
	GetPromptByModule(module string) (*model.AIPrompt, error)
	InsertCanonical(module, version, content, hash, description string) error
	UpdateCanonicalContent(module, version, content, hash, expectedHash string) error
}

// realStore wires syncStore to the repository package functions.
type realStore struct{}

func (realStore) GetPromptByModule(m string) (*model.AIPrompt, error) {
	return repository.GetPromptByModule(m)
}

func (realStore) InsertCanonical(m, v, c, h, d string) error {
	return repository.InsertCanonical(m, v, c, h, d)
}

func (realStore) UpdateCanonicalContent(m, v, c, h, expectedHash string) error {
	return repository.UpdateCanonicalContent(m, v, c, h, expectedHash)
}

// SyncCanonical aligns DB rows to the in-memory Canonical registry.
// Called on startup with database.DB; the db parameter is kept for API
// symmetry and future use (e.g. passing a transaction).
//
// Decision per module:
//
//	DB 无该模块 → InsertCanonical（补种初始值）
//	DB 已存在   → noop；仅 today_question 的已知旧出厂版且正文未漂移时升级
//
// Errors are logged per-module; a failing module never blocks startup.
func SyncCanonical(db *sql.DB) error {
	return syncCanonicalWith(realStore{})
}

// syncCanonicalWith is the testable core of SyncCanonical.
func syncCanonicalWith(store syncStore) error {
	for module, def := range canonical {
		row, err := store.GetPromptByModule(module)
		if err != nil {
			log.Printf("[prompt-sync] module=%s action=skip reason=db_error err=%v", module, err)
			continue
		}

		switch {
		case row == nil:
			if err := store.InsertCanonical(module, def.Version, def.Content, def.Hash, def.Description); err != nil {
				log.Printf("[prompt-sync] module=%s action=skip reason=insert_error err=%v", module, err)
				continue
			}
			log.Printf("[prompt-sync] module=%s action=insert version=%s hash=%s", module, def.Version, shortHash(def.Hash))

		case module == "today_question" && !row.IsCustomized &&
			(row.Version == "v1-evidence-only" || row.Version == "v2-insufficient-ai-advice") &&
			row.CanonicalHash != "" && HashContent(row.Content) == row.CanonicalHash:
			if err := store.UpdateCanonicalContent(module, def.Version, def.Content, def.Hash, row.CanonicalHash); err != nil {
				log.Printf("[prompt-sync] module=%s action=skip reason=update_error err=%v", module, err)
				continue
			}
			log.Printf("[prompt-sync] module=%s action=upgrade version=%s", module, def.Version)

		default:
			// 管理员维护版及其他模块继续依赖后台手动「采用出厂新版」。
			log.Printf("[prompt-sync] module=%s action=noop reason=row_exists version=%s", module, row.Version)
		}
	}
	return nil
}

func shortHash(h string) string {
	if len(h) < 8 {
		return h
	}
	return h[:8] + "..."
}

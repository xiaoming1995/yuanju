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
	UpdateCanonicalContent(module, version, content, hash string) error
}

// realStore wires syncStore to the repository package functions.
type realStore struct{}

func (realStore) GetPromptByModule(m string) (*model.AIPrompt, error) {
	return repository.GetPromptByModule(m)
}

func (realStore) InsertCanonical(m, v, c, h, d string) error {
	return repository.InsertCanonical(m, v, c, h, d)
}

func (realStore) UpdateCanonicalContent(m, v, c, h string) error {
	return repository.UpdateCanonicalContent(m, v, c, h)
}

// SyncCanonical aligns DB rows to the in-memory Canonical registry.
// Called on startup with database.DB; the db parameter is kept for API
// symmetry and future use (e.g. passing a transaction).
//
// Decision per module:
//
//	DB 无该模块 → InsertCanonical（补种初始值）
//	DB 已存在   → noop（维护版神圣，永不覆盖；升级靠后台手动「采用」）
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

		default:
			// 维护版神圣：已存在的行永不被出厂版自动覆盖。
			// 升级靠后台手动「采用出厂新版」；漂移由 DriftStatus 在读取时暴露。
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

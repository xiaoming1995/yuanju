# pkg/prompt — Canonical AI prompt registry

This package centralizes AI prompt templates as code-side canonical definitions, and provides a startup sync mechanism that aligns the DB `ai_prompts` table to the registry while preserving admin customizations.

## Why this exists

Prompts evolve with code. When a new code path expects the AI to emit `question_focus` / `decision_advice` (structured output), the prompt that asks for those fields must be updated *in sync*. Previously the DB-stored prompt silently shadowed the code's intended version, causing schema mismatches after deploy. This registry makes the code the source of truth while still letting admins customize via UI.

## How it works

```
init()
  └─ pkg/prompt/canonical_<module>.go
       └─ Register(module, Definition{ Version, Description, Content })

main.go startup
  ├─ database.Migrate(ModeStartup)
  └─ prompt.SyncCanonical(database.DB)
       └─ for module, def := range canonical:
              switch DB row state:
                 nil                  → INSERT canonical (is_customized=false)
                 is_customized=true   → skip (admin owns it)
                 version == def.Ver   → noop (already aligned)
                 else                 → UPDATE to canonical
```

## Adding a new module

1. Create `pkg/prompt/canonical_<module>.go`:

```go
package prompt

func init() {
    Register("<module>", Definition{
        Version:     "v1-initial",
        Description: "What this prompt is for",
        Content:     myModuleCanonicalContent,
    })
}

const myModuleCanonicalContent = `... Go template syntax with {{.SomeField}} placeholders ...`
```

2. Wire the service to use the prompt — typically the existing service already calls `repository.GetPromptByModule(module)`; the DB row will be populated by SyncCanonical on next startup.

3. Add a test in `canonical_test.go` (or a sibling test file) asserting:
   - `MustGet("<module>")` returns the registered Definition
   - Content contains expected template variables
   - Content does NOT contain forbidden keywords (if a validator like `ValidateYearNarrative` exists)

## Bumping the version

When you change the prompt body of an existing module:

1. Edit `Content` in the corresponding `canonical_<module>.go`.
2. **Bump the `Version` string** — use a descriptive identifier like `"v2-question-aware"` or `"2026-05-20-decision-first"`. Don't reuse old version strings.
3. On the next deploy, SyncCanonical will detect the version mismatch and upgrade unedited DB rows. Admin-customized rows are skipped (their `is_customized=true` protects them). If admin wants to pick up the new version, they click "重置为系统默认" in the admin UI.

## API surface

- `Register(module string, def Definition)` — call from `init()` only. Panics on duplicate module.
- `MustGet(module string) Definition` — panics if module unregistered. Use when caller knows the module exists.
- `Lookup(module string) (Definition, bool)` — safe variant. Use in handlers that might receive arbitrary input.
- `Has(module string) bool` — convenience.
- `SyncCanonical(db *sql.DB) error` — called once at startup from `main.go`. Always returns nil (failures are logged per-module).

## Customization protection invariants

The system guarantees:

1. A row with `is_customized=true` is **never** overwritten by SyncCanonical.
2. An admin PUT to `/api/admin/prompts/:module` automatically sets `is_customized=true`.
3. An admin POST to `/api/admin/prompts/:module/reset` restores canonical content + clears the flag.
4. PUT/reset to an unregistered module returns 404.

## Migration history note

Migration `00011_ai_prompt_versioning.sql` added the `version`, `is_customized`, `canonical_hash` columns and marked **all existing rows as `is_customized=true`** as a conservative safety net — no row's content was changed during the migration. New modules added to the canonical registry after migration get their row INSERTed on next startup with `is_customized=false`.

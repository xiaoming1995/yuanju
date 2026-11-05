package prompt

import (
	"errors"
	"testing"

	"yuanju/internal/model"
)

// fakeStore implements syncStore entirely in memory.
type fakeStore struct {
	rows      map[string]*model.AIPrompt
	inserts   []string
	updates   []string
	getErr    map[string]error // module → error to return from GetPromptByModule
	insertErr map[string]error // module → error to return from InsertCanonical
	updateErr map[string]error // module → error to return from UpdateCanonicalContent
	// dbErr, if non-nil, is returned by GetPromptByModule for all modules.
	dbErr error
}

func newFakeStore() *fakeStore {
	today := MustGet("today_question")
	return &fakeStore{
		rows:      map[string]*model.AIPrompt{"today_question": {Module: "today_question", Version: today.Version, Content: today.Content, CanonicalHash: today.Hash}},
		getErr:    map[string]error{},
		insertErr: map[string]error{},
		updateErr: map[string]error{},
	}
}

func TestSyncCanonical_InsertsTodayQuestion(t *testing.T) {
	store := newFakeStore()
	delete(store.rows, "today_question")
	store.rows["compatibility"] = &model.AIPrompt{Module: "compatibility"}
	if err := syncCanonicalWith(store); err != nil {
		t.Fatal(err)
	}
	if len(store.inserts) != 1 || store.inserts[0] != "today_question" {
		t.Fatalf("today question prompt missing: %v", store.inserts)
	}
}

func (f *fakeStore) GetPromptByModule(m string) (*model.AIPrompt, error) {
	if err := f.getErr[m]; err != nil {
		return nil, err
	}
	if f.dbErr != nil {
		return nil, f.dbErr
	}
	return f.rows[m], nil
}

func (f *fakeStore) InsertCanonical(m, v, c, h, d string) error {
	if err := f.insertErr[m]; err != nil {
		return err
	}
	f.inserts = append(f.inserts, m)
	f.rows[m] = &model.AIPrompt{Module: m, Version: v, Content: c, CanonicalHash: h, Description: d, IsCustomized: false}
	return nil
}

func (f *fakeStore) UpdateCanonicalContent(m, v, c, h, expectedHash string) error {
	if err := f.updateErr[m]; err != nil {
		return err
	}
	if f.rows[m] == nil || f.rows[m].IsCustomized || f.rows[m].CanonicalHash != expectedHash {
		return errors.New("canonical prompt changed before update")
	}
	f.updates = append(f.updates, m)
	old := f.rows[m]
	// Replace with a fresh struct copy so the caller's `row` pointer (returned from
	// GetPromptByModule earlier) keeps its old field values, matching real DB semantics.
	f.rows[m] = &model.AIPrompt{
		ID:            old.ID,
		Module:        old.Module,
		Content:       c,
		Description:   old.Description,
		Version:       v,
		IsCustomized:  old.IsCustomized,
		CanonicalHash: h,
		CreatedAt:     old.CreatedAt,
		UpdatedAt:     old.UpdatedAt,
	}
	return nil
}

func TestSyncCanonicalUpgradesOnlyUntouchedTodayQuestionPrompt(t *testing.T) {
	store := newFakeStore()
	oldContent := "old factory today question prompt"
	store.rows["today_question"] = &model.AIPrompt{Module: "today_question", Version: "v1-evidence-only", Content: oldContent, CanonicalHash: HashContent(oldContent)}
	if err := syncCanonicalWith(store); err != nil {
		t.Fatal(err)
	}
	if len(store.updates) != 1 || store.updates[0] != "today_question" || store.rows["today_question"].Version != MustGet("today_question").Version {
		t.Fatalf("factory prompt not upgraded: updates=%v row=%+v", store.updates, store.rows["today_question"])
	}
	store.updates = nil
	store.rows["today_question"] = &model.AIPrompt{Module: "today_question", Version: "v2-insufficient-ai-advice", Content: "admin edited", CanonicalHash: HashContent("old factory"), IsCustomized: false}
	if err := syncCanonicalWith(store); err != nil || len(store.updates) != 0 {
		t.Fatalf("drifted prompt overwritten: updates=%v err=%v", store.updates, err)
	}
	store.rows["today_question"] = &model.AIPrompt{Module: "today_question", Version: "v1-evidence-only", Content: oldContent, CanonicalHash: HashContent(oldContent), IsCustomized: true}
	if err := syncCanonicalWith(store); err != nil || len(store.updates) != 0 {
		t.Fatalf("customized prompt overwritten: updates=%v err=%v", store.updates, err)
	}
}

// TestSyncCanonical_InsertsMissingModule verifies that a module absent from
// the DB is inserted with the canonical content.
func TestSyncCanonical_InsertsMissingModule(t *testing.T) {
	store := newFakeStore() // empty — no rows
	if err := syncCanonicalWith(store); err != nil {
		t.Fatal(err)
	}
	if len(store.inserts) != 1 || store.inserts[0] != "compatibility" {
		t.Fatalf("expected insert of compatibility, got %v", store.inserts)
	}
	if len(store.updates) != 0 {
		t.Errorf("expected no updates, got %v", store.updates)
	}
	// Row should now exist with canonical version
	canonical := MustGet("compatibility")
	if store.rows["compatibility"].Version != canonical.Version {
		t.Errorf("inserted row version=%s, want %s", store.rows["compatibility"].Version, canonical.Version)
	}
}

// TestSyncCanonical_SkipsCustomizedRow verifies that a row flagged
// is_customized=true is never overwritten.
func TestSyncCanonical_SkipsCustomizedRow(t *testing.T) {
	store := newFakeStore()
	store.rows["compatibility"] = &model.AIPrompt{
		Module:       "compatibility",
		Version:      "old",
		IsCustomized: true,
		Content:      "admin edited",
	}
	if err := syncCanonicalWith(store); err != nil {
		t.Fatal(err)
	}
	if len(store.inserts) != 0 || len(store.updates) != 0 {
		t.Errorf("customized row should not be touched; inserts=%v updates=%v", store.inserts, store.updates)
	}
	if store.rows["compatibility"].Version != "old" {
		t.Error("customized row version mutated unexpectedly")
	}
	if store.rows["compatibility"].Content != "admin edited" {
		t.Error("customized row content mutated unexpectedly")
	}
}

// TestSyncCanonical_DoesNotOverwriteExistingRow verifies the new model:
// any existing row (even a non-customized, version-mismatched one) is left
// untouched by startup sync. Upgrades happen only via admin "adopt factory".
func TestSyncCanonical_DoesNotOverwriteExistingRow(t *testing.T) {
	store := newFakeStore()
	store.rows["compatibility"] = &model.AIPrompt{
		Module:       "compatibility",
		Version:      "v2-old",
		IsCustomized: false,
		Content:      "old content",
	}
	if err := syncCanonicalWith(store); err != nil {
		t.Fatal(err)
	}
	if len(store.updates) != 0 {
		t.Errorf("existing row must never be overwritten, got updates=%v", store.updates)
	}
	if len(store.inserts) != 0 {
		t.Errorf("existing row must not be re-inserted, got inserts=%v", store.inserts)
	}
	if store.rows["compatibility"].Version != "v2-old" {
		t.Errorf("existing row version must stay v2-old, got %s", store.rows["compatibility"].Version)
	}
	if store.rows["compatibility"].Content != "old content" {
		t.Errorf("existing row content must stay unchanged, got %s", store.rows["compatibility"].Content)
	}
}

// TestSyncCanonical_NoOpOnAlignedRow verifies that a non-customized row
// already at the canonical version is left untouched.
func TestSyncCanonical_NoOpOnAlignedRow(t *testing.T) {
	store := newFakeStore()
	canonical := MustGet("compatibility")
	store.rows["compatibility"] = &model.AIPrompt{
		Module:        "compatibility",
		Version:       canonical.Version,
		IsCustomized:  false,
		Content:       canonical.Content,
		CanonicalHash: canonical.Hash,
	}
	if err := syncCanonicalWith(store); err != nil {
		t.Fatal(err)
	}
	if len(store.inserts) != 0 || len(store.updates) != 0 {
		t.Errorf("aligned row should be noop; inserts=%v updates=%v", store.inserts, store.updates)
	}
}

// TestSyncCanonical_DBErrorOnGetSkipsModuleNoPanic verifies that a GetPromptByModule
// error does not panic or return error; module is skipped and processing continues.
func TestSyncCanonical_DBErrorOnGetSkipsModuleNoPanic(t *testing.T) {
	store := newFakeStore()
	store.getErr["compatibility"] = errors.New("simulated db down")
	if err := syncCanonicalWith(store); err != nil {
		t.Fatalf("SyncCanonical should not return error even on DB failure, got: %v", err)
	}
	if len(store.inserts) != 0 {
		t.Errorf("expected no inserts when Get fails, got %v", store.inserts)
	}
	// No panic + returns nil = pass
}

// TestSyncCanonical_DBErrorOnInsertSkipsModuleNoPanic verifies that an InsertCanonical
// error does not panic or return error; module is skipped and processing continues.
func TestSyncCanonical_DBErrorOnInsertSkipsModuleNoPanic(t *testing.T) {
	store := newFakeStore()
	// Row is missing, so SyncCanonical tries InsertCanonical
	store.insertErr["compatibility"] = errors.New("simulated insert failure")
	if err := syncCanonicalWith(store); err != nil {
		t.Fatalf("SyncCanonical should not return error even on insert failure, got: %v", err)
	}
	// The error is logged and skipped, inserts list should remain empty
	if len(store.inserts) != 0 {
		t.Errorf("expected no inserts when insert fails, got %v", store.inserts)
	}
}

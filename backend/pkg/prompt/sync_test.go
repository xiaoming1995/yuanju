package prompt

import (
	"testing"

	"yuanju/internal/model"
)

// fakeStore implements syncStore entirely in memory.
type fakeStore struct {
	rows    map[string]*model.AIPrompt
	inserts []string
	updates []string
	// dbErr, if non-nil, is returned by GetPromptByModule for all modules.
	dbErr error
}

func newFakeStore() *fakeStore {
	return &fakeStore{rows: map[string]*model.AIPrompt{}}
}

func (f *fakeStore) GetPromptByModule(m string) (*model.AIPrompt, error) {
	if f.dbErr != nil {
		return nil, f.dbErr
	}
	return f.rows[m], nil
}

func (f *fakeStore) InsertCanonical(m, v, c, h, d string) error {
	f.inserts = append(f.inserts, m)
	f.rows[m] = &model.AIPrompt{Module: m, Version: v, Content: c, CanonicalHash: h, Description: d, IsCustomized: false}
	return nil
}

func (f *fakeStore) UpdateCanonicalContent(m, v, c, h string) error {
	f.updates = append(f.updates, m)
	row := f.rows[m]
	row.Version = v
	row.Content = c
	row.CanonicalHash = h
	return nil
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

// TestSyncCanonical_UpgradesStaleAlignedRow verifies that a non-customized
// row with an outdated version is upgraded to the canonical version.
func TestSyncCanonical_UpgradesStaleAlignedRow(t *testing.T) {
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
	if len(store.updates) != 1 || store.updates[0] != "compatibility" {
		t.Fatalf("expected 1 update for compatibility, got %v", store.updates)
	}
	if len(store.inserts) != 0 {
		t.Errorf("expected no inserts, got %v", store.inserts)
	}
	canonical := MustGet("compatibility")
	if store.rows["compatibility"].Version != canonical.Version {
		t.Errorf("expected version upgraded to %s, got %s", canonical.Version, store.rows["compatibility"].Version)
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

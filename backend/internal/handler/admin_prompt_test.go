package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"yuanju/internal/repository"
	"yuanju/pkg/prompt"
)

// These tests require the live database (the repository layer reads database.DB).
// They are integration tests in spirit — skip if DB is unavailable.

func TestAdminSavePrompt_SetsCustomizedTrue(t *testing.T) {
	if !canConnectDB(t) {
		t.Skip("DB not available")
	}
	module := "compatibility"

	// reset state: ensure is_customized=false at start
	def := prompt.MustGet(module)
	if err := repository.ResetToCanonical(module, def.Version, def.Content, def.Hash); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// PUT new content
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/prompts/:module", UpdatePrompt)

	body, _ := json.Marshal(map[string]string{"content": def.Content + "\n# admin tweak"})
	req := httptest.NewRequest(http.MethodPut, "/prompts/"+module, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// verify is_customized flipped
	got, err := repository.GetPromptByModule(module)
	if err != nil || got == nil {
		t.Fatalf("get failed: %v", err)
	}
	if !got.IsCustomized {
		t.Errorf("expected is_customized=true after admin save, got false")
	}
}

func TestResetPromptToCanonical_RewritesAndUnflags(t *testing.T) {
	if !canConnectDB(t) {
		t.Skip("DB not available")
	}
	module := "compatibility"

	// arrange: ensure row exists in customized state with different content
	def := prompt.MustGet(module)
	if err := repository.UpdatePrompt(module, "admin tweaked content "+def.Version); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	_ = repository.SetCustomized(module, true)

	// act: hit reset endpoint
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/prompts/:module/reset", ResetPromptToCanonical)

	req := httptest.NewRequest(http.MethodPost, "/prompts/"+module+"/reset", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// assert: content + version + customized restored
	got, err := repository.GetPromptByModule(module)
	if err != nil || got == nil {
		t.Fatalf("get after reset failed: %v", err)
	}
	if got.Content != def.Content {
		t.Errorf("reset did not restore canonical content")
	}
	if got.Version != def.Version {
		t.Errorf("reset did not restore canonical version: got %q want %q", got.Version, def.Version)
	}
	if got.IsCustomized {
		t.Errorf("reset did not clear is_customized")
	}
	if got.CanonicalHash != def.Hash {
		t.Errorf("reset did not restore hash")
	}
}

func TestResetPromptToCanonical_UnknownModule_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/prompts/:module/reset", ResetPromptToCanonical)

	req := httptest.NewRequest(http.MethodPost, "/prompts/not-a-real-module/reset", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown module, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestAdminSavePrompt_UnknownModule_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/prompts/:module", UpdatePrompt)

	body := strings.NewReader(`{"content":"x"}`)
	req := httptest.NewRequest(http.MethodPut, "/prompts/not-a-real-module", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown module, got %d body=%s", w.Code, w.Body.String())
	}
}

func canConnectDB(t *testing.T) bool {
	t.Helper()
	// Catch nil-pointer panics that occur when database.DB is not initialized.
	ok := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				ok = false
			}
		}()
		if _, err := repository.GetPromptByModule("compatibility"); err != nil {
			ok = false
		}
	}()
	return ok
}

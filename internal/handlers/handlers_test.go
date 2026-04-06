package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, 200, map[string]string{"hello": "world"})

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", ct)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["hello"] != "world" {
		t.Errorf("expected hello=world, got %v", resp)
	}
}

func TestGenerateToken(t *testing.T) {
	token := generateToken()
	if len(token) != 64 {
		t.Errorf("expected 64 char hex token, got length %d", len(token))
	}
	// tokens should be unique
	token2 := generateToken()
	if token == token2 {
		t.Error("expected unique tokens")
	}
}

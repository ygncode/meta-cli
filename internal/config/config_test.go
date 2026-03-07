package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ygncode/meta-cli/internal/config"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.DefaultAccount != "default" {
		t.Errorf("expected default account, got %s", cfg.DefaultAccount)
	}
	if cfg.GraphAPIVersion != "v25.0" {
		t.Errorf("expected v25.0, got %s", cfg.GraphAPIVersion)
	}
	if cfg.WebhookPort != 8080 {
		t.Errorf("expected 8080, got %d", cfg.WebhookPort)
	}
}

func TestLoadMissing(t *testing.T) {
	// Set HOME to a temp dir with no config file
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Should return defaults
	if cfg.DefaultAccount != "default" {
		t.Errorf("expected defaults when no file, got %+v", cfg)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := config.Default()
	cfg.DefaultPage = "123456"
	cfg.RAGDir = "/docs"

	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmp, ".meta-cli", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if saved.DefaultPage != "123456" {
		t.Errorf("expected 123456, got %s", saved.DefaultPage)
	}

	// Load it back
	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.DefaultPage != "123456" {
		t.Errorf("expected 123456, got %s", loaded.DefaultPage)
	}
	if loaded.RAGDir != "/docs" {
		t.Errorf("expected /docs, got %s", loaded.RAGDir)
	}
}

func TestDefaultDebounceSeconds(t *testing.T) {
	cfg := config.Default()
	if cfg.DebounceSeconds != 3 {
		t.Errorf("expected default DebounceSeconds 3, got %d", cfg.DebounceSeconds)
	}
}

func TestSaveAndLoadAutoReply(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := config.Default()
	cfg.AutoReply = true
	cfg.HooksEndpoint = "http://127.0.0.1:18789/hooks/agent"
	cfg.HooksToken = "test_token"
	cfg.DebounceSeconds = 5
	cfg.PromptTemplate = "Hello {{.PSID}}"

	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !loaded.AutoReply {
		t.Error("expected AutoReply true")
	}
	if loaded.HooksEndpoint != "http://127.0.0.1:18789/hooks/agent" {
		t.Errorf("expected hooks endpoint, got %s", loaded.HooksEndpoint)
	}
	if loaded.HooksToken != "test_token" {
		t.Errorf("expected hooks token, got %s", loaded.HooksToken)
	}
	if loaded.DebounceSeconds != 5 {
		t.Errorf("expected debounce 5, got %d", loaded.DebounceSeconds)
	}
	if loaded.PromptTemplate != "Hello {{.PSID}}" {
		t.Errorf("expected prompt template, got %s", loaded.PromptTemplate)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".meta-cli")
	os.MkdirAll(dir, 0o700)
	os.WriteFile(filepath.Join(dir, "config.json"), []byte("{invalid json}"), 0o600)

	_, err := config.Load()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

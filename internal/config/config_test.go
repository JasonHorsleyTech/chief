package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Worktree.Setup != "" {
		t.Errorf("expected empty setup, got %q", cfg.Worktree.Setup)
	}
	if cfg.OnComplete.Push {
		t.Error("expected Push to be false")
	}
	if cfg.OnComplete.CreatePR {
		t.Error("expected CreatePR to be false")
	}
	if cfg.PromptsDir != "" {
		t.Errorf("expected empty PromptsDir, got %q", cfg.PromptsDir)
	}
	if cfg.RetryOnRateLimit {
		t.Error("expected RetryOnRateLimit to be false")
	}
	if cfg.RetryIntervalMinutes != 60 {
		t.Errorf("expected RetryIntervalMinutes to be 60, got %d", cfg.RetryIntervalMinutes)
	}
	if cfg.MaxRateLimitRetries != 3 {
		t.Errorf("expected MaxRateLimitRetries to be 3, got %d", cfg.MaxRateLimitRetries)
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Worktree.Setup != "" {
		t.Errorf("expected empty setup, got %q", cfg.Worktree.Setup)
	}
	if cfg.RetryIntervalMinutes != 60 {
		t.Errorf("expected RetryIntervalMinutes to be 60, got %d", cfg.RetryIntervalMinutes)
	}
	if cfg.MaxRateLimitRetries != 3 {
		t.Errorf("expected MaxRateLimitRetries to be 3, got %d", cfg.MaxRateLimitRetries)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	cfg := &Config{
		Worktree: WorktreeConfig{
			Setup: "npm install",
		},
		OnComplete: OnCompleteConfig{
			Push:     true,
			CreatePR: true,
		},
	}

	if err := Save(dir, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Worktree.Setup != "npm install" {
		t.Errorf("expected setup %q, got %q", "npm install", loaded.Worktree.Setup)
	}
	if !loaded.OnComplete.Push {
		t.Error("expected Push to be true")
	}
	if !loaded.OnComplete.CreatePR {
		t.Error("expected CreatePR to be true")
	}
}

func TestSaveAndLoadNewFields(t *testing.T) {
	dir := t.TempDir()

	cfg := &Config{
		PromptsDir:           "/custom/prompts",
		RetryOnRateLimit:     true,
		RetryIntervalMinutes: 30,
		MaxRateLimitRetries:  5,
	}

	if err := Save(dir, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.PromptsDir != "/custom/prompts" {
		t.Errorf("expected PromptsDir %q, got %q", "/custom/prompts", loaded.PromptsDir)
	}
	if !loaded.RetryOnRateLimit {
		t.Error("expected RetryOnRateLimit to be true")
	}
	if loaded.RetryIntervalMinutes != 30 {
		t.Errorf("expected RetryIntervalMinutes 30, got %d", loaded.RetryIntervalMinutes)
	}
	if loaded.MaxRateLimitRetries != 5 {
		t.Errorf("expected MaxRateLimitRetries 5, got %d", loaded.MaxRateLimitRetries)
	}
}

func TestLoadMissingNewFieldsUseDefaults(t *testing.T) {
	dir := t.TempDir()

	// Write a config that only has the old fields
	chiefDir := filepath.Join(dir, ".chief")
	if err := os.MkdirAll(chiefDir, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := "worktree:\n  setup: make install\n"
	if err := os.WriteFile(filepath.Join(chiefDir, "config.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.PromptsDir != "" {
		t.Errorf("expected empty PromptsDir, got %q", loaded.PromptsDir)
	}
	if loaded.RetryOnRateLimit {
		t.Error("expected RetryOnRateLimit to be false")
	}
	if loaded.RetryIntervalMinutes != 60 {
		t.Errorf("expected RetryIntervalMinutes 60, got %d", loaded.RetryIntervalMinutes)
	}
	if loaded.MaxRateLimitRetries != 3 {
		t.Errorf("expected MaxRateLimitRetries 3, got %d", loaded.MaxRateLimitRetries)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()

	if Exists(dir) {
		t.Error("expected Exists to return false for missing config")
	}

	// Create the config
	chiefDir := filepath.Join(dir, ".chief")
	if err := os.MkdirAll(chiefDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(chiefDir, "config.yaml"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !Exists(dir) {
		t.Error("expected Exists to return true for existing config")
	}
}

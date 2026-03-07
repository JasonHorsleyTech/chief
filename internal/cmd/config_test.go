package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/minicodemonkey/chief/internal/config"
)

// captureStdout redirects os.Stdout during fn and returns the captured output.
func captureStdout(fn func() error) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	fnErr := fn()

	w.Close()
	os.Stdout = old

	out, readErr := io.ReadAll(r)
	if readErr != nil {
		return "", readErr
	}

	return string(out), fnErr
}

func TestRunConfigNoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	opts := ConfigOptions{BaseDir: tmpDir}
	output, err := captureStdout(func() error {
		return RunConfig(opts)
	})
	if err != nil {
		t.Fatalf("RunConfig() returned error: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, ".chief", "config.yaml")
	if !strings.Contains(output, "# Config: "+expectedPath) {
		t.Errorf("Output should contain config path comment\ngot:\n%s", output)
	}

	if !strings.Contains(output, "# No config file found. Run 'chief config init' to create one.") {
		t.Errorf("Output should contain missing config note\ngot:\n%s", output)
	}
}

func TestRunConfigWithExistingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a config file
	cfg := config.Default()
	if err := config.Save(tmpDir, cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	opts := ConfigOptions{BaseDir: tmpDir}
	output, err := captureStdout(func() error {
		return RunConfig(opts)
	})
	if err != nil {
		t.Fatalf("RunConfig() returned error: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, ".chief", "config.yaml")
	if !strings.Contains(output, "# Config: "+expectedPath) {
		t.Errorf("Output should contain config path comment\ngot:\n%s", output)
	}

	if strings.Contains(output, "# No config file found") {
		t.Errorf("Output should NOT contain missing file note when config exists\ngot:\n%s", output)
	}
}

func TestRunConfigIncludesDefaultYAMLFields(t *testing.T) {
	tmpDir := t.TempDir()

	opts := ConfigOptions{BaseDir: tmpDir}
	output, err := captureStdout(func() error {
		return RunConfig(opts)
	})
	if err != nil {
		t.Fatalf("RunConfig() returned error: %v", err)
	}

	// Default config should include retry fields
	for _, field := range []string{"retryIntervalMinutes:", "maxRateLimitRetries:", "retryOnRateLimit:"} {
		if !strings.Contains(output, field) {
			t.Errorf("Output should contain YAML field %q\ngot:\n%s", field, output)
		}
	}
}

func TestRunConfigOutputIsValidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a config file with non-default values
	cfg := config.Default()
	cfg.RetryOnRateLimit = true
	cfg.MaxRateLimitRetries = 5
	if err := config.Save(tmpDir, cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	opts := ConfigOptions{BaseDir: tmpDir}
	output, err := captureStdout(func() error {
		return RunConfig(opts)
	})
	if err != nil {
		t.Fatalf("RunConfig() returned error: %v", err)
	}

	// Non-default values should appear in output
	if !strings.Contains(output, "retryOnRateLimit: true") {
		t.Errorf("Output should reflect saved config value\ngot:\n%s", output)
	}
	if !strings.Contains(output, "maxRateLimitRetries: 5") {
		t.Errorf("Output should reflect saved config value\ngot:\n%s", output)
	}
}

func TestRunConfigInitCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()

	opts := ConfigInitOptions{BaseDir: tmpDir}
	output, err := captureStdout(func() error {
		return RunConfigInit(opts)
	})
	if err != nil {
		t.Fatalf("RunConfigInit() returned error: %v", err)
	}

	cfgPath := filepath.Join(tmpDir, ".chief", "config.yaml")
	if !strings.Contains(output, cfgPath) {
		t.Errorf("Output should mention created file path\ngot:\n%s", output)
	}

	if !config.Exists(tmpDir) {
		t.Errorf("Config file should exist after init")
	}
}

func TestRunConfigInitFileContainsAllFields(t *testing.T) {
	tmpDir := t.TempDir()

	if err := RunConfigInit(ConfigInitOptions{BaseDir: tmpDir}); err != nil {
		t.Fatalf("RunConfigInit() returned error: %v", err)
	}

	cfgPath := filepath.Join(tmpDir, ".chief", "config.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	content := string(data)

	for _, field := range []string{"promptsDir", "retryOnRateLimit", "retryIntervalMinutes", "maxRateLimitRetries", "worktree", "onComplete"} {
		if !strings.Contains(content, field) {
			t.Errorf("Generated config should contain field %q\ngot:\n%s", field, content)
		}
	}
}

func TestRunConfigInitFileIsValidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	if err := RunConfigInit(ConfigInitOptions{BaseDir: tmpDir}); err != nil {
		t.Fatalf("RunConfigInit() returned error: %v", err)
	}

	// Load() must parse the generated file without error
	cfg, err := config.Load(tmpDir)
	if err != nil {
		t.Fatalf("config.Load() on generated file returned error: %v", err)
	}

	// Check defaults are preserved
	if cfg.RetryIntervalMinutes != 60 {
		t.Errorf("Expected RetryIntervalMinutes=60, got %d", cfg.RetryIntervalMinutes)
	}
	if cfg.MaxRateLimitRetries != 3 {
		t.Errorf("Expected MaxRateLimitRetries=3, got %d", cfg.MaxRateLimitRetries)
	}
}

func TestRunConfigInitErrorIfFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config first
	if err := RunConfigInit(ConfigInitOptions{BaseDir: tmpDir}); err != nil {
		t.Fatalf("First RunConfigInit() returned error: %v", err)
	}

	// Second call should fail
	err := RunConfigInit(ConfigInitOptions{BaseDir: tmpDir})
	if err == nil {
		t.Errorf("Expected error when config already exists, got nil")
	}
}

func TestRunConfigInitForceOverwrites(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config first
	if err := RunConfigInit(ConfigInitOptions{BaseDir: tmpDir}); err != nil {
		t.Fatalf("First RunConfigInit() returned error: %v", err)
	}

	// --force should succeed
	if err := RunConfigInit(ConfigInitOptions{BaseDir: tmpDir, Force: true}); err != nil {
		t.Errorf("RunConfigInit with Force=true should not error: %v", err)
	}
}

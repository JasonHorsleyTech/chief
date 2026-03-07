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

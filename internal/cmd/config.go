package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/minicodemonkey/chief/internal/config"
	"gopkg.in/yaml.v3"
)

// configInitTemplate is the commented config file written by `chief config init`.
const configInitTemplate = `# Chief configuration file
# Run 'chief config' to view current effective settings.

worktree:
  # Shell command to run when setting up a new worktree (e.g., "npm install")
  setup: ""

onComplete:
  # Push changes to the remote when all stories complete (true/false)
  push: false
  # Create a pull request when all stories complete (true/false)
  createPR: false

# Custom prompts directory path. Overrides the embedded prompts when set.
# Leave empty to use embedded defaults.
promptsDir: ""

# Automatically wait and retry when Claude hits an API rate limit (true/false)
retryOnRateLimit: false

# How long to wait (in minutes) before retrying after a rate limit
retryIntervalMinutes: 60

# Maximum number of rate-limit retries before stopping
maxRateLimitRetries: 3
`

// ConfigInitOptions contains options for the config init command.
type ConfigInitOptions struct {
	BaseDir string // Base directory for .chief/ (default: current directory)
	Force   bool   // Overwrite existing config file if true
}

// RunConfigInit creates a default commented config file at .chief/config.yaml.
// If the file already exists and Force is false, it returns an error.
func RunConfigInit(opts ConfigInitOptions) error {
	if opts.BaseDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		opts.BaseDir = cwd
	}

	cfgPath := filepath.Join(opts.BaseDir, ".chief", "config.yaml")

	if config.Exists(opts.BaseDir) && !opts.Force {
		return fmt.Errorf("config file already exists: %s\nUse --force to overwrite", cfgPath)
	}

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return fmt.Errorf("failed to create .chief directory: %w", err)
	}

	if err := os.WriteFile(cfgPath, []byte(configInitTemplate), 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Config file created: %s\n", cfgPath)
	return nil
}

// ConfigOptions contains configuration for the config command.
type ConfigOptions struct {
	BaseDir string // Base directory for .chief/ (default: current directory)
}

// RunConfig prints the current effective config as YAML to stdout.
// A comment header always shows the config file path.
// If no config file exists, an additional explanatory note is printed.
func RunConfig(opts ConfigOptions) error {
	if opts.BaseDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		opts.BaseDir = cwd
	}

	cfgPath := filepath.Join(opts.BaseDir, ".chief", "config.yaml")
	fmt.Printf("# Config: %s\n", cfgPath)

	if !config.Exists(opts.BaseDir) {
		fmt.Println("# No config file found. Run 'chief config init' to create one.")
	}

	fmt.Println()

	cfg, err := config.Load(opts.BaseDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	fmt.Print(string(data))
	return nil
}

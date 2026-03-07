package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/minicodemonkey/chief/internal/config"
	"gopkg.in/yaml.v3"
)

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

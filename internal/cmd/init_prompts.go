package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	chiefembed "github.com/minicodemonkey/chief/embed"
)

// InitPromptsOptions holds configuration for the init-prompts command.
type InitPromptsOptions struct {
	// Path is the directory to create. Defaults to ~/chief-prompts/ if empty.
	Path string
}

// RunInitPrompts creates a prompts directory pre-populated with all embedded
// default prompt templates, giving users a ready-to-edit starting point.
func RunInitPrompts(opts InitPromptsOptions) error {
	dir := opts.Path
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to determine home directory: %w", err)
		}
		dir = filepath.Join(home, "chief-prompts")
	}

	// Resolve to absolute path for display.
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Create the directory (including any parents).
	if err := os.MkdirAll(absDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", absDir, err)
	}

	// Write all embedded templates into the directory.
	templates := chiefembed.RawTemplates()
	// Write in a deterministic order that matches the acceptance criteria.
	fileOrder := []string{
		"prompt.txt",
		"init_prompt.txt",
		"edit_prompt.txt",
		"convert_prompt.txt",
		"detect_setup_prompt.txt",
	}
	for _, name := range fileOrder {
		content := templates[name]
		dest := filepath.Join(absDir, name)
		if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", name, err)
		}
	}

	fmt.Printf("Prompts directory initialised: %s\n", absDir)
	for _, name := range fileOrder {
		fmt.Printf("  %s\n", name)
	}
	fmt.Printf("Run: chief --prompts-dir %s\n", absDir)
	return nil
}

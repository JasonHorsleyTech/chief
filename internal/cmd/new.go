// Package cmd provides CLI command implementations for Chief.
// This includes new, edit, status, and list commands that can be
// run from the command line without launching the full TUI.
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/minicodemonkey/chief/embed"
	"github.com/minicodemonkey/chief/internal/loop"
	"github.com/minicodemonkey/chief/internal/prd"
)

// NewOptions contains configuration for the new command.
type NewOptions struct {
	Name     string        // PRD name (default: "main")
	Context  string        // Optional context to pass to the agent
	BaseDir  string        // Base directory for .chief/prds/ (default: current directory)
	Provider loop.Provider // Agent CLI provider (Claude or Codex)
	Start    bool          // After creation, launch TUI and start loop
	Auto     bool          // Non-interactive PRD creation (hands-off mode)
}

// RunNew creates a new PRD by launching an agent session.
func RunNew(opts NewOptions) error {
	// Set defaults
	if opts.Name == "" {
		opts.Name = "main"
	}
	if opts.BaseDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		opts.BaseDir = cwd
	}

	// Validate name (alphanumeric, -, _)
	if !isValidPRDName(opts.Name) {
		return fmt.Errorf("invalid PRD name %q: must contain only letters, numbers, hyphens, and underscores", opts.Name)
	}

	// Create directory structure: .chief/prds/<name>/
	prdDir := filepath.Join(opts.BaseDir, ".chief", "prds", opts.Name)
	if err := os.MkdirAll(prdDir, 0755); err != nil {
		return fmt.Errorf("failed to create PRD directory: %w", err)
	}

	// Check if prd.md already exists
	prdMdPath := filepath.Join(prdDir, "prd.md")
	if _, err := os.Stat(prdMdPath); err == nil {
		return fmt.Errorf("PRD already exists at %s. Use 'chief edit %s' to modify it", prdMdPath, opts.Name)
	}

	if opts.Provider == nil {
		return fmt.Errorf("new command requires Provider to be set")
	}

	if opts.Auto {
		// Non-interactive mode: agent makes all decisions
		prompt := embed.GetAutoInitPrompt(prdDir, opts.Context)
		fmt.Printf("Creating PRD in %s (automatic mode)...\n", prdDir)
		fmt.Printf("Using %s to generate PRD...\n", opts.Provider.Name())
		fmt.Println()

		if err := runNonInteractiveAgent(opts.Provider, opts.BaseDir, prompt); err != nil {
			return fmt.Errorf("%s session failed: %w", opts.Provider.Name(), err)
		}
	} else {
		// Interactive mode: agent asks questions
		prompt := embed.GetInitPrompt(prdDir, opts.Context)
		fmt.Printf("Creating PRD in %s...\n", prdDir)
		fmt.Printf("Launching %s to help you create your PRD...\n", opts.Provider.Name())
		fmt.Println()

		if err := runInteractiveAgent(opts.Provider, opts.BaseDir, prompt); err != nil {
			return fmt.Errorf("%s session failed: %w", opts.Provider.Name(), err)
		}
	}

	// Check if prd.md was created
	if _, err := os.Stat(prdMdPath); os.IsNotExist(err) {
		// Clean up empty directory to prevent broken picker entries
		os.Remove(prdDir)
		fmt.Println("\nNo prd.md was created. Run 'chief new' again to try again.")
		return nil
	}

	// Validate the created prd.md can be parsed
	if _, err := prd.ParseMarkdownPRD(prdMdPath); err != nil {
		fmt.Printf("\nWarning: prd.md was created but could not be parsed: %v\n", err)
		fmt.Println("You may need to edit it to match the expected format.")
	} else {
		fmt.Println("\nPRD created successfully!")
	}

	if !opts.Start {
		fmt.Printf("\nYour PRD is ready! Run 'chief' or 'chief %s' to start working on it.\n", opts.Name)
	}
	return nil
}

// runInteractiveAgent launches an interactive agent session in the specified directory.
func runInteractiveAgent(provider loop.Provider, workDir, prompt string) error {
	if provider == nil {
		return fmt.Errorf("interactive agent requires Provider to be set")
	}
	cmd := provider.InteractiveCommand(workDir, prompt)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runNonInteractiveAgent launches a non-interactive agent session for automatic PRD creation.
func runNonInteractiveAgent(provider loop.Provider, workDir, prompt string) error {
	if provider == nil {
		return fmt.Errorf("non-interactive agent requires Provider to be set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := provider.LoopCommand(ctx, prompt, workDir)
	// LoopCommand outputs stream-json which isn't human-readable; discard it.
	// The agent writes prd.md as a file side effect.
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// isValidPRDName checks if the name contains only valid characters.
func isValidPRDName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

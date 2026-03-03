package embed

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetPrompt(t *testing.T) {
	prdPath := "/path/to/prd.json"
	progressPath := "/path/to/progress.md"
	storyContext := `{"id":"US-001","title":"Test Story"}`
	prompt := GetPrompt("", prdPath, progressPath, storyContext, "US-001", "Test Story")

	// Verify all placeholders were substituted
	if strings.Contains(prompt, "{{PRD_PATH}}") {
		t.Error("Expected {{PRD_PATH}} to be substituted")
	}
	if strings.Contains(prompt, "{{PROGRESS_PATH}}") {
		t.Error("Expected {{PROGRESS_PATH}} to be substituted")
	}
	if strings.Contains(prompt, "{{STORY_CONTEXT}}") {
		t.Error("Expected {{STORY_CONTEXT}} to be substituted")
	}
	if strings.Contains(prompt, "{{STORY_ID}}") {
		t.Error("Expected {{STORY_ID}} to be substituted")
	}
	if strings.Contains(prompt, "{{STORY_TITLE}}") {
		t.Error("Expected {{STORY_TITLE}} to be substituted")
	}

	// Verify the commit message contains the exact story ID and title
	if !strings.Contains(prompt, "feat: US-001 - Test Story") {
		t.Error("Expected prompt to contain exact commit message 'feat: US-001 - Test Story'")
	}

	// Verify the PRD path appears in the prompt
	if !strings.Contains(prompt, prdPath) {
		t.Errorf("Expected prompt to contain PRD path %q", prdPath)
	}

	// Verify the progress path appears in the prompt
	if !strings.Contains(prompt, progressPath) {
		t.Errorf("Expected prompt to contain progress path %q", progressPath)
	}

	// Verify the story context is inlined in the prompt
	if !strings.Contains(prompt, storyContext) {
		t.Error("Expected prompt to contain inlined story context")
	}

	// Verify the prompt contains key instructions
	if !strings.Contains(prompt, "chief-complete") {
		t.Error("Expected prompt to contain chief-complete instruction")
	}

	if !strings.Contains(prompt, "ralph-status") {
		t.Error("Expected prompt to contain ralph-status instruction")
	}

	if !strings.Contains(prompt, "passes: true") {
		t.Error("Expected prompt to contain passes: true instruction")
	}
}

func TestGetPrompt_NoFileReadInstruction(t *testing.T) {
	prompt := GetPrompt("", "/path/prd.json", "/path/progress.md", `{"id":"US-001"}`, "US-001", "Test Story")

	// The prompt should NOT instruct Claude to read the PRD file
	if strings.Contains(prompt, "Read the PRD") {
		t.Error("Expected prompt to NOT contain 'Read the PRD' file-read instruction")
	}
}

func TestPromptTemplateNotEmpty(t *testing.T) {
	if promptTemplate == "" {
		t.Error("Expected promptTemplate to be embedded and non-empty")
	}
}

func TestGetPrompt_ChiefExclusion(t *testing.T) {
	prompt := GetPrompt("", "/path/prd.json", "/path/progress.md", `{"id":"US-001"}`, "US-001", "Test Story")

	// The prompt must instruct Claude to never stage or commit .chief/ files
	if !strings.Contains(prompt, ".chief/") {
		t.Error("Expected prompt to contain .chief/ exclusion instruction")
	}
	if !strings.Contains(prompt, "NEVER stage or commit") {
		t.Error("Expected prompt to explicitly say NEVER stage or commit .chief/ files")
	}
	// The commit step should not say "commit ALL changes" anymore
	if strings.Contains(prompt, "commit ALL changes") {
		t.Error("Expected prompt to NOT say 'commit ALL changes' — it should exclude .chief/ files")
	}
}

func TestGetConvertPrompt(t *testing.T) {
	prdFilePath := "/path/to/prds/main/prd.md"
	prompt := GetConvertPrompt("", prdFilePath, "US")

	// Verify the prompt is not empty
	if prompt == "" {
		t.Error("Expected GetConvertPrompt() to return non-empty prompt")
	}

	// Verify file path is substituted (not inlined content)
	if !strings.Contains(prompt, prdFilePath) {
		t.Error("Expected prompt to contain the PRD file path")
	}
	if strings.Contains(prompt, "{{PRD_FILE_PATH}}") {
		t.Error("Expected {{PRD_FILE_PATH}} to be substituted")
	}

	// Verify the old {{PRD_CONTENT}} placeholder is completely removed
	if strings.Contains(prompt, "{{PRD_CONTENT}}") {
		t.Error("Expected {{PRD_CONTENT}} placeholder to be completely removed")
	}

	// Verify ID prefix is substituted
	if strings.Contains(prompt, "{{ID_PREFIX}}") {
		t.Error("Expected {{ID_PREFIX}} to be substituted")
	}
	if !strings.Contains(prompt, "US-001") {
		t.Error("Expected prompt to contain US-001 when prefix is US")
	}

	// Verify key instructions are present
	if !strings.Contains(prompt, "JSON") {
		t.Error("Expected prompt to mention JSON")
	}

	if !strings.Contains(prompt, "userStories") {
		t.Error("Expected prompt to describe userStories structure")
	}

	if !strings.Contains(prompt, `"passes": false`) {
		t.Error("Expected prompt to specify passes: false default")
	}

	// Verify prompt instructs Claude to read the file
	if !strings.Contains(prompt, "Read the PRD file") {
		t.Error("Expected prompt to instruct Claude to read the PRD file")
	}
}

func TestGetConvertPrompt_CustomPrefix(t *testing.T) {
	prompt := GetConvertPrompt("", "/path/prd.md", "MFR")

	// Verify custom prefix is used, not hardcoded US
	if strings.Contains(prompt, "{{ID_PREFIX}}") {
		t.Error("Expected {{ID_PREFIX}} to be substituted")
	}
	if !strings.Contains(prompt, "MFR-001") {
		t.Error("Expected prompt to contain MFR-001 when prefix is MFR")
	}
	if !strings.Contains(prompt, "MFR-002") {
		t.Error("Expected prompt to contain MFR-002 when prefix is MFR")
	}
}

func TestGetInitPrompt(t *testing.T) {
	prdDir := "/path/to/.chief/prds/main"

	// Test with no context
	prompt := GetInitPrompt("", prdDir, "")
	if !strings.Contains(prompt, "No additional context provided") {
		t.Error("Expected default context message")
	}

	// Verify PRD directory is substituted
	if !strings.Contains(prompt, prdDir) {
		t.Errorf("Expected prompt to contain PRD directory %q", prdDir)
	}
	if strings.Contains(prompt, "{{PRD_DIR}}") {
		t.Error("Expected {{PRD_DIR}} to be substituted")
	}

	// Test with context
	context := "Build a todo app"
	promptWithContext := GetInitPrompt("", prdDir, context)
	if !strings.Contains(promptWithContext, context) {
		t.Error("Expected context to be substituted in prompt")
	}
}

func TestGetEditPrompt(t *testing.T) {
	prompt := GetEditPrompt("", "/test/path/prds/main")
	if prompt == "" {
		t.Error("Expected GetEditPrompt() to return non-empty prompt")
	}
	if !strings.Contains(prompt, "/test/path/prds/main") {
		t.Error("Expected prompt to contain the PRD directory path")
	}
}

// --- Override directory tests ---

func TestGetPrompt_EmptyDir_UsesEmbedded(t *testing.T) {
	prompt := GetPrompt("", "/path/prd.json", "/path/progress.md", `{}`, "US-001", "Title")
	if prompt == "" {
		t.Error("Expected non-empty prompt when promptsDir is empty")
	}
}

func TestGetPrompt_WithOverrideFile(t *testing.T) {
	dir := t.TempDir()
	overrideContent := "OVERRIDE prompt {{PRD_PATH}} {{STORY_ID}} {{STORY_TITLE}} {{PROGRESS_PATH}} {{STORY_CONTEXT}}"
	if err := os.WriteFile(filepath.Join(dir, "prompt.txt"), []byte(overrideContent), 0600); err != nil {
		t.Fatal(err)
	}

	prompt := GetPrompt(dir, "/my/prd.json", "/my/progress.md", `{}`, "US-999", "My Story")
	if !strings.Contains(prompt, "OVERRIDE prompt") {
		t.Error("Expected override file content to be used")
	}
	if !strings.Contains(prompt, "/my/prd.json") {
		t.Error("Expected template substitution to run on override content")
	}
	if !strings.Contains(prompt, "US-999") {
		t.Error("Expected STORY_ID substitution in override content")
	}
}

func TestGetPrompt_DirMissingFile_UsesEmbedded(t *testing.T) {
	dir := t.TempDir()
	// No prompt.txt in dir — should fall back to embedded.
	prompt := GetPrompt(dir, "/path/prd.json", "/path/progress.md", `{}`, "US-001", "Title")
	if prompt == "" {
		t.Error("Expected non-empty embedded prompt when override file is absent")
	}
	// Embedded prompt contains known content.
	if !strings.Contains(prompt, "chief-complete") {
		t.Error("Expected embedded prompt to be used when override file is absent")
	}
}

func TestGetInitPrompt_WithOverrideFile(t *testing.T) {
	dir := t.TempDir()
	overrideContent := "CUSTOM INIT {{PRD_DIR}} {{CONTEXT}}"
	if err := os.WriteFile(filepath.Join(dir, "init_prompt.txt"), []byte(overrideContent), 0600); err != nil {
		t.Fatal(err)
	}

	prompt := GetInitPrompt(dir, "/some/prd/dir", "build a todo app")
	if !strings.Contains(prompt, "CUSTOM INIT") {
		t.Error("Expected override init_prompt.txt to be used")
	}
	if !strings.Contains(prompt, "/some/prd/dir") {
		t.Error("Expected PRD_DIR substitution")
	}
}

func TestGetEditPrompt_WithOverrideFile(t *testing.T) {
	dir := t.TempDir()
	overrideContent := "CUSTOM EDIT {{PRD_DIR}}"
	if err := os.WriteFile(filepath.Join(dir, "edit_prompt.txt"), []byte(overrideContent), 0600); err != nil {
		t.Fatal(err)
	}

	prompt := GetEditPrompt(dir, "/edit/prd/dir")
	if !strings.Contains(prompt, "CUSTOM EDIT") {
		t.Error("Expected override edit_prompt.txt to be used")
	}
	if !strings.Contains(prompt, "/edit/prd/dir") {
		t.Error("Expected PRD_DIR substitution in override")
	}
}

func TestGetConvertPrompt_WithOverrideFile(t *testing.T) {
	dir := t.TempDir()
	overrideContent := "CUSTOM CONVERT {{PRD_FILE_PATH}} {{ID_PREFIX}}-001"
	if err := os.WriteFile(filepath.Join(dir, "convert_prompt.txt"), []byte(overrideContent), 0600); err != nil {
		t.Fatal(err)
	}

	prompt := GetConvertPrompt(dir, "/path/prd.md", "TS")
	if !strings.Contains(prompt, "CUSTOM CONVERT") {
		t.Error("Expected override convert_prompt.txt to be used")
	}
	if !strings.Contains(prompt, "/path/prd.md") {
		t.Error("Expected PRD_FILE_PATH substitution in override")
	}
	if !strings.Contains(prompt, "TS-001") {
		t.Error("Expected ID_PREFIX substitution in override")
	}
}

func TestGetDetectSetupPrompt_EmptyDir_UsesEmbedded(t *testing.T) {
	prompt := GetDetectSetupPrompt("")
	if prompt == "" {
		t.Error("Expected non-empty detect setup prompt")
	}
}

func TestGetDetectSetupPrompt_WithOverrideFile(t *testing.T) {
	dir := t.TempDir()
	overrideContent := "CUSTOM DETECT SETUP PROMPT"
	if err := os.WriteFile(filepath.Join(dir, "detect_setup_prompt.txt"), []byte(overrideContent), 0600); err != nil {
		t.Fatal(err)
	}

	prompt := GetDetectSetupPrompt(dir)
	if prompt != overrideContent {
		t.Errorf("Expected override content %q, got %q", overrideContent, prompt)
	}
}

func TestGetDetectSetupPrompt_DirMissingFile_UsesEmbedded(t *testing.T) {
	dir := t.TempDir()
	// No detect_setup_prompt.txt — should fall back to embedded.
	embedded := GetDetectSetupPrompt("")
	prompt := GetDetectSetupPrompt(dir)
	if prompt != embedded {
		t.Error("Expected embedded detect setup prompt when override file is absent")
	}
}

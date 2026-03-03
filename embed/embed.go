// Package embed provides embedded prompt templates used by Chief.
// All prompts are embedded at compile time using Go's embed directive.
package embed

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"
)

//go:embed prompt.txt
var promptTemplate string

//go:embed init_prompt.txt
var initPromptTemplate string

//go:embed edit_prompt.txt
var editPromptTemplate string

//go:embed convert_prompt.txt
var convertPromptTemplate string

//go:embed detect_setup_prompt.txt
var detectSetupPromptTemplate string

// loadTemplate returns the content from promptsDir/<filename> if promptsDir is
// non-empty and the file exists and is readable; otherwise returns the embedded
// fallback string. It never returns an error: a missing override is not a fault.
func loadTemplate(promptsDir, filename, fallback string) string {
	if promptsDir == "" {
		return fallback
	}
	data, err := os.ReadFile(filepath.Join(promptsDir, filename))
	if err != nil {
		// File absent or unreadable — use embedded default silently.
		return fallback
	}
	return string(data)
}

// GetPrompt returns the agent prompt with the PRD path, progress path, and
// current story context substituted. The storyContext is the JSON of the
// current story to work on, inlined directly into the prompt so that the
// agent does not need to read the entire prd.json file.
func GetPrompt(promptsDir, prdPath, progressPath, storyContext, storyID, storyTitle string) string {
	tmpl := loadTemplate(promptsDir, "prompt.txt", promptTemplate)
	result := strings.ReplaceAll(tmpl, "{{PRD_PATH}}", prdPath)
	result = strings.ReplaceAll(result, "{{PROGRESS_PATH}}", progressPath)
	result = strings.ReplaceAll(result, "{{STORY_CONTEXT}}", storyContext)
	result = strings.ReplaceAll(result, "{{STORY_ID}}", storyID)
	return strings.ReplaceAll(result, "{{STORY_TITLE}}", storyTitle)
}

// GetInitPrompt returns the PRD generator prompt with the PRD directory and optional context substituted.
func GetInitPrompt(promptsDir, prdDir, context string) string {
	if context == "" {
		context = "No additional context provided. Ask the user what they want to build."
	}
	tmpl := loadTemplate(promptsDir, "init_prompt.txt", initPromptTemplate)
	result := strings.ReplaceAll(tmpl, "{{PRD_DIR}}", prdDir)
	return strings.ReplaceAll(result, "{{CONTEXT}}", context)
}

// GetEditPrompt returns the PRD editor prompt with the PRD directory substituted.
func GetEditPrompt(promptsDir, prdDir string) string {
	tmpl := loadTemplate(promptsDir, "edit_prompt.txt", editPromptTemplate)
	return strings.ReplaceAll(tmpl, "{{PRD_DIR}}", prdDir)
}

// GetConvertPrompt returns the PRD converter prompt with the file path and ID prefix substituted.
// Claude reads the file itself using file-reading tools instead of receiving inlined content.
// The idPrefix determines the story ID convention (e.g., "US" → US-001, "MFR" → MFR-001).
func GetConvertPrompt(promptsDir, prdFilePath, idPrefix string) string {
	tmpl := loadTemplate(promptsDir, "convert_prompt.txt", convertPromptTemplate)
	result := strings.ReplaceAll(tmpl, "{{PRD_FILE_PATH}}", prdFilePath)
	return strings.ReplaceAll(result, "{{ID_PREFIX}}", idPrefix)
}

// GetDetectSetupPrompt returns the prompt for detecting project setup commands.
func GetDetectSetupPrompt(promptsDir string) string {
	return loadTemplate(promptsDir, "detect_setup_prompt.txt", detectSetupPromptTemplate)
}

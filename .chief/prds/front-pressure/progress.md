## Codebase Patterns
- Config structs follow the pattern: nested struct type + field on Config with yaml tag matching camelCase config key
- Tests use `t.TempDir()` for isolated file system operations
- All config types are in `internal/config/config.go` (single file)
- Test package is `package config` (same package, not `_test`)
- Parser uses `extractStoryID(text, startTag, endTag)` to extract content between XML-like tags; reuse this for new tag types
- Parser test file is `package loop` (same package, not `_test`)
- Multiline text in stream-json test fixtures must use `\n` escape sequences in JSON strings (literal newlines break JSON parsing)

---

## 2026-03-04 - US-001
- What was implemented: Added `FrontPressureConfig` struct with `Enabled bool` field to `internal/config/config.go`. Added `FrontPressure FrontPressureConfig` field on the `Config` struct with `yaml:"frontPressure"` tag.
- Files changed:
  - `internal/config/config.go` - added `FrontPressureConfig` type and `FrontPressure` field on `Config`
  - `internal/config/config_test.go` - added `TestFrontPressureMarshalUnmarshal` and `TestFrontPressureDefaultsToFalse` tests
- **Learnings for future iterations:**
  - Config follows a consistent pattern: each feature area gets its own `XxxConfig` struct with yaml tags
  - The `Default()` function returns `&Config{}` which gives zero-values (false for bool, empty for string)
  - Tests use temp dirs and Save/Load functions for round-trip testing
  - The `prd.json` may have an `inProgress: true` field added by the chief orchestrator - safe to remove when setting `passes: true`
---

## 2026-03-04 - US-002
- What was implemented: Added `DismissedConcerns []string` field to `UserStory` in `internal/prd/types.go` with `json:"dismissedConcerns,omitempty"` tag. Added three tests to `prd_test.go` covering the new field.
- Files changed:
  - `internal/prd/types.go` - added `DismissedConcerns []string` field to `UserStory` struct
  - `internal/prd/prd_test.go` - added `TestDismissedConcerns_EmptyOmittedFromJSON`, `TestDismissedConcerns_RoundTrip`, `TestDismissedConcerns_LegacyPRDDeserializesWithEmptySlice`
- **Learnings for future iterations:**
  - `prd_test.go` uses `package prd` (same package), not `package prd_test`
  - The `omitempty` tag on a `[]string` field causes nil/empty slices to be omitted from JSON - verified via `map[string]interface{}` inspection
  - Legacy PRDs without new fields deserialize cleanly due to Go's zero-value initialization (nil slice for `[]string`)
  - Test file already imports `encoding/json`, `os`, `path/filepath` - no new imports needed for these tests
---

## 2026-03-04 - US-003
- What was implemented: Added `EventFrontPressure` constant to `EventType` iota in `internal/loop/parser.go`. Added `"FrontPressure"` case to `String()` method. Added detection of `<front-pressure>...</front-pressure>` tags in `parseAssistantMessage()` using the existing `extractStoryID()` helper. Added four new tests to `parser_test.go` and added `EventFrontPressure` to the `TestEventTypeString` table.
- Files changed:
  - `internal/loop/parser.go` - added `EventFrontPressure` constant, `String()` case, and tag detection in `parseAssistantMessage()`
  - `internal/loop/parser_test.go` - added `EventFrontPressure` to string table, added `TestParseLineFrontPressurePresent`, `TestParseLineFrontPressureAbsent`, `TestParseLineFrontPressureMalformed`, `TestParseLineFrontPressureMultiline`
- **Learnings for future iterations:**
  - `extractStoryID()` is a general-purpose tag extractor - reuse it for any `<tag>content</tag>` pattern
  - The concern detection is placed BEFORE the `ralph-status` check in `parseAssistantMessage()` - order matters since first match wins
  - `extractStoryID()` already calls `strings.TrimSpace()` on the extracted content - no need to trim again at the call site
  - Stream-json test fixtures with multiline text must encode newlines as `\n` (JSON escape), not literal newlines
---

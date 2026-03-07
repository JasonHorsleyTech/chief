## Codebase Patterns
- Config struct is in `internal/config/config.go`; tests are in `config_test.go` in the same package
- `Default()` must explicitly set non-zero defaults (e.g., `RetryIntervalMinutes: 60`)
- `Load()` starts from `Default()` then unmarshals YAML over it, so missing keys keep their defaults automatically
- `Save()` uses `yaml.Marshal` — no changes needed there when adding fields
- The config file lives at `.chief/config.yaml` relative to the project root (baseDir)

---

## 2026-03-07 - US-001
- What was implemented: Added four new fields to the `Config` struct: `PromptsDir string`, `RetryOnRateLimit bool`, `RetryIntervalMinutes int` (default: 60), `MaxRateLimitRetries int` (default: 3). Updated `Default()` to return non-zero defaults for `RetryIntervalMinutes` and `MaxRateLimitRetries`. Added two new test functions: `TestSaveAndLoadNewFields` and `TestLoadMissingNewFieldsUseDefaults`.
- Files changed: `internal/config/config.go`, `internal/config/config_test.go`
- **Learnings for future iterations:**
  - The `Load()` function uses `cfg := Default()` before unmarshaling, so YAML fields that are missing will automatically keep their defaults — no special handling needed
  - The existing `Save()` with `yaml.Marshal` automatically serializes all struct fields including newly added ones
  - CLI commands live in `internal/cmd/` — future US-002/US-003 will need to add a `config` subcommand there
  - The `cmd/chief/` directory contains the main entry point; examine how existing subcommands are wired up before adding new ones
---

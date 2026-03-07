## Codebase Patterns
- Config struct is in `internal/config/config.go`; tests are in `config_test.go` in the same package
- `Default()` must explicitly set non-zero defaults (e.g., `RetryIntervalMinutes: 60`)
- `Load()` starts from `Default()` then unmarshals YAML over it, so missing keys keep their defaults automatically
- `Save()` uses `yaml.Marshal` — no changes needed there when adding fields
- The config file lives at `.chief/config.yaml` relative to the project root (baseDir)
- Subcommands are implemented in `internal/cmd/` as `Run<Name>` functions; wired up in `cmd/chief/main.go` via `findSubcmd()` switch
- `findSubcmd()` returns the first non-flag positional arg; nested sub-subcommands (e.g., `config init`) require manual arg parsing within the subcommand handler
- Tests that capture stdout use `os.Pipe()` to redirect `os.Stdout`; must restore `os.Stdout` after — do not run in parallel
- `config.Exists(baseDir)` is the exported way to check if `.chief/config.yaml` exists

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

## 2026-03-07 - US-002
- What was implemented: Added `chief config` subcommand that prints the current effective config as YAML. Added `chief config --help` which shows usage. The command prints a `# Config: <path>` header and a note if no config file exists, then the full config as YAML including defaults.
- Files changed: `internal/cmd/config.go` (new), `internal/cmd/config_test.go` (new), `cmd/chief/main.go`
- **Learnings for future iterations:**
  - New subcommands need: a `Run<Name>` function + options struct in `internal/cmd/`, a `run<Name>()` dispatch function in `main.go`, and a `case "<name>":` in the `findSubcmd()` switch
  - `runConfig()` in `main.go` already has a `switch subCmd` ready for `case "init":` — US-003 only needs to add that case and implement `RunConfigInit` in `internal/cmd/config.go`
  - `gopkg.in/yaml.v3` is available as a dependency and can be imported in `internal/cmd/` files
  - `config.Exists(baseDir)` + `config.Load(baseDir)` are the two config package functions needed for the show-config command
  - Test output capture via `os.Pipe()` works reliably as long as tests are sequential (no `t.Parallel()`)
---

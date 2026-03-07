## Codebase Patterns
- `IsRateLimitError` lives in `internal/loop/ratelimit.go`; uses `strings.ToLower` + `strings.Contains` for case-insensitive matching against a `rateLimitPatterns` slice
- Rate-limit retry uses an outer loop in `runIterationWithRetry`; crash retry is the inner loop in `runIterationWithCrashRetry`; rate-limit detection flag (`rateLimitDetected`) is set in `processOutput` and checked after each inner loop run
- `EventRateLimitWaiting` carries `RetryAt time.Time`, `AttemptNumber int`, `MaxAttempts int` fields; zero-valued for all other event types
- `LoopStateRateLimitWaiting` added to `LoopState` enum in `manager.go`; manager sets it on `EventRateLimitWaiting` and resets to `LoopStateRunning` on `EventRetrying`/`EventIterationStart`
- `Manager.SetConfig()` derives `rateLimitRetryConfig` from `config.Config` fields automatically
- TUI `AppState` enum mirrors `LoopState` but is separate; `StateRateLimitWaiting` added to `AppState` in `internal/tui/app.go`
- TUI countdown uses `rateLimitCountdownTickMsg` (1s tick) stored as `extraCmd` in `handleLoopEvent`; `extraCmd` is batched with `listenForManagerEvents()` at the end of the function
- `renderDetailsPanel` checks `a.state == StateRateLimitWaiting` before error state to show the countdown panel
- Activity line countdown text is computed fresh from `time.Until(a.rateLimitRetryAt)` on every render — no stored countdown value needed
- When transitioning back from `StateRateLimitWaiting` (on `EventRetrying` or `EventIterationStart`), `tickElapsed()` is restarted via `extraCmd` because the elapsed ticker stops when state is not `StateRunning`


- Config struct is in `internal/config/config.go`; tests are in `config_test.go` in the same package
- `Default()` must explicitly set non-zero defaults (e.g., `RetryIntervalMinutes: 60`)
- `Load()` starts from `Default()` then unmarshals YAML over it, so missing keys keep their defaults automatically
- `Save()` uses `yaml.Marshal` — no changes needed there when adding fields
- The config file lives at `.chief/config.yaml` relative to the project root (baseDir)
- Subcommands are implemented in `internal/cmd/` as `Run<Name>` functions; wired up in `cmd/chief/main.go` via `findSubcmd()` switch
- `findSubcmd()` returns the first non-flag positional arg; nested sub-subcommands (e.g., `config init`) require manual arg parsing within the subcommand handler
- Tests that capture stdout use `os.Pipe()` to redirect `os.Stdout`; must restore `os.Stdout` after — do not run in parallel
- `config.Exists(baseDir)` is the exported way to check if `.chief/config.yaml` exists
- For config init with comments: use a raw string `const` template rather than `yaml.Marshal` (which strips comments); write via `os.WriteFile`

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

## 2026-03-07 - US-003
- What was implemented: Added `chief config init` sub-subcommand. Implemented `RunConfigInit` function in `internal/cmd/config.go` with a `configInitTemplate` const string (raw YAML with inline comments for every field). Handles existing file (error unless `--force`). Added 5 tests covering: file creation, all fields present, valid YAML parseable by `Load()`, error on existing file, `--force` overwrite. Wired `case "init":` into `runConfig()` in `main.go` with `--force` flag parsing.
- Files changed: `internal/cmd/config.go`, `internal/cmd/config_test.go`, `cmd/chief/main.go`, `.chief/prds/config-and-retry-loop/prd.json`
- **Learnings for future iterations:**
  - `yaml.Marshal` does not preserve comments — use a hand-written template string for commented config files
  - Sub-subcommand flags (like `--force` for `config init`) must be parsed from the `remaining` slice inside `runConfig()` in `main.go`
  - `os.MkdirAll(filepath.Dir(cfgPath), 0o755)` is needed before `os.WriteFile` to ensure `.chief/` exists
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

## 2026-03-07 - US-004
- What was implemented: Updated `printHelp()` in `cmd/chief/main.go` to surface the config command. Added `config` to the Commands list with an inline note about `.chief/config.yaml`. Added a `Config:` section at the end of help output mentioning the config file path, `chief config` to view settings, `chief config init` to create a config file, and `chief config --help` for subcommand details.
- Files changed: `cmd/chief/main.go`, `.chief/prds/config-and-retry-loop/prd.json`
- **Learnings for future iterations:**
  - `printHelp()` in `cmd/chief/main.go` is a raw multi-line string literal — edits are straightforward text additions
  - The `config` subcommand was already wired in `findSubcmd()` switch but not mentioned in the help output; always keep Commands list in sync with the switch cases
---

## 2026-03-07 - US-006
- What was implemented: Added `IsRateLimitError(output string) bool` in `internal/loop/ratelimit.go`. Matches known rate-limit/quota patterns case-insensitively: `"rate limit"`, `"rate_limit"`, `"quota"`, `"overloaded"`, `"529"`, `"429"`, `"usage limit"`. Tests cover 8 true cases and 4 false cases.
- Files changed: `internal/loop/ratelimit.go` (new), `internal/loop/ratelimit_test.go` (new), `.chief/prds/config-and-retry-loop/prd.json`
- **Learnings for future iterations:**
  - New helper functions in `internal/loop/` can go in their own file; the package is `loop` (no subdirectory needed)
  - Pattern matching uses `strings.ToLower` + `strings.Contains` for simple, robust case-insensitive matching — no regex needed
  - US-007 will use `IsRateLimitError` in `runIteration` to decide whether to enter a rate-limit waiting state vs. treating the error as a crash
---

## 2026-03-07 - US-007
- What was implemented: Added `RateLimitRetryConfig` struct and rate-limit retry logic to the loop runner. Added `EventRateLimitWaiting` event type with `RetryAt`, `AttemptNumber`, `MaxAttempts` fields. Added `LoopStateRateLimitWaiting` to `LoopState`. Rate-limit detection happens in `processOutput` by calling `IsRateLimitError` on raw output lines. The outer retry loop (`runIterationWithRetry`) handles rate-limit waits; crash retries stay in inner `runIterationWithCrashRetry`. `Manager.SetConfig()` auto-derives `rateLimitRetryConfig` from config fields. Rate-limit retry is disabled by default (opt-in via `retryOnRateLimit: true` in config).
- Files changed: `internal/loop/parser.go`, `internal/loop/loop.go`, `internal/loop/manager.go`, `internal/loop/loop_test.go`, `.chief/prds/config-and-retry-loop/prd.json`
- **Learnings for future iterations:**
  - Rate-limit retry is a separate outer loop from crash retry — don't conflate the two
  - `processOutput` runs in a goroutine but `wg.Wait()` in `runIteration` ensures it completes before the error is inspected; safe to check `rateLimitDetected` after `runIteration` returns
  - `time.After(0)` fires immediately — setting `RetryIntervalMinutes: 0` enables fast integration tests
  - `LoopStateRateLimitWaiting` is set by the manager's event forwarding goroutine based on `EventRateLimitWaiting`; it reverts to `LoopStateRunning` on next `EventRetrying` or `EventIterationStart`
  - US-008 (TUI countdown) will react to `EventRateLimitWaiting` events and use `RetryAt` to compute the countdown
---

## 2026-03-07 - US-008
- What was implemented: Added rate-limit countdown display to the TUI. Added `StateRateLimitWaiting` to `AppState` enum with its `String()`, styles, and activity/state style mappings. Added `rateLimitRetryAt`, `rateLimitAttemptNumber`, `rateLimitMaxAttempts` fields to `App` struct. Added `rateLimitCountdownTickMsg` type and `tickRateLimitCountdown()` function (1s tick). Handles `EventRateLimitWaiting` in `handleLoopEvent` to set state and start tick; handles transition back on `EventRetrying`/`EventIterationStart`. Dashboard activity line and narrow activity line show "Rate limit — retrying in H:MM:SS  (Attempt N/M)" when in waiting state. Details panel shows a `renderRateLimitPanel` with countdown and attempt info. Log viewer displays rate-limit waiting events with ⏳ prefix. Footer shortcuts adapted for the waiting state.
- Files changed: `internal/tui/app.go`, `internal/tui/styles.go`, `internal/tui/dashboard.go`, `internal/tui/log.go`, `.chief/prds/config-and-retry-loop/prd.json`
- **Learnings for future iterations:**
  - Use `extraCmd tea.Cmd` local variable in `handleLoopEvent` and batch it at the end with `tea.Batch(a.listenForManagerEvents(), autoActionCmd, extraCmd)` — `tea.Batch` handles nil commands
  - Countdown text is computed via `time.Until(a.rateLimitRetryAt)` on every render — no need to store a countdown value
  - The `elapsedTickMsg` handler only re-arms when `a.state == StateRunning`; when coming back from rate-limit waiting, explicitly restart it via `extraCmd = tickElapsed()`
  - `renderDetailsPanel` should check rate-limit state before error state to avoid hiding the countdown behind the error panel
---

## 2026-03-07 - US-005
- What was implemented: Wired `PromptsDir` config field as a fallback in `runTUIWithOptions`, `runNew`, and `runEdit` in `cmd/chief/main.go`. When no `--prompts-dir` CLI flag is provided, the config is loaded from cwd and `cfg.PromptsDir` is used if it exists and is a valid directory. CLI flag still takes precedence when both are set.
- Files changed: `cmd/chief/main.go`, `.chief/prds/config-and-retry-loop/prd.json`
- **Learnings for future iterations:**
  - The `config` package is already imported in `main.go` — no new imports needed when adding config loading
  - `runTUIWithOptions` is called recursively for post-exit TUI restarts; setting `opts.PromptsDir` early in the function means it propagates through all recursive calls correctly
  - Config-based `PromptsDir` should be validated with `os.Stat` (same as CLI path) to silently ignore misconfigured/missing paths
  - The three places to wire prompts dir: `runTUIWithOptions` (TUI mode), `runNew` (new PRD creation), `runEdit` (PRD editing)
---

## 2026-03-07 - US-009
- What was implemented: Added a "Configuration" section to `README.md` documenting all config fields in a markdown table and including an example YAML snippet showing retry settings enabled for overnight use. The section explains the rate-limit countdown TUI behavior as well.
- Files changed: `README.md`, `.chief/prds/config-and-retry-loop/prd.json`
- **Learnings for future iterations:**
  - README edits are straightforward — insert the new section between existing `## Requirements` and `## License` headings
  - The config fields to document live in `internal/cmd/config.go` (`configInitTemplate`) and `internal/config/config.go` (`Default()`)
  - A markdown table works well for documenting config fields with their defaults and descriptions
---

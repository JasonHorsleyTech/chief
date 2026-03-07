# PRD: Expanded Config System + Rate-Limit Retry Loop

## Introduction

Chief currently has a minimal config system (`.chief/config.yaml`) with only three settings, accessible only through the TUI settings overlay. As chief grows to support more users with different workflows, it needs a richer, more discoverable configuration system.

This PRD covers two tightly related improvements:

1. **Expanded Config System** — Add new configuration fields (including retry settings, prompts path overrides, and behavior defaults), make the config file easy to initialize and edit from the CLI, and surface it in `--help` output so users know it exists.

2. **Rate-Limit Retry Loop** — When a running Claude session hits an API rate limit or quota exhaustion error, chief should wait a configurable number of minutes and automatically retry rather than halting overnight work.

A key insight: the existing `RetryConfig` in `internal/loop/loop.go` handles short-delay crash recovery (retries with seconds of backoff). This feature adds a *separate*, long-delay rate-limit recovery — fundamentally different in purpose and timescale.

---

## Goals

- Add `retryOnRateLimit` and `retryIntervalMinutes` to the config schema so users can configure automated overnight recovery
- Add a `promptsDir` config field so users can point chief at a custom prompts directory without a CLI flag
- Make the config file easily discoverable: visible in `--help` and `chief --help` output
- Add a `chief config` subcommand for viewing the active config and opening/initializing the config file
- When a rate-limit error is detected, show a clear countdown in the TUI and retry automatically when the interval expires
- Keep all new config fields optional with sensible defaults (feature is opt-in, default off)

---

## User Stories

### US-001: Expand config schema with new fields
**Priority:** 1
**Description:** As a developer, I need the config struct to support the new settings so they can be parsed from YAML, defaulted, and passed through the system.

**Acceptance Criteria:**
- [ ] `Config` struct in `internal/config/config.go` adds:
  - `PromptsDir string` (default: `""`, meaning use embedded)
  - `RetryOnRateLimit bool` (default: `false`)
  - `RetryIntervalMinutes int` (default: `60`)
  - `MaxRateLimitRetries int` (default: `3`)
- [ ] `Default()` returns correct zero/false values for new fields
- [ ] `Load()` correctly deserializes new fields from YAML; missing keys use defaults
- [ ] `Save()` serializes all fields including new ones
- [ ] `go test ./internal/config/...` passes

### US-002: Add `chief config` subcommand
**Priority:** 2
**Description:** As a user, I want to run `chief config` to see my current config and know where the file lives, so I can understand what's configured without hunting for the file.

**Acceptance Criteria:**
- [ ] `chief config` prints the current effective config as YAML to stdout, including default values
- [ ] Output includes a comment header showing the config file path that was loaded (e.g., `# Config: /path/to/project/.chief/config.yaml`)
- [ ] If no config file exists, output includes a note: `# No config file found. Run 'chief config init' to create one.`
- [ ] `chief config --help` shows usage
- [ ] `go test ./cmd/...` passes (or existing test suite passes)

### US-003: Add `chief config init` to scaffold a config file
**Priority:** 3
**Description:** As a new user, I want to run `chief config init` to generate a commented config file with all available options, so I can understand what's configurable and edit it to my needs.

**Acceptance Criteria:**
- [ ] `chief config init` creates `.chief/config.yaml` in the current directory
- [ ] Generated file includes every config field with its default value
- [ ] Each field has an inline comment explaining what it does and valid values
- [ ] If `.chief/config.yaml` already exists, command prints an error and exits non-zero without overwriting
- [ ] `--force` flag overwrites existing config
- [ ] Generated file is valid YAML that `Load()` can parse without error
- [ ] `go test ./cmd/...` passes

### US-004: Surface config in `--help` output
**Priority:** 4
**Description:** As a user reading `chief --help`, I want to see a mention of the config file so I know it exists and how to learn more about it.

**Acceptance Criteria:**
- [ ] `chief --help` output includes a section or note like: `Config: Run 'chief config' to view settings, 'chief config init' to create a config file.`
- [ ] The config file path (`.chief/config.yaml`) is mentioned
- [ ] `chief config --help` is also mentioned or discoverable from `chief --help`

### US-005: Wire `promptsDir` config field through the system
**Priority:** 5
**Description:** As a user, I want to set `promptsDir` in my config file instead of always passing `--prompts-dir` on the CLI, so my custom prompts are used automatically.

**Acceptance Criteria:**
- [ ] When `PromptsDir` is set in config and no `--prompts-dir` flag is provided, the config value is used
- [ ] CLI flag `--prompts-dir` takes precedence over config value when both are set
- [ ] Empty `PromptsDir` in config falls back to embedded prompts (existing behavior)
- [ ] `go test ./...` passes

### US-006: Detect rate-limit and quota errors from Claude output
**Priority:** 6
**Description:** As a developer, I need the loop runner to detect when Claude has returned a rate-limit or quota error so the system can enter a waiting state instead of treating it as a crash.

**Acceptance Criteria:**
- [ ] A function `IsRateLimitError(output string) bool` (or equivalent) is defined in `internal/loop/`
- [ ] Function returns `true` for known rate-limit/quota patterns in Claude's stream-json output (e.g., messages containing `"rate limit"`, `"quota"`, `"overloaded"`, `"529"`, `"429"`, `"usage limit"`)
- [ ] Function returns `false` for normal process exits and other error types
- [ ] Unit tests cover at least 5 known error string patterns (true cases) and 3 non-matching cases (false cases)
- [ ] `go test ./internal/loop/...` passes

### US-007: Implement rate-limit waiting state in the loop runner
**Priority:** 7
**Description:** As a user running chief overnight, I want the loop to pause and retry after a rate-limit error instead of stopping, so I can recover quota and continue work automatically.

**Acceptance Criteria:**
- [ ] When `IsRateLimitError` returns true AND `config.RetryOnRateLimit` is true AND retry attempts < `config.MaxRateLimitRetries`, the loop enters a `StateRateLimitWaiting` state
- [ ] The loop waits `config.RetryIntervalMinutes` minutes before retrying
- [ ] A new event type `EventRateLimitWaiting` is emitted with fields: `RetryAt time.Time`, `AttemptNumber int`, `MaxAttempts int`
- [ ] After the wait, the loop retries the current story from the beginning (same behavior as existing crash retry)
- [ ] If `MaxRateLimitRetries` is exhausted, the loop emits `EventError` and stops (does not loop forever)
- [ ] If `config.RetryOnRateLimit` is false, existing behavior is preserved (loop stops on rate-limit error)
- [ ] `go test ./internal/loop/...` passes

### US-008: Show rate-limit countdown in the TUI
**Priority:** 8
**Description:** As a user, I want to see a clear countdown in the TUI when chief is waiting for a rate limit to expire, so I know it's not hung and understand when it will retry.

**Acceptance Criteria:**
- [ ] When `EventRateLimitWaiting` is received, the TUI displays a visible "Rate limit — retrying in X:XX:XX" countdown for the affected PRD/loop
- [ ] Countdown updates every second using a `tea.Tick`-style timer
- [ ] Countdown shows: attempt number and max attempts (e.g., "Attempt 1/3")
- [ ] When the wait completes and the retry begins, the display returns to normal running state
- [ ] The countdown is visible in the dashboard view (does not require switching to log view)
- [ ] Other PRDs/loops continue running normally and are not blocked by a waiting loop

### US-009: Document config fields and retry behavior
**Priority:** 9
**Description:** As a user reading the README or running `chief --help`, I want to understand the retry feature and how to enable it, so I can configure it for overnight use.

**Acceptance Criteria:**
- [ ] `README.md` includes a "Configuration" section documenting all config fields with examples
- [ ] `README.md` includes an example config snippet showing retry settings enabled
- [ ] `chief config init` output already serves as inline documentation (covered by US-003)

---

## Functional Requirements

- FR-1: `Config` struct must support `PromptsDir`, `RetryOnRateLimit`, `RetryIntervalMinutes`, and `MaxRateLimitRetries` fields
- FR-2: `chief config` prints effective config with file path header to stdout
- FR-3: `chief config init` creates a fully-commented `.chief/config.yaml` scaffold
- FR-4: `--help` output references the config system and `chief config` subcommand
- FR-5: CLI `--prompts-dir` flag takes precedence over `config.PromptsDir`; config value used when flag is absent
- FR-6: Rate-limit detection must cover HTTP 429/529 responses and quota-exhaustion message patterns in Claude stream output
- FR-7: Rate-limit retry is entirely opt-in — `RetryOnRateLimit` defaults to `false`
- FR-8: The rate-limit wait loop must be cancellable (e.g., user quits chief mid-wait)
- FR-9: Rate-limit retries are tracked separately from crash retries (`RetryConfig.MaxRetries`)
- FR-10: TUI countdown must tick every second and not block the main event loop
- FR-11: A PRD loop in rate-limit waiting state must not block other PRD loops from running

---

## Non-Goals

- No global `~/.config/chief/config.yaml` in this iteration — local project config only
- No interactive `chief config edit` that opens a TUI editor — just scaffold + manual edit
- No smart backoff for rate limits (fixed interval only, no exponential growth)
- No notification (email, Slack, etc.) when a rate limit is hit
- No persisting retry state across chief process restarts (if you kill chief during a wait, it does not resume)
- No changes to the existing short-delay crash retry behavior (`RetryConfig`)
- No new config fields beyond what's listed (scope control)

---

## Design Considerations

### Config file format (generated by `chief config init`)
```yaml
# Chief configuration — .chief/config.yaml
# Run 'chief config' to view effective settings.

# Path to a directory of custom prompt files.
# When set, chief looks here before falling back to built-in prompts.
# Example: promptsDir: ".chief/my-prompts"
promptsDir: ""

# Automatically retry when Claude hits a rate limit or quota error.
# Useful for overnight runs. Default: false (disabled).
retryOnRateLimit: false

# How many minutes to wait before retrying after a rate limit.
# Default: 60 (one hour).
retryIntervalMinutes: 60

# Maximum number of rate-limit retries before giving up.
# Default: 3
maxRateLimitRetries: 3

worktree:
  # Shell command to run after creating a new worktree.
  # Example: setup: "npm install"
  setup: ""

onComplete:
  # Automatically push the branch to origin when a story completes.
  push: false
  # Automatically create a GitHub PR when a story completes.
  createPR: false
```

### TUI countdown display
The dashboard row for a waiting loop should show something like:
```
[PRD: my-feature]  ⏸ Rate limit — retrying in 0:47:23  (attempt 1/3)
```
Reuse the existing status styling patterns from `internal/tui/app.go`.

---

## Technical Considerations

- **Existing RetryConfig**: `internal/loop/loop.go` already has `RetryConfig` with `MaxRetries`, `RetryDelays`, and `Enabled`. The new rate-limit retry is a separate mechanism — do not conflate the two. Rate-limit retry uses minute-scale delays; crash retry uses second-scale delays.
- **Error detection**: Claude is spawned as a subprocess with stream-json output. Rate-limit errors may appear as JSON `type: "result"` messages with error content, or as stderr. Check both streams.
- **Context cancellation**: The wait loop must select on `ctx.Done()` so chief can be stopped cleanly mid-wait.
- **`chief config` subcommand**: Use the existing CLI flag parsing pattern in `cmd/chief/main.go`. Check how `chief init-prompts` was added as a reference for adding subcommands.
- **`PromptsDir` threading**: `promptsDir` is already supported as a CLI flag (`--prompts-dir`). Config integration just means reading the config value when the flag is unset. Check `internal/prompts/` for where this is resolved.
- **TUI timer**: Use `tea.Tick(time.Second, ...)` pattern for the countdown. Store `RetryAt time.Time` in the loop's TUI state and compute remaining time each tick.

---

## Success Metrics

- A user can run `chief config init`, edit one line to set `retryOnRateLimit: true`, and have chief recover from a rate limit without any manual intervention
- `chief --help` mentions the config system — new users can discover it without reading the README
- A chief process hitting a rate limit at 11 PM with `retryIntervalMinutes: 60` and `maxRateLimitRetries: 3` can recover and complete additional stories by 2 AM without user interaction
- Config file scaffold (`chief config init`) is self-documenting — no external docs needed to understand available settings

---

## Open Questions

- Should `chief config init` also run during `chief init-prompts` setup flow, or stay independent?
- Should the TUI settings overlay be expanded to show the new retry fields, or is YAML editing sufficient for now?
- Should `chief config` subcommand support `chief config set retryOnRateLimit true` one-liner edits in a future iteration?

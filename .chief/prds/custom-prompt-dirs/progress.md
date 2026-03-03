## Codebase Patterns
- CLI flag parsing uses a hand-rolled switch-case loop in `parseTUIFlags()` (cmd/chief/main.go), not the standard `flag` package. New flags follow the same space-separated and `=`-separated patterns.
- Value-taking flags validate inline and call `os.Exit(1)` on error — no error returns.
- `TUIOptions` is the top-level struct passed to `runTUIWithOptions`; fields added here naturally thread through all recursive calls.
- The `embed` package (root-level `/embed/`) contains all five prompt loaders: `GetPrompt`, `GetInitPrompt`, `GetEditPrompt`, `GetConvertPrompt`, `GetDetectSetupPrompt`.
- All five `Get*` functions in `embed/embed.go` now take `promptsDir string` as their first parameter. Pass `""` to use embedded defaults; pass a directory path to enable overrides.
- Override file lookup uses a private `loadTemplate(promptsDir, filename, fallback)` helper — no error on missing override file, silent fallback to embedded.
- `loop.Manager` has `SetPromptsDir(string)` that must be called before any `Start` — it is captured in the `promptBuilderForPRD` closure when `Start` is called. Changing it after a loop starts has no effect on that loop.
- `tui.App.SetPromptsDir(string)` delegates to `a.manager.SetPromptsDir` — the right place to call it is immediately after `tui.NewAppWithOptions` in `main.go`.
- The orchestrator injects `"inProgress": true` into prd.json; always remove it (alongside setting `"passes": true`) when marking a story complete.

---

## 2026-03-03 - US-001
- Added `PromptsDir string` field to `TUIOptions` struct in `cmd/chief/main.go`
- Added `--prompts-dir <path>` and `--prompts-dir=<path>` flag parsing in `parseTUIFlags()`
- Directory existence validation: uses `os.Stat` + `info.IsDir()`; prints `prompts directory not found: <path>` and exits with non-zero on failure
- Updated `printHelp()` to list `--prompts-dir <path>` under Global Options
- Files changed: `cmd/chief/main.go`
- **Learnings for future iterations:**
  - The `parseTUIFlags()` loop only runs for TUI mode (after subcommand routing). `--prompts-dir` before a subcommand like `new` or `edit` will NOT be parsed by this function — those stories (US-004, US-005) need separate handling in `runNew`/`runEdit` or a pre-pass.
  - `runTUIWithOptions` receives the full `*TUIOptions`, so `PromptsDir` is available for all downstream threading in later stories without extra plumbing.
---

## 2026-03-03 - US-002
- Added `promptsDir string` as the first parameter to all five `Get*` functions in `embed/embed.go`
- Introduced private `loadTemplate(promptsDir, filename, fallback string) string` helper: reads `<promptsDir>/<filename>` when dir is non-empty; silently falls back to embedded string on any error (missing file, permission denied, etc.)
- Updated all existing call sites to pass `""` as `promptsDir`: `internal/loop/loop.go`, `internal/cmd/edit.go`, `internal/cmd/new.go`, `internal/prd/generator.go`
- `GetDetectSetupPrompt` had no call sites yet; its signature was updated to `GetDetectSetupPrompt(promptsDir string)`
- Added unit tests covering: empty dir (embedded used), override file present (override used with template substitution), specific file absent from dir (embedded used)
- Files changed: `embed/embed.go`, `embed/embed_test.go`, `internal/loop/loop.go`, `internal/cmd/edit.go`, `internal/cmd/new.go`, `internal/prd/generator.go`
- **Learnings for future iterations:**
  - All call sites currently pass `""` — subsequent stories (US-003 through US-006) replace `""` with the actual `PromptsDir` from their respective option structs.
  - The `prd.json` file had an `"inProgress": true` field added by the orchestrator; need to remove it (along with `"passes": false`) when marking complete.
---

## 2026-03-03 - US-003
- Added `promptsDir string` parameter to `promptBuilderForPRD(prdPath, promptsDir string)` in `internal/loop/loop.go`; passes it through to `embed.GetPrompt`
- Added `promptsDir string` field to `loop.Manager` struct in `internal/loop/manager.go`
- Added `Manager.SetPromptsDir(string)` setter following the same pattern as `SetBaseDir`
- In `Manager.Start`, reads `m.promptsDir` under `m.mu.RLock` and passes it to `promptBuilderForPRD` so each loop iteration picks up the override
- Added `promptsDir string` field to `tui.App` struct in `internal/tui/app.go`
- Added `App.SetPromptsDir(string)` setter that stores the value and delegates to `a.manager.SetPromptsDir`
- In `cmd/chief/main.go` `runTUIWithOptions`, calls `app.SetPromptsDir(opts.PromptsDir)` when non-empty, immediately after `NewAppWithOptions`
- Files changed: `internal/loop/loop.go`, `internal/loop/manager.go`, `internal/tui/app.go`, `cmd/chief/main.go`
- **Learnings for future iterations:**
  - `promptBuilderForPRD` is called in `Manager.Start` — so `SetPromptsDir` must be called before the user presses play (before `Start`). Since `SetPromptsDir` is called right after app creation in `main.go`, this is always satisfied.
  - `NewLoopWithEmbeddedPrompt` (used in direct non-manager paths) now passes `""` to `promptBuilderForPRD` — if that code path ever needs prompts-dir support it will need updating.
  - `m.mu.RLock` already held when reading `m.retryConfig` in `Start`; I reused the same read-lock block to also read `m.promptsDir` — clean and race-free.
---

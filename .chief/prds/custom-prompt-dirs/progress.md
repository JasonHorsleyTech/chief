## Codebase Patterns
- `embed.RawTemplates()` returns a `map[string]string` of all five embedded template filenames to their raw (unsubstituted) content — use this when writing defaults to disk (e.g. `init-prompts`), not the `Get*` functions which perform variable substitution.
- CLI flag parsing uses a hand-rolled switch-case loop in `parseTUIFlags()` (cmd/chief/main.go), not the standard `flag` package. New flags follow the same space-separated and `=`-separated patterns.
- Value-taking flags validate inline and call `os.Exit(1)` on error — no error returns.
- `TUIOptions` is the top-level struct passed to `runTUIWithOptions`; fields added here naturally thread through all recursive calls.
- The `embed` package (root-level `/embed/`) contains all five prompt loaders: `GetPrompt`, `GetInitPrompt`, `GetEditPrompt`, `GetConvertPrompt`, `GetDetectSetupPrompt`.
- All five `Get*` functions in `embed/embed.go` now take `promptsDir string` as their first parameter. Pass `""` to use embedded defaults; pass a directory path to enable overrides.
- Override file lookup uses a private `loadTemplate(promptsDir, filename, fallback)` helper — no error on missing override file, silent fallback to embedded.
- `loop.Manager` has `SetPromptsDir(string)` that must be called before any `Start` — it is captured in the `promptBuilderForPRD` closure when `Start` is called. Changing it after a loop starts has no effect on that loop.
- `tui.App.SetPromptsDir(string)` delegates to `a.manager.SetPromptsDir` — the right place to call it is immediately after `tui.NewAppWithOptions` in `main.go`.
- The orchestrator injects `"inProgress": true` into prd.json; always remove it (alongside setting `"passes": true`) when marking a story complete.
- `findSubcmd()` in `cmd/chief/main.go` finds the first non-flag positional arg (skipping value-taking flags). `extractGlobalPromptsDir()` validates and returns `--prompts-dir`. Both are reusable for future subcommand routing and global flag extraction.

---

## 2026-03-03 - US-007
- Added `RawTemplates() map[string]string` to `embed/embed.go` — returns all five embedded template filenames mapped to their raw (unsubstituted) content
- Created `internal/cmd/init_prompts.go` with `InitPromptsOptions{Path string}` and `RunInitPrompts(opts)` that: defaults to `~/chief-prompts/`, resolves absolute path, calls `os.MkdirAll`, writes all five files, prints success message with path, filenames, and `chief --prompts-dir` hint
- Added `case "init-prompts":` to `switch findSubcmd()` in `main()` in `cmd/chief/main.go`
- Added `runInitPrompts()` in `cmd/chief/main.go` that finds the subcommand position and reads the optional path argument
- Updated `printHelp()` to add `init-prompts [path]` under Commands and two new examples: `chief --prompts-dir ~/chief-prompts` and `chief init-prompts`
- Files changed: `embed/embed.go`, `internal/cmd/init_prompts.go`, `cmd/chief/main.go`
- **Learnings for future iterations:**
  - Subcommands with hyphens (like `init-prompts`) work fine in `findSubcmd()` because they don't start with `-`.
  - `RawTemplates()` must return copies of the embedded vars (not call `loadTemplate`) so no directory lookup is attempted — the raw embedded strings are always returned.
  - `os.UserHomeDir()` is the right way to resolve `~` in Go rather than parsing `$HOME` directly.
---

## 2026-03-03 - US-006
- Added `PromptsDir string` field to `prd.ConvertOptions` in `internal/prd/generator.go`
- Added `promptsDir string` parameter to `runClaudeConversion`; passes it to `embed.GetConvertPrompt`
- Added `PromptsDir string` field to `cmd.ConvertOptions` in `internal/cmd/new.go`
- `RunConvertWithOptions` now forwards `opts.PromptsDir` to `prd.ConvertOptions`
- `RunNew` now calls `RunConvertWithOptions` (instead of `RunConvert`) passing `opts.PromptsDir`
- Inline conversion in `main.go` `runTUIWithOptions` now passes `opts.PromptsDir` in `prd.ConvertOptions`
- `embed.GetDetectSetupPrompt` already had `promptsDir string` as its first parameter (added in US-002); no call sites in production code exist yet, so no call-site updates were needed
- Files changed: `internal/prd/generator.go`, `internal/cmd/new.go`, `cmd/chief/main.go`
- **Learnings for future iterations:**
  - `runClaudeConversion` in `internal/prd/generator.go` is a private helper called by `Convert`; to thread a new option through it you must add the parameter to both the call site in `Convert` and the function signature.
  - `RunConvert(prdDir string)` is a convenience wrapper; prefer `RunConvertWithOptions` when you need to thread extra options — avoid adding parameters to `RunConvert` to keep the simple API stable.
  - `GetDetectSetupPrompt` has no production call sites — the function signature update in US-002 is sufficient for now.
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

## 2026-03-03 - US-005
- Added `PromptsDir string` field to `cmd.EditOptions` in `internal/cmd/edit.go`
- Changed `embed.GetEditPrompt("", prdDir)` to `embed.GetEditPrompt(opts.PromptsDir, prdDir)` in `cmd.RunEdit`
- Refactored `runEdit()` in `cmd/chief/main.go`: added `extractGlobalPromptsDir()` call, found "edit" position in os.Args to correctly skip leading global flags, parsed name and edit-specific flags from args following "edit"
- Forwarded `PromptsDir: opts.PromptsDir` in the `PostExitEdit` handler inside `runTUIWithOptions`
- Files changed: `internal/cmd/edit.go`, `cmd/chief/main.go`
- **Learnings for future iterations:**
  - The same "find subcommand position" pattern used for `runNew()` (US-004) was reused verbatim for `runEdit()` — it's the right approach for any subcommand that can be preceded by global flags.
  - `runEdit()` formerly started at `os.Args[2]`; starting there breaks `chief --prompts-dir /foo edit` since it would see `--prompts-dir` as an unknown flag — always find the subcommand's index first.
---

## 2026-03-03 - US-004
- Added `PromptsDir string` field to `cmd.NewOptions` in `internal/cmd/new.go`
- Changed `embed.GetInitPrompt("", ...)` to `embed.GetInitPrompt(opts.PromptsDir, ...)` in `cmd.RunNew`
- Added `findSubcmd()` helper in `cmd/chief/main.go` that returns the first non-flag argument, skipping value-taking flags (`--prompts-dir`, `--max-iterations`, `-n`) and their values
- Added `extractGlobalPromptsDir()` helper in `cmd/chief/main.go` that scans all args for `--prompts-dir` with validation (exits on bad path)
- Replaced `switch os.Args[1]` routing in `main()` with `switch findSubcmd()` so `chief --prompts-dir /foo new` correctly routes to `runNew()`
- Updated `runNew()` to call `extractGlobalPromptsDir()` and to find "new" by position in args (not always at index 1) so name/context args are parsed correctly even after global flags
- Forwarded `opts.PromptsDir` in both the `PostExitInit` handler and the first-time setup `RunNew` call inside `runTUIWithOptions`
- Files changed: `internal/cmd/new.go`, `cmd/chief/main.go`
- **Learnings for future iterations:**
  - `--help` and `--version` were previously matched in the `switch os.Args[1]` block; after moving to `findSubcmd()` (which skips `-`-prefixed args), they fall through to `parseTUIFlags()` which handles them — no special-casing needed.
  - `runEdit()` still uses `os.Args[i]` starting from index 2; US-005 will need the same "find subcommand position" fix if `chief --prompts-dir /foo edit` should work.
  - `extractGlobalPromptsDir()` is also used by `runNew()`; US-005 should reuse it for `runEdit()`.
---

## Codebase Patterns
- CLI flag parsing uses a hand-rolled switch-case loop in `parseTUIFlags()` (cmd/chief/main.go), not the standard `flag` package. New flags follow the same space-separated and `=`-separated patterns.
- Value-taking flags validate inline and call `os.Exit(1)` on error — no error returns.
- `TUIOptions` is the top-level struct passed to `runTUIWithOptions`; fields added here naturally thread through all recursive calls.
- The `embed` package (root-level `/embed/`) contains all five prompt loaders: `GetPrompt`, `GetInitPrompt`, `GetEditPrompt`, `GetConvertPrompt`, `GetDetectSetupPrompt`.

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

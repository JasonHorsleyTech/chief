# PRD: Custom Prompts Directory Override

## Introduction

Chief embeds five prompt templates at compile time that drive every Claude
invocation in the Ralph loop. This feature lets users override any or all of
those prompts by pointing Chief at a local directory containing replacement
files. Running `chief --prompts-dir /path/to/dir` causes every operation in
that session to prefer the files in that directory over the built-in defaults,
enabling rapid experimentation with agent instructions without recompiling.

A companion `chief init-prompts [path]` subcommand scaffolds the directory by
copying all five embedded prompts into it, giving users a ready-to-edit
starting point.

## Goals

- Accept a `--prompts-dir <path>` global flag that applies to all Claude
  invocations in the session (TUI run, `chief new`, `chief edit`, PRD
  conversion, and first-time setup detection).
- For each of the five embedded prompts, use the override file if present in
  the prompts directory; fall back to the embedded default if the file is
  absent.
- Hard-fail with a descriptive error if `--prompts-dir` is given but the
  directory does not exist or is not readable. Never silently ignore a bad
  path.
- Provide `chief init-prompts [path]` to scaffold the override directory by
  copying all five embedded prompts into it. Default path: `~/chief-prompts`.

## User Stories

### US-001: Parse `--prompts-dir` global flag and validate directory
**Priority:** 1
**Description:** As a user, I want Chief to accept a `--prompts-dir <path>`
flag so I can tell it where my custom prompts live before any Claude
invocations happen.

**Acceptance Criteria:**
- [ ] `--prompts-dir <path>` is accepted anywhere in the CLI argument list
      before a subcommand or PRD name (e.g. `chief --prompts-dir /foo`,
      `chief --prompts-dir /foo new`, `chief --prompts-dir /foo my-prd`).
- [ ] If the flag is present and `<path>` does not exist or is not a readable
      directory, Chief prints a clear error (`prompts directory not found:
      <path>`) and exits with a non-zero code. It does not fall back silently.
- [ ] If the flag is absent, Chief behaves exactly as it does today.
- [ ] The parsed path is stored in `TUIOptions.PromptsDir string` and threaded
      into all downstream option structs throughout the session (including
      recursive `runTUIWithOptions` calls triggered by post-exit actions).
- [ ] `--help` output lists the new flag with a one-line description.
- [ ] Typecheck passes.

---

### US-002: Override-aware prompt loading in the `embed` package
**Priority:** 2
**Description:** As a developer, I need the `embed` package to check a
prompts directory for each template file before using the compiled-in default,
so all five prompts can be independently overridden.

**Acceptance Criteria:**
- [ ] Each of the five existing `Get*` functions gains a `promptsDir string`
      as its first parameter: `GetPrompt`, `GetInitPrompt`, `GetEditPrompt`,
      `GetConvertPrompt`, `GetDetectSetupPrompt`.
- [ ] When `promptsDir` is non-empty, the function attempts to read the
      corresponding file from that directory using the same filename as the
      embedded file (`prompt.txt`, `init_prompt.txt`, `edit_prompt.txt`,
      `convert_prompt.txt`, `detect_setup_prompt.txt`).
- [ ] If the override file is present and readable, its content is used as the
      template instead of the embedded string. Template variable substitution
      (`{{PRD_PATH}}`, `{{STORY_ID}}`, etc.) still runs normally on the
      override content.
- [ ] If the override file is absent from the directory (but the directory
      itself exists and was already validated), the function falls back to the
      embedded default without error.
- [ ] When `promptsDir` is empty (`""`), behaviour is identical to today.
- [ ] All existing call sites are updated to pass `""` as `promptsDir` so they
      compile; subsequent stories replace `""` with the real value.
- [ ] Unit tests cover: empty dir (uses embedded), dir with override file
      (uses override), dir missing a specific file (uses embedded for that
      file).
- [ ] Typecheck passes.

---

### US-003: Thread `PromptsDir` through the agent loop
**Priority:** 3
**Description:** As a user, I want the Ralph loop agent prompt to be read from
my custom directory when I launch the TUI with `--prompts-dir`, so every
Claude agent iteration uses my override instructions.

**Acceptance Criteria:**
- [ ] `tui.NewAppWithOptions` (or an equivalent setter on `tui.App`) accepts
      `PromptsDir string` and stores it.
- [ ] `Manager` gains a `PromptsDir string` field and a setter (e.g.
      `SetPromptsDir`); `tui.App` calls the setter before starting any loop.
- [ ] `promptBuilderForPRD` (in `loop.go`) accepts `promptsDir string` and
      passes it to `embed.GetPrompt`.
- [ ] `Manager.Start` passes `m.PromptsDir` when constructing the
      `promptBuilderForPRD` closure.
- [ ] Running `chief --prompts-dir /my/dir` and pressing play in the TUI
      causes each Claude iteration to use `/my/dir/prompt.txt` if it exists.
- [ ] Typecheck passes.

---

### US-004: Thread `PromptsDir` through `chief new`
**Priority:** 4
**Description:** As a user, I want `chief new` (run directly or triggered from
within the TUI) to use my custom init prompt when `--prompts-dir` is set.

**Acceptance Criteria:**
- [ ] `cmd.NewOptions` gains a `PromptsDir string` field.
- [ ] `main.go` `runNew()` reads the global `PromptsDir` (parsed from
      `--prompts-dir`) and sets `opts.PromptsDir` before calling `cmd.RunNew`.
- [ ] `cmd.RunNew` passes `opts.PromptsDir` to `embed.GetInitPrompt`.
- [ ] When a post-exit `PostExitInit` action re-invokes `runNew` from within
      `runTUIWithOptions`, the same `PromptsDir` from the original `TUIOptions`
      is forwarded.
- [ ] Running `chief --prompts-dir /my/dir new` uses `/my/dir/init_prompt.txt`
      if present.
- [ ] Typecheck passes.

---

### US-005: Thread `PromptsDir` through `chief edit`
**Priority:** 5
**Description:** As a user, I want `chief edit` to use my custom edit prompt
when `--prompts-dir` is set.

**Acceptance Criteria:**
- [ ] `cmd.EditOptions` gains a `PromptsDir string` field.
- [ ] `main.go` `runEdit()` reads the global `PromptsDir` and sets
      `opts.PromptsDir` before calling `cmd.RunEdit`.
- [ ] `cmd.RunEdit` passes `opts.PromptsDir` to `embed.GetEditPrompt`.
- [ ] When a post-exit `PostExitEdit` action re-invokes `runEdit` from within
      `runTUIWithOptions`, the `PromptsDir` from the original `TUIOptions` is
      forwarded.
- [ ] Running `chief --prompts-dir /my/dir edit` uses
      `/my/dir/edit_prompt.txt` if present.
- [ ] Typecheck passes.

---

### US-006: Thread `PromptsDir` through PRD conversion and setup detection
**Priority:** 6
**Description:** As a user, I want PRD conversion (`prd.md` → `prd.json`) and
first-time setup detection to also use my custom prompts when `--prompts-dir`
is set.

**Acceptance Criteria:**
- [ ] `prd.ConvertOptions` gains a `PromptsDir string` field.
- [ ] `prd.Convert` passes `opts.PromptsDir` to `embed.GetConvertPrompt`.
- [ ] `cmd.ConvertOptions` (in `internal/cmd/new.go`) gains `PromptsDir
      string` and forwards it to `prd.ConvertOptions` in
      `RunConvertWithOptions`.
- [ ] `cmd.RunNew` passes `PromptsDir` through to `RunConvertWithOptions`.
- [ ] The inline conversion step in `main.go` `runTUIWithOptions` also
      receives `opts.PromptsDir` in `prd.ConvertOptions`.
- [ ] Wherever `embed.GetDetectSetupPrompt()` is called, the call site is
      updated to pass `PromptsDir` (trace the call from
      `tui/first_time_setup.go` or wherever it lives).
- [ ] Running `chief --prompts-dir /my/dir` uses `/my/dir/convert_prompt.txt`
      and `/my/dir/detect_setup_prompt.txt` if present, for all conversion and
      setup-detection operations.
- [ ] Typecheck passes.

---

### US-007: Add `chief init-prompts [path]` subcommand
**Priority:** 7
**Description:** As a user, I want a single command that creates my prompts
directory pre-populated with all of Chief's current default prompts, so I can
start customising from a known baseline without hunting for the source files.

**Acceptance Criteria:**
- [ ] `chief init-prompts` (no path) creates `~/chief-prompts/` and writes all
      five prompt files into it.
- [ ] `chief init-prompts /custom/path` creates the specified directory and
      writes all five prompt files into it.
- [ ] The five files written are named exactly: `prompt.txt`,
      `init_prompt.txt`, `edit_prompt.txt`, `convert_prompt.txt`,
      `detect_setup_prompt.txt`. Content is identical to the current embedded
      defaults.
- [ ] If the target directory already exists, Chief writes (overwrites) each
      file without error. It does not ask for confirmation.
- [ ] On success, Chief prints the resolved absolute path of the directory and
      a list of the files written, e.g.:
      ```
      Prompts directory initialised: /Users/you/chief-prompts
        prompt.txt
        init_prompt.txt
        edit_prompt.txt
        convert_prompt.txt
        detect_setup_prompt.txt
      Run: chief --prompts-dir /Users/you/chief-prompts
      ```
- [ ] The subcommand is wired into the `switch os.Args[1]` block in `main.go`.
- [ ] `--help` output lists `init-prompts` under Commands.
- [ ] Typecheck passes.

---

### US-008: Update `--help` and documentation
**Priority:** 8
**Description:** As a user, I want `chief --help` to document the new flag and
subcommand so I can discover them without reading source code.

**Acceptance Criteria:**
- [ ] `--help` output includes `--prompts-dir <path>` under Global Options with
      a brief description: "Load prompt overrides from directory (hard-fails if
      path is invalid)".
- [ ] `--help` output includes `init-prompts [path]` under Commands with a
      brief description: "Scaffold a prompts directory with default templates".
- [ ] Usage examples section includes at least two examples:
      `chief --prompts-dir ~/chief-prompts` and `chief init-prompts`.
- [ ] Typecheck passes.

---

## Functional Requirements

- FR-1: `--prompts-dir <path>` is a global flag accepted before any subcommand
  or positional argument.
- FR-2: When `--prompts-dir` is given, the directory must exist and be
  readable; otherwise Chief exits with a non-zero code and a human-readable
  error before spawning any Claude process.
- FR-3: For each of the five embedded prompts, if a file with the matching
  name exists in the override directory, it replaces the embedded template for
  the entire session. Files absent from the directory use the embedded default.
- FR-4: Template variable substitution continues to work on override content
  (the same `strings.ReplaceAll` logic applies).
- FR-5: The override directory setting persists through all recursive
  `runTUIWithOptions` calls (post-exit new/edit flows).
- FR-6: `chief init-prompts [path]` writes the five embedded prompt files to
  the target directory (default `~/chief-prompts`), overwriting any existing
  files silently.
- FR-7: `chief init-prompts` does not require `--prompts-dir` and is
  independent of the override system.

## Non-Goals

- No support for per-PRD prompt overrides (one global directory per session
  only).
- No merging or patching of prompts (override file fully replaces the default).
- No validation of override file contents (Chief does not check that required
  template variables like `{{PRD_PATH}}` are present).
- No hot-reload of prompt files while the TUI is running.
- No UI in the TUI to browse or select prompt files.
- No changes to how Chief stores or ships its embedded prompts.

## Technical Considerations

- The five embedded files live in `embed/` and are compiled in via
  `//go:embed`. Their filenames are the canonical names for the override
  directory lookup.
- To expose embedded content for `init-prompts`, the `embed` package will need
  to export the raw template strings (or a `GetRaw*(name string) string`
  helper) so `main.go` can write them to disk.
- `PromptsDir` must be threaded through `TUIOptions → tui.App → Manager →
  promptBuilderForPRD` and separately through `NewOptions`, `EditOptions`, and
  `prd.ConvertOptions`. A package-level global must not be used.
- `main.go` currently parses flags with a hand-rolled loop. The `--prompts-dir`
  flag follows the same `arg == "--prompts-dir"` + `i++` pattern already used
  by `--max-iterations`.

## Success Metrics

- A user can run `chief init-prompts`, edit one of the resulting files, then
  run `chief --prompts-dir ~/chief-prompts` and observe their custom
  instructions being used by the agent — confirmed by the changed behaviour.
- No regression: running `chief` without `--prompts-dir` behaves identically
  to the current release.
- All existing tests pass; new unit tests in `embed` cover the override
  resolution logic.

## Open Questions

- Should `chief init-prompts` fail if any of the five files already exist, or
  always overwrite? (Current spec: always overwrite silently — revisit if
  users report accidental overwrites.)
- The `detect_setup_prompt.txt` call site needs to be located (likely in
  `tui/first_time_setup.go`). If it does not accept a directory yet, it will
  be wired in as part of US-006.

# Chief

<p align="center">
  <img src="assets/hero.png" alt="Chief" width="500">
</p>

Build big projects with Claude. Chief breaks your work into tasks and runs Claude Code in a loop until they're done.

**[Documentation](https://minicodemonkey.github.io/chief/)** · **[Quick Start](https://minicodemonkey.github.io/chief/guide/quick-start)**

![Chief TUI](https://minicodemonkey.github.io/chief/images/tui-screenshot.png)

## Install

```bash
brew install minicodemonkey/chief/chief
```

Or via install script:

```bash
curl -fsSL https://raw.githubusercontent.com/MiniCodeMonkey/chief/refs/heads/main/install.sh | sh
```

## Usage

```bash
# Create a new project
chief new

# Launch the TUI and press 's' to start
chief
```

Chief runs Claude in a [Ralph Wiggum loop](https://ghuntley.com/ralph/): each iteration starts with a fresh context window, but progress is persisted between runs. This lets Claude work through large projects without hitting context limits.

## How It Works

1. **Describe your project** as a series of tasks
2. **Chief runs Claude** in a loop, one task at a time
3. **One commit per task** — clean git history, easy to review

See the [documentation](https://minicodemonkey.github.io/chief/concepts/how-it-works) for details.

## Requirements

- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) installed and authenticated

## Configuration

Chief reads settings from `.chief/config.yaml` in your project directory. Run `chief config init` to create the file with all options documented inline, or `chief config` to view your current settings.

### Config fields

| Field | Default | Description |
|-------|---------|-------------|
| `worktree.setup` | `""` | Shell command run when setting up a new worktree (e.g. `"npm install"`) |
| `onComplete.push` | `false` | Push to remote when all stories complete |
| `onComplete.createPR` | `false` | Open a pull request when all stories complete |
| `promptsDir` | `""` | Path to a custom prompts directory; falls back to embedded prompts when empty |
| `retryOnRateLimit` | `false` | Automatically wait and retry when Claude hits an API rate limit |
| `retryIntervalMinutes` | `60` | Minutes to wait before retrying after a rate limit |
| `maxRateLimitRetries` | `3` | Maximum number of rate-limit retries before stopping |

### Example: overnight runs with rate-limit retry

```yaml
# .chief/config.yaml

# Automatically recover from API rate limits so chief can run overnight
retryOnRateLimit: true
retryIntervalMinutes: 60   # wait 1 hour between retries
maxRateLimitRetries: 5     # try up to 5 times before giving up

onComplete:
  push: true       # push commits when done
  createPR: true   # open a pull request automatically
```

When `retryOnRateLimit` is enabled, chief displays a live countdown in the TUI showing when the next retry will happen (e.g. _"Rate limit — retrying in 0:58:12  (Attempt 1/5)"_). Other PRDs continue running normally during the wait.

## License

MIT

## Acknowledgments

- [snarktank/ralph](https://github.com/snarktank/ralph) — The original Ralph implementation that inspired this project
- [Geoffrey Huntley](https://ghuntley.com/ralph/) — For coining the "Ralph Wiggum loop" pattern
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling

# claude-limits

A CLI tool that displays your Claude Code usage limits — 5-hour session, 7-day weekly, per-model sub-buckets, and extra usage — with visual progress bars, time-until-reset, and a compact colorized statusline for shell prompts.

## Example output

```
Claude Usage
─────────────────────────────
Session (5h)      ████░░░░░░   40%  resets in 2h 15m
Weekly  (7d)      ██████░░░░   60%  resets in 3d 4h 30m
  Opus only       ██████████  100%  resets in 3d 4h 30m
  Sonnet only     ██░░░░░░░░   20%  resets in 3d 4h 30m
  OAuth apps      ░░░░░░░░░░    0%  resets in 3d 4h 30m
  Cowork          ░░░░░░░░░░    0%  resets in 3d 4h 30m
Extra usage       ███░░░░░░░   35%  $3450 / $10000
```

Sub-buckets and extra usage are only shown when the API returns data for them.

## Installation

```sh
go install github.com/doron/claude-limits@latest
```

Or build from source:

```sh
git clone https://github.com/doron/claude-limits.git
cd claude-limits
go build -o claude-limits .
```

## Prerequisites

You must be logged into Claude Code. The tool reads your OAuth token from:

1. **macOS Keychain** (preferred) — the `Claude Code-credentials` service entry
2. **Credentials file** — `~/.claude/.credentials.json`

Set `CLAUDE_CONFIG_DIR` to override the default `~/.claude` config directory.

## Usage

```sh
claude-limits
```

### Statusline output

Use `--statusline` for a compact, colorized single-line output suitable for shell prompts, tmux status bars, and starship custom modules:

```sh
claude-limits --statusline
```

Example output (colors applied in terminal):

```
5h:40% (2h15m) 7d:60% (3d4h30m) Op:100% Sn:20% OA:0% CW:0% Ex:$34.50/$100
```

**Bucket labels:**

| Label | Bucket |
|-------|--------|
| `5h`  | 5-hour session |
| `7d`  | 7-day weekly (all models) |
| `Op`  | 7-day Opus only |
| `Sn`  | 7-day Sonnet only |
| `OA`  | 7-day OAuth apps |
| `CW`  | 7-day Cowork |
| `Ex`  | Extra usage (overage) |

**Color coding:**

| Color   | Threshold |
|---------|-----------|
| Cyan    | < 50% |
| Yellow  | 50–79% |
| Magenta | 80%+ |
| Green   | Extra usage (always) |

Buckets are omitted when the API returns no data for them. Extra usage (`Ex:$used/$limit`) is only shown when extra usage is enabled on your account.

#### Starship integration

Add a custom module to `~/.config/starship.toml`:

```toml
[custom.claude]
command = "claude-limits --statusline"
when = true
format = "[$output]($style) "
style = "bold"
```

#### tmux status bar

Add to `~/.tmux.conf`:

```sh
set -g status-right "#(claude-limits --statusline)"
set -g status-interval 60
```

### JSON output

Use `--json` for machine-readable output, useful for scripting or piping into `jq`:

```sh
claude-limits --json
```

```json
{
  "session": {
    "percent": 40,
    "resets_at": "2026-03-04T14:13:00Z",
    "resets_in_seconds": 7980
  },
  "weekly": {
    "percent": 60,
    "resets_at": "2026-03-07T17:00:00Z",
    "resets_in_seconds": 277200
  },
  "weekly_opus": {
    "percent": 100,
    "resets_at": "2026-03-07T17:00:00Z",
    "resets_in_seconds": 277200
  },
  "weekly_sonnet": {
    "percent": 20,
    "resets_at": "2026-03-07T17:00:00Z",
    "resets_in_seconds": 277200
  },
  "weekly_oauth_apps": {
    "percent": 0,
    "resets_at": "2026-03-07T17:00:00Z",
    "resets_in_seconds": 277200
  },
  "weekly_cowork": {
    "percent": 0,
    "resets_at": "2026-03-07T17:00:00Z",
    "resets_in_seconds": 277200
  },
  "extra_usage": {
    "is_enabled": true,
    "monthly_limit": 10000,
    "used_credits": 3450,
    "percent": 34.5
  }
}
```

All fields are omitted when the API returns no data for that bucket.

## API response cache

To avoid hitting the API on every invocation, responses are cached locally in `~/.claude/.usage-cache.json` (or `$CLAUDE_CONFIG_DIR/.usage-cache.json`). The default TTL is **5 minutes**.

Override the TTL with the `CLAUDE_LIMITS_CACHE_TTL` environment variable (in seconds):

```sh
# Use a 60-second cache
CLAUDE_LIMITS_CACHE_TTL=60 claude-limits

# Disable caching entirely
CLAUDE_LIMITS_CACHE_TTL=0 claude-limits
```

Cache writes are atomic and best-effort — a failure to write the cache never prevents the tool from returning results.

## Running tests

```sh
go test ./...
```

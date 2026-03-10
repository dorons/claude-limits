# claude-limits

A CLI tool that displays your Claude Code usage limits — 5-hour session and 7-day weekly utilization — with visual progress bars and time-until-reset.

## Example output

```
Claude Usage
─────────────────────────────
Session (5h)  ████░░░░░░   40%  resets in 2h 15m
Weekly  (7d)  ██████████  100%  resets in 3d 4h 30m
```

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
    "percent": 72,
    "resets_at": "2026-03-07T17:00:00Z",
    "resets_in_seconds": 277200
  }
}
```

Keys are omitted when the API returns no data for that bucket.

## API response cache

To avoid hitting the API on every invocation, responses are cached locally in `~/.claude/.usage-cache.json` (or `$CLAUDE_CONFIG_DIR/.usage-cache.json`). The default TTL is **3 minutes**.

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

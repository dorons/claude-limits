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

No flags or arguments needed. The tool queries the Anthropic usage API and prints current utilization for both the 5-hour session window and the 7-day weekly window.

## Running tests

```sh
go test ./...
```

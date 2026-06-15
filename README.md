# agent-tools

Fast native CLI tools for AI agent sessions. Single static Go binaries, no runtime dependencies, composable via Unix pipes.

## Tools

| Tool | Status | Purpose |
|------|--------|---------|
| [`gh-merge-wait`](./cmd/gh-merge-wait/) | planned | Squash-merge a PR with automatic retry/poll — no manual polling loops |
| [`gh-action-version`](./cmd/gh-action-version/) | planned | Look up the latest Node.js-24-compatible version of GitHub Actions |
| [`git-pr-branch`](./cmd/git-pr-branch/) | planned | Open/close PR branches against main with one command |
| [`patch-verify`](./cmd/patch-verify/) | implemented | Apply a safe literal string replacement and print a unified diff |
| [`issue-ship`](./cmd/issue-ship/) | planned | Drive the full triage → branch → PR → merge → cleanup pipeline |

## Install

```bash
go install github.com/bmaltais/agent-tools/cmd/gh-merge-wait@latest
go install github.com/bmaltais/agent-tools/cmd/gh-action-version@latest
go install github.com/bmaltais/agent-tools/cmd/git-pr-branch@latest
go install github.com/bmaltais/agent-tools/cmd/patch-verify@latest
go install github.com/bmaltais/agent-tools/cmd/issue-ship@latest
```

## Build

```bash
go build ./...
go test ./...
go vet ./...
```

## patch-verify

Safe literal replacement tool for agent-driven edits.

Usage:

```bash
patch-verify [--all] [--dry-run] <file|- > <old-string> <new-string>
```

Behavior:

- Replaces the first occurrence by default
- Fails non-zero when no match is found
- Fails non-zero when multiple matches are found unless `--all` is set
- `--all` replaces all occurrences
- `--dry-run` prints the unified diff without writing the file
- Colorizes diff output when writing to a terminal; emits plain diff when piped
- With `<file>` set to `-`, reads from stdin and writes diff plus patched content to stdout

## Design Principles

- Single static binary per tool — no Node, Python, or JVM
- Sub-100ms execution for typical inputs
- Composable with standard Unix pipes
- Authenticated via `gh auth` / `GITHUB_TOKEN` — no extra credential setup

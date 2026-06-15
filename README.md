# agent-tools

Fast native CLI tools for AI agent sessions. Single static Go binaries, no runtime dependencies, composable via Unix pipes.

## Tools

| Tool | Status | Purpose |
|------|--------|---------|
| [`gh-merge-wait`](./cmd/gh-merge-wait/) | planned | Squash-merge a PR with automatic retry/poll — no manual polling loops |
| [`gh-action-version`](./cmd/gh-action-version/) | planned | Look up the latest Node.js-24-compatible version of GitHub Actions |
| [`git-pr-branch`](./cmd/git-pr-branch/) | planned | Open/close PR branches against main with one command |
| [`patch-verify`](./cmd/patch-verify/) | implemented | Apply a safe literal string replacement and print a unified diff |
| [`issue-ship`](./cmd/issue-ship/) | implemented | Drive the full branch → PR → merge → cleanup pipeline for a GitHub issue |

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

## Registry

Agent discovery metadata lives in a dense machine-readable registry.

- Canonical source: `tools/registry.json`
- Key legend: `tools/schema-legend.md`
- Schema contract: `tools/registry.schema.json`
- Copilot runtime overlay: `tools/overlays/copilot.json`

Validation is enforced in CI via `.github/workflows/registry-ci.yml`.

## Releases

Releases are built from tags on main with `.github/workflows/release.yml`.

- Build matrix: linux/amd64, linux/arm64, darwin/arm64
- Artifact naming: `agent-tools_<tool>-<version>_<os>_<arch>`
- Integrity outputs: `SHA256SUMS` and `release-manifest.json`

## Planning

- V1 milestones: `docs/V1-PLAN.md`
- Execution checklist: `docs/IMPLEMENTATION-CHECKLIST.md`

## issue-ship

Idempotent pipeline driver for a GitHub issue: branch → PR → merge → cleanup.
Re-running from any stage picks up where it left off.

Usage:

```bash
issue-ship [--dry-run] [--method squash|rebase|merge] [--from-stage branch|pr|merge|cleanup] <owner/repo> <issue-number>
```

Behavior:

- Detects the current pipeline stage automatically (branch, PR open, PR merged, cleanup) and advances from there
- Branch name: `feat/issue-{N}-{slug}` where slug is the first 5 words of the issue title, kebab-cased
- PR body includes `Closes #{N}` to auto-close the issue on merge
- `--method` controls the merge strategy (default: `squash`)
- `--dry-run` prints each stage action without executing any git or gh commands
- `--from-stage` forces a restart from a named stage, useful for recovery

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

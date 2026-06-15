# agent-tools

Fast native CLI tools for AI agent sessions. Single static Go binaries, zero runtime dependencies, composable via Unix pipes.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/bmaltais/agent-tools/main/install.sh | bash
```

- Installs to `~/.local/bin` — make sure that directory is on your `$PATH`
- Only **GA-status** tools are installed; tools in `plan` or other pre-release states are skipped
- Override the install directory: `AGENT_TOOLS_BIN_DIR=/usr/local/bin bash install.sh`

## Tools

| Tool | Status | Purpose |
|------|--------|---------|
| [`gh-merge-wait`](./cmd/gh-merge-wait/) | implemented | Squash-merge a PR and poll until merged, with automatic 502 retry and REST fallback |
| [`gh-action-version`](./cmd/gh-action-version/) | implemented | Look up the latest release of GitHub Actions matching a given runtime (e.g. node24) |
| [`git-pr-branch`](./cmd/git-pr-branch/) | implemented | Open/close PR branch lifecycle as two single commands |
| [`patch-verify`](./cmd/patch-verify/) | implemented | Apply a safe literal string replacement and print a unified diff |
| [`issue-ship`](./cmd/issue-ship/) | implemented | Drive the full branch → PR → merge → cleanup pipeline for a GitHub issue |
| [`gh-triage-apply`](./cmd/gh-triage-apply/) | implemented | Apply labels and post a triage comment to one or more GitHub issues in one command |
| [`gh-pr-comments`](./cmd/gh-pr-comments/) | implemented | Fetch inline and conversation PR comments as a unified JSON array |

## Install from source

```bash
go install github.com/bmaltais/agent-tools/cmd/gh-merge-wait@latest
go install github.com/bmaltais/agent-tools/cmd/gh-action-version@latest
go install github.com/bmaltais/agent-tools/cmd/git-pr-branch@latest
go install github.com/bmaltais/agent-tools/cmd/patch-verify@latest
go install github.com/bmaltais/agent-tools/cmd/issue-ship@latest
go install github.com/bmaltais/agent-tools/cmd/gh-triage-apply@latest
go install github.com/bmaltais/agent-tools/cmd/gh-pr-comments@latest
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

## gh-triage-apply

Apply a label set and post a triage comment to one or more GitHub issues in a single command.

Usage:

```bash
gh-triage-apply [--dry-run] --labels <label,...> (--comment <text> | --comment-file <path|->) <owner/repo> <issue-number> [<issue-number>...]
```

Behavior:

- Accepts one or more issue numbers; processes each sequentially
- All errors are collected and reported — execution continues past individual failures (fail-aggregate, not fail-fast)
- Applies labels via `gh issue edit --add-label` then posts the comment via `gh issue comment`, per issue
- `--comment` provides the body inline; `--comment-file <path>` reads from a file; `--comment-file -` reads from stdin (body is read once, reused for all issues)
- `--dry-run` prints what would be done without calling `gh`
- Exits non-zero if any issue failed, including the `gh` error output
- Authenticated via `gh auth` / `GITHUB_TOKEN` — no extra credential setup

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

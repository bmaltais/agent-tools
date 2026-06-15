# V1 Plan

Goal: ship Copilot-first AI-native CLI tool ecosystem with dense machine-readable discovery, strict registry validation, and tag-driven verifiable releases.

## Milestone 1: Registry Foundation

Exit criteria:
- tools/registry.json exists as canonical source of truth
- tools/schema-legend.md defines compact keys
- tools/registry.schema.json validates registry shape
- tools/overlays/copilot.json defines Copilot runtime adapter hints

Deliverables:
- Dense key registry records for all current tools
- Schema major field and semantic version in registry
- Agent-neutral core + runtime overlay split

## Milestone 2: CI Enforcement

Exit criteria:
- CI fails on invalid registry schema
- CI fails if registry references unknown binaries
- CI fails on schema major mismatch between registry and overlay

Deliverables:
- GitHub workflow for strict registry validation on PR and main
- Deterministic validation script embedded in CI

## Milestone 3: Release Pipeline

Exit criteria:
- Tag push on main builds binaries for linux/amd64, linux/arm64, darwin/arm64
- Artifact names match contract: agent-tools_<tool>-<version>_<os>_<arch>
- SHA256SUMS generated and published
- release-manifest.json published with schema major + registry hash

Deliverables:
- Tag-triggered release workflow
- Main-ancestor guard for tags
- Release assets uploaded to GitHub Release

## Milestone 4: Copilot Consumption Path

Exit criteria:
- Session startup can load local registry + legend
- Optional remote refresh checks manifest schema major + registry hash
- Refresh failure keeps last-known-good and marks stale state

Deliverables:
- Loader contract doc for Copilot integration
- Runtime stale-state warning payload contract
- High-risk operation gate list (install/update/registry mutation/remote adapter generation)

## Milestone 5: Operational Hardening

Exit criteria:
- Registry evolution policy documented and enforced (major/minor rules)
- Basic adapter compatibility tests exist for supported schema major
- Release manifest consumers can verify artifact integrity from one fetch

Deliverables:
- Compatibility checks for adapter major support
- Regression tests for parser/loader behavior
- Playbook for v1.1 windows matrix expansion

## Non-goals (v1)

- Human-first UX optimization
- Signature attestations and policy enforcement (v2)
- Immediate multi-runtime parity beyond Copilot overlays
- Full automation of all high-risk stale-state override workflows

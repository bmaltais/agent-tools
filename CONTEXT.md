# Agent Tools

This context defines the language for a repository whose purpose is to reduce AI agent tool-call count through purpose-built native CLIs.

## Language

**AI Agent**:
An autonomous coding/runtime assistant that performs tasks by invoking tools and commands.
_Avoid_: User, operator

**Agent Tool**:
A CLI command designed to collapse multi-step agent workflows into fewer calls with deterministic behavior.
_Avoid_: Script, helper

**Copilot Runtime**:
The first supported AI agent environment where Agent Tools must be discoverable and usable.
_Avoid_: Generic IDE support

**Cross-Agent Support**:
Future compatibility of Agent Tools across multiple AI agent runtimes beyond Copilot.
_Avoid_: Immediate multi-runtime parity

**AI-First Interface**:
A command interface optimized for machine invocation over interactive human ergonomics.
_Avoid_: Human-first UX

**Release Artifact**:
A versioned binary published from tagged main and intended as the canonical installation unit.
_Avoid_: Ad-hoc local builds

**Installation Surface**:
The supported path an agent runtime uses to acquire Agent Tools.
_Avoid_: Unspecified install method

**Tool Metadata Contract**:
A machine-readable description of each Agent Tool used by runtimes to decide selection and invocation.
_Avoid_: Prose-only docs, hardcoded rules

**Dense Metadata Style**:
Minimal-token, high-signal metadata phrasing optimized for LLM parsing rather than human readability.
_Avoid_: Fluffy explanations, narrative prose

**Tool Registry**:
A single canonical JSON file containing the full Tool Metadata Contract set for all Agent Tools.
_Avoid_: Split metadata across prose files

**Runtime Overlay**:
A thin runtime-specific mapping layer that adapts core Tool Registry records to a given agent runtime contract.
_Avoid_: Runtime-specific fields in core registry

**Compact Key Schema**:
A Tool Registry encoding style that uses short field keys and dense values to minimize token usage.
_Avoid_: Verbose field names

**Schema Legend**:
A minimal read-once mapping document that defines Compact Key Schema field meanings.
_Avoid_: Repeating verbose explanations inside registry records

**Registry CI Gate**:
A mandatory CI validation step that fails builds when Tool Registry schema or tool binary references are invalid.
_Avoid_: Warning-only validation

**Registry Schema Version**:
A semantic version value that defines compatibility expectations for the entire Tool Registry contract.
_Avoid_: Unversioned registry

**Adapter Compatibility Contract**:
Runtime adapters declare supported registry major version and fail fast on unsupported major values.
_Avoid_: Silent fallback across breaking schema changes

**Release Integrity Baseline**:
Release distribution includes binaries, SHA256 checksums, and a machine-readable manifest for deterministic verification.
_Avoid_: Binary-only releases

**Release Manifest**:
A machine-readable release index containing binary identity and verification metadata for runtime installers/adapters.
_Avoid_: Human-only release notes as install metadata source

**Release Matrix v1**:
The minimum supported build target set for v1 release artifacts.
_Avoid_: Unbounded target expansion in initial workflow

**Artifact Naming Contract**:
Deterministic machine-parseable file naming for release binaries.
_Avoid_: Ad-hoc naming patterns

**Session Load Policy**:
Runtime boot behavior for loading Tool Registry and Schema Legend at session start.
_Avoid_: Network-only startup dependency

**Stale Runtime State**:
A runtime condition where local registry is used after remote refresh failure or unresolved update drift.
_Avoid_: Silent unknown freshness

**Structured Refresh Warning**:
A machine-parseable warning emitted when refresh/update verification fails while continuing on last-known-good metadata.
_Avoid_: Unstructured log-only warning

**Operation-Risk Gating**:
An enforcement model where stale-state blocking applies only to explicitly high-risk operations.
_Avoid_: Global stale-state blocking

## Relationships

- An **AI Agent** invokes one or more **Agent Tool** commands to complete a task.
- **Copilot Runtime** is the initial target environment for **Agent Tool** discoverability and usage.
- **Cross-Agent Support** extends the same **Agent Tool** set to additional runtimes after Copilot-first goals are met.
- **AI-First Interface** constrains how each **Agent Tool** is designed and documented.
- **Release Artifact** is the canonical **Installation Surface** for runtime consumption.
- `go install` is a fallback **Installation Surface** for development and local experimentation.
- **Copilot Runtime** selects an **Agent Tool** using the **Tool Metadata Contract**.
- **Tool Metadata Contract** uses **Dense Metadata Style** and includes intent, command, inputs, outputs, preconditions, side effects, failure modes, and examples.
- **Tool Registry** is the design-time source of truth for metadata and runtime discovery.
- Runtime help output is execution-time reference; it does not replace **Tool Registry**.
- **Tool Registry** stays agent-neutral and runtime-independent.
- **Runtime Overlay** carries Copilot-specific and future runtime-specific execution hints.
- Runtime adapters are generated or loaded at session start from **Tool Registry** + **Runtime Overlay**.
- **Tool Registry** uses **Compact Key Schema** with dense values.
- **Schema Legend** is loaded once per session before runtime tool selection.
- **Registry CI Gate** enforces schema validity and binary-reference integrity before merge/release.
- **Registry Schema Version** governs evolution for the full registry.
- Additive schema changes map to minor version increments; breaking key changes/removals map to major increments.
- **Adapter Compatibility Contract** requires explicit major-version support declaration.
- **Release Integrity Baseline** is required in v1 for all tagged releases.
- **Release Manifest** includes binary name, os, arch, sha256, registry schema major, and registry hash.
- **Release Matrix v1** is linux/amd64, linux/arm64, and darwin/arm64.
- **Artifact Naming Contract** is `agent-tools_<tool>-<version>_<os>_<arch>`.
- **Session Load Policy** is local-first with optional remote refresh when version/hash drift is detected.
- Freshness checks use manifest schema major and registry hash.
- On refresh failure, runtime continues with local last-known-good registry and marks **Stale Runtime State**.
- A **Structured Refresh Warning** is emitted whenever stale continuation is entered.
- **Operation-Risk Gating** keeps normal tool execution non-blocking while applying stale-state enforcement to high-risk operations.

## Example dialogue

> **Dev:** "Should we prioritize human-friendly prompts in this command?"
> **Domain expert:** "No. This repo is **AI-First Interface**. Prioritize deterministic machine use in **Copilot Runtime** first, then expand via **Cross-Agent Support**."

## Flagged ambiguities

- "user" was used to mean both human maintainers and autonomous agents — resolved: the primary runtime actor is **AI Agent**.
- "support all agents" could imply immediate parity — resolved: **Copilot Runtime** first, then **Cross-Agent Support** in later phases.
- "install path" was ambiguous between source-based and binary-based distribution — resolved: **Release Artifact** is canonical, with `go install` as fallback.
- "tool choice logic" was ambiguous between static prompts and semantic matching — resolved: selection is driven by **Tool Metadata Contract**.
- "good metadata" was ambiguous between human-documentation style and machine-optimized style — resolved: use **Dense Metadata Style**.
- "metadata location" was ambiguous between binary help text and repository data — resolved: use single **Tool Registry** at `tools/registry.json`.
- "schema scope" was ambiguous between Copilot-first fields and long-term portability — resolved: use agent-neutral core schema plus optional **Runtime Overlay** per runtime.
- "schema readability" was ambiguous between long keys and token-efficient keys — resolved: use **Compact Key Schema** plus one **Schema Legend**.
- "registry validation" was ambiguous between warning and enforcement — resolved: use strict **Registry CI Gate** that fails on invalid schema or unknown binaries.
- "schema evolution" was ambiguous between flexible parsing and strict compatibility — resolved: use **Registry Schema Version** with major/minor rules and explicit **Adapter Compatibility Contract**.
- "release trust model" was ambiguous between binaries-only and verifiable distribution — resolved: use **Release Integrity Baseline** with **Release Manifest** and checksums; signatures deferred to v2.
- "release target scope" was ambiguous between broad parity and minimal coverage — resolved: use **Release Matrix v1** for linux/amd64, linux/arm64, darwin/arm64; defer windows to v1.1.
- "artifact naming" was ambiguous between human-friendly and machine-parseable patterns — resolved: enforce **Artifact Naming Contract**.
- "registry load source" was ambiguous between local-only and network-first — resolved: use **Session Load Policy** with local cache first and verified optional remote refresh.
- "refresh failure behavior" was ambiguous between fail-fast and silent fallback — resolved: continue with last-known-good local registry, set **Stale Runtime State**, emit **Structured Refresh Warning**.
- "stale enforcement scope" was ambiguous between time-based global lock and selective safety checks — resolved: use **Operation-Risk Gating** with stale blocking only for high-risk operations.

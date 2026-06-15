# Schema Versioning Policy

This document defines the rules for evolving the Tool Registry schema (`tools/registry.json`) safely across versions. Runtime adapters (overlays) declare the schema major they were built for; a mismatch between registry `sm` and overlay `sm` causes an immediate load failure — there is no silent fallback.

## Schema Version Fields

The registry carries two version fields:

| Field | Key | Meaning |
|-------|-----|---------|
| Semantic version | `sv` | Full semver string (`MAJOR.MINOR.PATCH`) |
| Schema major | `sm` | Integer; drives adapter compatibility checks |

The overlay (`tools/overlays/<runtime>.json`) carries a matching `sm` field. The loader enforces `registry.sm == overlay.sm` on every load.

## When to Bump Major (`sm`)

Bump `sm` (and the `MAJOR` component of `sv`) when a change **breaks existing adapter logic** — i.e., an adapter built against the previous schema major would silently mis-parse, skip, or misroute tool records.

### Breaking changes (require major bump)

- **Removing a field** that adapters read (e.g. dropping `bin` from tool records)
- **Renaming a key** (e.g. renaming compact key `id` → `name`)
- **Changing the type** of an existing field (e.g. `sm` from `int` to `string`)
- **Changing the semantics** of an existing field without renaming it (e.g. `risk` values gaining a new meaning that old adapters would map incorrectly)
- **Restructuring a nested object** in a way that breaks existing field paths (e.g. flattening `a[].k` / `a[].t` into a different shape)

**Example:** Removing the `bin` field from tool records and replacing it with a `bins` array would break any adapter that reads `t[].bin` to resolve the binary path. This requires a major bump: `sm: 1 → 2`, `sv: 1.x.y → 2.0.0`.

### Non-breaking changes (minor bump only)

- **Adding a new optional field** that old adapters will simply ignore
- **Adding a new tool record** to the `t` array
- **Extending an enum** with a new allowed value (e.g. a new `risk` level) where old adapters treat unknown values safely (e.g. by defaulting to high risk)
- **Updating metadata-only fields** that do not affect routing logic (e.g. `ts` timestamp, `in` description text)

**Example:** Adding an optional `tags []string` field to tool records is non-breaking — adapters that do not know the field will skip it. Minor bump only: `sv: 1.0.0 → 1.1.0`, `sm` stays `1`.

## Changelog Entry Format for Registry-Breaking Changes

When a breaking change is published, the `CHANGELOG.md` entry **must** include a machine-parseable `BREAKING` block in addition to human prose. Adapters and release tooling may scan for this block.

### Format

```
## [X.0.0] - YYYY-MM-DD

### Breaking Changes

<!-- BREAKING schema_major=X -->
- field: <compact-key-path>
  change: <removed|renamed|type-changed|semantic-changed|restructured>
  from: <previous shape or value, one line>
  to: <new shape or value, one line>
  migration: <what adapter authors must do>
<!-- /BREAKING -->
```

### Rules

- The `<!-- BREAKING schema_major=X -->` open tag and `<!-- /BREAKING -->` close tag are **required** on their own lines.
- `schema_major` must equal the new `sm` value in `registry.json`.
- Each broken field is a YAML-style list entry under the tag block.
- `field` uses dot-notation with array wildcard: e.g. `t[].bin`, `t[].a[].k`.
- `change` is one of: `removed`, `renamed`, `type-changed`, `semantic-changed`, `restructured`.
- `migration` is a single imperative sentence describing what adapter authors must change.

### Example

```
## [2.0.0] - 2027-01-15

The `bin` field on tool records has been replaced by a `bins` map to support
multi-binary tools. All adapters must update binary resolution logic.

### Breaking Changes

<!-- BREAKING schema_major=2 -->
- field: t[].bin
  change: removed
  from: "bin": "<string>"
  to: "bins": {"<os>/<arch>": "<string>"}
  migration: Replace t[i].bin reads with t[i].bins[os+"/"+arch] lookups.
<!-- /BREAKING -->
```

## Adapter Compatibility Enforcement

The loader (`internal/registry`) enforces compatibility at load time:

1. **Local load**: `registry.sm` must equal `overlay.sm`. Any mismatch returns an error immediately — the session does not start.
2. **Remote refresh**: The remote manifest's `schema_major` must equal the locally-loaded registry `sm`. A mismatch marks the state as stale with code `refresh_schema_major_mismatch` and aborts the refresh.

Adapter authors who publish overlays for a new schema major must also ship an updated overlay with a matching `sm` value. There is no version negotiation — fail fast is the contract.

## v1.1 Roadmap

<!-- TODO v1.1: windows matrix expansion -->
The v1 release matrix covers `linux/amd64`, `linux/arm64`, and `darwin/arm64`. Support for `windows/amd64` is planned for v1.1. When adding the Windows target, the release workflow (`release.yml`) build matrix must be extended and the smoke-check action must include a Windows runner. No schema changes are required for this expansion.

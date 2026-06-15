# Implementation Checklist

Use this as issue-ready execution slices.

## Slice 1: Registry Artifacts

- [x] Add tools/registry.json (canonical dense source)
- [x] Add tools/schema-legend.md (read-once key map)
- [x] Add tools/registry.schema.json (validation contract)
- [x] Add tools/overlays/copilot.json (runtime overlay)
- [x] Add parser/loader package for registry + overlay
- [x] Add unit tests for registry parsing and duplicate tool-id detection

## Slice 2: CI Gate

- [x] Add .github/workflows/registry-ci.yml
- [x] Enforce schema validation with jsonschema
- [x] Enforce binary path existence check for all registry tool records
- [x] Enforce registry schema major == overlay schema major
- [x] Add local make target or script for parity with CI validation

## Slice 3: Release Workflow

- [x] Add .github/workflows/release.yml
- [x] Enforce tag commit ancestry on main
- [x] Build matrix: linux/amd64, linux/arm64, darwin/arm64
- [x] Enforce artifact naming contract
- [x] Generate SHA256SUMS
- [x] Generate release-manifest.json with schema major + registry hash
- [x] Publish release assets
- [ ] Add post-release smoke check action (download + verify one binary)

## Slice 4: Copilot Runtime Integration

- [x] Define session-start loader behavior (local first, optional remote refresh)
- [x] Implement stale-state marker and structured warning payload
- [ ] Implement operation-risk gating for high-risk actions
- [ ] Define and implement manual override contract with audit fields
- [x] Add integration test for stale refresh fallback path

## Slice 5: Governance and Evolution

- [ ] Add schema version bump policy document and examples
- [ ] Add adapter compatibility tests by schema major
- [ ] Add changelog entry format for registry-breaking changes
- [ ] Add v1.1 issue for windows matrix expansion

## Suggested Issue Titles

- feat: add registry loader with overlay merge
- test: add registry parser and compatibility tests
- chore: add local registry validate command
- feat: add stale-state and structured warning runtime contract
- feat: implement risk-gated operation policy
- docs: define schema bump and compatibility policy
- feat: add release smoke verification step

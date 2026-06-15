---
name: Registry Loader Slice
about: Implement registry loader + overlay merge + stale signaling
labels: [enhancement, ready-for-agent]
---

## Summary
Implement runtime loader for local-first registry + overlay with optional refresh and stale-state signaling.

## Scope
- [ ] Loader reads tools/registry.json + tools/overlays/<runtime>.json
- [ ] Optional remote refresh path with hash/schema checks
- [ ] Last-known-good fallback on refresh failure
- [ ] Structured stale warning payload

## Acceptance Criteria
- [ ] Local-only load succeeds offline
- [ ] Remote refresh updates state when manifest hash differs and verification passes
- [ ] Refresh failure keeps local state and marks stale=true
- [ ] Unit tests cover success and fallback branches

## Out of Scope
- Full runtime command wiring
- Signature verification

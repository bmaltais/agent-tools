---
name: Release Hardening Slice
about: Add release smoke verification and governance follow-ups
labels: [enhancement, ready-for-agent]
---

## Summary
Harden release pipeline and governance around registry compatibility.

## Scope
- [ ] Post-release smoke check (download + checksum verify + basic execution)
- [ ] Adapter compatibility tests by schema major
- [ ] Schema bump policy doc with major/minor examples
- [ ] Track windows matrix expansion for v1.1

## Acceptance Criteria
- [ ] Release smoke step verifies at least one published binary per tag
- [ ] Compatibility tests fail on unsupported schema major
- [ ] Governance docs describe breaking vs additive changes
- [ ] Windows expansion issue exists with clear entry criteria

## Out of Scope
- New runtime adapter implementation
- v2 signature attestations

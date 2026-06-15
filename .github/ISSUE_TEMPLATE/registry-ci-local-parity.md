---
name: Registry CI Parity Slice
about: Keep local validate command and CI gate identical
labels: [enhancement, ready-for-agent]
---

## Summary
Ensure strict registry validation logic is shared across local and CI paths.

## Scope
- [ ] Single validation script used by both Makefile and GitHub workflow
- [ ] Schema validation
- [ ] Binary path existence check for each registry tool
- [ ] Overlay-to-registry tool mapping check
- [ ] Schema major consistency check

## Acceptance Criteria
- [ ] `make validate-registry` passes locally when metadata is valid
- [ ] CI fails on invalid schema, unknown binary, or major mismatch
- [ ] No duplicated validation logic between local and CI

## Out of Scope
- Runtime loader behavior
- Release packaging

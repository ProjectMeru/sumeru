## Summary

<!-- What changed and why -->

## Checklist

- [ ] `make` from `sumeru/` (lint, test, build)
- [ ] Sensitive fields → `core/security/fields.go`; bypass → `AuditedBypass` / `WithElevated`
- [ ] Auth/session cookies → `core/server/web/cookie_helpers.go`
- [ ] Tests under `test/` or `core/swc/tests/` where behavior changed

**If applicable:** `make generate` · `make test-modules` · `make test-integration`

## Notes

<!-- Breaking changes, follow-ups, linked issues -->

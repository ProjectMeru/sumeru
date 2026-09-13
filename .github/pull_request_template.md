## Summary

<!-- What changed and why? -->

## Standards checklist

- [ ] `cd sumeru && make check` (lint + security gates + tests)
- [ ] `make update-graphify` from monorepo root (`sumeru_erp/`) if Go or SWC changed
- [ ] `rg -i 'odoo|openerp' sumeru/ --glob '!**/node_modules/**'` — no branding hits
- [ ] New sensitive fields added only in `sumeru/core/security/fields.go`
- [ ] Security bypass uses `orm.AuditedBypass` / `WithElevated` (not raw `ContextWithBypass`)
- [ ] Tests added under `sumeru/test/` (see monorepo `AGENTS.md`)

## Verification

- [ ] `make swc-test` when `core/swc/` changed
- [ ] `make generate` when imports or `sumeru.conf.example` addons path changed

**Areas touched:** <!-- e.g. ORM, server/web, SWC, module loader, addons/base -->

## CI

Pull requests run **Go build**, **Go test**, **Go lint** (incl. security bypass script), **SWC check/test**, and **generate drift** (`cmd/sumeru/zimports.go`).

**Integration** (PostgreSQL + install `base` + tagged tests) runs on push to `main` or `dev` only.

## Notes

<!-- Breaking changes, follow-ups, linked issues -->

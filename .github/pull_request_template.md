## Summary

<!-- What changed and why? -->

## Verification

- [ ] `make check` (or `make swc-check` + `go test ./...`) from the `sumeru` module root
- [ ] `make swc-test` when `core/swc/` changed
- [ ] `make generate` when imports or `sumeru.conf.example` addons path changed

**Areas touched:** <!-- e.g. ORM, server/web, SWC, module loader, addons/base -->

## CI

Pull requests run **Go build**, **Go test**, **SWC check/test**, and **generate drift** (`cmd/sumeru/zimports.go`).

**Integration** (PostgreSQL + install `base` + tagged tests) runs on push to `main` or `dev` only.

## Notes

<!-- Breaking changes, follow-ups, linked issues -->

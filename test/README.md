# Sumeru tests

All Go test cases live under `test/` only. The only exception is `export_test.go` files colocated with source packages — they expose unexported symbols to external tests, not test cases themselves.

SWC tests live under `core/swc/tests/` (see `core/swc/tests/README.md`).

## Layout

```
test/
  harness/          Shared helpers (RepoRoot, ActivateModules, temp addons)
  module/           Module test suite (static, unit, addons, integration tiers)
  core/             Tests for core/ packages (migrated from colocated *_test.go)
  addons/           Tests for addon packages
  acceptance/       Platform capability smoke checks
  integration/      DB-gated tests (-tags=integration)
  ...               Domain packages (api, orm, parser, web, etc.)
```

## Running tests

```bash
make check                 # SWC coverage + module static gate + go test ./test/...
make test-modules          # Module suite tiers 0–2 (static + unit + addon)
make test-modules-static   # Convention validation for all discovered addons
make test-integration      # PostgreSQL integration (requires Docker)
make test-coverage         # Full repo coverage profile
```

## Module test suite tiers

| Tier | Path | Makefile target |
|------|------|-----------------|
| 0 Static | `test/module/static/` | `test-modules-static` |
| 1 Unit | `test/module/unit/` | `test-modules-unit` |
| 2 Addon | `test/module/addons/`, `test/addons/` | `test-modules-addon` |
| 3 Integration | `test/module/integration/`, `test/integration/` | `test-modules-integration` |

## Coverage policy

- **Target:** 90% statement coverage on `go test ./test/... -coverpkg=./...` (full repo)
- **Current baseline:** ~42% (unit + sqlmock); DB-backed ORM/module/web paths need PostgreSQL integration CI
- **SWC:** 90% on scoped globs in `core/swc/vitest.config.ts`; CI runs `npm run test:coverage`
- **Go gate:** `scripts/check_go_coverage.sh` — default `GO_COVERAGE_MIN=90`; CI uses `GO_COVERAGE_MIN=42` until integration tier is enabled
- **Measure:** `make test-coverage` or `GO_COVERAGE_MIN=90 make test-coverage` for strict gate

## Writing new tests

1. Add `*_test.go` under the matching `test/` subtree — never under `core/` or `addons/`
2. Use `test/harness` for repo root, model activation, and temp addon fixtures
3. For unexported symbols, add exports to the source package's `export_test.go`
4. Integration tests must use `//go:build integration`

# Sumeru tests

Standards index: [AGENTS.md](../../AGENTS.md) (monorepo root). Test layout rule: [test-layout.mdc](../../.cursor/rules/test-layout.mdc).

All Go test cases live under `test/` only. Colocated `testexports.go` files in source packages expose symbols for external tests in `test/core/*` and `test/addons/*`; they are not test cases themselves.

SWC tests live under `core/swc/tests/` (see `core/swc/tests/README.md`).

## Layout policy (golden rule)

| Kind | Allowed location |
|------|------------------|
| Go `*_test.go` | `sumeru/test/**` only |
| SWC `*.test.ts` | `sumeru/core/swc/tests/**` only |
| Test export hooks | `**/testexports.go` next to source (not tests) |

CI and local `make standards` run [`scripts/check_test_layout.sh`](../scripts/check_test_layout.sh) to reject colocated Go tests or SWC tests under `core/swc/src/`.

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
make                       # full standard: lint + SWC + Go tests + coverage gate + build
make check                 # same as make
make test-modules          # Module suite tiers 0–2 (static + unit + addon)
make test-modules-static   # Convention validation for all discovered addons
make test-integration      # PostgreSQL integration (requires Docker)
make test-go               # Full repo coverage profile + gate
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
- **Measure:** `make test-go` or `GO_COVERAGE_MIN=90 make test-go` for strict gate

## Writing new tests

1. Add `*_test.go` under the matching `test/` subtree — never under `core/` or `addons/`
2. Use `test/harness` for repo root, model activation, and temp addon fixtures
3. For symbols needed only from external tests, add exports to the source package's `testexports.go`
4. Integration tests must use `//go:build integration`

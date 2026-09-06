# Contributing to Sumeru

Thanks for helping improve Sumeru. This guide covers where to put changes, how to develop locally, and what we expect in pull requests.

By participating, you agree to follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Where to put your work

Use the right repository so core and standard apps stay pullable for everyone:

| Change type                                                       | Repository                 | Notes                                                                                      |
| ----------------------------------------------------------------- | -------------------------- | ------------------------------------------------------------------------------------------ |
| Engine, ORM, server, web shell, kernel addons (`base`, `mail`, …) | **`sumeru`** (this repo)   | Prefer `sumeru/core/sdk` from addon Go code; avoid new direct imports of `sumeru/core/orm` |
| Shared business apps (CRM, Sales, Inventory, …)                   | **`sumeru_addons`**        | Depends only on `sumeru`                                                                   |
| Customer-specific modules, branding, local runner                 | **`sumeru_custom_addons`** | Keep custom code under `addons/`; do not fork core for one-off features                    |

Most application teams **pull** `sumeru` and `sumeru_addons` and develop only in `sumeru_custom_addons`.

## Development setup

1. Clone the three siblings (see [README.md](README.md#quick-start-recommended)).
2. Create a PostgreSQL database and configure `sumeru_custom_addons/sumeru.conf`.
3. From `sumeru_custom_addons`:

   ```bash
   make replace-sumeru
   make replace-sumeru-addons
   make generate
   make run
   ```

4. After pulling core or standard addons:

   ```bash
   cd ../sumeru && git pull
   cd ../sumeru_addons && git pull
   cd ../sumeru_custom_addons && make generate
   ```

Core-only work can use `make generate` / `make run` from this repo (see README).

## Code generation

- **This repo:** `make generate` runs `go generate ./cmd/sumeru` and refreshes `cmd/sumeru/zimports.go` from `sumeru.conf.example`.
- **Custom workspace:** `make generate` writes `addonimports/zimports.go` from that workspace’s INI — do not generate custom imports into the `sumeru` tree.

Scaffold a new **core-tree** addon:

```bash
make bp NAME=my_module
make generate
```

For custom modules, from **`sumeru_custom_addons`**: `make new MODULE=my_module` (runs `sumeru-bp` then `make generate`).

## Runtime and globals

Prefer `sumeru/core/runtime.Runtime` for new injectable surfaces (DB, registry, events). Package-level singletons (`orm.DB`, `orm.Registry`, …) remain for bootstrap compatibility, but **do not add new package-level process globals** — put shared state on `Runtime` (or an explicit constructor) and migrate call sites incrementally. Tests should construct an isolated `runtime.New(...)` when practical.

## Dead code policy

Remove a symbol only when:

1. It has zero in-repo callers (or only self-references).
2. A documented replacement already exists (e.g. `BuildWhereWithRecordRules` replaced `MergeRuleDomainsIntoSearch`).
3. Public SDK / extension hooks (`core/sdk`, `RegisterObjectAction`) are kept even if unused in-tree.

Do not delete half-built features that still write data (e.g. outbox enqueue) unless the feature is abandoned.

## Logging

Use **`sumeru/core/applog`** only: `Info` / `Warn` / `Debug` / `Error` with `Event`, or the thin helpers `InfoMsg`, `WarnMsg`, `DebugMsg`, **`ErrorCode`**, and **`WarnCode`**. Failures should carry a stable `error_code` (see `sumeru/core/errcode`) plus a human `message`; never log passwords, tokens, sids, or API keys (context is auto-scrubbed). Before `SetupFromConfig` (config load, path resolve), use `BootstrapFatal` for fatal errors. See the logging guide in `sumeru_docs/core/guides/logging.md`. Do not import stdlib `log` or call `fmt.Printf` for operational logging in `core/server` or `core/module`. Stdout is always on when logging is enabled; `log_file` is optional. Do not add Zap or other logging libraries.

## Testing

From the `sumeru` module root, before opening a PR:

```bash
make lint          # swc-check + go vet + golangci-lint
go test ./test/... -count=1
go build ./...
```

Add or update tests under `test/` when you change ORM, server, or module behavior.

## Continuous integration

GitHub Actions runs on every pull request and push to `main` / `dev`:

| Job          | What it checks                                            | Local equivalent                                        |
| ------------ | --------------------------------------------------------- | ------------------------------------------------------- |
| **Go build** | `go build ./...`                                          | `go build ./...`                                        |
| **Go test**  | `go test ./... -count=1`                                  | `go test ./test/...` or `make check`                    |
| **Go lint**  | `go vet` + golangci-lint                                  | `make lint` (includes SWC typecheck)                    |
| **Go vuln**  | `govulncheck ./...`                                       | `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` |
| **SWC**      | `npm run check` + coverage tests in `core/swc`            | `make swc-check` / `make swc-test`                      |
| **Generate** | `make generate` — `cmd/sumeru/zimports.go` must not drift | `make generate` then review diff                        |

On **push to `main` or `dev` only**, an **integration** job boots PostgreSQL, installs the `base` module with `sumeru.conf.ci`, and runs `go test -tags=integration ./test/integration/...`.

Reproduce integration locally:

```bash
# Option A: docker-compose.test.yml (port 5433 — adjust sumeru.conf.ci db_port)
docker compose -f docker-compose.test.yml up -d --wait
go run ./cmd/sumeru -- -c sumeru.conf.ci -i base --stop-after-init
make test-integration

# Option B: CI-style Postgres on localhost:5432 with sumeru.conf.ci as committed
```

See the [Actions tab](https://github.com/ProjectMeru/sumeru/actions/workflows/ci.yml) for workflow runs. Dependabot opens weekly Go/npm and monthly GitHub Actions update PRs.

## Pull requests

- Keep diffs focused; one concern per PR when practical.
- Match existing naming and layout (`sys.*` / `core.*`, addon folder = technical name).
- Do **not** commit local secrets, `sumeru.conf` with real passwords, or credentials.
- Do not commit generated custom-workspace `addonimports/` unless that project explicitly tracks them.
- Describe _why_ the change is needed and how you verified it (commands, ports, modules installed).
- For security-sensitive findings, follow [SECURITY.md](SECURITY.md) instead of a public PR discussion of exploits.

## Questions

Open a GitHub issue on [ProjectMeru/sumeru](https://github.com/ProjectMeru/sumeru) for design or bug discussion. For vulnerabilities, use the private process in [SECURITY.md](SECURITY.md).

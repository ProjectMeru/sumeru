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

See the [Actions tab](https://github.com/ProjectMeru/sumeru/actions/workflows/ci.yml) for workflow runs. Dependabot opens weekly Go/npm and monthly GitHub Actions update PRs.

## Security elevation

Kernel paths that must bypass record rules (module install, cron, outbox drain, setup bootstrap) must use **`orm.WithElevated(ctx, reason, fn)`** with a **stable reason string** — never user input, tokens, or dynamic SQL.

- **`WithElevated`** time-boxes work (15 minutes), sets bypass on the callback context, and writes `security_elevate` / `security_elevate_done` / `security_elevate_fail` rows to `sys.audit` (best-effort). User/RPC contexts must not carry bypass unless inside that boundary (`RejectSmuggledUserBypass`).
- **RPC and ORM CRUD** (`search`, `read`, `write`, `create`, `unlink`, `read_group`, `onchange`, `call`) enforce model ACL and record rules on the session/API-key context from `rpcSecurityContext` — never `ContextWithBypass` on the request.
- **`AuditedBypass`** is for ORM internals and code already running inside an elevated boundary (sequences, side effects, i18n). Do not call it on `r.Context()` in web handlers or at addon hook entry.
- **`BackgroundBypass`** in tests/config is not audited; do not use it in handlers or addons.
- CI enforces raw `ContextWithBypass` and addon `AuditedBypass` via [`scripts/check_security_bypass.sh`](scripts/check_security_bypass.sh).

| Reason | Use |
|--------|-----|
| `module.install` | Install / uninstall / activate modules |
| `cron.run` | Scheduler tick |
| `outbox.drain` | Outbox publisher |
| `automation.server_action` | Server action event hooks |
| `digest.hook` | Digest cron hook |
| `setup.bootstrap` | First-time setup handler |
| `schema.sync` | Registry schema sync |

## Developer mode (SWC)

System administrators see a **bug icon** in the web top bar (`features.debugMenu` in SWC bootstrap). Use it to:

- **Enable developer mode** (`?debug=1`) — client arch/RPC logging, debug drawer, optional field inspector.
- **Enable developer mode (assets)** (`?debug=assets`) — same with asset cache bust on reload.
- **Disable developer mode** — clears the query param and session flag.

With developer mode on, the top-bar **bug menu** shows sectioned actions (Record, User interface, Security, Tools). Each form field label gets an **info icon**; hover it for a technical popover (field, model, domain, modifiers). **Metadata** and **Data** open modals; **Access rights** opens the secondary debug drawer (collapsed by default). The **field inspector** toggle enables click-to-select on the field widget without blocking normal input when off. **SWC Vision** and **Open metrics** remain admin-gated.

## Realtime bus channels (SWC WebSocket)

Clients connect to `GET /web/swc/bus` (session cookie). Wire frames are JSON:

- Client → server: `subscribe` / `unsubscribe` (with `channels`, optional `last_event_id`), `ping`.
- Server → client: `event` (`id`, `channel`, `payload`), `pong`, `error`.

**Channel ACL** (server enforces before subscribe):

| Prefix | Rule |
|--------|------|
| `user/{uid}/…` | Session uid must match `{uid}` |
| `group/{xmlid}` | User must belong to group |
| `record/{model}/{id}` | Read access on record |
| `company/{id}` | User allowed company |
| `model/{model}` | Model read ACL |

Events are persisted in `sys.bus.event`; PostgreSQL `NOTIFY sumeru_bus` fan-out supports multiple app processes. Payloads carry record ids only — never field values from elevated writes.

## Mail thread and notifications

Models with `mail_thread` on the embedded model tag auto-subscribe creator and common assignee fields on create. Chatter uses `mail.message` subtypes (`mail.message.subtype`). Followers (`mail.follower`) drive `mail.notification` rows and optional HTML email via the mail queue. @mentions in chatter bodies notify mentioned users by login.

## Auth providers and MFA

Configure IdPs under **Settings → Security → Authentication providers** (`sys.auth.provider`; PKCE on start/callback). **Linked identities** lists `core.user.identity` rows. The login page shows enabled providers as SSO buttons. Local password login stays available unless system parameter `auth.local_enabled` is `false` and at least one provider is enabled. TOTP: users enroll under **Settings → Account security**; login uses `totp_enabled` / `totp_secret` with an HMAC-signed pending-MFA cookie before session creation; trusted devices use a signed cookie bound to the user id. SAML is not implemented — use OIDC-capable IdPs. When `jwks_url` is set, the callback verifies the `id_token` signature against JWKS (RS256/ES256) before linking the user.

## View modifier expressions (SWC)

Dynamic `invisible` / `readonly` / `required` expressions in form and list arch are evaluated client-side with a **frozen allowlist** of identifiers: record field names, `user_id`, `company_id`, and `context` (object). Expressions must be boolean JavaScript fragments (for example `state == 'done'`), not statements. Tokens such as `function`, `=>`, `[`, `` ` ``, or `;` are rejected. Static arch flags still apply when an expression is missing or invalid. List column expressions that reference record fields are evaluated without a row context (static arch flags apply). Action `context` on the workspace payload is not wired yet — `context` is an empty object until then.

## Pull requests

- Keep diffs focused; one concern per PR when practical.
- Match existing naming and layout (`sys.*` / `core.*`, addon folder = technical name).
- Do **not** commit local secrets, `sumeru.conf` with real passwords, or credentials.
- Do not commit generated custom-workspace `addonimports/` unless that project explicitly tracks them.
- Describe _why_ the change is needed and how you verified it (commands, ports, modules installed).
- For security-sensitive findings, follow [SECURITY.md](SECURITY.md) instead of a public PR discussion of exploits.

## Questions

Open a GitHub issue on [ProjectMeru/sumeru](https://github.com/ProjectMeru/sumeru) for design or bug discussion. For vulnerabilities, use the private process in [SECURITY.md](SECURITY.md).

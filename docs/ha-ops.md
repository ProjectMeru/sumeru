# High availability operations (Phase 6)

Sumeru sessions are **DB-backed** (`sys.session`), so sticky sessions are not required for auth.

## Requirements before multi-instance

1. **`csrf_secret` / `SUMERU_CSRF_SECRET`** — shared across all app processes (see `sumeru.conf.example`).
2. **Shared PostgreSQL** — one primary; optional `db_read_replica_dsn` for read RPCs.
3. **`metrics_scrape_token`** — scrape `/metrics` without an admin browser session.
4. **Probes** — liveness `GET /api/health`; readiness `GET /api/ready` (DB ping).
5. **Queue** — in-process pub/sub does not cross instances. Use `queue.SetPublishMirrorHook` (or an external broker) so outbox/bus events fan out; WebSocket bus remains per-instance unless a shared pub/sub is wired.

## Rolling deploy

1. Backup DB.
2. Apply additive schema / module updates (`docs/ddl-policy.md`).
3. Roll instances behind the load balancer; drain with SIGTERM (20s shutdown).

## SSO / IdP

Native SAML/OIDC is not shipped yet. See [sso.md](sso.md). Until then use strong passwords, system-admin gated password changes (`orm.SetUserPassword`), and the login-link notify action (not a tokenized reset).

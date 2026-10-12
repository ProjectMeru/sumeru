# High availability operations

Sumeru sessions are **DB-backed** (`sys.session`), so HTTP load balancers do **not** need sticky sessions for cookie auth. Session cookies are browser-scoped; idle timeout ~8h with a 24h DB `expires_at` ceiling.

## Before running multiple app instances

1. **`csrf_secret` / `SUMERU_CSRF_SECRET`** — identical on every process (`sumeru.conf.example`).
2. **One PostgreSQL primary** — all instances share the same database; optional `db_read_replica_dsn` for read-heavy RPC paths.
3. **`metrics_scrape_token` / `SUMERU_METRICS_SCRAPE_TOKEN`** — scrape `GET /metrics` without an admin browser session.
4. **Health probes** — liveness `GET /api/health`; readiness `GET /api/ready` (DB ping).
5. **SWC bus** — events persist to `sys.bus.event` and fan out with **PostgreSQL NOTIFY** (`StartBusListener` on each process). WebSocket clients attach to one instance; NOTIFY delivers cross-instance fan-out when all nodes share the same DB.
6. **Optional queue mirror** — in-process `core/queue` topics can use `queue.SetPublishMirrorHook` (e.g. Redis) if you add workers outside the monolith; not required for standard SWC bus.

## Rolling deploy

1. Backup the database ([ops-runbook.md](ops-runbook.md)).
2. Apply additive schema / module updates per [ddl-policy.md](ddl-policy.md).
3. Roll instances behind the load balancer; allow SIGTERM drain (~20s shutdown).

## Identity

OIDC/OAuth and TOTP are built in — see [sso.md](sso.md). SAML is not shipped.

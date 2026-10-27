# Operations runbook (single-node pilot)

Alpha kernel — use for pilots and staging only. See [platform-status.md](platform-status.md).

## Deploy

1. Secrets via env: `SUMERU_DB_PASSWORD`, `SUMERU_CSRF_SECRET`, optional `SUMERU_METRICS_SCRAPE_TOKEN`, SMTP vars for mail notify.
2. Production-ish flags: `dev_mode=false`; set `rate_limit_rpm` (defaults to 120 when unset and not in dev).
3. Start: `./sumeru -c sumeru.conf` or `docker compose -f docker-compose.prod.yml up --build`.
4. Probes: `GET /api/health` (live), `GET /api/ready` (database).

### HTTPS and HTTP/2 (RPC + web)

Set **`tls_cert_file`** and **`tls_key_file`** in `sumeru.conf` (see `sumeru.conf.example`). The process listens with TLS; Go negotiates **HTTP/2** via ALPN on that listener. Terminate TLS at a reverse proxy instead when you prefer — ensure the proxy speaks HTTP/2 to clients and HTTP/1.1 or h2 to the app as appropriate.

Integrators call **`POST /api/rpc/auth.login`** (database + tenant headers, no Bearer), then **`POST /api/rpc/{model}/{method}`** with **`Authorization: Bearer <api_key>`** from the login response.

Optional shell presence (off by default):

- `close_session_on_last_tab = true` — SWC tab lease + `POST /web/session/touch` heartbeat; when all tabs stop, session ends after `session_browser_grace_seconds`.
- When **false** (default), closing tabs does not revoke the session until idle/DB expiry.

## Backup / restore

- `pg_dump` / managed Postgres backups.
- Restore DB before rolling back binaries if schema moved forward ([ddl-policy.md](ddl-policy.md)).

## Incident response

| Action | How |
| ------ | --- |
| Revoke all sessions for a user | Delete `sys.session` rows for `user_id` or use admin logout flows |
| Force re-login everywhere | Rotate sessions table or invalidate cookies via logout |
| CSRF invalidation | Changing `csrf_secret` requires restart of **all** instances |
| Observability | JSON logs (`request_id`); `GET /metrics` with scrape token |

Session timing: ~8h idle behavior with 24h DB ceiling; see `core/server/web` session resolution.

## Rollback

1. Point traffic to the previous binary/image.
2. If DDL or `-u` module sync advanced schema, restore DB or run authored reverse SQL — sync does not drop columns automatically.

## Related

- Multi-instance: [ha-ops.md](ha-ops.md)
- Auth / IdP: [sso.md](sso.md)

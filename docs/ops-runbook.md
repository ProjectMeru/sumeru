# Operations runbook (single-node pilot)

## Deploy

1. Set secrets via env: `SUMERU_DB_PASSWORD`, `SUMERU_CSRF_SECRET`, optional `SUMERU_METRICS_SCRAPE_TOKEN`.
2. `dev_mode=false`, `rate_limit_rpm` ≥ 120 (auto-defaulted when not in dev).
3. Start: `./sumeru -c sumeru.conf` or `docker compose -f docker-compose.prod.yml up --build`.
4. Probes: `/api/health` (live), `/api/ready` (DB).

## Backup / restore

- Use `pg_dump` / `pg_restore` (or your managed Postgres backup).
- Restore DB before rolling back application binaries if schema advanced (see `ddl-policy.md`).

## Incident

- Revoke sessions: delete rows from `sys.session` or destroy cookie via logout.
- Rotate `csrf_secret` only with a full restart of all instances (invalidates CSRF tokens).
- Check `/metrics` (Bearer scrape token) and JSON logs (`request_id`).

## Rollback

1. Point LB to previous image/binary.
2. If DDL was applied, restore DB or run reverse SQL from the change ticket.

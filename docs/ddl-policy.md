# Schema sync and DDL policy

The ORM materializes registered models into PostgreSQL with **additive** sync in `core/orm/schema_sync.go`:

- Creates missing tables and columns
- Ensures indexes, unique indexes, and foreign keys (best-effort)
- Does **not** drop or rename columns/tables automatically

Module install/update runs scoped sync via `SyncRegistrySchemaForModule` for models owned or extended by that module.

## Environments

| Environment | Policy |
|-------------|--------|
| Fresh DB / CI | Install (`-i`) and startup sync are source of truth |
| Shared staging / production | **Forward-only** additive changes; destructive DDL is a manual, reviewed migration |

## Production rules

1. **Backup** before module `-u` or manual DDL.
2. Prefer **additive** model changes (new nullable columns or defaults applied in Go).
3. **Renames/drops:** ship explicit SQL, apply in a maintenance window, then align Go models. Do not expect sync to remove old columns.
4. Review manifest and model diffs before `-u` on production.
5. Use production `db_sslmode`, pool limits, and secrets from `sumeru.conf.example`.

## Rollback

There is no automatic schema rollback. Restore from backup or run reverse SQL you authored. Rolling back the binary alone is usually safe if only new columns were added; it is unsafe if required columns were dropped manually.

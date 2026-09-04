# Schema sync and DDL policy

Sumeru materializes the ORM registry into PostgreSQL with **additive** schema sync (`core/orm/schema_sync.go`):

- Creates missing tables and columns
- Ensures indexes, unique indexes, and foreign keys (best-effort)
- Does **not** drop or rename columns/tables automatically

## Environments

| Environment | Policy |
|-------------|--------|
| Fresh setup / CI | `SyncRegistrySchema` / module install sync is the source of truth |
| Shared staging / production | Treat sync as **forward-only**. Destructive changes require a reviewed SQL migration run book and backup |

## Production rules

1. **Backup** before any manual DDL or module upgrade that alters schema.
2. Prefer **additive** model changes (new columns nullable or with Go-side defaults).
3. For renames/drops: ship a documented SQL script, apply in a maintenance window, then update Go models to match. Do not rely on sync to reverse changes.
4. Module install/update (`-i` / `-u`) still runs scoped sync; review module diffs before production update.
5. Keep `db_sslmode` and pool settings production-appropriate (`sumeru.conf.example`).

## Rollback

Schema sync has no automatic rollback. Rollback = restore from backup (or reverse SQL you authored). Application binary rollback alone may fail if the DB already has newer columns (usually safe) or if you manually dropped required columns (unsafe).

## Future

A recorded migration ledger (versioned SQL + apply tracking) may replace ad-hoc scripts; until then this document is the operational contract.

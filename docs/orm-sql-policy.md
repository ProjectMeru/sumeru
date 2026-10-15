# ORM and raw SQL policy (kernel)

Sumeru separates **business data access** from **infrastructure SQL**.

## Tier A — RPC and business mutations (ORM only)

All [`POST /api/rpc/{model}/{method}`](../core/server/web/rpc_json.go) model operations go through [`core/server/api`](../core/server/api/) into [`core/orm`](../core/orm/) (`Search`, `Create`, `UpdateRecordByID`, `Unlink`, `ReadGroup`, etc.). Record rules, model ACL, and field ACL apply. [`POST /api/rpc/auth.login`](../core/server/web/rpc_auth.go) uses ORM helpers in [`user_auth.go`](../core/orm/user_auth.go).

## Tier B — Web handlers on business models

Handlers that touch tenant/business models should use ORM helpers, not ad-hoc `UPDATE`/`SELECT`. Examples migrated to ORM:

- Active company switch → [`SetUserActiveCompany`](../core/orm/user_active_company.go)
- Event log listing → `SearchPage` on `app.log`
- Menu child walk → `SearchPage` on `sys.menu`

## Tier C — Infrastructure SQL (allowlisted)

These paths may use parameterized SQL via `orm.DB` with [`MustQuotedTableName`](../core/orm/db_wrapper.go):

| Area | Files | Reason |
|------|-------|--------|
| Sessions | `core/server/web/auth.go`, `session_browser.go` | Cookie session store |
| Login | — | Credential verify via `orm.AuthenticateLocalUser` (ORM read on `core.user`) |
| Health | `core/server/web/rpc_json.go` | `Ping()` only |
| Tests | `core/server/web/testexports.go` | Test session fixtures |
| ORM kernel | `core/orm/*` | Query builder implementation |
| Module sync | `core/module/*` | Elevated install/sync |
| Engine menus | `core/engine/render/menus.go` | Performance (review when changing) |

CI enforces that **new** raw SQL in `core/server/web` stays within the allowlist ([`scripts/check_sql_allowlist.sh`](../scripts/check_sql_allowlist.sh)).

## Verification

```bash
cd sumeru && bash scripts/check_sql_allowlist.sh
cd sumeru && bash scripts/check_sql_safety.sh
```

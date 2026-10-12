# Platform status (kernel)

Maturity of the **`sumeru` repository** (Go engine + `addons/`). Business apps in [sumeru_addons](https://github.com/ProjectMeru/sumeru_addons) are tracked separately.

**Stage:** **Alpha (platform kernel)** — development, demos, and custom modules. Not for production or regulated workloads. JSON-RPC payloads and UX may change between releases.

## Capability matrix

| Area | Status | Notes |
| ---- | ------ | ----- |
| Modules & ORM | **Ready** | Manifests, install/update, compute, domains, multi-company, elevated context |
| Security | **Ready** | Groups, record rules, model ACL, field ACL UI |
| Views & SWC | **Ready** | List, form, kanban, graph, pivot, calendar, gantt, map, cohort, hierarchy, activity; view inherit |
| JSON-RPC | **Ready** | `POST /api/rpc` (session or API key) — primary integration API |
| Reports, import/export | **Ready** | Document reports; import wizard; async export / bulk jobs (`platform`) |
| Mail & activities | **Partial** | `mail` addon: messages, followers, notifications, outbound queue, activity board; no inbound IMAP |
| Realtime | **Ready** | `sys.bus.event`, PG NOTIFY, SWC WebSocket reconnect |
| Portal | **Partial** | `/portal` shell and share links |
| Auth | **Ready** | Password, OIDC/OAuth providers, TOTP, trusted device — [sso.md](sso.md) |
| Automation | **Ready** | Cron, server actions (`automation` addon) |
| Currency, UoM, properties | **Ready** | `currency`, `uom` addons; monetary fields; JSONB properties + `sys.property.definition` |
| IM | **Ready** | Internal chat (`im` addon), separate from record mail |
| Studio, website CMS, offline PWA | **Roadmap** | Not in kernel today |
| Inbound mail gateway | **Not planned** | Outbound + in-app messaging only |
| Public REST API | **Not planned** | Use JSON-RPC and HTTP helpers (export, content, health) |

## Kernel addons (shipped in this repo)

| Module | Role |
| ------ | ---- |
| `base` | Users, companies, partners, security, auth providers |
| `platform` | Attachments, sequences, parameters, reports, import/export (auto_install) |
| `currency`, `uom` | Currency catalog, rates, units of measure |
| `geo`, `i18n` | Geography, translations |
| `im` | Internal messaging |
| `mail` | Record messages and activities |
| `contacts`, `calendar` | Reference apps |
| `automation`, `audit`, `digest` | Technical automation and audit |
| `sumeru_ai` | Optional AI hooks (`auto_import: false`) |

Manifest `depends` and `auto_install` flags are authoritative; run `make generate` after dependency changes in downstream workspaces.

## Integrations

- **JSON-RPC:** `POST /api/rpc` with browser session or API key.
- **HTTP:** Content download, export endpoints, health/readiness — see `core/server/web` routes and [architecture-layers.md](architecture-layers.md).

Extend via addons and RPC, not a separate REST surface.

# Sumeru

**Modular open-source ERP — Go backend, PostgreSQL, and a modern web workspace.**

[![CI](https://github.com/ProjectMeru/sumeru/actions/workflows/ci.yml/badge.svg)](https://github.com/ProjectMeru/sumeru/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.26.2+-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Pre-Alpha](https://img.shields.io/badge/Status-Pre--Alpha-critical)](https://github.com/ProjectMeru/sumeru)
[![Docs](https://img.shields.io/badge/Docs-projectmeru.github.io-informational)](https://projectmeru.github.io/sumeru/docs/)

![Sumeru settings and apps hub](core/engine/assets/img/hero_img.png)

*Settings hub — configure companies, users, localization, and open installed apps.*

> [!CAUTION]
>
> **Pre-alpha software** — not for production or commercial use.
>
> - Do not deploy live business workloads; stability and data integrity are not guaranteed.
> - APIs, data models, and behavior may change without notice.
> - Use for local development, evaluation, and feedback only.

Sumeru is a modular ERP platform built in **Go**. Install **apps** (addons) for CRM, sales, inventory, and more; persist data in **PostgreSQL**; define screens with XML **views**; and run day-to-day work in a **SWC** workspace client with a shared control bar for search, filters, and saved searches across list, form, kanban, graph, pivot, calendar, gantt, map, and cohort views.

This repository is the **core engine** (`module sumeru`). Most teams keep it pull-only and run the server from **[sumeru_custom_addons](https://github.com/ProjectMeru/sumeru_custom_addons)**.

## Key features

- **Modular apps** — manifests, XML views/menus, Go models, install and update from disk
- **PostgreSQL ORM** — model sync on startup, record rules, multi-company support
- **Collection control bar** — search, filters, multi group-by, custom domain rules, and saved-search favorites on collection and analysis views
- **Workspace views** — list, form, kanban, graph, pivot, calendar, gantt, map, and cohort
- **Saved searches** — persist filter/search/group state per user (`swc.saved.search` + favorites API)
- **SWC client** — TypeScript sources in `core/swc/`; bundles built locally (not committed to git)
- **JSON-RPC API** — `POST /api/rpc` for integrations (session or API key)
- **Stable addon SDK** — prefer **`sumeru/core/sdk`** over direct **`sumeru/core/orm`** imports
- **Plain CSS shell** — design tokens and `sum-*` layout under `core/engine/assets/css/`

## Quick start

**Prerequisites:** [Go 1.26.2+](https://go.dev/dl/), [Node.js](https://nodejs.org/) (npm — builds the SWC UI), [PostgreSQL](https://www.postgresql.org/)

Clone the three sibling repositories, configure the workspace, and run:

```bash
mkdir -p ~/sumeru_erp && cd ~/sumeru_erp
git clone git@github.com:ProjectMeru/sumeru.git
git clone git@github.com:ProjectMeru/sumeru_addons.git
git clone git@github.com:ProjectMeru/sumeru_custom_addons.git

# Create a PostgreSQL database matching db_name in your INI, e.g.:
#   psql -c "CREATE DATABASE sumeru;"

cd sumeru_custom_addons
cp sumeru.conf.example sumeru.conf   # edit db_*, http_port, addons_path
make setup    # go.mod, imports, SWC + login JS bundles
make run      # regenerate imports; rebuild assets when missing or stale
```

Open **`http://localhost:8080`** (or your `http_port`). `/` redirects to **`/web/apps`**.

**Day-to-day updates:**

```bash
cd ../sumeru && git pull
cd ../sumeru_addons && git pull
cd ../sumeru_custom_addons && make run
```

After pulling SWC changes in `sumeru`, `make run` rebuilds client bundles automatically when TypeScript sources are newer than the on-disk JS.

Full workspace details: **[sumeru_custom_addons README](https://github.com/ProjectMeru/sumeru_custom_addons/blob/main/README.md)**.

### Core-only (optional)

When you only need kernel apps under `sumeru/addons/`:

```bash
cd sumeru
cp sumeru.conf.example sumeru.conf
make setup && make run
```

Install sample data, then serve: `go run ./cmd/sumeru -- -c sumeru.conf -i company,user --stop-after-init` then `make run`.

## Development workflow

Most development happens in **`sumeru_custom_addons`**. The core repo uses the same Makefile targets when working on kernel addons or SWC directly.

### Custom workspace (`sumeru_custom_addons`)

| Step | Command | What it does |
| ---- | ------- | ------------ |
| First time | `make setup` | Config, go.mod replaces, import generation, SWC + login JS |
| Daily dev | `make run` or `make dev` | Regenerate imports, rebuild assets if stale, start HTTP server |
| After SWC/CSS edits | `make swc` | Force rebuild workspace bundle + login JS |
| Pull upstream | `git pull` in `sumeru` / `sumeru_addons`, then `make run` | Rebuilds assets when sources changed |
| Tests | `make check` | SWC typecheck + Go tests |
| Binary | `make build` | Imports + assets + `bin/sumeru-erp` |

### Core repo (`sumeru`)

| Step | Command | What it does |
| ---- | ------- | ------------ |
| First time | `make setup` | `sumeru.conf`, SWC assets, `cmd/sumeru/zimports.go` |
| Daily dev | `make run` | `generate` + `assets` + `go run` |
| Force client rebuild | `make swc` | Always rebuild SWC + login JS |
| Incremental assets | `make assets` | Build only when bundles missing or sources changed |
| Tests | `make check` | `swc-check` + `go test ./...` |
| Binary | `make build` | `generate` + `assets` + `./sumeru` binary |

```text
make run  →  generate (Go imports)  →  assets (SWC + login JS)  →  go run server
```

## Client assets

The browser UI is **not stored in git**. Sources live under **`core/swc/src/`** (TypeScript); build outputs are gitignored (see `.gitignore`).

| Output | Purpose |
| ------ | ------- |
| `core/engine/assets/swc/swc.js` | Workspace UI (`/static/swc/swc.js`) |
| `core/engine/assets/js/sumeru-password-toggle.js` | Login / setup password visibility |
| `core/engine/assets/js/sumeru-password-match.js` | Setup password confirmation |

Build with **`make swc`** (always rebuild) or let **`make assets`** / **`make run`** build when bundles are missing or SWC sources changed. A fresh clone needs **Node.js** — run **`make setup`** or **`make run`** once before opening the app.

## Architecture

Sumeru splits across three repositories so you can pull engine and standard apps without mixing customer code.

```text
sumeru_custom_addons
  make setup / make run
    → generate (addonimports/zimports.go)
    → assets (SWC + login JS → sumeru/core/engine/assets/)
    → go run HTTP server
  loads sumeru (core) + sumeru_addons via addons_path in sumeru.conf
```

| Repository | Role |
| ---------- | ---- |
| **[sumeru](https://github.com/ProjectMeru/sumeru)** | Core engine + kernel apps (`base`, `mail`, …) |
| **[sumeru_addons](https://github.com/ProjectMeru/sumeru_addons)** | Standard business apps (CRM, Sales, Inventory, …) |
| **[sumeru_custom_addons](https://github.com/ProjectMeru/sumeru_custom_addons)** | Your workspace: custom addons, INI, generated imports, `make run` |

**Entry binary:** `cmd/sumeru/main.go` → `sumeru/core/server`. Library code under `core/` has no `main`.

## Tech stack

| Layer | Technology |
| ----- | ---------- |
| Server | Go 1.26.2+, structured logging (`log/slog`) |
| Database | PostgreSQL |
| Modules | Go addons + XML views/menus + manifest sync |
| Workspace UI | SWC (TypeScript) — sources in `core/swc/` |
| Client build | Node.js + npm (esbuild); `make assets` / `make swc` |
| Styling | Plain CSS (`core/engine/assets/css/`) |
| API | JSON-RPC at `POST /api/rpc`, health at `GET /api/health` |

CLI flags: `-c` config, `-d` database, `-p` port, `-i` install modules, `-u` update modules. See [Tooling docs](https://projectmeru.github.io/sumeru/docs/guides/build/tooling.html).

## Documentation

| Resource | Contents |
| -------- | -------- |
| [Documentation home](https://projectmeru.github.io/sumeru/docs/) | Guides, reference, and tutorials |
| [Installation](https://projectmeru.github.io/sumeru/docs/guides/start/installation.html) | First-time setup |
| [Configuration](https://projectmeru.github.io/sumeru/docs/guides/start/configuration.html) | `sumeru.conf` keys and paths |
| [JSON-RPC API](https://projectmeru.github.io/sumeru/docs/reference/json-rpc.html) | RPC methods, auth, errors |
| [SWC architecture](https://projectmeru.github.io/sumeru/docs/guides/concepts/swc-architecture.html) | Workspace client |
| [Creating an addon](https://projectmeru.github.io/sumeru/docs/guides/build/creating-an-addon.html) | Module authoring |
| [Tooling](https://projectmeru.github.io/sumeru/docs/guides/build/tooling.html) | Makefile, import-gen, CLI |
| [sumeru_addons README](https://github.com/ProjectMeru/sumeru_addons/blob/main/README.md) | Standard business apps |
| [sumeru_custom_addons README](https://github.com/ProjectMeru/sumeru_custom_addons/blob/main/README.md) | Workspace runner |

Configuration template: **`sumeru.conf.example`** in this repo.

## Project layout

| Path | Purpose |
| ---- | ------- |
| `core/orm/` | PostgreSQL models, CRUD, registry |
| `core/engine/` | View XML, HTML render, templates, assets |
| `core/server/` | INI config, HTTP handlers |
| `core/module/` | Addon discovery, install/update |
| `core/sdk/` | Stable Go API for addons |
| `core/swc/` | Workspace UI source (TypeScript) |
| `cmd/sumeru/` | Server binary + generated imports |
| `addons/` | Kernel apps shipped with core |
| `test/` | Unit and integration tests |

## Contributing

See **[CONTRIBUTING.md](CONTRIBUTING.md)** for where to put changes, the generate/test loop, and PR expectations. Please follow the **[Code of Conduct](CODE_OF_CONDUCT.md)**.

## Security

Report vulnerabilities privately — see **[SECURITY.md](SECURITY.md)**. Do not open public issues for undisclosed security problems.

## License

Licensed under the [Apache License, Version 2.0](LICENSE).

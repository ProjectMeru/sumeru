# Sumeru

**Modular open-source ERP/CRM platform** - Go backend, PostgreSQL, installable apps, and a TypeScript workspace client for enterprise operations.

[![CI](https://github.com/ProjectMeru/sumeru/actions/workflows/ci.yml/badge.svg)](https://github.com/ProjectMeru/sumeru/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.26.9+-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Alpha](<https://img.shields.io/badge/Status-Alpha%20(platform%20kernel)-orange>)](https://github.com/ProjectMeru/sumeru)
[![Docs](https://img.shields.io/badge/Docs-projectmeru.github.io-informational)](https://projectmeru.github.io/sumeru/docs/)

## Table of contents

- [Sumeru](#sumeru)
  - [Table of contents](#table-of-contents)
  - [Overview](#overview)
  - [Platform maturity](#platform-maturity)
  - [Features](#features)
  - [Preview](#preview)
  - [Architecture](#architecture)
  - [Repository ecosystem](#repository-ecosystem)
  - [Getting started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Recommended — custom workspace](#recommended--custom-workspace)
    - [Alternative — core only](#alternative--core-only)
    - [First run](#first-run)
  - [Configuration](#configuration)
  - [Development](#development)
  - [Documentation](#documentation)
  - [Community](#community)

## Overview

Sumeru is a **modular ERP and CRM engine** for teams that want a single deployable stack: **PostgreSQL** for data, **XML views** for screens and menus, **Go models** for business logic, and a **SWC** web workspace for list, form, kanban, and analysis work. Install **apps** (addons) from disk; extend behavior through manifests and the **`sumeru/core/sdk`** surface instead of forking core.

This repository (`module sumeru`) is the **kernel**: ORM, HTTP/JSON-RPC, rendering shell, and platform apps (`base`, `platform`, `mail`, `currency`, …). Standard CRM/ERP apps ship in **[sumeru_addons](https://github.com/ProjectMeru/sumeru_addons)**. Most integrators run the server from **[sumeru_custom_addons](https://github.com/ProjectMeru/sumeru_custom_addons)** (paths, config, generated imports).

> [!CAUTION]
>
> **Alpha — platform kernel** — not for production or regulated workloads.
>
> - The **platform layer** (ORM, security, workspace, mail/bus, automation) is usable for **development, demos, and custom modules**.
> - Schemas, RPC payloads, and UX may still change between releases.
> - Business app depth and operational hardening live in sibling repos and your deployment process.

## Platform maturity

Status reflects **this repository only** (engine + kernel addons), not full ERP product parity in **sumeru_addons**.

| Area                      | Status      | Notes                                                                                                            |
| ------------------------- | ----------- | ---------------------------------------------------------------------------------------------------------------- |
| Modules & ORM             | **Ready**   | Manifests, install/update, compute, domains, multi-company, elevated context                                     |
| Security                  | **Ready**   | Groups, record rules, model ACL, field ACL UI                                                                    |
| Views & SWC               | **Ready**   | List, form, kanban, graph, pivot, calendar, gantt, map, cohort, hierarchy, activity; view inherit                |
| JSON-RPC                  | **Ready**   | `POST /api/rpc` with session or API key — **the supported integration surface** (no separate REST layer planned) |
| Reports, import/export    | **Ready**   | Document reports; import wizard; async export / bulk jobs                                                        |
| Mail & activities         | **Partial** | Messages, followers, notifications, outbound mail queue, activity board; no inbound IMAP gateway                 |
| Realtime                  | **Ready**   | Bus channels, durable events, client reconnect                                                                   |
| Portal                    | **Partial** | Portal shell and share links; not a full customer portal product                                                 |
| Auth                      | **Ready**   | Password login, OAuth/OIDC providers, TOTP, trusted device                                                       |
| Automation                | **Ready**   | Cron, server actions, workflow hooks (`automation` addon)                                                        |
| Currency, UoM, properties | **Ready**   | Dedicated addons; monetary fields; JSONB properties + definition admin                                           |

Integrations should use **JSON-RPC** (`/api/rpc`) and documented HTTP routes (content, export, health). A resource-oriented REST API is **not** on the roadmap.

## Features

| Area               | Capabilities                                                                             |
| ------------------ | ---------------------------------------------------------------------------------------- |
| **Apps & modules** | Manifests, dependency order, install/update from `addons_path`; XML data and views       |
| **Data layer**     | PostgreSQL ORM, model sync on startup, record rules, multi-company                       |
| **Workspace**      | List, form, kanban, graph, pivot, calendar, gantt, map, cohort, hierarchy, activity      |
| **Collection UX**  | Search, filters, group-by, custom domains, saved-search favorites, command palette       |
| **Integrations**   | **JSON-RPC** `POST /api/rpc` (session or API key); mail and notifications; import/export |
| **Client**         | TypeScript SWC in `core/swc/`; server-rendered shell + SPA workspace                     |
| **Extensibility**  | Prefer **`sumeru/core/sdk`** for addons; `sum-*` design tokens in CSS                    |

## Preview

User flow: **sign in → apps catalog → settings and workspace**.

**Sign-in** — enterprise login shell with optional company logo and tagline on the left panel.

![Enterprise login with branding panel](core/engine/assets/img/screenshots/login_screen.png)

**Apps** — browse, install, and open CRM/ERP modules from the catalog.

![Installed and available apps](core/engine/assets/img/screenshots/apps_imgs.png)

**Settings & workspace** — configure the organization and work in list, form, and analysis views.

![Settings hub and workspace chrome](core/engine/assets/img/screenshots/hero_img.png)

## Architecture

Sumeru runs as a **Go monolith**: HTTP handlers and JSON-RPC call into the ORM; addons register models and actions; the SWC client loads inside an HTML shell from `core/engine/render`.

| Layer                | Location                                        |
| -------------------- | ----------------------------------------------- |
| UI shell & routes    | `core/server/web`, `core/engine/render`         |
| Workspace client     | `core/swc`                                      |
| Business logic & RPC | `core/server/api`, `core/orm`, `core/sdk`       |
| Modules              | `core/module`, `addons/`, external addon repos  |
| Security             | Access CSV/XML, record rules, groups, field ACL |

Full layer diagram and request flow: **[docs/architecture-layers.md](docs/architecture-layers.md)**.

## Repository ecosystem

| Repository                                                                      | Role                                                    |
| ------------------------------------------------------------------------------- | ------------------------------------------------------- |
| **[sumeru](https://github.com/ProjectMeru/sumeru)**                             | Core engine + kernel apps                               |
| **[sumeru_addons](https://github.com/ProjectMeru/sumeru_addons)**               | Standard CRM/ERP apps                                   |
| **[sumeru_custom_addons](https://github.com/ProjectMeru/sumeru_custom_addons)** | Workspace: `sumeru.conf`, generated imports, `make run` |

## Getting started

### Prerequisites

- [Go 1.26.9+](https://go.dev/dl/)
- [Node.js](https://nodejs.org/) (npm — builds SWC bundles; not committed to git)
- [PostgreSQL](https://www.postgresql.org/)
- Git
- Optional: Docker for `make test-integration`

### Recommended — custom workspace

Clone the three sibling repositories, create a database matching `db_name`, then run from the workspace repo:

```bash
mkdir -p ~/sumeru_erp && cd ~/sumeru_erp
git clone git@github.com:ProjectMeru/sumeru.git
git clone git@github.com:ProjectMeru/sumeru_addons.git
git clone git@github.com:ProjectMeru/sumeru_custom_addons.git

# psql -c "CREATE DATABASE sumeru;"

cd sumeru_custom_addons
cp sumeru.conf.example sumeru.conf   # edit db_*, http_port, addons_path
make setup && make run
```

Open **`http://localhost:8080`** (or your `http_port`). Details: [sumeru_custom_addons README](https://github.com/ProjectMeru/sumeru_custom_addons/blob/main/README.md).

### Alternative — core only

Kernel apps under `sumeru/addons/`:

```bash
cd sumeru
cp sumeru.conf.example sumeru.conf   # addons_path = addons
make setup && make run
```

Optional sample data:

```bash
go run ./cmd/sumeru -- -c sumeru.conf -i company,user --stop-after-init
make run
```

### First run

| Step        | Behavior                                                                                       |
| ----------- | ---------------------------------------------------------------------------------------------- |
| Empty DB    | Browser **`/setup`** wizard (company + admin); see `setup_*` keys in **`sumeru.conf.example`** |
| After setup | **`/web/login`**; set **`dev_mode = true`** for local UI iteration                             |
| Signed in   | `/` → **`/web/apps`**                                                                          |

## Configuration

Copy **`sumeru.conf.example`** → **`sumeru.conf`** next to `go.mod` (or in `sumeru_custom_addons` for the workspace flow).

| Setting                      | Required | Notes                             |
| ---------------------------- | -------- | --------------------------------- |
| `db_*`, `db_sslmode`         | Yes      | PostgreSQL connection             |
| `http_port`                  | Yes      | Default `8080`                    |
| `addons_path`                | Yes      | Comma-separated addon directories |
| `dev_mode`                   | No       | `true` for local development      |
| `log_enabled` / `log_stdout` | No       | See example for file logging      |

Reference: [configuration guide](https://projectmeru.github.io/sumeru/docs/guides/start/configuration.html).

## Development

Run **`make`** from **`sumeru/`** (or workspace Makefile in **`sumeru_custom_addons/`**) before opening a PR.

| Command                 | Purpose                                         |
| ----------------------- | ----------------------------------------------- |
| `make setup`            | `sumeru.conf` scaffold; `generate` + SWC assets |
| `make run` / `make dev` | `go run ./cmd/sumeru -c sumeru.conf`            |
| `make swc`              | Rebuild workspace + login JS bundles            |
| `make` / `make check`   | Lint, tests, `go build ./...` (CI parity)       |
| `make build`            | Production binary `./sumeru`                    |
| `make help`             | Modules, integration tests, i18n, shell REPL    |

Extra CLI flags: **`EXTRA_RUN_FLAGS`** (e.g. `make run EXTRA_RUN_FLAGS='-p 9090'`). See [tooling docs](https://projectmeru.github.io/sumeru/docs/guides/build/tooling.html).

SWC sources live in `core/swc/src/`; bundles land in `core/engine/assets/` when you run `make setup` or `make run`.

## Documentation

| Resource      | Link                                                                                                    |
| ------------- | ------------------------------------------------------------------------------------------------------- |
| Site          | [projectmeru.github.io/sumeru/docs/](https://projectmeru.github.io/sumeru/docs/)                        |
| Architecture  | [docs/architecture-layers.md](docs/architecture-layers.md)                                              |
| Configuration | [guides/start/configuration](https://projectmeru.github.io/sumeru/docs/guides/start/configuration.html) |
| Tooling & CLI | [guides/build/tooling](https://projectmeru.github.io/sumeru/docs/guides/build/tooling.html)             |

## Community

|                  |                                                                             |
| ---------------- | --------------------------------------------------------------------------- |
| **Contributing** | [CONTRIBUTING.md](CONTRIBUTING.md) — layout, generate loop, PR expectations |
| **Security**     | [SECURITY.md](SECURITY.md) — responsible disclosure                         |
| **License**      | [Apache 2.0](LICENSE)                                                       |

# Sumeru

**Modular open-source ERP/CRM platform** — Go backend, PostgreSQL, installable apps, and a TypeScript workspace client.

[![CI](https://github.com/ProjectMeru/sumeru/actions/workflows/ci.yml/badge.svg)](https://github.com/ProjectMeru/sumeru/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.27.0+-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Alpha](https://img.shields.io/badge/Status-Alpha%20(platform%20kernel)-orange)](https://github.com/ProjectMeru/sumeru)
[![Docs](https://img.shields.io/badge/Docs-projectmeru.github.io-informational)](https://projectmeru.github.io/sumeru/docs/)

## Overview

Sumeru is a **modular ERP/CRM engine**: **PostgreSQL**, **XML views**, **Go models** (`sumeru/core/sdk`), and the **SWC** web workspace. This repo is the **kernel** (ORM, JSON-RPC, shell, platform apps). Standard apps ship in **[sumeru_addons](https://github.com/ProjectMeru/sumeru_addons)**; most teams run from **[sumeru_custom_addons](https://github.com/ProjectMeru/sumeru_custom_addons)**.

> [!CAUTION]
> **Alpha (platform kernel)** — not for production or regulated use. See **[docs/platform-status.md](docs/platform-status.md)** for what is ready, partial, or on the roadmap (integrations are **JSON-RPC** only).

## Quick start

**Prerequisites:** Go 1.27.0+, Node.js (SWC build), PostgreSQL.

```bash
mkdir -p ~/sumeru_erp && cd ~/sumeru_erp
git clone git@github.com:ProjectMeru/sumeru.git
git clone git@github.com:ProjectMeru/sumeru_addons.git
git clone git@github.com:ProjectMeru/sumeru_custom_addons.git

cd sumeru_custom_addons
cp sumeru.conf.example sumeru.conf   # db_*, http_port, addons_path
make setup && make run
```

Open `http://localhost:8080`. Empty DB → **`/setup`**; then **`/web/login`** → **`/web/apps`**. Workspace details: [sumeru_custom_addons README](https://github.com/ProjectMeru/sumeru_custom_addons/blob/main/README.md).

**Core only:** `cd sumeru && cp sumeru.conf.example sumeru.conf && make setup && make run` (`addons_path = addons`).

## Preview

![Login](core/engine/assets/img/screenshots/login_screen.png)

![Apps catalog](core/engine/assets/img/screenshots/apps_imgs.png)

![Workspace](core/engine/assets/img/screenshots/hero_img.png)

## Development

From `sumeru/` (or `make` targets in `sumeru_custom_addons/`): **`make`** before PRs; **`make run`** for the server; **`make help`** for modules, tests, and REPL.

| Doc | Link |
| --- | --- |
| Platform status & capabilities | [docs/platform-status.md](docs/platform-status.md) |
| Architecture | [docs/architecture-layers.md](docs/architecture-layers.md) |
| Auth / SSO (OIDC, TOTP) | [docs/sso.md](docs/sso.md) |
| Ops (single node, HA, DDL) | [docs/ops-runbook.md](docs/ops-runbook.md) · [docs/ha-ops.md](docs/ha-ops.md) · [docs/ddl-policy.md](docs/ddl-policy.md) |
| Configuration | [sumeru.conf.example](sumeru.conf.example) · [site guide](https://projectmeru.github.io/sumeru/docs/guides/start/configuration.html) |
| Tooling | [site guide](https://projectmeru.github.io/sumeru/docs/guides/build/tooling.html) |
| Full docs site | [projectmeru.github.io/sumeru/docs/](https://projectmeru.github.io/sumeru/docs/) |

## Community

[CONTRIBUTING.md](CONTRIBUTING.md) · [SECURITY.md](SECURITY.md) · [Apache 2.0](LICENSE)

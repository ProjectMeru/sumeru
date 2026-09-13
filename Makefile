# Sumeru kernel — `make` or `make all` runs the full local standard.
.PHONY: help all check lint standards audit-test vet lint-go \
	check-sql check-logs test test-go swc-test swc-check swc-build swc-deps swc assets \
	setup run dev build build-check generate \
	test-modules test-modules-static test-modules-unit test-modules-addon \
	test-integration test-db \
	bp css db-check i18n-export i18n-import module shell

EXTRA_RUN_FLAGS ?=
TEST_DSN ?= host=localhost port=5433 user=postgres password=postgres dbname=sumeru_test sslmode=disable
GO_TEST_FLAGS ?= -count=1 -timeout 15m
GO_COVERAGE_MIN ?= 42
GOLANGCI := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

SWC_DIR := core/swc
SWC_BUNDLE := core/engine/assets/swc/swc.js
SWC_JS_TOGGLE := core/engine/assets/js/sumeru-password-toggle.js
SWC_JS_MATCH := core/engine/assets/js/sumeru-password-match.js
SWC_JS_APPS_PAGE := core/engine/assets/js/apps-page.js
SWC_ASSET_INPUTS := $(SWC_DIR)/esbuild.config.mjs $(SWC_DIR)/sum-compile.mjs $(SWC_DIR)/package.json

# =============================================================================
# Standard — default goal
# =============================================================================

all check: lint test build-check

lint: standards swc-check vet lint-go audit-test

test: swc-test test-go

# =============================================================================
# Static analysis
# =============================================================================

standards: check-sql check-logs
	@bash scripts/check_security_bypass.sh

check-sql:
	@bash scripts/check_sql_safety.sh

check-logs:
	@bash scripts/check_no_stdlog.sh

audit-test:
	go test ./test/core/security/... $(GO_TEST_FLAGS)

vet:
	go vet ./...

lint-go:
	$(GOLANGCI) run --timeout=10m

build-check:
	go build ./...

# =============================================================================
# Tests — Go (one target for full suite + coverage gate)
# =============================================================================

test-go:
	go test ./test/... -coverpkg=./... -coverprofile=coverage.out $(GO_TEST_FLAGS)
	GO_COVERAGE_MIN=$(GO_COVERAGE_MIN) bash scripts/check_go_coverage.sh coverage.out

test-modules-static:
	go test ./test/module/static/... ./test/core/importgen/... $(GO_TEST_FLAGS)

test-modules-unit:
	go test ./test/module/unit/... $(GO_TEST_FLAGS)

test-modules-addon:
	go test ./test/module/addons/... ./test/addons/... $(GO_TEST_FLAGS)

test-modules: test-modules-static test-modules-unit test-modules-addon

test-db:
	docker compose -f docker-compose.test.yml up -d --wait

test-integration: test-db
	SUMERU_TEST_DSN='$(TEST_DSN)' go test -tags=integration \
		./test/integration/... ./test/module/integration/... $(GO_TEST_FLAGS)

# =============================================================================
# SWC & client assets
# =============================================================================

swc-deps:
	cd $(SWC_DIR) && npm install

swc-check: swc-deps
	cd $(SWC_DIR) && npm run check

swc-test: swc-deps
	cd $(SWC_DIR) && npm run test:coverage

swc-build: swc-deps
	cd $(SWC_DIR) && npm run build

swc: swc-build

assets:
	@if [ ! -f $(SWC_BUNDLE) ] || [ ! -f $(SWC_JS_TOGGLE) ] || [ ! -f $(SWC_JS_MATCH) ] || [ ! -f $(SWC_JS_APPS_PAGE) ]; then \
		echo "Building SWC assets (bundles missing)..."; \
		$(MAKE) swc-build; \
	elif find $(SWC_DIR)/src $(SWC_ASSET_INPUTS) -type f -newer $(SWC_BUNDLE) 2>/dev/null | grep -q .; then \
		echo "Building SWC assets (sources changed)..."; \
		$(MAKE) swc-build; \
	else \
		echo "SWC assets up to date"; \
	fi

# =============================================================================
# Dev & production binary
# =============================================================================

setup:
	@test -f sumeru.conf || cp sumeru.conf.example sumeru.conf
	$(MAKE) assets generate

run dev: generate assets
	go run ./cmd/sumeru -- -c sumeru.conf $(EXTRA_RUN_FLAGS)

build: generate assets
	go build -o sumeru ./cmd/sumeru

generate:
	go generate ./cmd/sumeru

# =============================================================================
# CLI tools
# =============================================================================

bp:
	@test -n "$(NAME)" || (echo 'usage: make bp NAME=my_module' >&2 && exit 1)
	go run ./cmd/sumeru-bp -name $(NAME)

css:
	@echo "No CSS build step — edit core/engine/assets/css/*.css"

db-check:
	go run ./cmd/sumeru-db-check -- -c sumeru.conf

i18n-export:
	go run ./cmd/sumeru-i18n -- -c sumeru.conf export -o translations.csv

i18n-import:
	go run ./cmd/sumeru-i18n -- -c sumeru.conf import -i translations.csv

module:
	go run ./cmd/sumeru-module -- -c sumeru.conf $(ARGS)

shell:
	go run ./cmd/sumeru-shell -- -c sumeru.conf

# =============================================================================
# Help
# =============================================================================

help:
	@echo "Sumeru Makefile — run from the sumeru/ directory (kernel repo root)"
	@echo ""
	@echo "STANDARD (pre-PR / matches CI)"
	@echo "  make                 full gate: lint + test + build-check  (default goal)"
	@echo "  make all             same as make"
	@echo "  make check           same as make"
	@echo ""
	@echo "LINT (static analysis — no full test suite)"
	@echo "  make lint            standards + swc-check + vet + golangci-lint + audit-test"
	@echo "  make standards       check-sql + check-logs + security bypass script"
	@echo "  make check-sql       reject raw SQL fmt.Sprintf in core/server/web"
	@echo "  make check-logs      reject stdlib log/fmt.Print in core/"
	@echo "  make audit-test      go test ./test/core/security/..."
	@echo "  make vet             go vet ./..."
	@echo "  make lint-go         golangci-lint only"
	@echo "  make build-check     go build ./..."
	@echo ""
	@echo "TEST"
	@echo "  make test            swc-test + test-go"
	@echo "  make test-go         go test ./test/... + coverage gate (GO_COVERAGE_MIN=$(GO_COVERAGE_MIN))"
	@echo "  make swc-test        vitest with coverage (core/swc)"
	@echo "  make test-modules    module suite tiers 0–2 (static + unit + addon)"
	@echo "  make test-modules-static   addon convention validation only"
	@echo "  make test-modules-unit     module unit tier only"
	@echo "  make test-modules-addon    addon tests tier only"
	@echo "  make test-integration      PostgreSQL tests (-tags=integration; needs Docker)"
	@echo "  make test-db               start docker-compose.test.yml postgres"
	@echo ""
	@echo "SWC & ASSETS"
	@echo "  make swc-check       TypeScript typecheck (core/swc)"
	@echo "  make swc-build       esbuild bundle + login JS → core/engine/assets/"
	@echo "  make swc             alias for swc-build (force rebuild)"
	@echo "  make assets          build SWC bundles if missing or sources changed"
	@echo "  make swc-deps        npm install in core/swc (used by swc-* targets)"
	@echo ""
	@echo "DEV & BUILD"
	@echo "  make setup           sumeru.conf + assets + generate (first-time bootstrap)"
	@echo "  make dev             run server (alias: run)"
	@echo "  make run             go run ./cmd/sumeru -c sumeru.conf"
	@echo "  make build           production binary → ./sumeru"
	@echo "  make generate        refresh cmd/sumeru/zimports.go"
	@echo ""
	@echo "TOOLS"
	@echo "  make bp NAME=x       scaffold a kernel addon"
	@echo "  make module ARGS='list'   module CLI (install, list, …)"
	@echo "  make shell           ORM REPL"
	@echo "  make db-check        database connectivity check"
	@echo "  make i18n-export     export translations.csv"
	@echo "  make i18n-import     import translations.csv"
	@echo "  make css             reminder: edit core/engine/assets/css/*.css directly"
	@echo ""
	@echo "VARIABLES"
	@echo "  EXTRA_RUN_FLAGS      passed to make run (e.g. -p 9090 -d mydb)"
	@echo "  TEST_DSN             integration test postgres DSN"
	@echo "  GO_COVERAGE_MIN      min Go coverage % for test-go (default $(GO_COVERAGE_MIN))"
	@echo "  GO_TEST_FLAGS        default: $(GO_TEST_FLAGS)"

.DEFAULT_GOAL := all

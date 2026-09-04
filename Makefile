.PHONY: help setup dev build css run generate bp check-sql check-logs db-check \
	i18n-export i18n-import module shell test-db test-integration test-coverage \
	test-modules test-modules-static test-modules-unit test-modules-addon test-modules-integration \
	swc swc-build assets swc-check swc-test check lint

# Extra flags for `make run`, e.g. `make run EXTRA_RUN_FLAGS='-p 9090 -d sumeru_staging'`
EXTRA_RUN_FLAGS ?=

SWC_DIR := core/swc
SWC_BUNDLE := core/engine/assets/swc/swc.js
SWC_JS_TOGGLE := core/engine/assets/js/sumeru-password-toggle.js
SWC_JS_MATCH := core/engine/assets/js/sumeru-password-match.js
SWC_JS_APPS_PAGE := core/engine/assets/js/apps-page.js
SWC_ASSET_INPUTS := $(SWC_DIR)/esbuild.config.mjs $(SWC_DIR)/sum-compile.mjs $(SWC_DIR)/package.json

check-sql:
	@bash scripts/check_sql_safety.sh

check-logs:
	@bash scripts/check_no_stdlog.sh

# Match CI go-lint: go vet + golangci-lint v2 (see .golangci.yml).
# Use go run so a stale v1 binary on PATH does not break the target.
lint:
	go vet ./...
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run --timeout=10m

generate:
	go generate ./cmd/sumeru

# Build SWC workspace bundle + login JS (always rebuild).
swc-build:
	cd $(SWC_DIR) && npm install && npm run build

swc: swc-build

# Build client assets when missing or when SWC sources changed (used by run/build).
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

swc-check:
	cd $(SWC_DIR) && npm install && npm run check

swc-test:
	cd $(SWC_DIR) && npm install && npm run test:coverage

# First-time local bootstrap: config, client bundles, Go imports.
setup:
	@test -f sumeru.conf || cp sumeru.conf.example sumeru.conf
	$(MAKE) assets
	$(MAKE) generate

# Dev server: imports + client assets + Go server.
run: generate assets
	go run ./cmd/sumeru -- -c sumeru.conf $(EXTRA_RUN_FLAGS)

dev: run

# Production-style binary next to Makefile.
build: generate assets
	go build -o sumeru ./cmd/sumeru

check: swc-check lint test-modules-static
	go test ./test/... -count=1

test-modules-static:
	go test ./test/module/static/... ./test/core/importgen/... -count=1

test-modules-unit:
	go test ./test/module/unit/... -count=1

test-modules-addon:
	go test ./test/module/addons/... ./test/addons/... -count=1

test-modules: test-modules-static test-modules-unit test-modules-addon

test-modules-integration: test-db
	SUMERU_TEST_DSN='host=localhost port=5433 user=postgres password=postgres dbname=sumeru_test sslmode=disable' \
		go test -tags=integration ./test/integration/... ./test/module/integration/... -count=1

test-coverage:
	go test ./test/... -coverpkg=./... -coverprofile=coverage.out -count=1
	@bash scripts/check_go_coverage.sh coverage.out

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

test-db:
	docker compose -f docker-compose.test.yml up -d --wait

test-integration: test-db
	SUMERU_TEST_DSN='host=localhost port=5433 user=postgres password=postgres dbname=sumeru_test sslmode=disable' \
		go test -tags=integration ./test/integration/... ./test/module/integration/... -count=1

help:
	@echo "Sumeru Makefile — common dev flow:"
	@echo "  make setup   - sumeru.conf (if missing), SWC assets, go generate"
	@echo "  make run     - generate + assets + go run (alias: make dev)"
	@echo "  make build   - generate + assets + go build -o sumeru"
	@echo ""
	@echo "Client (SWC + login JS under core/engine/assets/):"
	@echo "  make assets  - build bundles when missing or sources changed"
	@echo "  make swc     - always rebuild SWC + login JS"
	@echo "  make swc-check / swc-test - TypeScript check / vitest"
	@echo ""
	@echo "Go / addons:"
	@echo "  make generate - refresh cmd/sumeru/zimports.go"
	@echo "  make bp NAME=x - scaffold kernel addon (then make generate)"
	@echo "  make lint    - go vet + golangci-lint (matches CI go-lint)"
	@echo "  make check   - swc-check + lint + test-modules-static + go test ./test/..."
	@echo "  make test-modules - static + unit + addon module suite tiers"
	@echo "  make test-coverage - full repo coverage with 90% gate"
	@echo "  make module  - module CLI (ARGS='list' | 'install sales' | ...)"
	@echo "  make shell   - ORM REPL"
	@echo ""
	@echo "Other: db-check | i18n-export | i18n-import | test-integration | check-sql | check-logs"
	@echo "Vars: EXTRA_RUN_FLAGS='-p 9090 -d mydb'"
	@echo "Prerequisites: Go 1.26.6+, Node.js (npm), PostgreSQL — see README.md"

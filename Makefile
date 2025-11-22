ROOT := $(CURDIR)
GO ?= go
GOFUMPT_VERSION ?= v0.9.2
GOLANGCI_LINT_VERSION ?= v1.64.8
ARTIFACT_DIR := $(ROOT)/specs/001-go-agno-rewrite/artifacts
COVER_DIR := $(ROOT)/specs/001-go-agno-rewrite/artifacts/coverage
BENCH_DIR := $(ROOT)/specs/001-go-agno-rewrite/artifacts/bench
LOG_DIR := $(ARTIFACT_DIR)/logs
COVER_PROFILE ?= $(COVER_DIR)/coverage.out
COVER_FUNC ?= $(COVER_DIR)/coverage.txt
BENCH_OUTPUT ?= $(BENCH_DIR)/bench.txt
DIST_DIR ?= $(ROOT)/dist
RELEASE_PLATFORMS ?= linux/amd64 linux/arm64 darwin/arm64
BENCH_BASELINE ?= $(BENCH_DIR)/python_baseline.txt
FIXTURE_SOURCE_DIR ?= $(ROOT)/specs/001-go-agno-rewrite/contracts/fixtures-src
FIXTURE_DEST_DIR ?= $(ROOT)/specs/001-go-agno-rewrite/contracts/fixtures
VERIFY_ONLY ?= false
GOCACHE_DIR ?= $(ROOT)/.cache/go-build
DEFAULT_GOMEMLIMIT ?= 2GiB
DEFAULT_GOGC ?= 120
GO_ENV_BASE := GOCACHE=$(GOCACHE_DIR)
BENCH_ENV := $(GO_ENV_BASE) GOMEMLIMIT=$${GOMEMLIMIT:-$(DEFAULT_GOMEMLIMIT)} GOGC=$${GOGC:-$(DEFAULT_GOGC)}

.PHONY: help fmt lint test providers-test coverage bench gen-fixtures release constitution-check tidy audit-no-python docs-build docs-serve docs-check

help: ## Show available targets
	@echo "Available targets:"
	@grep -E '^[a-zA-Z0-9_.-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*##"} {printf "  %-22s %s\n", $$1, $$2}'

fmt: | $(GOCACHE_DIR) ## Format Go code with gofumpt
	@echo "==> gofumpt ./..."
	@cd $(ROOT)/go && $(GO_ENV_BASE) $(GO) run mvdan.cc/gofumpt@$(GOFUMPT_VERSION) -w .

lint: | $(GOCACHE_DIR) ## Run golangci-lint with configured linters
	@echo "==> golangci-lint ./..."
	@cd $(ROOT)/go && command -v golangci-lint >/dev/null || { echo "golangci-lint not installed; run '$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)'"; exit 1; }
	@cd $(ROOT)/go && $(GO_ENV_BASE) golangci-lint run ./...

test: | $(GOCACHE_DIR) ## Run unit and package tests
	@echo "==> go test ./..."
	@cd $(ROOT)/go && $(GO_ENV_BASE) $(GO) test ./...

providers-test: | $(GOCACHE_DIR) ## Run provider integration tests (env-gated)
	@echo "==> providers integration tests (env-gated)"
	@cd $(ROOT)/go && $(GO_ENV_BASE) $(GO) test ./tests/providers

coverage: | $(GOCACHE_DIR) ## Generate coverage profile and summary
	@echo "==> coverage profile -> $(COVER_PROFILE)"
	@mkdir -p $(COVER_DIR)
	@cd $(ROOT)/go && $(GO_ENV_BASE) $(GO) test ./... -coverpkg=./... -coverprofile=$(COVER_PROFILE) -covermode=atomic
	@cd $(ROOT)/go && $(GO_ENV_BASE) $(GO) tool cover -func=$(COVER_PROFILE) > $(COVER_FUNC)

bench: | $(GOCACHE_DIR) ## Run benchmarks and summarize with benchstat
	@echo "==> benchmark -> $(BENCH_OUTPUT)"
	@mkdir -p $(BENCH_DIR)
	@cd $(ROOT)/go && $(BENCH_ENV) $(GO) test -run=^$$ -bench=. -benchmem ./... | tee $(BENCH_OUTPUT)
	@command -v benchstat >/dev/null || { echo "benchstat not installed; run '$(GO) install golang.org/x/perf/cmd/benchstat@latest'"; exit 1; }
	@if [ -n "$(BENCH_BASELINE)" ] && [ -f "$(BENCH_BASELINE)" ]; then benchstat $(BENCH_BASELINE) $(BENCH_OUTPUT) > $(BENCH_DIR)/benchstat.txt; else benchstat $(BENCH_OUTPUT) > $(BENCH_DIR)/benchstat.txt; fi

gen-fixtures: | $(GOCACHE_DIR) ## Copy sanitized fixtures from precomputed references
	@echo "==> fixture generation from precomputed reference (pure Go)"
	@cd $(ROOT)/go && if [ "$(VERIFY_ONLY)" = "true" ]; then \
		$(GO_ENV_BASE) $(GO) run ./scripts/gen_fixtures --source=$(FIXTURE_SOURCE_DIR) --dest=$(FIXTURE_DEST_DIR) --verify-only; \
	else \
		$(GO_ENV_BASE) $(GO) run ./scripts/gen_fixtures --source=$(FIXTURE_SOURCE_DIR) --dest=$(FIXTURE_DEST_DIR); \
	fi

audit-no-python: ## Ensure no cgo or Python subprocess usage is present
	@echo "==> auditing for forbidden cgo/Python bridges"
	@command -v rg >/dev/null || { echo "ripgrep (rg) is required for audit-no-python"; exit 1; }
	@cd $(ROOT)/go && if rg -n 'import \"C\"' .; then echo "cgo usage detected"; exit 1; else echo "no cgo imports detected"; fi
	@cd $(ROOT)/go && if rg -n 'exec\\.Command\\(\"(python|python3)\"' .; then echo "python subprocess detected; remove runtime dependency"; exit 1; else echo "no python subprocess detected"; fi
	@cd $(ROOT)/go && if rg -n '/agno' .; then echo "python bridge detected; remove runtime dependency"; exit 1; else echo "no runtime python bridges detected"; fi

release: | $(GOCACHE_DIR) ## Build release binaries into dist/ for common platforms
	@echo "==> building release binaries"
	@mkdir -p $(DIST_DIR)
	@cd $(ROOT)/go; \
	for plat in $(RELEASE_PLATFORMS); do \
		OS=$${plat%%/*}; ARCH=$${plat##*/}; \
		OUT="$(DIST_DIR)/agno-$${OS}-$${ARCH}"; \
		echo "  -> $$OUT"; \
		CGO_ENABLED=0 GOOS=$$OS GOARCH=$$ARCH $(GO_ENV_BASE) $(GO) build -o $$OUT ./cmd/agno || exit $$?; \
	done
	@cd $(DIST_DIR) && sha256sum agno-* > sha256sums.txt
	@echo "Artifacts:"
	@cd $(DIST_DIR) && ls -1 agno-* sha256sums.txt

constitution-check: | $(LOG_DIR) ## Run the full constitution check suite and log outputs
	@echo "==> constitution-check (logs -> $(LOG_DIR))"
	@$(MAKE) --no-print-directory fmt 2>&1 | tee $(LOG_DIR)/fmt.log
	@$(MAKE) --no-print-directory lint 2>&1 | tee $(LOG_DIR)/lint.log
	@$(MAKE) --no-print-directory test 2>&1 | tee $(LOG_DIR)/test.log
	@$(MAKE) --no-print-directory providers-test 2>&1 | tee $(LOG_DIR)/providers-test.log
	@$(MAKE) --no-print-directory coverage 2>&1 | tee $(LOG_DIR)/coverage.log
	@$(MAKE) --no-print-directory bench 2>&1 | tee $(LOG_DIR)/bench.log
	@$(MAKE) --no-print-directory audit-no-python 2>&1 | tee $(LOG_DIR)/audit-no-python.log
	@echo "==> constitution-check completed (see $(LOG_DIR))"

tidy: ## Tidy Go module dependencies
	@cd $(ROOT)/go && $(GO_ENV_BASE) $(GO) mod tidy

$(LOG_DIR):
	@mkdir -p $(LOG_DIR) $(COVER_DIR) $(BENCH_DIR)

$(GOCACHE_DIR):
	@mkdir -p $(GOCACHE_DIR)

docs-build: ## Build VitePress documentation site
	@echo "==> docs: build"
	@cd $(ROOT) && if [ ! -f "docs/package.json" ]; then echo "docs/package.json not found; initialize docs/ first"; exit 1; fi
	@cd $(ROOT) && command -v pnpm >/dev/null 2>&1 && pnpm install --filter ./docs || npm install --prefix ./docs
	@cd $(ROOT) && command -v pnpm >/dev/null 2>&1 && pnpm --filter ./docs run docs:build || npm --prefix ./docs run docs:build

docs-serve: ## Serve VitePress docs locally for preview
	@echo "==> docs: dev"
	@cd $(ROOT) && command -v pnpm >/dev/null 2>&1 && pnpm --filter ./docs run docs:dev || npm --prefix ./docs run docs:dev

docs-check: ## Run docs build and path safety checks
	@echo "==> docs: check (build + path scan)"
	@cd $(ROOT) && if [ ! -x "scripts/check-docs-paths.sh" ]; then echo "scripts/check-docs-paths.sh not found or not executable"; exit 1; fi
	@cd $(ROOT) && $(ROOT)/scripts/check-docs-paths.sh
	@$(MAKE) --no-print-directory docs-build

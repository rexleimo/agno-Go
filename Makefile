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
BENCH_BASELINE ?=
FIXTURE_SOURCE_DIR ?= $(ROOT)/specs/001-go-agno-rewrite/contracts/fixtures-src
FIXTURE_DEST_DIR ?= $(ROOT)/specs/001-go-agno-rewrite/contracts/fixtures
VERIFY_ONLY ?= false

.PHONY: help fmt lint test providers-test coverage bench gen-fixtures release constitution-check tidy audit-no-python

help: ## Show available targets
	@echo "Available targets:"
	@grep -E '^[a-zA-Z0-9_.-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*##"} {printf "  %-22s %s\n", $$1, $$2}'

fmt: ## Format Go code with gofumpt
	@echo "==> gofumpt ./..."
	@cd $(ROOT)/go && $(GO) run mvdan.cc/gofumpt@$(GOFUMPT_VERSION) -w .

lint: ## Run golangci-lint with configured linters
	@echo "==> golangci-lint ./..."
	@cd $(ROOT)/go && command -v golangci-lint >/dev/null || { echo "golangci-lint not installed; run '$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)'"; exit 1; }
	@cd $(ROOT)/go && golangci-lint run ./...

test: ## Run unit and package tests
	@echo "==> go test ./..."
	@cd $(ROOT)/go && $(GO) test ./...

providers-test: ## Run provider integration tests (env-gated)
	@echo "==> providers integration tests (env-gated)"
	@cd $(ROOT)/go && $(GO) test ./tests/providers

coverage: ## Generate coverage profile and summary
	@echo "==> coverage profile -> $(COVER_PROFILE)"
	@mkdir -p $(COVER_DIR)
	@cd $(ROOT)/go && $(GO) test ./... -coverpkg=./... -coverprofile=$(COVER_PROFILE) -covermode=atomic
	@cd $(ROOT)/go && $(GO) tool cover -func=$(COVER_PROFILE) > $(COVER_FUNC)

bench: ## Run benchmarks and summarize with benchstat
	@echo "==> benchmark -> $(BENCH_OUTPUT)"
	@mkdir -p $(BENCH_DIR)
	@cd $(ROOT)/go && $(GO) test -run=^$$ -bench=. -benchmem ./... | tee $(BENCH_OUTPUT)
	@command -v benchstat >/dev/null || { echo "benchstat not installed; run '$(GO) install golang.org/x/perf/cmd/benchstat@latest'"; exit 1; }
	@if [ -n "$(BENCH_BASELINE)" ]; then benchstat $(BENCH_BASELINE) $(BENCH_OUTPUT) > $(BENCH_DIR)/benchstat.txt; else benchstat $(BENCH_OUTPUT) > $(BENCH_DIR)/benchstat.txt; fi

gen-fixtures: ## Copy sanitized fixtures from precomputed references
	@echo "==> fixture generation from precomputed reference (pure Go)"
	@cd $(ROOT)/go && if [ "$(VERIFY_ONLY)" = "true" ]; then \
		$(GO) run ./scripts/gen_fixtures.go --source=$(FIXTURE_SOURCE_DIR) --dest=$(FIXTURE_DEST_DIR) --verify-only; \
	else \
		$(GO) run ./scripts/gen_fixtures.go --source=$(FIXTURE_SOURCE_DIR) --dest=$(FIXTURE_DEST_DIR); \
	fi

audit-no-python: ## Ensure no cgo or Python subprocess usage is present
	@echo "==> auditing for forbidden cgo/Python bridges"
	@command -v rg >/dev/null || { echo "ripgrep (rg) is required for audit-no-python"; exit 1; }
	@cd $(ROOT)/go && if rg -n 'import \"C\"' .; then echo "cgo usage detected"; exit 1; else echo "no cgo imports detected"; fi
	@cd $(ROOT)/go && if rg -n 'exec\\.Command\\(\"(python|python3)\"' .; then echo "python subprocess detected; remove runtime dependency"; exit 1; else echo "no python subprocess detected"; fi
	@cd $(ROOT)/go && if rg -n '/agno' .; then echo "python bridge detected; remove runtime dependency"; exit 1; else echo "no runtime python bridges detected"; fi

release: ## Build release binaries into dist/
	@echo "==> building release binaries"
	@mkdir -p $(DIST_DIR)
	@cd $(ROOT)/go && $(GO) build -o $(DIST_DIR)/agno ./cmd/agno
	@echo "Artifacts:"
	@echo "  binary: $(DIST_DIR)/agno"

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
	@cd $(ROOT)/go && $(GO) mod tidy

$(LOG_DIR):
	@mkdir -p $(LOG_DIR) $(COVER_DIR) $(BENCH_DIR)

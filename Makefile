SHELL := /bin/sh

GO ?= go
GOLANGCI_LINT ?= golangci-lint
GORELEASER ?= goreleaser

ACTIONLINT_VERSION ?= v1.7.12
GOIMPORTS_VERSION ?= v0.44.0
GOVULNCHECK_VERSION ?= v1.1.4

.PHONY: help
help: ## Show available targets.
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "%-22s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: quick
quick: fmt-check lint test build ## Run fast local checks.

.PHONY: check
check: actionlint mod-check fmt-check lint vet build goreleaser-check test test-race cover govulncheck ## Run CI-equivalent checks.

.PHONY: fmt
fmt: ## Format Go files.
	$(GO)fmt -w .
	$(GO) run golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION) -w .

.PHONY: fmt-check
fmt-check: ## Check Go formatting and imports.
	@test -z "$$($(GO)fmt -l .)"
	@output="$$($(GO) run golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION) -l .)"; \
	if [ -n "$$output" ]; then printf '%s\n' "$$output"; exit 1; fi

.PHONY: mod-check
mod-check: ## Check go.mod and go.sum are tidy.
	$(GO) mod tidy
	git diff --exit-code go.mod go.sum

.PHONY: lint
lint: ## Run golangci-lint.
	$(GOLANGCI_LINT) run

.PHONY: vet
vet: ## Run go vet.
	$(GO) vet ./...

.PHONY: test
test: ## Run tests.
	$(GO) test ./...

.PHONY: test-race
test-race: ## Run race tests.
	$(GO) test -race ./...

.PHONY: cover
cover: ## Run coverage and print summary.
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

.PHONY: build
build: ## Build all packages.
	$(GO) build ./...

.PHONY: actionlint
actionlint: ## Lint GitHub Actions workflows.
	$(GO) run github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION) .github/workflows/ci.yml .github/workflows/release.yml

.PHONY: govulncheck
govulncheck: ## Check reachable Go vulnerabilities.
	$(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

.PHONY: goreleaser-check
goreleaser-check: ## Check GoReleaser config.
	$(GORELEASER) check

.PHONY: snapshot
snapshot: ## Build local GoReleaser snapshot artifacts.
	$(GORELEASER) release --snapshot --clean

.PHONY: clean
clean: ## Remove generated local artifacts.
	rm -rf dist coverage.out

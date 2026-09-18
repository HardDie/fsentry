# Common developer commands. Run `make help` for the list.

GO           := go
PKG          ?= .
PKGSITE_ADDR ?= localhost:8081
GOLANGCI_LINT ?= golangci-lint
# '^$$' so Make does not eat the empty-regexp '$'.
BENCH_RUN    := '^$$'

.DEFAULT_GOAL := help

.PHONY: help all test test-integration test-all bench \
	vet fmt tidy lint doc doc-all docs-site examples ci clean

## help: Show this list
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //' | awk -F': ' '{printf "  %-22s %s\n", $$1, $$2}'

## all: Unit tests (same as make test)
all: test

## test: Unit tests with race detector
test:
	$(GO) test -race -count=1 ./...

## test-integration: Tests with the integration build tag
test-integration:
	$(GO) test -tags=integration -count=1 ./...

## test-all: Unit then integration tests
test-all: test test-integration

## bench: Benchmarks with allocs/op (does not run unit tests)
bench:
	$(GO) test -bench=. -benchmem -run $(BENCH_RUN) ./...

## vet: go vet all packages
vet:
	$(GO) vet ./...

## fmt: Fail if any Go file needs gofmt
fmt:
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then echo "$$files"; echo "run: gofmt -w on the files above"; exit 1; fi

## tidy: go mod download and tidy
tidy:
	$(GO) mod download
	$(GO) mod tidy

## lint: golangci-lint
lint:
	$(GOLANGCI_LINT) run ./...

## doc: Package summary (PKG=.)
doc:
	$(GO) doc $(PKG)

## doc-all: Package + all exports (PKG=.)
doc-all:
	$(GO) doc -all $(PKG)

## docs-site: HTML godoc at PKGSITE_ADDR (default localhost:8081)
docs-site:
	$(GO) run golang.org/x/pkgsite/cmd/pkgsite@latest -http $(PKGSITE_ADDR)

## examples: Run all example programs
examples:
	@set -e; for d in examples/*/; do \
		if [ -f "$$d/main.go" ]; then echo "==> $$d"; $(GO) run "./$$d"; fi; \
	done

## ci: What CI should run (unit, integration, vet, fmt)
ci: test test-integration vet fmt

## clean: Remove Go build artifacts
clean:
	$(GO) clean ./...

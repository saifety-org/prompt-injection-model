# CI and local checks must use pinned modules, never a developer's go.work.
export GOWORK := off
export GOFLAGS := -mod=readonly
GOLANGCI_LINT_VERSION := v2.14.0
GOLANGCI_LINT ?= $(CURDIR)/bin/golangci-lint

.PHONY: fmt-check lint lint-install ci-test ci-build

fmt-check:
	bash scripts/check-format.sh

lint-install:
	GOBIN=$(CURDIR)/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint: fmt-check vet
	@$(GOLANGCI_LINT) version | grep -F 'version $(GOLANGCI_LINT_VERSION:v%=%) ' >/dev/null || { echo 'Run make lint-install (requires $(GOLANGCI_LINT_VERSION))'; exit 1; }
	$(GOLANGCI_LINT) run --config .golangci.yml ./...

.PHONY: test vet build

test:
	go test -race ./...

vet:
	go vet ./...

build:
	go build ./...

ci-test:
	go test -race -count=1 -timeout=5m ./...

ci-build: build

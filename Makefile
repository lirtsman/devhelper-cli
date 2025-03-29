VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X github.com/lirtsman/devhelper-cli/cmd.Version=$(VERSION) -X github.com/lirtsman/devhelper-cli/cmd.BuildDate=$(BUILD_DATE) -X github.com/lirtsman/devhelper-cli/cmd.Commit=$(COMMIT)"
GOBIN := $(shell go env GOPATH)/bin
COVERAGE_DIR := ./coverage
COVERAGE_PROFILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html
GOLANGCI_LINT_VERSION := v1.59.1

.PHONY: all
all: build

.PHONY: build
build:
	go build $(LDFLAGS) -o devhelper-cli

.PHONY: install
install: build
	cp devhelper-cli $(GOBIN)/devhelper-cli

.PHONY: clean
clean:
	rm -f devhelper-cli
	go clean
	rm -rf $(COVERAGE_DIR)

.PHONY: test
test:
	go test -v ./...

.PHONY: test-coverage
test-coverage: prepare-coverage
	go test -coverprofile=$(COVERAGE_PROFILE) ./...
	@echo "Coverage profile written to $(COVERAGE_PROFILE)"

.PHONY: test-coverage-html
test-coverage-html: test-coverage
	go tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	@echo "HTML coverage report generated at $(COVERAGE_HTML)"
	@echo "Open $(COVERAGE_HTML) in your browser to view the report"

.PHONY: test-coverage-func
test-coverage-func: test-coverage
	go tool cover -func=$(COVERAGE_PROFILE)

.PHONY: prepare-coverage
prepare-coverage:
	mkdir -p $(COVERAGE_DIR)

.PHONY: lint
lint: install-lint-tools
	go vet ./...
	# Temporarily commenting out golangci-lint due to compatibility issues with Go 1.24
	# $(GOBIN)/golangci-lint run

.PHONY: install-lint-tools
install-lint-tools:
	@if ! which golangci-lint > /dev/null || [ "$$(golangci-lint --version | awk '{print $$4}')" != "$(GOLANGCI_LINT_VERSION)" ]; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	else \
		echo "golangci-lint is already installed at the correct version"; \
	fi

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: deps
deps:
	go get ./...
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)

.PHONY: release
release: clean build
	@echo "Built release version $(VERSION)"

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all              - Build the application (default)"
	@echo "  build            - Build the application"
	@echo "  install          - Install the application to GOPATH/bin"
	@echo "  clean            - Remove build artifacts"
	@echo "  test             - Run tests"
	@echo "  test-coverage    - Run tests with coverage and output to file"
	@echo "  test-coverage-html - Generate HTML coverage report"
	@echo "  test-coverage-func - Show function-level coverage stats"
	@echo "  lint             - Run linters"
	@echo "  fmt              - Format code"
	@echo "  deps             - Install dependencies"
	@echo "  release          - Build a release version"
	@echo "  help             - Show this help message" 
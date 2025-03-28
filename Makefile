VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X bitbucket.org/shielddev/shielddev-cli/cmd.Version=$(VERSION) -X bitbucket.org/shielddev/shielddev-cli/cmd.BuildDate=$(BUILD_DATE) -X bitbucket.org/shielddev/shielddev-cli/cmd.Commit=$(COMMIT)"
GOBIN := $(shell go env GOPATH)/bin

.PHONY: all
all: build

.PHONY: build
build:
	go build $(LDFLAGS) -o shielddev-cli

.PHONY: install
install: build
	cp shielddev-cli $(GOBIN)/shielddev-cli

.PHONY: clean
clean:
	rm -f shielddev-cli
	go clean

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	go vet ./...
	$(GOBIN)/golangci-lint run

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: deps
deps:
	go get ./...
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: release
release: clean build
	@echo "Built release version $(VERSION)"

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all        - Build the application (default)"
	@echo "  build      - Build the application"
	@echo "  install    - Install the application to GOPATH/bin"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linters"
	@echo "  fmt        - Format code"
	@echo "  deps       - Install dependencies"
	@echo "  release    - Build a release version"
	@echo "  help       - Show this help message" 
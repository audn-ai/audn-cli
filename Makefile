# audn-cli Makefile

# Variables
BINARY_NAME=audn
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
MAIN_PACKAGE=main.go
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.CommitHash=$(COMMIT_HASH)'"

# OS/Arch detection
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
OS := $(shell echo $(UNAME_S) | tr '[:upper:]' '[:lower:]')
ARCH := $(UNAME_M)

ifeq ($(ARCH),x86_64)
	ARCH := amd64
endif

ifeq ($(OS),darwin)
	PLATFORM := darwin
else ifeq ($(OS),linux)
	PLATFORM := linux
else ifeq ($(OS),windows)
	PLATFORM := windows
	BINARY_NAME := $(BINARY_NAME).exe
endif

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo "$(GREEN)audn-cli Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)Usage:$(NC)"
	@echo "  make [target]"
	@echo ""
	@echo "$(YELLOW)Targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the binary for current OS/Arch
	@echo "$(YELLOW)Building $(BINARY_NAME) for $(PLATFORM)/$(ARCH)...$(NC)"
	@CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)Build complete: ./$(BINARY_NAME)$(NC)"

.PHONY: build-all
build-all: ## Build binaries for all supported platforms
	@echo "$(YELLOW)Building for all platforms...$(NC)"
	@mkdir -p dist
	
	@echo "Building linux/amd64..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	
	@echo "Building linux/arm64..."
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	
	@echo "Building darwin/amd64..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	
	@echo "Building darwin/arm64..."
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	
	@echo "Building windows/amd64..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	
	@echo "$(GREEN)All builds complete in ./dist/$(NC)"

.PHONY: install
install: build ## Install the binary to /usr/local/bin
	@echo "$(YELLOW)Installing $(BINARY_NAME) to /usr/local/bin...$(NC)"
	@sudo mv $(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)Installation complete$(NC)"

.PHONY: uninstall
uninstall: ## Uninstall the binary from /usr/local/bin
	@echo "$(YELLOW)Uninstalling $(BINARY_NAME)...$(NC)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)Uninstallation complete$(NC)"

.PHONY: test
test: ## Run tests
	@echo "$(YELLOW)Running tests...$(NC)"
	@$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "$(GREEN)Tests complete$(NC)"

.PHONY: test-coverage
test-coverage: test ## Run tests with coverage report
	@echo "$(YELLOW)Generating coverage report...$(NC)"
	@$(GO) tool cover -html=coverage.txt -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: lint
lint: ## Run linter
	@echo "$(YELLOW)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "$(RED)golangci-lint not installed. Install it with:$(NC)"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
	fi

.PHONY: fmt
fmt: ## Format code
	@echo "$(YELLOW)Formatting code...$(NC)"
	@$(GO) fmt ./...
	@echo "$(GREEN)Formatting complete$(NC)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(YELLOW)Running go vet...$(NC)"
	@$(GO) vet ./...
	@echo "$(GREEN)Vet complete$(NC)"

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	@echo "$(YELLOW)Tidying modules...$(NC)"
	@$(GO) mod tidy
	@echo "$(GREEN)Modules tidied$(NC)"

.PHONY: mod-download
mod-download: ## Download go modules
	@echo "$(YELLOW)Downloading modules...$(NC)"
	@$(GO) mod download
	@echo "$(GREEN)Modules downloaded$(NC)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning...$(NC)"
	@rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	@rm -rf dist/
	@rm -f coverage.txt coverage.html
	@echo "$(GREEN)Clean complete$(NC)"

.PHONY: run
run: build ## Build and run the binary
	@echo "$(YELLOW)Running $(BINARY_NAME)...$(NC)"
	@./$(BINARY_NAME)

.PHONY: dev
dev: fmt vet test build ## Run full development cycle

.PHONY: ci
ci: mod-tidy fmt vet lint test build ## Run CI checks

.PHONY: release
release: ## Create a new release (requires VERSION parameter)
ifndef VERSION
	@echo "$(RED)Error: VERSION is required$(NC)"
	@echo "Usage: make release VERSION=v1.0.0"
	@exit 1
endif
	@echo "$(YELLOW)Creating release $(VERSION)...$(NC)"
	@git tag -a cli/$(VERSION) -m "Release $(VERSION)"
	@git push origin cli/$(VERSION)
	@echo "$(GREEN)Release $(VERSION) created and pushed$(NC)"

.PHONY: release-dry-run
release-dry-run: ## Dry run of release process
	@echo "$(YELLOW)Dry run of release process...$(NC)"
	@echo "Current version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Commit hash: $(COMMIT_HASH)"
	@$(MAKE) build-all
	@echo "$(GREEN)Dry run complete$(NC)"

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	@docker build -t audn-cli:$(VERSION) .
	@echo "$(GREEN)Docker image built: audn-cli:$(VERSION)$(NC)"

.PHONY: docker-run
docker-run: docker-build ## Build and run Docker container
	@echo "$(YELLOW)Running Docker container...$(NC)"
	@docker run --rm -it audn-cli:$(VERSION)

.PHONY: deps-check
deps-check: ## Check for dependency updates
	@echo "$(YELLOW)Checking for dependency updates...$(NC)"
	@$(GO) list -u -m all

.PHONY: deps-update
deps-update: ## Update dependencies to latest versions
	@echo "$(YELLOW)Updating dependencies...$(NC)"
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

.PHONY: security
security: ## Run security checks
	@echo "$(YELLOW)Running security checks...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(RED)gosec not installed. Install it with:$(NC)"; \
		echo "  go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

.PHONY: version
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Commit Hash: $(COMMIT_HASH)"
	@echo "Platform: $(PLATFORM)/$(ARCH)"

# Default target
.DEFAULT_GOAL := help
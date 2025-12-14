.PHONY: all build test clean help lint fmt docker-build docker-test docker-shell deb

# Binary name
BINARY := sekai-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Docker settings (for docker-build)
DOCKER_GO_VERSION := 1.21-alpine
BUILD_DIR := ./build
DEB_DIR := $(BUILD_DIR)/deb

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

all: docker-build ## Build using Docker

# =============================================================================
# Docker-based builds (primary)
# =============================================================================

docker-build: ## Build binary using Docker
	@mkdir -p $(BUILD_DIR)
	docker run --rm \
		-v "$(PWD)":/app \
		-w /app \
		-e CGO_ENABLED=0 \
		golang:$(DOCKER_GO_VERSION) \
		go build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)" \
		-o $(BUILD_DIR)/$(BINARY) ./cmd/sekai-cli
	@echo "Built $(BUILD_DIR)/$(BINARY)"

docker-build-all: ## Build for all platforms using Docker
	@mkdir -p $(BUILD_DIR)
	docker run --rm \
		-v "$(PWD)":/app \
		-w /app \
		-e CGO_ENABLED=0 \
		golang:$(DOCKER_GO_VERSION) \
		sh -c '\
			GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/sekai-cli && \
			GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/sekai-cli && \
			GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/sekai-cli && \
			GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/sekai-cli && \
			GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/sekai-cli \
		'
	@echo "Built binaries for all platforms in $(BUILD_DIR)/"

docker-test: ## Run tests using Docker
	docker run --rm \
		-v "$(PWD)":/app \
		-w /app \
		golang:$(DOCKER_GO_VERSION) \
		go test -v ./...

docker-test-coverage: ## Run tests with coverage using Docker
	@mkdir -p coverage
	docker run --rm \
		-v "$(PWD)":/app \
		-w /app \
		golang:$(DOCKER_GO_VERSION) \
		sh -c 'go test -coverprofile=/app/coverage/coverage.out ./... && go tool cover -html=/app/coverage/coverage.out -o /app/coverage/coverage.html'
	@echo "Coverage report: coverage/coverage.html"

docker-lint: ## Run linter using Docker
	docker run --rm \
		-v "$(PWD)":/app \
		-w /app \
		golang:$(DOCKER_GO_VERSION) \
		go vet ./...

docker-fmt: ## Format code using Docker
	docker run --rm \
		-v "$(PWD)":/app \
		-w /app \
		golang:$(DOCKER_GO_VERSION) \
		gofmt -s -w .

docker-fmt-check: ## Check code formatting using Docker
	docker run --rm \
		-v "$(PWD)":/app \
		-w /app \
		golang:$(DOCKER_GO_VERSION) \
		sh -c 'if [ -n "$$(gofmt -l .)" ]; then echo "Code not formatted:"; gofmt -l .; exit 1; fi'

docker-shell: ## Open shell in Docker container
	docker run --rm -it \
		-v "$(PWD)":/app \
		-w /app \
		golang:$(DOCKER_GO_VERSION) \
		sh

docker-deps: ## Verify dependencies using Docker
	docker run --rm \
		-v "$(PWD)":/app \
		-w /app \
		golang:$(DOCKER_GO_VERSION) \
		sh -c 'go mod verify && go mod tidy'

# =============================================================================
# Aliases for convenience
# =============================================================================

build: docker-build ## Alias for docker-build
test: docker-test ## Alias for docker-test
lint: docker-lint ## Alias for docker-lint
fmt: docker-fmt ## Alias for docker-fmt

# =============================================================================
# Integration Tests (requires kiranet network from sekin)
# =============================================================================

integration-test: ## Run integration tests against sekin-sekai-1 container
	/usr/local/go/bin/go test -v -count=1 ./test/integration/...

integration-test-single: ## Run single integration test (usage: make integration-test-single TEST=TestGovNetworkProperties)
ifndef TEST
	$(error TEST is required. Usage: make integration-test-single TEST=TestGovNetworkProperties)
endif
	/usr/local/go/bin/go test -v -count=1 -run $(TEST) ./test/integration/...

integration-test-module: ## Run module tests (usage: make integration-test-module MODULE=Status)
ifndef MODULE
	$(error MODULE is required. Usage: make integration-test-module MODULE=Status)
endif
	/usr/local/go/bin/go test -v -count=1 -run "Test$(MODULE)" ./test/integration/...

# =============================================================================
# Debian Package
# =============================================================================

deb: docker-build ## Build Debian package (.deb)
	@echo "Building Debian package..."
	@rm -rf $(DEB_DIR)
	@mkdir -p $(DEB_DIR)/DEBIAN
	@mkdir -p $(DEB_DIR)/usr/bin
	@mkdir -p $(DEB_DIR)/etc/sekai-cli
	@mkdir -p $(DEB_DIR)/usr/share/doc/sekai-cli
	@cp $(BUILD_DIR)/$(BINARY) $(DEB_DIR)/usr/bin/$(BINARY)
	@chmod 755 $(DEB_DIR)/usr/bin/$(BINARY)
	@cp README.md $(DEB_DIR)/usr/share/doc/sekai-cli/
	@echo '{}' > $(DEB_DIR)/etc/sekai-cli/config.json.example
	@echo "Package: sekai-cli" > $(DEB_DIR)/DEBIAN/control
	@echo "Version: $(VERSION)" >> $(DEB_DIR)/DEBIAN/control
	@echo "Section: utils" >> $(DEB_DIR)/DEBIAN/control
	@echo "Priority: optional" >> $(DEB_DIR)/DEBIAN/control
	@echo "Architecture: amd64" >> $(DEB_DIR)/DEBIAN/control
	@echo "Recommends: docker.io | docker-ce" >> $(DEB_DIR)/DEBIAN/control
	@echo "Maintainer: KIRA Network <hello@kira.network>" >> $(DEB_DIR)/DEBIAN/control
	@echo "Description: Command-line interface for SEKAI blockchain" >> $(DEB_DIR)/DEBIAN/control
	@echo " sekai-cli is a command-line tool for interacting with SEKAI blockchain." >> $(DEB_DIR)/DEBIAN/control
	@echo "/etc/sekai-cli/config.json.example" > $(DEB_DIR)/DEBIAN/conffiles
	@dpkg-deb --build $(DEB_DIR) $(BUILD_DIR)/$(BINARY)_$(VERSION)_amd64.deb
	@echo "Built $(BUILD_DIR)/$(BINARY)_$(VERSION)_amd64.deb"

# =============================================================================
# Cleanup
# =============================================================================

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -rf coverage
	rm -rf $(DEB_DIR)
	docker compose down --remove-orphans 2>/dev/null || true

# =============================================================================
# Development helpers
# =============================================================================

new-module: ## Create a new module (usage: make new-module NAME=mymodule)
ifndef NAME
	$(error NAME is required. Usage: make new-module NAME=mymodule)
endif
	@mkdir -p pkg/modules/$(NAME)
	@echo 'package $(NAME)' > pkg/modules/$(NAME)/$(NAME).go
	@echo '' >> pkg/modules/$(NAME)/$(NAME).go
	@echo 'import "github.com/kiracore/sekai-cli/internal/executor"' >> pkg/modules/$(NAME)/$(NAME).go
	@echo '' >> pkg/modules/$(NAME)/$(NAME).go
	@echo 'type Module struct {' >> pkg/modules/$(NAME)/$(NAME).go
	@echo '    exec executor.Executor' >> pkg/modules/$(NAME)/$(NAME).go
	@echo '}' >> pkg/modules/$(NAME)/$(NAME).go
	@echo '' >> pkg/modules/$(NAME)/$(NAME).go
	@echo 'func New(exec executor.Executor) *Module {' >> pkg/modules/$(NAME)/$(NAME).go
	@echo '    return &Module{exec: exec}' >> pkg/modules/$(NAME)/$(NAME).go
	@echo '}' >> pkg/modules/$(NAME)/$(NAME).go
	@echo 'package $(NAME)' > pkg/modules/$(NAME)/types.go
	@echo "Created pkg/modules/$(NAME)/"

tree: ## Show project structure
	@find . -type f -name "*.go" | grep -v vendor | sort

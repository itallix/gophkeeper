# Build variables
BINARY_NAME=cli
BUILD_DIR=bin
BUILD_VERSION="v0.0.1"
BUILD_DATE=$(shell date +"%Y/%m/%d %H:%M")
BUILD_COMMIT=$(shell git rev-parse HEAD)
MAIN_PACKAGE=./cmd/client
SERVER_PACKAGE=./cmd/server
DATABASE_DSN="postgres://postgres:P@ssw0rd@localhost/gophkeeper?sslmode=disable"

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

LDFLAGS := -ldflags "-X 'main.buildVersion=$(BUILD_VERSION)' -X 'main.buildDate=$(BUILD_DATE)' -X 'main.buildCommit=$(BUILD_COMMIT)'"

# Build targets for different architectures
LINUX_AMD64=$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
LINUX_ARM64=$(BUILD_DIR)/$(BINARY_NAME)-linux-arm64
WINDOWS_AMD64=$(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
MACOS_AMD64=$(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
MACOS_ARM64=$(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64

.DEFAULT_GOAL := help
.PHONY: all clean up down migrate-up migrate-down clean-data lint test linux windows macos
all: clean linux windows macos;

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Clean build directory
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

up:
	@docker-compose up

down:
	@docker-compose down

migrate-up:
	migrate -database $(DATABASE_DSN) -path db/migrations up

migrate-down:
	migrate -database $(DATABASE_DSN) -path db/migrations down

clean-data:
	sudo rm -rf ./db/data/

lint:
	golangci-lint run --fix

test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	# go tool cover -func=coverage.out

.PHONY: server
server:
	$(GOBUILD) -o ./bin/server $(SERVER_PACKAGE)

# Build for Linux (AMD64)
linux-amd64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(LINUX_AMD64) $(MAIN_PACKAGE)

# Build for Linux (ARM64)
linux-arm64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(LINUX_ARM64) $(MAIN_PACKAGE)

# Build for Windows (AMD64)
windows-amd64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(WINDOWS_AMD64) $(MAIN_PACKAGE)

# Build for macOS (AMD64)
macos-amd64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(MACOS_AMD64) $(MAIN_PACKAGE)

# Build for macOS (ARM64/M1)
macos-arm64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(MACOS_ARM64) $(MAIN_PACKAGE)

# Combined targets
linux: linux-amd64 linux-arm64
windows: windows-amd64
macos: macos-amd64 macos-arm64

# Create compressed archives for each target
.PHONY: compress
compress: all
	cd $(BUILD_DIR) && \
	tar czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
	tar czf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
	zip $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe && \
	tar czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
	tar czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64

.PHONY: proto
proto:
	protoc --go_out=pkg/generated --go_opt=paths=source_relative \
		--go-grpc_out=pkg/generated --go-grpc_opt=paths=source_relative \
		api/proto/v1/*.proto

# Show help
.PHONY: help
help:
	@echo ""
	@echo "Available targets:"
	@echo "  all          - Build for all platforms"
	@echo "  linux        - Build for Linux (AMD64 and ARM64)"
	@echo "  windows      - Build for Windows (AMD64)"
	@echo "  macos        - Build for macOS (AMD64 and ARM64)"
	@echo "  clean        - Clean build directory"
	@echo "  test         - Run tests"
	@echo "  compress     - Create compressed archives"
	@echo "  server       - Build server for current platform"
	@echo ""
	@echo "Individual architecture targets:"
	@echo "  linux-amd64  - Build for Linux AMD64"
	@echo "  linux-arm64  - Build for Linux ARM64"
	@echo "  windows-amd64- Build for Windows AMD64"
	@echo "  macos-amd64  - Build for macOS AMD64"
	@echo "  macos-arm64  - Build for macOS ARM64 (M1)"
	@echo ""
	@echo " Docker:"
	@echo "   up           - Start services"
	@echo "   down         - Stop services"
	@echo ""
	@echo " Database:"
	@echo "   migrate-up   - Run migrations up"
	@echo "   migrate-down - Run migrations down"
	@echo "   clean-data   - Clean database data"

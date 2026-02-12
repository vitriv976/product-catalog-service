.PHONY: help proto migrate setup-db test run clean docker-up docker-down

# Configuration
PROTO_DIR=proto
PROTO_FILES=$(shell find $(PROTO_DIR) -name '*.proto')
GO_BIN=$(shell pwd)/../../go/bin
PROTOC_BIN=$(shell pwd)/../../protoc/bin
GOPATH=$(shell pwd)/../../gopath
# Use "docker compose" (V2 plugin) or fall back to "docker-compose" (V1 standalone)
DOCKER_COMPOSE := $(shell command -v docker-compose 2>/dev/null && echo docker-compose || echo "docker compose")

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	@export PATH=$(GO_BIN):$(GOPATH)/bin:$(PROTOC_BIN):$$PATH && \
		protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)
	@echo "Protobuf code generated successfully"

docker-up: ## Start Spanner emulator
	@echo "Starting Spanner emulator..."
	@$(DOCKER_COMPOSE) up -d
	@echo "Waiting for Spanner emulator to be ready..."
	@sleep 5
	@echo "Spanner emulator is ready"

docker-down: ## Stop Spanner emulator
	@echo "Stopping Spanner emulator..."
	@$(DOCKER_COMPOSE) down
	@echo "Spanner emulator stopped"

setup-db: docker-up ## Create Spanner database and tables (no spanner-cli needed)
	@echo "Creating Spanner database..."
	@export PATH=$(GO_BIN):$$PATH && \
		export GOPATH=$(GOPATH) && \
		export SPANNER_EMULATOR_HOST=localhost:9010 && \
		go run scripts/setup_db.go
	@echo "Database setup complete"

migrate: setup-db ## Alias for setup-db (creates database and schema)

test: docker-up ## Run all tests
	@echo "Running tests..."
	@export PATH=$(GO_BIN):$$PATH && \
		export GOPATH=$(GOPATH) && \
		export SPANNER_EMULATOR_HOST=localhost:9010 && \
		go test -v ./...
	@echo "Tests completed"

run: docker-up ## Start the gRPC server
	@echo "Starting gRPC server..."
	@export PATH=$(GO_BIN):$$PATH && \
		export GOPATH=$(GOPATH) && \
		export SPANNER_EMULATOR_HOST=localhost:9010 && \
		go run cmd/server/main.go

build: ## Build the server binary
	@echo "Building server..."
	@export PATH=$(GO_BIN):$$PATH && \
		export GOPATH=$(GOPATH) && \
		go build -o bin/server cmd/server/main.go
	@echo "Server built successfully: bin/server"

clean: ## Clean generated files and build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@find $(PROTO_DIR) -name '*.pb.go' -delete
	@echo "Cleaned successfully"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@export PATH=$(GO_BIN):$$PATH && \
		export GOPATH=$(GOPATH) && \
		go mod download
	@echo "Dependencies downloaded"

tidy: ## Tidy go.mod and go.sum
	@echo "Tidying dependencies..."
	@export PATH=$(GO_BIN):$$PATH && \
		export GOPATH=$(GOPATH) && \
		go mod tidy
	@echo "Dependencies tidied"

fmt: ## Format Go code
	@echo "Formatting code..."
	@export PATH=$(GO_BIN):$$PATH && \
		go fmt ./...
	@echo "Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	@export PATH=$(GO_BIN):$$PATH && \
		export GOPATH=$(GOPATH) && \
		go vet ./...
	@echo "go vet completed"

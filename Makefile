.PHONY: proto build run test docker clean deps migrate lint migrate-up migrate-down migrate-rollback migrate-status migrate-create seed fresh refresh reset release

# Application
APP_NAME=go-grst-boilerplate
BUILD_DIR=bin
MIGRATE_CMD=cmd/migrate/main.go

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Generate code from proto files (when using protokit)
proto:
	@echo "Generating proto files..."
	protoc \
		--proto_path=contract \
		--proto_path=$(shell go env GOPATH)/src \
		--go_out=handler/grpc \
		--go_opt=paths=source_relative \
		--go-grpc_out=handler/grpc \
		--go-grpc_opt=paths=source_relative \
		contract/*.proto
	@echo "Proto generation completed"

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) main.go
	@echo "Build completed: $(BUILD_DIR)/$(APP_NAME)"

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	$(GORUN) main.go

# Run the application with hot reload (requires air)
dev:
	@echo "Running $(APP_NAME) with hot reload..."
	air

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -cover ./...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies downloaded"

# ==================== Database Migrations ====================

# Run all pending migrations
migrate: migrate-up

migrate-up:
	@echo "Running migrations..."
	$(GORUN) $(MIGRATE_CMD) up

# Rollback all migrations
migrate-down:
	@echo "Rolling back all migrations..."
	$(GORUN) $(MIGRATE_CMD) down

# Rollback last migration
migrate-rollback:
	@echo "Rolling back last migration..."
	$(GORUN) $(MIGRATE_CMD) rollback

# Show current migration version
migrate-status:
	@echo "Checking migration status..."
	$(GORUN) $(MIGRATE_CMD) status

# Create new migration (usage: make migrate-create name=create_users_table)
migrate-create:
	@echo "Creating migration: $(name)..."
	$(GORUN) $(MIGRATE_CMD) create $(name)

# Run database seeders
seed:
	@echo "Running seeders..."
	$(GORUN) $(MIGRATE_CMD) seed

# Fresh migration (drop all and migrate)
fresh:
	@echo "Running fresh migration..."
	$(GORUN) $(MIGRATE_CMD) fresh

# Fresh migration with seed
fresh-seed:
	@echo "Running fresh migration with seed..."
	$(GORUN) $(MIGRATE_CMD) fresh --seed

# Refresh migrations (rollback all and migrate)
refresh:
	@echo "Refreshing migrations..."
	$(GORUN) $(MIGRATE_CMD) refresh

# Refresh migrations with seed
refresh-seed:
	@echo "Refreshing migrations with seed..."
	$(GORUN) $(MIGRATE_CMD) refresh --seed

# Reset database (rollback all)
reset:
	@echo "Resetting database..."
	$(GORUN) $(MIGRATE_CMD) reset

# ==================== End Migrations ====================

# Lint the code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Format the code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned"

# Docker build
docker:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .
	@echo "Docker image built: $(APP_NAME):latest"

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run --rm -p 3000:3000 -p 50051:50051 --env-file .env $(APP_NAME):latest

# Docker compose up
compose-up:
	docker-compose up -d

# Docker compose down
compose-down:
	docker-compose down

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Development tools installed"

# ==================== Release ====================

# Create a new release (usage: make release VERSION=1.0.0)
release:
ifndef VERSION
	$(error VERSION is not set. Usage: make release VERSION=1.0.0)
endif
	@echo "Creating release v$(VERSION)..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: Working directory is not clean. Commit or stash changes first."; \
		exit 1; \
	fi
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)
	@echo "Release v$(VERSION) created and pushed"

# Create a release candidate (usage: make release-rc VERSION=1.0.0-rc.1)
release-rc:
ifndef VERSION
	$(error VERSION is not set. Usage: make release-rc VERSION=1.0.0-rc.1)
endif
	@echo "Creating release candidate v$(VERSION)..."
	git tag -a v$(VERSION) -m "Release candidate v$(VERSION)"
	git push origin v$(VERSION)
	@echo "Release candidate v$(VERSION) created and pushed"

# Delete a release tag (usage: make release-delete VERSION=1.0.0)
release-delete:
ifndef VERSION
	$(error VERSION is not set. Usage: make release-delete VERSION=1.0.0)
endif
	@echo "Deleting release v$(VERSION)..."
	git tag -d v$(VERSION) || true
	git push origin :refs/tags/v$(VERSION) || true
	@echo "Release v$(VERSION) deleted"

# ==================== End Release ====================

# Help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Application:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make dev            - Run with hot reload (requires air)"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make deps           - Download dependencies"
	@echo "  make lint           - Lint the code"
	@echo "  make fmt            - Format the code"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Database Migrations:"
	@echo "  make migrate        - Run all pending migrations"
	@echo "  make migrate-down   - Rollback all migrations"
	@echo "  make migrate-rollback - Rollback last migration"
	@echo "  make migrate-status - Show current migration version"
	@echo "  make migrate-create name=<name> - Create new migration"
	@echo "  make seed           - Run database seeders"
	@echo "  make fresh          - Drop all and re-migrate"
	@echo "  make fresh-seed     - Drop all, migrate, and seed"
	@echo "  make refresh        - Rollback all and re-migrate"
	@echo "  make refresh-seed   - Rollback all, migrate, and seed"
	@echo "  make reset          - Rollback all migrations"
	@echo ""
	@echo "Docker:"
	@echo "  make docker         - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo "  make compose-up     - Start with docker-compose"
	@echo "  make compose-down   - Stop docker-compose"
	@echo ""
	@echo "Release (Semantic Versioning):"
	@echo "  make release VERSION=x.y.z    - Create and push a release tag"
	@echo "  make release-rc VERSION=x.y.z-rc.1 - Create release candidate"
	@echo "  make release-delete VERSION=x.y.z - Delete a release tag"
	@echo ""
	@echo "Other:"
	@echo "  make proto          - Generate code from proto files"
	@echo "  make install-tools  - Install development tools"

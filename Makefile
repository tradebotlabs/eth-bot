.PHONY: all build run test clean dev web-dev web-build docker-build lint fmt

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Binary name
BINARY_NAME=eth-bot
BINARY_DIR=bin

# Main package
MAIN_PKG=./cmd/bot

# Build flags
LDFLAGS=-ldflags "-s -w"

all: clean build

build:
	@echo "Building..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PKG)
	@echo "Build complete: $(BINARY_DIR)/$(BINARY_NAME)"

run:
	@echo "Running..."
	$(GORUN) $(MAIN_PKG)

dev:
	@echo "Running in development mode..."
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air

test:
	@echo "Running tests..."
	$(GOTEST) -v -race -cover ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@rm -rf web/dist web/node_modules

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

lint:
	@echo "Linting..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOLINT) run ./...

# Web commands
web-install:
	@echo "Installing web dependencies..."
	cd web && npm install

web-dev:
	@echo "Running web in development mode..."
	cd web && npm run dev

web-build:
	@echo "Building web..."
	cd web && npm run build

web-preview:
	@echo "Previewing web build..."
	cd web && npm run preview

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t eth-trading-bot .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env eth-trading-bot

docker-compose-up:
	@echo "Starting with docker-compose..."
	docker-compose up -d

docker-compose-down:
	@echo "Stopping docker-compose..."
	docker-compose down

# Database commands
db-migrate:
	@echo "Running database migrations..."
	$(GORUN) $(MAIN_PKG) migrate

db-reset:
	@echo "Resetting database..."
	rm -f data/trading.db
	$(GORUN) $(MAIN_PKG) migrate

# Backtest commands
backtest:
	@echo "Running backtest..."
	$(GORUN) $(MAIN_PKG) backtest

# Help
help:
	@echo "ETH Trading Bot - Available commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build          - Build the binary"
	@echo "  make run            - Run the application"
	@echo "  make dev            - Run with hot reload (requires air)"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make deps           - Download and tidy dependencies"
	@echo ""
	@echo "Web:"
	@echo "  make web-install    - Install web dependencies"
	@echo "  make web-dev        - Run web dev server"
	@echo "  make web-build      - Build web for production"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo ""
	@echo "Database:"
	@echo "  make db-migrate     - Run migrations"
	@echo "  make db-reset       - Reset database"

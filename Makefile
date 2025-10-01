.PHONY: help format clean build generate-client run deps tidy

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

format: ## Format code using gofmt
	go fmt ./...

clean: ## Clean build artifacts
	rm -rf ./bin

build: ## Build the example application
	go build -o ./bin/judgeval ./examples

deps: ## Download dependencies
	go mod download

tidy: ## Tidy up dependencies
	go mod tidy

generate-client: ## Generate API client from OpenAPI spec
	python3 scripts/generate-client.py
	make format

run: ## Run the example application with environment variables
	@if [ -f .env ]; then \
		set -a && . ./.env && set +a && go run ./examples; \
	else \
		go run ./examples; \
	fi

test: ## Run tests
	go test ./...

install: ## Install the library (for local development)
	go install ./...

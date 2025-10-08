.PHONY: help format clean build generate-client run deps tidy

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

run: ## Run the example application with environment variables (usage: make run <example_name>)
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		echo "Error: Please specify an example to run"; \
		echo "Usage: make run <example_name>"; \
		echo "Available examples:"; \
		echo "  simple-chat"; \
		echo "  manual-otel"; \
		exit 1; \
	fi
	@if [ -f .env ]; then \
		set -a && . ./.env && set +a && cd examples/$(filter-out $@,$(MAKECMDGOALS)) && go run .; \
	else \
		cd examples/$(filter-out $@,$(MAKECMDGOALS)) && go run .; \
	fi

test: ## Run tests
	go test ./...

install: ## Install the library (for local development)
	go install ./...

# Handle arguments passed to make run
%:
	@:

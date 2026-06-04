.PHONY: help all build build-servers run run-local docker run-docker clean

BIN_DIR=bin

TOCK_MCP_SERVER=$(BIN_DIR)/mcp-server-tock-web

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

all: build

build: build-servers

build-servers: ## Build mcp servers
	mkdir -p $(BIN_DIR)
	go mod tidy
	go build -o ./$(TOCK_MCP_SERVER) .

run: build-servers ## Run tock-mcp-server
	$(TOCK_MCP_SERVER) -config ./config.yaml -addr :8083

run-local: build-servers ## Run tock-mcp-server
	$(TOCK_MCP_SERVER) -config ./tmp/config.yaml -addr :8083

docker: ## Build Docker image for tock-mcp-server
	docker build -t tock-mcp-server:local .

run-docker: docker ## Run tock-mcp-server in Docker
	docker run --rm -p 8083:8083 \
      -v "$(pwd)/config.yaml:/config/config.yaml:ro" \
      tock-mcp-server:local

run-mcp-instpector: ## Run mcp-inspector
	npx @modelcontextprotocol/inspector

clean: ## Clean build artifacts
	go clean
	rm -rf $(BIN_DIR)
.PHONY: help all build build-servers run run-local docker run-docker tests clean lintchart security install-security-tools govulncheck gosec

BIN_DIR=bin

TOCK_MCP_SERVER=$(BIN_DIR)/mcp-server-tock-web

# Variables
CHART=k8s/chart
chartversion?=`awk '/^version/ {print $$NF}' ${CHART}/Chart.yaml`
appversion?=`awk '/^appVersion/ {print $$NF}' ${CHART}/Chart.yaml`

# Colors for output
BLUE = \033[0;34m
GREEN = \033[0;32m
RED = \033[0;31m
NC = \033[0m # No Color

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

tests: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	go clean
	rm -rf $(BIN_DIR)

lint-chart: ## Lint the chart
	@echo "$(GREEN)Lint Helm Chart ...$(NC)"
	helm lint ${CHART}
	@echo "$(GREEN)✓ Lint completed$(NC)"

chart-version: ## Display chart version
	@echo "$(BLUE)Chart Application Version:$(appversion)  Chart Version:${chartversion}$(NC)"

security: govulncheck gosec

install-security-tools: ## Install security tools
	@echo "$(GREEN)Installing security tools ...$(NC)"
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "$(GREEN)✓ Security tools installed$(NC)"

govulncheck: ## Run govulncheck security checks
	@echo "$(GREEN)Running govulncheck ...$(NC)"
	govulncheck ./...
	@echo "$(GREEN)✓ govulncheck completed$(NC)"

gosec: ## Run gosec security checks
	@echo "$(GREEN)Running gosec ...$(NC)"
	gosec ./...
	@echo "$(GREEN)✓ gosec completed$(NC)"
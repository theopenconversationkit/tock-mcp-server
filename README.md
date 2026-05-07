# tock-mcp-server
MCP server for T.O.C.K. web-connector

## Building the server

```bash
make build-servers
```

## Running the server

```bash
make run-tock-mcp
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/mcp` | JSON-RPC (stateless, used by most MCP clients) |
| `GET`  | `/mcp` | SSE stream (server-sent events, used by streaming-capable clients) |
| `GET`  | `/health` | Health check — returns `{"status":"ok"}` |

Both transports support all standard MCP methods including `ping`, `initialize`, and `tools/call`.

## Configuration

The server expects a YAML configuration file with the following structure:

```yaml
tock:
  # Base URL of the Tock instance (no trailing slash)
  base_url: "https://<bot-api-host>"
  # Tock namespace
  namespace: "app"
  # Bot name
  bot: "howtonet"
  # Web-connector identifier
  connector: "web"
  # User ID sent to Tock with every request
  user_id: "mcp-user-001"
  # Optional extra HTTP headers forwarded to the Tock web-connector.
  # The value is the default used when the MCP caller does not provide one.
  # Leave empty ("") to make the header optional with no default.
  # extra_headers:
  #  X-Toki-Origin: "github-copilot"   # always sent; can be overridden by the caller
  #  X-Toki-Filter: ""                 # no default — provided by the caller if needed

server:
  # HTTP listen address of the MCP server
  addr: ":8083"
  # Name of the MCP tool exposed to AI clients.
  # Defaults to "ask_tock" if empty or omitted.
  tool_name: "ask_tock"
  # Description of the MCP tool shown to AI clients.
  # This text helps the AI decide when and how to call the tool.
  # Defaults to a built-in description if empty or omitted.
  tool_description: "Ask a question to the Tock documentary chatbot (RAG). Returns the text response and links to source documents."
```

> **Tip:** Customise `tool_name` and `tool_description` to match the domain of your Tock bot so that AI assistants can better decide when to invoke the tool.

## Docker image (scratch)

The image is built with a `scratch` runtime and expects a configuration file available at `/config/config.yaml`.

`config.yaml`
```yaml
tock:
  # Base URL of the Tock instance (no trailing slash)
  base_url: "https://<bot-api-endpoint>"
  # Tock namespace
  namespace: "app"
  # Bot name
  bot: "howtonet"
  # Web-connector identifier
  connector: "web"
  # User ID sent to Tock with every request
  user_id: "mcp-user-001"

server:
  # HTTP listen address of the MCP server
  addr: ":8083"
  # Name and description of the MCP tool (optional — defaults shown below)
  tool_name: "ask_tock"
  tool_description: "Ask a question to the Tock documentary chatbot (RAG). Returns the text response and links to source documents."
```

```bash
docker run --rm -p 8083:8083 \
  -v "$(pwd)/config.yaml:/config/config.yaml:ro" \
  ghcr.io/sacquatella/tock-mcp-server:v0.5.0
```

Build locally:

```bash
docker build -t tock-mcp-server:local .
```

Run locally with a mounted config file:

```bash
docker run --rm -p 8083:8083 \
  -v "$(pwd)/config.yaml:/config/config.yaml:ro" \
  tock-mcp-server:local
```

## Deploying on Kubernetes (ConfigMap)

### with manifest
The container does not embed `config.yaml`, so each environment can mount its own ConfigMap.

Use the provided manifest:

```bash
kubectl apply -f k8s/tock-mcp-server.yaml
```

Example (extract):

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tock-mcp-config
data:
  config.yaml: |
    tock:
      base_url: "https://demo.tock.ai"
      namespace: "sacquatella"
      bot: "howtonet"
      connector: "howtonetweb"
      user_id: "mcp-user-001"
    server:
      addr: ":8083"
      tool_name: "ask_tock"
      tool_description: "Ask a question to the Tock documentary chatbot (RAG). Returns the text response and links to source documents."
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tock-mcp-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tock-mcp-server
  template:
    metadata:
      labels:
        app: tock-mcp-server
    spec:
      containers:
        - name: tock-mcp-server
          image: ghcr.io/sacquatella/tock-mcp-server:v0.5.0
          args: ["-config", "/config/config.yaml", "-addr", ":8083"]
          volumeMounts:
            - name: config
              mountPath: /config
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: tock-mcp-config
```

### With Helm chart

The Helm chart is published to the GHCR OCI registry on every version tag and is also available locally in `k8s/chart/`.

**Install from the OCI registry (recommended):**

```bash
helm install tock-mcp oci://ghcr.io/sacquatella/charts/tock-mcp-server \
  --version <version> -f values.yaml
```

**Install from local source:**

```bash
helm install tock-mcp ./k8s/chart -f values.yaml
```

```yaml
config:
  tock:
    base_url: "https://tock.my-company.com"
    namespace: "app"
    bot: "my-bot"
    connector: "web"
    user_id: "mcp-prod"
  server:
    tool_name: "ask_tock"
    tool_description: "Ask a question to the Tock documentary chatbot (RAG). Returns the text response and links to source documents."

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: tock-mcp.my-company.com
      paths:
        - path: /mcp
          pathType: Prefix
  tls:
    - secretName: tock-mcp-tls
      hosts:
        - tock-mcp.my-company.com
```

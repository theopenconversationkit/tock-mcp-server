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

## Docker image (scratch)

The image is built with a `scratch` runtime and expects a configuration file available at `/etc/tock/config.yaml`.

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
          image: ghcr.io/sacquatella/tock-mcp-server:<tag>
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

## Container CI/CD

A GitHub Actions workflow is available in `.github/workflows/container.yml`:
- it builds the image on pushes and pull requests,
- it only pushes to GHCR when the ref is a tag.

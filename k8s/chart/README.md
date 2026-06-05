# tock-mcp-server Helm Chart

Helm chart for the **Tock MCP Server** — a [Model Context Protocol](https://modelcontextprotocol.io) server bridging AI agents with the [Tock](https://doc.tock.ai) web-connector.

| Chart version | App version | Image |
|---------------|-------------|---|
| 0.1.2         | 0.7.1       | `ghcr.io/theopenconversationkit/tock-mcp-server` |

---

## Prerequisites

- Kubernetes ≥ 1.23
- Helm ≥ 3.10
- (optional) An Ingress controller (nginx, traefik…) if `ingress.enabled: true`
- (optional) [cert-manager](https://cert-manager.io) for automatic TLS

---

## Installation

```bash
# From local checkout
helm install tock-mcp ./k8s/chart \
  --set image.tag=v0.7.1 \
  --set config.tock.base_url=https://your-tock-instance
```

```bash
# Dry-run / preview
helm install tock-mcp ./k8s/chart --dry-run --debug \
  --set image.tag=v0.7.1
```

---

## Configuration

All parameters are set in [`values.yaml`](values.yaml) and can be overridden with `--set` or a custom values file.

### Image

| Parameter | Description | Default |
|---|---|---|
| `image.repository` | Container image | `ghcr.io/theopenconversationkit/tock-mcp-server` |
| `image.tag` | Image tag (empty = Chart `appVersion`) | `""` |
| `image.pullPolicy` | Pull policy | `IfNotPresent` |
| `imagePullSecrets` | Pull secrets for private registries | `[]` |
| `replicaCount` | Number of replicas | `1` |

### Tock configuration

Injected as a `ConfigMap` mounted at `/config/config.yaml` inside the container.

| Parameter                 | Description                                                       | Default                                                                                                          |
|---------------------------|-------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------|
| `config.tock.base_url`    | Base URL of the Tock API                                          | `https://demo.tock.ai`                                                                                           |
| `config.tock.namespace`   | Tock namespace                                                    | `app`                                                                                                            |
| `config.tock.bot`         | Bot name                                                          | `howtonet`                                                                                                       |
| `config.tock.connector`   | Web-connector identifier                                          | `howtonetweb`                                                                                                    |
| `config.tock.user_id`     | User ID sent to Tock                                              | `mcp-user-001`                                                                                                   |
| `config.server.addr`      | HTTP listen address                                               | `:8083`                                                                                                          |
| `config.server.tool_name` | MCP tool name for defined connector                               | `ask_tock`                                                                                                       |
| `config.server.tool_description` | Highly detailed description guiding Agent when to use the tool... | `Ask a question to the Tock documentary chatbot (RAG). Returns the text response and links to source documents.` |
| `input_question_description` | Detailed guidance on what to include in the question...           | `Question to ask the Tock chatbot (RAG). Include context, error messages, version, environment, or desired objective if available` |

### Service

| Parameter | Description | Default |
|---|---|---|
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `8083` |

### Ingress

| Parameter | Description | Default |
|---|---|---|
| `ingress.enabled` | Enable Ingress resource | `false` |
| `ingress.className` | `ingressClassName` (e.g. `nginx`, `traefik`) | `""` |
| `ingress.annotations` | Extra annotations (e.g. cert-manager) | `{}` |
| `ingress.hosts` | List of hosts and paths | see below |
| `ingress.tls` | TLS configuration blocks | `[]` |

Default host example:

```yaml
ingress:
  hosts:
    - host: tock-mcp.example.com
      paths:
        - path: /mcp
          pathType: Prefix
```

### Resources & security

| Parameter | Description | Default |
|---|---|---|
| `resources.requests.cpu` | CPU request | `50m` |
| `resources.requests.memory` | Memory request | `32Mi` |
| `resources.limits.cpu` | CPU limit | `200m` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `podSecurityContext` | Pod-level security context | `runAsNonRoot: true`, uid/gid 65532 |
| `securityContext` | Container-level security context | `readOnlyRootFilesystem: true`, drop ALL |

---

## Examples

### Minimal — no ingress

```bash
helm install tock-mcp ./k8s/chart \
  --set image.tag=v0.6.0 \
  --set config.tock.base_url=https://tock.my-company.com \
  --set config.tock.namespace=myns \
  --set config.tock.bot=mybot \
  --set config.tock.connector=myconnector
```

### With ingress (nginx) and cert-manager TLS

```bash
helm install tock-mcp ./k8s/chart \
  --set image.tag=v0.6.0 \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set "ingress.annotations.cert-manager\.io/cluster-issuer=letsencrypt-prod" \
  --set "ingress.hosts[0].host=tock-mcp.my-company.com" \
  --set "ingress.hosts[0].paths[0].path=/mcp" \
  --set "ingress.hosts[0].paths[0].pathType=Prefix" \
  --set "ingress.tls[0].secretName=tock-mcp-tls" \
  --set "ingress.tls[0].hosts[0]=tock-mcp.my-company.com"
```

### With a custom values file (recommended for production)

```bash
helm install tock-mcp ./k8s/chart -f values-prod.yaml
```

`values-prod.yaml` example:

```yaml
image:
  tag: v0.6.0

config:
  tock:
    base_url: "https://tock.my-company.com"
    namespace: "prod-ns"
    bot: "prod-bot"
    connector: "prod-connector"
    user_id: "mcp-prod"
  server:
    addr: ":8083"

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

---

## Upgrade

```bash
helm upgrade tock-mcp ./k8s/chart \
  --set image.tag=v0.6.0
```

> **Note:** When `config.*` values change, the Deployment automatically rolls out thanks to the `checksum/config` annotation.

## Uninstall

```bash
helm uninstall tock-mcp
```

---

## Chart structure

```
k8s/chart/
├── Chart.yaml
├── values.yaml
└── templates/
    ├── _helpers.tpl      # named templates (fullname, labels…)
    ├── configmap.yaml    # config.yaml mounted at /config/
    ├── deployment.yaml   # Deployment with checksum rolling update
    ├── ingress.yaml      # optional Ingress (ingress.enabled)
    ├── service.yaml      # ClusterIP service
    └── NOTES.txt
```


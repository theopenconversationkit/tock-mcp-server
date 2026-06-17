# tock-mcp-server Helm Chart

Helm chart for the **Tock MCP Server** — a [Model Context Protocol](https://modelcontextprotocol.io) server bridging AI agents with the [Tock](https://doc.tock.ai) web-connector.

| Chart version | App version | Image |
|---------------|-------------|---|
| 0.7.3         | 0.8.0       | `ghcr.io/theopenconversationkit/tock-mcp-server` |

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
| `config.tock.timeout`     | Max time to wait for a Tock API response. Must be lower than `write_timeout`. Go duration format. | `""` → `25s`                                                                                |
| `config.server.addr`      | HTTP listen address                                               | `:8083`                                                                                                          |
| `config.server.tool_name` | MCP tool name for defined connector                               | `ask_tock`                                                                                                       |
| `config.server.tool_description` | Highly detailed description guiding Agent when to use the tool... | `Ask a question to the Tock documentary chatbot (RAG). Returns the text response and links to source documents.` |
| `config.server.input_question_description` | Detailed guidance on what to include in the question...           | `Question to ask the Tock chatbot (RAG). Include context, error messages, version, environment, or desired objective if available` |

### Server timeouts

All fields use the [Go duration format](https://pkg.go.dev/time#ParseDuration) (`"5s"`, `"1m"`, `"1m30s"`, …). Leave empty to use the built-in server default.

| Parameter | Default | Description |
|---|---|---|
| `config.server.read_header_timeout` | `""` → `5s` | Maximum time to read request headers |
| `config.server.read_timeout` | `""` → `15s` | Maximum time to read the full request body |
| `config.server.write_timeout` | `""` → `30s` | Maximum time to write the response |
| `config.server.idle_timeout` | `""` → `60s` | Maximum keep-alive idle time between requests |
| `config.server.shutdown_timeout` | `""` → `10s` | Graceful shutdown deadline for in-flight requests |

### OAuth 2.1

When `config.oauth.enabled` is `true`, every request to `/mcp` must carry a valid Bearer token. The `/health` endpoint is never protected.

| Parameter | Default | Description |
|---|---|---|
| `config.oauth.enabled` | `false` | Enable Bearer token validation on `/mcp` |
| `config.oauth.issuer` | `""` | Expected `iss` claim; also used for OIDC Discovery when `jwks_url` is empty |
| `config.oauth.jwks_url` | `""` | Explicit JWKS endpoint URL (overrides OIDC Discovery) |
| `config.oauth.audience` | `""` | Expected `aud` claim — leave empty to skip |
| `config.oauth.required_scopes` | `[]` | Scopes that must appear in the token's `scope` claim |

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

### Trust store for private CA

Use this when the MCP server must call a Tock URL signed by an internal CA stored in a Secret.

| Parameter | Description | Default |
|---|---|---|
| `trustStore.enabled` | Mount the Secret and set `SSL_CERT_FILE` | `false` |
| `trustStore.secretName` | Secret name containing the CA bundle file | `""` |
| `trustStore.secretKey` | Secret key filename, e.g. `ca.crt` | `ca.crt` |
| `trustStore.mountPath` | Mount directory inside the pod | `/etc/ssl/certs/custom-ca` |

Example:

```yaml
trustStore:
  enabled: true
  secretName: tock-private-ca
  secretKey: ca.crt
```

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
    # Optional: override default HTTP timeouts
    # write_timeout: "60s"
    # shutdown_timeout: "30s"
  # Optional: enable OAuth 2.1 Bearer token validation
  # oauth:
  #   enabled: true
  #   issuer: "https://auth.my-company.com/realms/prod"
  #   audience: "tock-mcp-server"
  #   required_scopes: ["read"]

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

> **Note:** When `config.*` values change, the Deployment automatically rolls out thanks to the `checksum/config` annotation. If you update the CA Secret content, restart the pod to reload the trust store.

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

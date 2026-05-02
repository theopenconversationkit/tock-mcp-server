# Configuring tock-mcp-server with GitHub Copilot

This document explains how to register the `tock-mcp-server` as an MCP tool in GitHub Copilot
and how to leverage the `extra_headers` feature to pass custom HTTP headers to the Tock
web-connector API.

---

## 1. Server-side configuration (`config.yaml`)

The `extra_headers` block declares which HTTP headers are forwarded to Tock.
Each entry maps a **header name** to a **default value** (use an empty string for no default).

```yaml
tock:
  base_url: "https://demo.tock.ai"
  namespace: "sacquatella"
  bot: "howtonet"
  connector: "howtonetweb"
  user_id: "mcp-user-001"
  extra_headers:
    # Always sent with this fixed value — Copilot does not need to pass it.
    X-Toki-Origin: "github-copilot"
    # No default: Copilot may pass a value at call time to filter results by category.
    X-Toki-Filter: ""

server:
  addr: ":8083"
```

**Merge rules:**

| Situation | Header sent to Tock |
|---|---|
| Default set in config, caller does not pass a value | Config default |
| Default set in config, caller overrides it | Caller's value |
| No default in config, caller passes a value | Caller's value |
| No default in config, caller omits it | Header **not sent** |
| Caller passes a key not declared in `extra_headers` | **Silently ignored** (allowlist) |

---

## 2. Registering the server in GitHub Copilot

### VS Code — `.vscode/mcp.json`

Place this file at the root of your workspace (or in `~/.vscode/`for a global registration):

```json
{
  "servers": {
    "tock-mcp-server": {
      "type": "http",
      "url": "http://localhost:8083/mcp"
    }
  }
}
```

### JetBrains (GoLand, IntelliJ…) — `~/.config/github-copilot/mcp.json`

```json
{
  "servers": {
    "tock-mcp-server": {
      "type": "http",
      "url": "http://localhost:8083/mcp"
    }
  }
}
```

> **Note:** start the server before opening Copilot chat:
> ```bash
> make run-tock-mcp
> # or with a custom config
> ./bin/mcp-server-tock-web -config ./config.yaml -addr :8083
> ```

---

## 3. Teaching Copilot how to use extra headers

Create `.github/copilot-instructions.md` at the root of your repository.
Copilot reads this file automatically and applies the instructions in every chat session.

```markdown
## Tock MCP Tool — ask_tock

When a user asks a question that requires documentation, technical knowledge, or FAQ lookup,
use the `ask_tock` tool from the `tock-mcp-server`.

### Headers

Always include the `headers` argument when calling `ask_tock`.

| Header         | Description                                   | Example values                                                      |
|----------------|-----------------------------------------------|---------------------------------------------------------------------|
| `X-Toki-Filter`| Restricts Tock search to a specific category  | `"category:faq"`, `"category:documentation"`, `"category:api"`     |
| `X-Toki-Origin`| Identifies the caller (default already set)   | omit — the server sets it automatically                             |

Omit `X-Toki-Filter` entirely if the question does not match a specific category.
```

---

## 4. Call examples

### Minimal call — no extra headers needed

The server automatically adds `X-Toki-Origin: github-copilot` from its default configuration.

```json
{
  "question": "What is the architecture of the platform?"
}
```

---

### With `X-Toki-Filter` to restrict results to FAQ entries

```json
{
  "question": "How do I reset my password?",
  "headers": {
    "X-Toki-Filter": "category:faq"
  }
}
```

---

### With `X-Toki-Filter` restricted to documentation

```json
{
  "question": "How does the web-connector authentication work?",
  "headers": {
    "X-Toki-Filter": "category:documentation"
  }
}
```

---

### Overriding the default origin (e.g. from a CI pipeline)

```json
{
  "question": "What are the breaking changes in the latest release?",
  "headers": {
    "X-Toki-Filter": "category:api",
    "X-Toki-Origin": "ci-pipeline"
  }
}
```

---

### Raw HTTP call (curl)

```bash
curl -X POST http://localhost:8083/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "ask_tock",
      "arguments": {
        "question": "How do I configure a web connector?",
        "headers": {
          "X-Toki-Filter": "category:documentation"
        }
      }
    }
  }'
```

---

## 5. End-to-end flow

```
Copilot / MCP client
  │
  │  ask_tock(question="...", headers={"X-Toki-Filter": "category:faq"})
  ▼
tock-mcp-server
  │  merges headers:
  │    X-Toki-Origin : "github-copilot"   ← server default (config.yaml)
  │    X-Toki-Filter : "category:faq"     ← provided by the caller
  │
  │  POST <base_url>/io/<namespace>/<bot>/<connector>
  │  Content-Type: application/json
  │  X-Toki-Origin: github-copilot
  │  X-Toki-Filter: category:faq
  ▼
Tock web-connector API
  │
  │  { "responses": [ ... ] }
  ▼
tock-mcp-server  →  formats response as Markdown  →  Copilot
```


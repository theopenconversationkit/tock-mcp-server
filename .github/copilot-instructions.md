## Tock MCP Tool — ask_tock

When a user asks a question that requires documentation, technical knowledge, or FAQ lookup,
use the `ask_tock` tool from the `tock-mcp-server`.

### Headers

Always include the `headers` argument when calling `ask_tock`.

| Header | Description | Values |
|---|---|---|
| `X-Toki-Filter` | Restricts Tock search to a specific category | `"category:faq"`, `"category:documentation"`, `"category:api"` — omit if no category applies |
| `X-Toki-Origin` | Identifies the caller (already set as default server-side, no need to pass) | — |

### Example call

```json
{
  "question": "How do I reset my password?",
  "headers": {
    "X-Toki-Filter": "category:faq"
  }
}
```

If the user's question does not match a specific category, omit `X-Toki-Filter` entirely:

```json
{
  "question": "What is the architecture of the platform?"
}
```


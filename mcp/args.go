package mcp

// AskTockArgs holds the input parameters for the ask_tock MCP tool.
type AskTockArgs struct {
	Question string            `json:"question"          jsonschema:"Question to ask the Tock chatbot (RAG). Include context, error messages, version, environment, or desired objective if available."`
	Headers  map[string]string `json:"headers,omitempty" jsonschema:"Optional values for the extra HTTP headers declared in the server configuration (e.g. X-Toki-Filter, X-Toki-Origin). Unknown keys are silently ignored."`
}

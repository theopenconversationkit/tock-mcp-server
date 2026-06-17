package mcp

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/theopenconversationkit/tock-mcp-server/config"
	"github.com/theopenconversationkit/tock-mcp-server/tock"
)

// RegisterTool registers the ask_tock tool on the given MCP server.
// It uses the provided Tock client to query the Tock API and formats responses as Markdown.
// The tool name, description, and input question description are all configurable via
// the ServerConfig (config.yaml: server.tool_name, server.tool_description, server.input_question_description).
func RegisterTool(server *mcp.Server, tockClient *tock.Client, serverCfg config.ServerConfig) {
	// Use the name and description from the config file, or fall back to sensible defaults.
	toolName := serverCfg.ToolName
	if strings.TrimSpace(toolName) == "" {
		toolName = "ask_tock"
	}
	toolDescription := serverCfg.ToolDescription
	if strings.TrimSpace(toolDescription) == "" {
		toolDescription = "Ask a question to the Tock documentary chatbot (RAG). Returns the text response and links to source documents."
	}
	inputQuestionDescription := serverCfg.InputQuestionDescription
	if strings.TrimSpace(inputQuestionDescription) == "" {
		inputQuestionDescription = "Question to ask the Tock chatbot (RAG). Include context, error messages, version, environment, or desired objective if available."
	}

	// Build the input schema dynamically so that inputQuestionDescription
	// from the config file is reflected in the tool definition exposed to AI clients.
	inputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"question": map[string]interface{}{
				"type":        "string",
				"description": inputQuestionDescription,
			},
			"headers": map[string]interface{}{
				"type":                 "object",
				"description":          "Optional values for the extra HTTP headers declared in the server configuration (e.g. X-Toki-Filter, X-Toki-Origin). Unknown keys are silently ignored.",
				"additionalProperties": map[string]interface{}{"type": "string"},
			},
		},
		"required":             []string{"question"},
		"additionalProperties": false,
	}

	// Register the ask_tock tool: takes a question string and returns the
	// Tock RAG answer formatted as Markdown text.
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        toolName,
			Description: toolDescription,
			InputSchema: inputSchema,
			Annotations: &mcp.ToolAnnotations{
				ReadOnlyHint:    true,
				DestructiveHint: new(false),
				IdempotentHint:  true,
			},
		},
		func(ctx context.Context, req *mcp.CallToolRequest, args AskTockArgs) (*mcp.CallToolResult, any, error) {
			if strings.TrimSpace(args.Question) == "" {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: "Question cannot be empty."},
					},
				}, nil, nil
			}

			log.Printf("[ask_tock] question: %q headers: %v", args.Question, args.Headers)

			// Créer un contexte avec un timeout explicite pour l'appel Tock,
			// car le WriteTimeout du serveur HTTP ne cancel pas le ctx du handler.
			//tockCtx, cancel := context.WithTimeout(ctx, serverCfg.WriteTimeout)
			//defer cancel()

			tockResp, err := tockClient.Ask(ctx, args.Question, args.Headers)
			if err != nil {
				if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
					log.Printf("[ask_tock] timeout: Tock API did not respond within %s", serverCfg.WriteTimeout)
				} else {
					log.Printf("[ask_tock] error calling Tock: %v", err)
				}
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Tock error: %v", err)},
					},
				}, nil, nil
			}

			formatted := tock.FormatResponse(tockResp)
			log.Printf("[ask_tock] response: %q", formatted)

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: formatted},
				},
			}, nil, nil
		},
	)
}

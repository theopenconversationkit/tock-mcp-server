package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

// ── Configuration ────────────────────────────────────────────────────────────

// TockConfig holds the connection parameters for the Tock web-connector API.
type TockConfig struct {
	BaseURL      string            `yaml:"base_url"`      // Base URL of the Tock instance, without trailing slash.
	Namespace    string            `yaml:"namespace"`     // Tock namespace (usually the organisation name).
	Bot          string            `yaml:"bot"`           // Bot identifier within the namespace.
	Connector    string            `yaml:"connector"`     // Web-connector identifier exposed by the bot.
	UserID       string            `yaml:"user_id"`       // User ID sent to Tock with every request.
	ExtraHeaders map[string]string `yaml:"extra_headers"` // Optional HTTP headers forwarded to Tock (name → default value; empty string means no default).
}

// ServerConfig holds the HTTP server parameters.
type ServerConfig struct {
	Addr            string `yaml:"addr"`             // Listen address, e.g. ":8083".
	ToolName        string `yaml:"tool_name"`        // Name of the MCP tool exposed to AI clients. Defaults to "ask_tock" if empty.
	ToolDescription string `yaml:"tool_description"` // Description of the MCP tool shown to AI clients. Falls back to a built-in default if empty.
}

// Config is the top-level configuration structure loaded from the YAML file.
type Config struct {
	Tock   TockConfig   `yaml:"tock"`
	Server ServerConfig `yaml:"server"`
}

// loadConfig reads and parses the YAML configuration file at path.
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// ── Tock API types ────────────────────────────────────────────────────────────

// TockQuery is the request payload sent to the Tock web-connector.
type TockQuery struct {
	Query  string `json:"query"`  // Natural-language question from the user.
	UserID string `json:"userId"` // Identifier of the caller passed to Tock.
}

// TockButton represents a clickable action returned by Tock (link, postback, quick-reply).
type TockButton struct {
	Title   string `json:"title"`             // Display label shown to the user.
	Payload string `json:"payload,omitempty"` // Postback payload (postback/quick_reply buttons).
	URL     string `json:"url,omitempty"`     // Target URL (web_url buttons).
	Type    string `json:"type"`              // Button type: "web_url", "postback", or "quick_reply".
}

// TockFile holds metadata for a file or image attachment returned by Tock.
type TockFile struct {
	URL  string `json:"url"`  // Public URL of the file.
	Name string `json:"name"` // Display name of the file.
	Type string `json:"type"` // MIME type or file category.
}

// TockCard is a rich-content card (title, subtitle, optional image and buttons).
type TockCard struct {
	Title    string       `json:"title,omitempty"`
	SubTitle string       `json:"subTitle,omitempty"`
	File     *TockFile    `json:"file,omitempty"`
	Buttons  []TockButton `json:"buttons,omitempty"`
}

// TockCarousel wraps a list of TockCard items returned as a carousel.
type TockCarousel struct {
	Cards []TockCard `json:"cards,omitempty"`
}

// TockMessage is a single message element within a Tock response.
// A response can mix plain text, cards, carousels, and standalone buttons.
type TockMessage struct {
	Text     string        `json:"text,omitempty"`
	Buttons  []TockButton  `json:"buttons,omitempty"`
	Card     *TockCard     `json:"card,omitempty"`
	Carousel *TockCarousel `json:"carousel,omitempty"`
}

// TockResponse is the top-level payload returned by the Tock web-connector.
type TockResponse struct {
	Responses []TockMessage `json:"responses"` // Ordered list of message elements.
}

// ── Tock client ───────────────────────────────────────────────────────────────

// TockClient wraps an HTTP client and the Tock connection configuration.
type TockClient struct {
	cfg  TockConfig
	http *http.Client
}

// NewTockClient creates a TockClient ready to query the given Tock configuration.
func NewTockClient(cfg TockConfig) *TockClient {
	return &TockClient{cfg: cfg, http: &http.Client{}}
}

// Ask sends question to the Tock web-connector and returns the structured response.
// The URL is built from the configured namespace, bot and connector identifiers.
// callHeaders provides per-call values for the extra headers declared in ExtraHeaders;
// they override the default values from the configuration.
// Only headers declared in ExtraHeaders are forwarded (unknown keys in callHeaders are ignored).
func (t *TockClient) Ask(ctx context.Context, question string, callHeaders map[string]string) (*TockResponse, error) {
	// Build the endpoint: <base_url>/io/<namespace>/<bot>/<connector>
	url := fmt.Sprintf(
		"%s/io/%s/%s/%s",
		strings.TrimRight(t.cfg.BaseURL, "/"),
		t.cfg.Namespace,
		t.cfg.Bot,
		t.cfg.Connector,
	)

	body, err := json.Marshal(TockQuery{Query: question, UserID: t.cfg.UserID})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Merge extra headers: config default is overridden by the call-time value when non-empty.
	// Only header names declared in ExtraHeaders are forwarded (allowlist).
	for name, defaultVal := range t.cfg.ExtraHeaders {
		val := defaultVal
		if v, ok := callHeaders[name]; ok && v != "" {
			val = v
		}
		if val != "" {
			req.Header.Set(name, val)
		}
	}

	resp, err := t.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Tock call: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Treat any non-2xx status as a hard error and surface the body for debugging.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tock status=%d body=%s", resp.StatusCode, string(raw))
	}

	var tockResp TockResponse
	if err := json.Unmarshal(raw, &tockResp); err != nil {
		return nil, fmt.Errorf("parse Tock response: %w", err)
	}

	return &tockResp, nil
}

// ── Format Tock response → readable text ─────────────────────────────────────

// formatTockResponse converts a TockResponse into a plain-text / Markdown string
// suitable for display in an MCP text content block.
func formatTockResponse(r *TockResponse) string {
	var sb strings.Builder

	for _, msg := range r.Responses {
		if msg.Text != "" {
			sb.WriteString(msg.Text)
			sb.WriteString("\n")
		}

		if msg.Card != nil {
			formatCard(&sb, msg.Card)
		}

		// Flatten carousels: each card is rendered sequentially.
		if msg.Carousel != nil {
			for _, card := range msg.Carousel.Cards {
				formatCard(&sb, &card)
			}
		}

		for _, btn := range msg.Buttons {
			formatButton(&sb, btn)
		}
	}

	return strings.TrimSpace(sb.String())
}

// formatCard appends a card's title, subtitle, optional file link and buttons to sb.
func formatCard(sb *strings.Builder, card *TockCard) {
	if card.Title != "" {
		sb.WriteString(fmt.Sprintf("**%s**\n", card.Title))
	}
	if card.SubTitle != "" {
		sb.WriteString(fmt.Sprintf("%s\n", card.SubTitle))
	}
	if card.File != nil && card.File.URL != "" {
		sb.WriteString(fmt.Sprintf(" %s : %s\n", card.File.Name, card.File.URL))
	}
	for _, btn := range card.Buttons {
		formatButton(sb, btn)
	}
}

// formatButton appends a single button line to sb.
// web_url buttons are rendered as Markdown links; others as list items.
func formatButton(sb *strings.Builder, btn TockButton) {
	switch btn.Type {
	case "web_url":
		sb.WriteString(fmt.Sprintf(" [%s](%s)\n", btn.Title, btn.URL))
	case "postback", "quick_reply":
		sb.WriteString(fmt.Sprintf("▶ %s\n", btn.Title))
	default:
		sb.WriteString(fmt.Sprintf("• %s\n", btn.Title))
	}
}

// ── MCP Tool args ─────────────────────────────────────────────────────────────

// AskTockArgs holds the input parameters for the ask_tock MCP tool.
type AskTockArgs struct {
	Question string            `json:"question"          jsonschema:"The question to ask the documentary chatbot"`
	Headers  map[string]string `json:"headers,omitempty" jsonschema:"Optional values for the extra HTTP headers declared in the server configuration (e.g. X-Toki-Filter, X-Toki-Origin). Unknown keys are silently ignored."`
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	configPath := flag.String("config", "config.yaml", "path to the YAML configuration file")
	addrFlag := flag.String("addr", "", "HTTP listen address for the MCP server (overrides config)")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// CLI flag takes precedence over the value in the config file.
	if *addrFlag != "" {
		cfg.Server.Addr = *addrFlag
	}
	// Fall back to a sensible default if neither the file nor the flag provides an address.
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = ":8083"
	}

	tockClient := NewTockClient(cfg.Tock)

	// Create the MCP server instance with its name and version metadata.
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "tock-web-mcp-server",
			Version: "1.0.0",
		},
		nil,
	)

	// Use the name and description from the config file, or fall back to sensible defaults.
	toolName := cfg.Server.ToolName
	if strings.TrimSpace(toolName) == "" {
		toolName = "ask_tock"
	}
	toolDescription := cfg.Server.ToolDescription
	if strings.TrimSpace(toolDescription) == "" {
		toolDescription = "Ask a question to the Tock documentary chatbot (RAG). Returns the text response and links to source documents."
	}

	// Register the ask_tock tool: takes a question string and returns the
	// Tock RAG answer formatted as Markdown text.
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        toolName,
			Description: toolDescription,
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

			tockResp, err := tockClient.Ask(ctx, args.Question, args.Headers)
			if err != nil {
				// Surface the error as an MCP tool error rather than crashing the server.
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Tock error: %v", err)},
					},
				}, nil, nil
			}

			formatted := formatTockResponse(tockResp)
			log.Printf("[ask_tock] response: %q", formatted)

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: formatted},
				},
			}, nil, nil
		},
	)

	// StreamableHTTPHandler exposes the MCP server over HTTP.
	// SSE (GET /mcp) and JSON-RPC (POST /mcp) are both supported.
	// JSONResponse is left to false so that GET /mcp opens a proper SSE stream
	// and ping/notifications are handled natively by the SDK.
	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return server },
		nil,
	)

	// Wrap the handler with a request logger before registering on the mux.
	logged := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
	})

	mux := http.NewServeMux()
	mux.Handle("/mcp", logged)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	log.Printf("tock-web MCP server listening on http://localhost%s/mcp", cfg.Server.Addr)
	log.Printf("→ POST /mcp  JSON-RPC (stateless)")
	log.Printf("→ GET  /mcp  SSE stream (server-sent events)")
	log.Printf("→ Tock endpoint: %s/io/%s/%s/%s",
		cfg.Tock.BaseURL, cfg.Tock.Namespace, cfg.Tock.Bot, cfg.Tock.Connector)

	if err := http.ListenAndServe(cfg.Server.Addr, mux); err != nil {
		log.Fatal(err)
	}
}

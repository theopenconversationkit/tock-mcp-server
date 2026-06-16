package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/theopenconversationkit/tock-mcp-server/api"
	"github.com/theopenconversationkit/tock-mcp-server/config"
	mcp_pkg "github.com/theopenconversationkit/tock-mcp-server/mcp"
	"github.com/theopenconversationkit/tock-mcp-server/tock"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to the YAML configuration file")
	addrFlag := flag.String("addr", "", "HTTP listen address for the MCP server (overrides config)")
	flag.Parse()

	// Load configuration.
	cfg, err := config.Load(*configPath)
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

	// Create Tock client.
	tockClient := tock.NewClient(cfg.Tock)

	// Create the MCP server instance with its name and version metadata.
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "tock-web-mcp-server",
			Version: "1.0.0",
		},
		nil,
	)

	// Register the ask_tock tool.
	mcp_pkg.RegisterTool(server, tockClient, cfg.Server)

	// Create the streamable HTTP handler for the MCP server.
	handler := api.NewStreamableHandler(
		func(r *http.Request) *mcp.Server { return server },
	)

	// Start the HTTP server.
	if err := api.StartServer(
		handler,
		cfg.Server.Addr,
		cfg.Tock.BaseURL,
		cfg.Tock.Namespace,
		cfg.Tock.Bot,
		cfg.Tock.Connector,
		cfg.OAuth,
	); err != nil {
		log.Fatal(err)
	}
}

package api

import (
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StartServer creates and starts the MCP HTTP server with the given handler.
// It exposes:
// - POST /mcp  for JSON-RPC (stateless)
// - GET  /mcp  for SSE stream (server-sent events)
// - GET  /health for health checks
func StartServer(handler http.Handler, addr string, tockBaseURL, namespace, bot, connector string) error {
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

	log.Printf("tock-web MCP server listening on http://localhost%s/mcp", addr)
	log.Printf("→ POST /mcp  JSON-RPC (stateless)")
	log.Printf("→ GET  /mcp  SSE stream (server-sent events)")
	log.Printf("→ Tock endpoint: %s/io/%s/%s/%s", tockBaseURL, namespace, bot, connector)

	return http.ListenAndServe(addr, mux)
}

// NewStreamableHandler creates an MCP streamable HTTP handler.
// JSONResponse is set to false so that GET /mcp opens a proper SSE stream
// and ping/notifications are handled natively by the SDK.
func NewStreamableHandler(serverFunc func(r *http.Request) *mcp.Server) http.Handler {
	return mcp.NewStreamableHTTPHandler(serverFunc, nil)
}

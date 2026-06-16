package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/theopenconversationkit/tock-mcp-server/config"
)

// StartServer creates and starts the MCP HTTP server with the given handler.
// It exposes:
// - POST /mcp  for JSON-RPC (stateless)
// - GET  /mcp  for SSE stream (server-sent events)
// - GET  /health for health checks
//
// When oauthCfg.Enabled is true, the /mcp endpoint is protected by OAuth 2.1
// Bearer token validation. The /health endpoint is never protected.
func StartServer(handler http.Handler, addr string, tockBaseURL, namespace, bot, connector string, oauthCfg config.OAuthConfig) error {
	// Build the OAuth middleware (no-op if disabled).
	oauthMiddleware, err := OAuthMiddleware(oauthCfg)
	if err != nil {
		return fmt.Errorf("oauth middleware init: %w", err)
	}

	// Wrap the handler with a request logger before registering on the mux.
	// Only log a static string to avoid log injection via user-controlled fields (G706).
	logged := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("incoming request on /mcp")
		handler.ServeHTTP(w, r)
	})

	// Apply OAuth middleware on the /mcp endpoint (no-op when OAuth is disabled).
	protected := oauthMiddleware(logged)

	mux := http.NewServeMux()
	mux.Handle("/mcp", protected)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	log.Printf("tock-web MCP server listening on http://localhost%s/mcp", addr)
	log.Printf("POST /mcp  JSON-RPC (stateless)")
	log.Printf("GET  /mcp  SSE stream (server-sent events)")
	log.Printf("Tock endpoint: %s/io/%s/%s/%s", tockBaseURL, namespace, bot, connector)

	// Use http.Server with explicit timeouts instead of http.ListenAndServe (G114).
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Run the server in a goroutine so we can listen for OS signals.
	serveErr := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			serveErr <- nil
			return
		}
		serveErr <- err
	}()

	// Block until SIGINT / SIGTERM or an unexpected server error.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		log.Printf("shutdown signal received, stopping HTTP server")
	}

	// Give in-flight requests up to 10 s to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return <-serveErr
}

// NewStreamableHandler creates an MCP streamable HTTP handler.
// JSONResponse is set to false so that GET /mcp opens a proper SSE stream
// and ping/notifications are handled natively by the SDK.
func NewStreamableHandler(serverFunc func(r *http.Request) *mcp.Server) http.Handler {
	return mcp.NewStreamableHTTPHandler(serverFunc, nil)
}

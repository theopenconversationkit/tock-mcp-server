package tock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/theopenconversationkit/tock-mcp-server/config"
)

// Client wraps an HTTP client and the Tock connection configuration.
type Client struct {
	cfg  config.TockConfig
	http *http.Client
}

// NewClient creates a Client ready to query the given Tock configuration.
func NewClient(cfg config.TockConfig) *Client {
	return &Client{cfg: cfg, http: &http.Client{}}
}

// Ask sends question to the Tock web-connector and returns the structured response.
// The URL is built from the configured namespace, bot and connector identifiers.
// callHeaders provides per-call values for the extra headers declared in ExtraHeaders;
// they override the default values from the configuration.
// Only headers declared in ExtraHeaders are forwarded (unknown keys in callHeaders are ignored).
func (t *Client) Ask(ctx context.Context, question string, callHeaders map[string]string) (*Response, error) {
	// Build the endpoint: <base_url>/io/<namespace>/<bot>/<connector>
	url := fmt.Sprintf(
		"%s/io/%s/%s/%s",
		strings.TrimRight(t.cfg.BaseURL, "/"),
		t.cfg.Namespace,
		t.cfg.Bot,
		t.cfg.Connector,
	)

	body, err := json.Marshal(Query{Query: question, UserID: t.cfg.UserID})
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
	// Matching is case-insensitive to avoid subtle config/runtime header casing issues.
	callOverrides := make(map[string]string, len(callHeaders))
	for name, value := range callHeaders {
		callOverrides[textproto.CanonicalMIMEHeaderKey(name)] = value
	}

	for name, defaultVal := range t.cfg.ExtraHeaders {
		canonicalName := textproto.CanonicalMIMEHeaderKey(name)
		val := defaultVal
		if v, ok := callOverrides[canonicalName]; ok && v != "" {
			val = v
		}
		if val != "" {
			req.Header.Set(canonicalName, val)
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

	var tockResp Response
	if err := json.Unmarshal(raw, &tockResp); err != nil {
		return nil, fmt.Errorf("parse Tock response: %w", err)
	}

	return &tockResp, nil
}

package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// TockConfig holds the connection parameters for the Tock web-connector API.
type TockConfig struct {
	BaseURL      string            `mapstructure:"base_url"`      // Base URL of the Tock instance, without trailing slash.
	Namespace    string            `mapstructure:"namespace"`     // Tock namespace (usually the organisation name).
	Bot          string            `mapstructure:"bot"`           // Bot identifier within the namespace.
	Connector    string            `mapstructure:"connector"`     // Web-connector identifier exposed by the bot.
	UserID       string            `mapstructure:"user_id"`       // User ID sent to Tock with every request.
	ExtraHeaders map[string]string `mapstructure:"extra_headers"` // Optional HTTP headers forwarded to Tock (name → default value; empty string means no default).
}

// ServerConfig holds the HTTP server parameters.
type ServerConfig struct {
	Addr                     string        `mapstructure:"addr"`                       // Listen address, e.g. ":8083".
	ToolName                 string        `mapstructure:"tool_name"`                  // Name of the MCP tool exposed to AI clients. Defaults to "ask_tock" if empty.
	ToolDescription          string        `mapstructure:"tool_description"`           // Description of the MCP tool shown to AI clients. Falls back to a built-in default if empty.
	InputQuestionDescription string        `mapstructure:"input_question_description"` // Description of the "question" input parameter. Falls back to a built-in default if empty.
	ReadHeaderTimeout        time.Duration `mapstructure:"read_header_timeout"`        // Max time to read request headers. Default: 5s.
	ReadTimeout              time.Duration `mapstructure:"read_timeout"`               // Max time to read the full request. Default: 15s.
	WriteTimeout             time.Duration `mapstructure:"write_timeout"`              // Max time to write the response. Default: 30s.
	IdleTimeout              time.Duration `mapstructure:"idle_timeout"`               // Max keep-alive idle time. Default: 60s.
	ShutdownTimeout          time.Duration `mapstructure:"shutdown_timeout"`           // Graceful shutdown deadline. Default: 10s.
}

// OAuthConfig holds the OAuth 2.1 Resource Server parameters.
// When Enabled is true, the server validates incoming Bearer tokens on protected endpoints.
type OAuthConfig struct {
	Enabled        bool     `mapstructure:"enabled"`         // Enable OAuth 2.1 token validation.
	Issuer         string   `mapstructure:"issuer"`          // Expected token issuer (iss claim). Also used for OIDC Discovery if JwksURL is empty.
	JwksURL        string   `mapstructure:"jwks_url"`        // Explicit JWKS endpoint URL. If empty, derived from Issuer via OIDC Discovery.
	Audience       string   `mapstructure:"audience"`        // Expected audience (aud claim). Leave empty to skip audience check.
	RequiredScopes []string `mapstructure:"required_scopes"` // Scopes that must be present in the token's "scope" claim.
}

// Config is the top-level configuration structure loaded using Viper.
type Config struct {
	Tock   TockConfig   `mapstructure:"tock"`
	Server ServerConfig `mapstructure:"server"`
	OAuth  OAuthConfig  `mapstructure:"oauth"`
}

// Load reads and parses the configuration file at path using Viper.
// Supports YAML, JSON, TOML, HCL formats (auto-detected by extension).
// Environment variables with prefix TOCK_MCP_ override file values.
// Example env mappings:
//   - TOCK_MCP_TOCK_BASE_URL → tock.base_url
//   - TOCK_MCP_TOCK_NAMESPACE → tock.namespace
//   - TOCK_MCP_SERVER_ADDR → server.addr
//
// Returns error if the config file cannot be read.
func Load(path string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	// Environment variable support with prefix
	v.SetEnvPrefix("TOCK_MCP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// LoadWithDefaults reads configuration from path but doesn't fail if file is missing.
// Useful for containerized environments where configuration comes from env vars.
// All values have sensible defaults; file values override defaults; env vars override all.
func LoadWithDefaults(path string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	// Environment variable support with prefix
	v.SetEnvPrefix("TOCK_MCP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Try to read config file, but don't fail if it doesn't exist
	if path != "" {
		v.SetConfigFile(path)
		_ = v.ReadInConfig() // Ignore error if file not found
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// setDefaults configures sensible default values for all config keys.
func setDefaults(v *viper.Viper) {
	v.SetDefault("tock.base_url", "http://localhost:8080")
	v.SetDefault("tock.namespace", "")
	v.SetDefault("tock.bot", "")
	v.SetDefault("tock.connector", "web_connector")
	v.SetDefault("tock.user_id", "mcp-user")
	v.SetDefault("tock.extra_headers", map[string]interface{}{})
	v.SetDefault("server.addr", ":8083")
	v.SetDefault("server.tool_name", "ask_tock")
	v.SetDefault("server.tool_description", "Ask a question to the Tock documentary chatbot (RAG). Returns the text response and links to source documents.")
	v.SetDefault("server.input_question_description", "Question to ask the Tock chatbot (RAG). Include context, error messages, version, environment, or desired objective if available.")
	v.SetDefault("server.read_header_timeout", 5*time.Second)
	v.SetDefault("server.read_timeout", 15*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("server.idle_timeout", 60*time.Second)
	v.SetDefault("server.shutdown_timeout", 10*time.Second)
	v.SetDefault("oauth.enabled", false)
	v.SetDefault("oauth.issuer", "")
	v.SetDefault("oauth.jwks_url", "")
	v.SetDefault("oauth.audience", "")
	v.SetDefault("oauth.required_scopes", []string{})
}

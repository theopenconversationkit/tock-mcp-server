package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/theopenconversationkit/tock-mcp-server/config"
)

// OAuthMiddleware returns an HTTP middleware that validates OAuth 2.1 Bearer tokens.
// It verifies the token signature via JWKS, checks the issuer and audience claims,
// and ensures required scopes are present.
//
// If cfg.Enabled is false, the returned middleware is a no-op passthrough.
func OAuthMiddleware(cfg config.OAuthConfig) (func(http.Handler) http.Handler, error) {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler { return next }, nil
	}

	if cfg.Issuer == "" && cfg.JwksURL == "" {
		return nil, fmt.Errorf("oauth: either issuer or jwks_url must be set when oauth is enabled")
	}

	ctx := context.Background()

	var verifier *oidc.IDTokenVerifier

	if cfg.JwksURL != "" {
		// Use explicit JWKS endpoint — no OIDC Discovery needed.
		keySet := oidc.NewRemoteKeySet(ctx, cfg.JwksURL)
		verifier = oidc.NewVerifier(cfg.Issuer, keySet, &oidc.Config{
			ClientID:          cfg.Audience,
			SkipClientIDCheck: cfg.Audience == "",
			SkipIssuerCheck:   cfg.Issuer == "",
		})
	} else {
		// Derive JWKS from OIDC Discovery at {issuer}/.well-known/openid-configuration.
		provider, err := oidc.NewProvider(ctx, cfg.Issuer)
		if err != nil {
			return nil, fmt.Errorf("oauth: oidc discovery failed for issuer %q: %w", cfg.Issuer, err)
		}
		verifier = provider.Verifier(&oidc.Config{
			ClientID:          cfg.Audience,
			SkipClientIDCheck: cfg.Audience == "",
		})
	}

	requiredScopes := cfg.RequiredScopes

	log.Printf("OAuth 2.1 Resource Server enabled (issuer=%q, audience=%q, scopes=%v)",
		cfg.Issuer, cfg.Audience, requiredScopes)

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Bearer token from Authorization header.
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, `{"error":"invalid Authorization header, expected Bearer token"}`, http.StatusUnauthorized)
				return
			}
			rawToken := parts[1]

			// Verify the token (signature, expiry, issuer, audience).
			idToken, err := verifier.Verify(r.Context(), rawToken)
			if err != nil {
				log.Printf("oauth: token verification failed: %v", err)
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Check required scopes if configured.
			if len(requiredScopes) > 0 {
				if err := checkScopes(idToken, requiredScopes); err != nil {
					log.Printf("oauth: scope check failed: %v", err)
					http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}

	return middleware, nil
}

// checkScopes extracts the "scope" claim from the token and verifies that
// all requiredScopes are present. The "scope" claim can be either a
// space-delimited string (RFC 6749) or a JSON array of strings.
func checkScopes(token *oidc.IDToken, requiredScopes []string) error {
	var claims struct {
		Scope  string   `json:"scope"`
		Scopes []string `json:"scopes"`
	}
	if err := token.Claims(&claims); err != nil {
		return fmt.Errorf("unable to parse scope claims: %w", err)
	}

	// Build a set of granted scopes from both possible claim formats.
	granted := make(map[string]bool)
	if claims.Scope != "" {
		for _, s := range strings.Fields(claims.Scope) {
			granted[s] = true
		}
	}
	for _, s := range claims.Scopes {
		granted[s] = true
	}

	for _, required := range requiredScopes {
		if !granted[required] {
			return fmt.Errorf("missing required scope: %s", required)
		}
	}
	return nil
}

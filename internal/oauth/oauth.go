// Package oauth implements the browser-based OAuth login flow (PKCE, RFC 8252
// loopback redirect) and its background token refresh: authorize -> code ->
// JWT (WorkOS) -> PAT (monolith OIDC exchange, RFC 8693, same call the MCP
// server makes).
//
// Depends only on internal/auth and the standard library — must not import
// internal/config or cli/* (the cli layer bridges config.Resolved into a Config).
package oauth

import (
	"net/http"
	"strings"
	"time"
)

// DefaultResource is the audience/resource indicator pinned into the JWT —
// must be registered as a Resource Indicator in the WorkOS dashboard. Stable
// across environments (unlike client_id), so it stays a constant.
const DefaultResource = "https://app.bitrise.io"

// DefaultIssuer is the WorkOS AuthKit domain hosting /oauth2/authorize and
// /oauth2/token (shared with the MCP server).
const DefaultIssuer = "https://oauth.bitrise.io"

// DefaultOIDCTokenEndpoint is the monolith endpoint that exchanges a WorkOS
// JWT for a Bitrise PAT (RFC 8693).
const DefaultOIDCTokenEndpoint = "https://app.bitrise.io/oidc/token"

// DefaultClientID is the CIMD URL identifying this client — the URL itself
// is the id, not a secret.
const DefaultClientID = "https://app.bitrise.io/.well-known/oauth-client/cli"

// defaultPATLifetime is the fallback when a token response omits expires_in
// (the monolith's PAT_EXPIRY is 1h). refreshSkew re-mints slightly before the
// real expiry so a token never goes stale mid-request.
const (
	defaultTimeout     = 30 * time.Second
	defaultPATLifetime = time.Hour
	refreshSkew        = 60 * time.Second
)

// Config carries the external inputs for the OAuth flow.
type Config struct {
	// Issuer is the WorkOS AuthKit domain hosting /oauth2/authorize and
	// /oauth2/token; empty means OAuth isn't configured (Login errors clearly).
	Issuer            string
	OIDCTokenEndpoint string
	ClientID          string
	Resource          string
}

// NewConfig builds a Config with the package-default Resource. client_id is
// passed in rather than a constant since it's a per-environment CIMD URL the
// config layer resolves.
func NewConfig(issuer, oidcTokenEndpoint, clientID string) Config {
	return Config{
		Issuer:            issuer,
		OIDCTokenEndpoint: oidcTokenEndpoint,
		ClientID:          clientID,
		Resource:          DefaultResource,
	}
}

func (c Config) httpClient() *http.Client {
	return &http.Client{Timeout: defaultTimeout}
}

func (c Config) authorizeEndpoint() string {
	return strings.TrimRight(c.Issuer, "/") + "/oauth2/authorize"
}

func (c Config) tokenEndpoint() string {
	return strings.TrimRight(c.Issuer, "/") + "/oauth2/token"
}

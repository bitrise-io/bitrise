package cmdutil

import (
	"os"

	"github.com/bitrise-io/bitrise/v2/internal/oauth"
)

// Env vars overriding the OAuth/web defaults below — rarely changed, mostly
// for pointing a build at a non-prod environment. Exported since tests set
// them directly.
const (
	EnvOAuthIssuer       = "BITRISE_OAUTH_ISSUER"
	EnvOIDCTokenEndpoint = "BITRISE_OIDC_TOKEN_ENDPOINT"
	EnvOAuthClientID     = "BITRISE_OAUTH_CLIENT_ID"
	EnvWebBaseURL        = "BITRISE_WEB_BASE_URL"
)

const defaultWebBaseURL = "https://app.bitrise.io"

// OAuthConfig builds the oauth.Config from env-var overrides, falling back
// to production defaults.
func OAuthConfig() oauth.Config {
	return oauth.NewConfig(
		firstNonEmpty(os.Getenv(EnvOAuthIssuer), oauth.DefaultIssuer),
		firstNonEmpty(os.Getenv(EnvOIDCTokenEndpoint), oauth.DefaultOIDCTokenEndpoint),
		firstNonEmpty(os.Getenv(EnvOAuthClientID), oauth.DefaultClientID),
	)
}

// ResolveWebBaseURL returns the resolved web base URL, overridable via
// BITRISE_WEB_BASE_URL.
func ResolveWebBaseURL() string {
	return firstNonEmpty(os.Getenv(EnvWebBaseURL), defaultWebBaseURL)
}

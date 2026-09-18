package cmdutil

import (
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
	"github.com/spf13/cobra"
)

// EnvRDEAPIBaseURL overrides the RDE API base URL — rarely changed, mostly
// for pointing at a non-prod environment. Exported since tests set it
// directly.
const EnvRDEAPIBaseURL = "BITRISE_RDE_API_BASE_URL"

// ResolveRDEAPIBaseURL returns the resolved RDE API base URL:
// BITRISE_RDE_API_BASE_URL, then the rde_api_base_url set via `bitrise
// config set` (global config file only — never a per-dir .bitrise-cli.yml,
// see internal/config.Resolve), then the built-in default. Any trailing
// slash is trimmed here, once, so every caller can concatenate a path onto
// the result without producing a double slash.
func ResolveRDEAPIBaseURL(cmd *cobra.Command) string {
	if v := os.Getenv(EnvRDEAPIBaseURL); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	if v := config.FromContext(cmd.Context()).RDEAPIBaseURL; v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return config.DefaultRDEAPIBaseURL
}

// NewRDEClient builds an *rdeapi.Client using the token resolved by
// liveToken and the resolved RDE API base URL.
func NewRDEClient(cmd *cobra.Command) (*rdeapi.Client, error) {
	tok, err := liveToken(cmd)
	if err != nil {
		return nil, err
	}
	return rdeapi.New(ResolveRDEAPIBaseURL(cmd), tok)
}

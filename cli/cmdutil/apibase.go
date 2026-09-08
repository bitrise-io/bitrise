package cmdutil

import (
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/spf13/cobra"
)

// EnvAPIBaseURL overrides the API base URL — rarely changed, mostly for
// pointing at a non-prod environment. Exported since tests set it directly.
const EnvAPIBaseURL = "BITRISE_API_BASE_URL"

// ResolveAPIBaseURL returns the resolved API base URL: BITRISE_API_BASE_URL,
// then the api_base_url set via `bitrise config set` (global config file only
// — never a per-dir .bitrise-cli.yml, see internal/config.Resolve), then the
// built-in default — matching ResolveWebBaseURL/ResolveRDEAPIBaseURL. Any
// trailing slash is trimmed here, once, so every caller can concatenate a
// path onto the result without producing a double slash.
func ResolveAPIBaseURL(cmd *cobra.Command) string {
	if v := os.Getenv(EnvAPIBaseURL); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	if v := config.FromContext(cmd.Context()).APIBaseURL; v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return config.DefaultAPIBaseURL
}

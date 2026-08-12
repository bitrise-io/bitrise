package cmdutil

import (
	"os"

	"github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/spf13/cobra"
)

// EnvWebBaseURL overrides the web base URL — rarely changed, mostly for
// pointing a build at a non-prod environment. Exported since tests set it
// directly.
const EnvWebBaseURL = "BITRISE_WEB_BASE_URL"

// ResolveWebBaseURL returns the resolved web base URL: BITRISE_WEB_BASE_URL,
// then the web_base_url set via `bitrise config set` (global config file only
// — never a per-dir .bitrise-cli.yml, see internal/config.Resolve), then the
// built-in default.
func ResolveWebBaseURL(cmd *cobra.Command) string {
	if v := os.Getenv(EnvWebBaseURL); v != "" {
		return v
	}
	if ctx := cmd.Context(); ctx != nil {
		if v := config.FromContext(ctx).WebBaseURL; v != "" {
			return v
		}
	}
	return config.DefaultWebBaseURL
}

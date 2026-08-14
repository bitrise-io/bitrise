package cmdutil

import (
	"os"
	"strings"

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
// built-in default. Any trailing slash is trimmed here, once, so every
// caller can concatenate a path onto the result without producing a double
// slash — the built-in default carries none, but a user-set value might.
func ResolveWebBaseURL(cmd *cobra.Command) string {
	if v := os.Getenv(EnvWebBaseURL); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	if v := config.FromContext(cmd.Context()).WebBaseURL; v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return config.DefaultWebBaseURL
}

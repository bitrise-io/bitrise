package cmdutil

import (
	"os"

	"github.com/bitrise-io/bitrise/v2/internal/config"
)

// EnvWebBaseURL overrides the web base URL — rarely changed, mostly for
// pointing a build at a non-prod environment. Exported since tests set it
// directly.
const EnvWebBaseURL = "BITRISE_WEB_BASE_URL"

const defaultWebBaseURL = "https://app.bitrise.io"

// ResolveWebBaseURL returns the resolved web base URL, overridable via
// BITRISE_WEB_BASE_URL.
func ResolveWebBaseURL() string {
	return config.FirstNonEmptyString(os.Getenv(EnvWebBaseURL), defaultWebBaseURL)
}

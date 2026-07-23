package cmdutil

import "os"

// EnvWebBaseURL overrides the web base URL — rarely changed, mostly for
// pointing a build at a non-prod environment. Exported since tests set it
// directly.
const EnvWebBaseURL = "BITRISE_WEB_BASE_URL"

const defaultWebBaseURL = "https://app.bitrise.io"

// ResolveWebBaseURL returns the resolved web base URL, overridable via
// BITRISE_WEB_BASE_URL.
func ResolveWebBaseURL() string {
	return firstNonEmpty(os.Getenv(EnvWebBaseURL), defaultWebBaseURL)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

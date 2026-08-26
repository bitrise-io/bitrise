package config

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise/v2/internal/baseurl"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

const (
	KeyAPIBaseURL         = "api_base_url"
	KeyWebBaseURL         = "web_base_url"
	KeyAppID              = "app_id"
	KeyDefaultWorkspaceID = "default_workspace_id"
	KeyOutput             = "output"
	KeyTheme              = "theme"
)

// Keys is the subset of Config's fields exposed to `bitrise config
// get/set/unset` — SetupVersion/LastCLIUpdateCheck/LastPluginUpdateChecks are
// deliberately excluded: the CLI writes those itself, they aren't user
// settings.
var Keys = []string{KeyAPIBaseURL, KeyWebBaseURL, KeyAppID, KeyDefaultWorkspaceID, KeyOutput, KeyTheme}

// URLKeys is the subset of Keys whose value Set validates as an absolute
// https URL (see validateURL) rather than accepting it as a plain
// identifier. Exported so tests can assert that invariant over every such
// key without hardcoding the list.
var URLKeys = []string{KeyAPIBaseURL, KeyWebBaseURL}

// Get returns the stored value of a known key.
func (c *Config) Get(key string) (string, error) {
	switch key {
	case KeyAPIBaseURL:
		return c.APIBaseURL, nil
	case KeyWebBaseURL:
		return c.WebBaseURL, nil
	case KeyAppID:
		return c.AppID, nil
	case KeyDefaultWorkspaceID:
		return c.DefaultWorkspaceID, nil
	case KeyOutput:
		return c.Output, nil
	case KeyTheme:
		return c.Theme, nil
	default:
		return "", unknownKeyErr(key)
	}
}

// Set assigns value to key. If validation fails, c is left untouched —
// callers can rely on Set being all-or-nothing.
func (c *Config) Set(key, value string) error {
	next := *c
	switch key {
	case KeyAPIBaseURL:
		if value != "" {
			if err := validateURL(KeyAPIBaseURL, value); err != nil {
				return err
			}
		}
		next.APIBaseURL = value
	case KeyWebBaseURL:
		if value != "" {
			if err := validateURL(KeyWebBaseURL, value); err != nil {
				return err
			}
		}
		next.WebBaseURL = value
	case KeyAppID:
		// An app slug, not a URL — nothing to validate locally; a wrong value
		// surfaces as a 404 from the API.
		next.AppID = value
	case KeyDefaultWorkspaceID:
		// A workspace slug; same reasoning as KeyAppID.
		next.DefaultWorkspaceID = value
	case KeyOutput:
		if value != "" {
			if _, err := output.ParseFormat(value); err != nil {
				return err
			}
		}
		next.Output = value
	case KeyTheme:
		if value != "" {
			if _, err := style.ParseTheme(value); err != nil {
				return err
			}
		}
		next.Theme = value
	default:
		return unknownKeyErr(key)
	}
	*c = next
	return nil
}

// Unset clears the value of key (equivalent to Set with an empty string).
func (c *Config) Unset(key string) error {
	return c.Set(key, "")
}

func unknownKeyErr(key string) error {
	return fmt.Errorf("unknown config key %q (valid keys: %s)", key, strings.Join(Keys, ", "))
}

// validateURL rejects a value that the actual consumer
// (webclient.New/bitriseapi.New, both backed by internal/baseurl) would
// reject or, worse for api_base_url, silently send the bearer token over
// plaintext http.
func validateURL(key, value string) error {
	_, err := baseurl.Validate(key, value)
	return err
}

package config

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	KeyAPIBaseURL = "api_base_url"
	KeyWebBaseURL = "web_base_url"
)

// Keys is the subset of Config's fields exposed to `bitrise config
// get/set/unset` — SetupVersion/LastCLIUpdateCheck/LastPluginUpdateChecks are
// deliberately excluded: the CLI writes those itself, they aren't user
// settings.
var Keys = []string{KeyAPIBaseURL, KeyWebBaseURL}

// Get returns the stored value of a known key.
func (c *Config) Get(key string) (string, error) {
	switch key {
	case KeyAPIBaseURL:
		return c.APIBaseURL, nil
	case KeyWebBaseURL:
		return c.WebBaseURL, nil
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

func validateURL(key, value string) error {
	u, err := url.Parse(value)
	if err != nil || !u.IsAbs() || u.Host == "" {
		return fmt.Errorf("%s %q must be an absolute URL", key, value)
	}
	return nil
}

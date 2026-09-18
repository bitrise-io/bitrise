// Package baseurl validates URLs that will carry credentials (a bearer
// token, a login password) to whatever host they name. Shared by
// internal/webclient, internal/bitriseapi, and internal/config, all of
// which need the same rule: https, except against a local test server.
package baseurl

import (
	"fmt"
	"net"
	"net/url"
)

// Validate checks that raw is an absolute URL using https — or plain http
// against a loopback host, so tests can point at an httptest server. label
// identifies raw in the returned error (e.g. "base URL", or a config key
// name like "api_base_url").
func Validate(label, raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || !u.IsAbs() || u.Host == "" {
		return nil, fmt.Errorf("%s %q must be an absolute URL", label, raw)
	}
	if u.Scheme != "https" && !isLoopbackHost(u.Hostname()) {
		return nil, fmt.Errorf("%s %q must use https (got %q)", label, raw, u.Scheme)
	}
	return u, nil
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

package goutil

import (
	"fmt"
	"net/url"
	"strings"
)

// ParsePackageNameFromURL - returns a Go package name/id (e.g. github.com/bitrise-tools/gows)
// from a git clone URL (e.g. https://github.com/bitrise-tools/gows.git)
func ParsePackageNameFromURL(remoteURL string) (string, error) {
	origRemoteURL := remoteURL
	if strings.HasPrefix(remoteURL, "git@") {
		remoteURL = "ssh://" + remoteURL
	}

	u, err := url.Parse(remoteURL)
	if err != nil {
		return "", fmt.Errorf("Failed to parse remote URL (%s): %s", origRemoteURL, err)
	}

	packagePth := u.Path
	packagePth = strings.TrimSuffix(packagePth, ".git")

	// in SSH git urls like "ssh://git@github.com:bitrise-io/go-utils.git" Go parses "github.com:bitrise-io" as the host
	// fix it by splitting it and replacing ":" with "/"
	hostSplits := strings.Split(u.Host, ":")
	host := hostSplits[0]
	if len(hostSplits) > 1 {
		if len(hostSplits) > 2 {
			return "", fmt.Errorf("More than one ':' found in the Host part of the URL (%s)", origRemoteURL)
		}
		packagePth = "/" + hostSplits[1] + packagePth
	}

	if host == "" {
		return "", fmt.Errorf("No Host found in URL (%s)", origRemoteURL)
	}
	if packagePth == "" || packagePth == "/" {
		return "", fmt.Errorf("No Path found in URL (%s)", origRemoteURL)
	}

	return host + packagePth, nil
}

package utils

import (
	"fmt"
	"regexp"
	"strings"

	ver "github.com/hashicorp/go-version"
)

func IsUpdateAvailable(currentVersion, latestVersion string) (bool, error) {
	if latestVersion == "" {
		return false, nil
	}

	re := regexp.MustCompile(`\d+`)
	components := re.FindAllString(currentVersion, -1)
	normalized := strings.Join(components, ".")
	locked, err := ver.NewSemver(normalized)

	if err != nil {
		return false, fmt.Errorf("error processing version (%s): normalized version (%s) not in semver format: %s", currentVersion, normalized, err)
	}

	latest, err := ver.NewSemver(latestVersion)
	if err != nil {
		return false, fmt.Errorf("error processing latest version (%s): %s", latestVersion, err)
	}

	switch len(components) {
	case 1:
		return locked.Segments()[0] < latest.Segments()[0], nil
	case 2:
		return locked.Segments()[0] < latest.Segments()[0] || locked.Segments()[1] < latest.Segments()[1], nil
	case 3:
		return locked.LessThan(latest), nil
	default:
		return false, nil
	}
}

func RepoReleasesURL(repoURL string) string {
	if strings.Contains(repoURL, "github") || strings.Contains(repoURL, "gitlab") {
		return repoURL + "/releases"
	}

	return repoURL
}

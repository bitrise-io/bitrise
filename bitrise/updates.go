package bitrise

import (
	"regexp"
	"strings"

	log "github.com/bitrise-io/bitrise/advancedlog"
	stepmanModels "github.com/bitrise-io/stepman/models"
	ver "github.com/hashicorp/go-version"
)

func isUpdateAvailable(stepInfo stepmanModels.StepInfoModel) bool {
	if stepInfo.LatestVersion == "" {
		return false
	}

	re := regexp.MustCompile(`\d+`)
	components := re.FindAllString(stepInfo.Version, -1)
	normalized := strings.Join(components, ".")
	locked, err := ver.NewSemver(normalized)

	if err != nil {
		log.Warnf("Error processing version (%s): normalized version (%s) not in semver format: %s", stepInfo.Version, normalized, err)
		return false
	}

	latest, err := ver.NewSemver(stepInfo.LatestVersion)
	if err != nil {
		log.Warnf("Error processing latest version (%s): %s", stepInfo.LatestVersion, err)
		return false
	}

	switch len(components) {
	case 1:
		return locked.Segments()[0] < latest.Segments()[0]
	case 2:
		return locked.Segments()[0] < latest.Segments()[0] || locked.Segments()[1] < latest.Segments()[1]
	case 3:
		return locked.LessThan(latest)
	default:
		return false
	}
}

func repoReleasesURL(repoURL string) string {
	if strings.Contains(repoURL, "github") || strings.Contains(repoURL, "gitlab") {
		return repoURL + "/releases"
	}

	return repoURL
}

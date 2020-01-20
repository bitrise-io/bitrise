package bitrise

import (
	"regexp"
	"strings"

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
	locked, _ := ver.NewSemver(normalized)
	latest, _ := ver.NewSemver(stepInfo.LatestVersion)

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

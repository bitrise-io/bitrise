package bitrise

import (
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/versions"
	stepmanModels "github.com/bitrise-io/stepman/models"
	ver "github.com/hashicorp/go-version"
)

func isUpdateAvailable(stepInfo stepmanModels.StepInfoModel) bool {
	if stepInfo.LatestVersion == "" {
		return false
	}

	if stepInfo.Version != stepInfo.EvaluatedVersion {
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

		}
	}

	res, err := versions.CompareVersions(stepInfo.Version, stepInfo.LatestVersion)
	if err != nil {
		log.Errorf("Failed to compare versions, err: %s", err)
	}

	return (res == 1)
}

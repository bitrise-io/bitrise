//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"github.com/bitrise-io/bitrise/v2/models"
)

// defaultTestToolConfig returns a default ToolConfigModel for testing
func defaultTestToolConfig() models.ToolConfigModel {
	return models.ToolConfigModel{
		Provider:                     "mise",
		ExperimentalFastInstall:      false,
		ExperimentalFastInstallForce: false,
	}
}

// fastInstallToolConfig returns a ToolConfigModel with fast install enabled
func fastInstallToolConfig() models.ToolConfigModel {
	return models.ToolConfigModel{
		Provider:                     "mise",
		ExperimentalFastInstall:      true,
		ExperimentalFastInstallForce: true,
	}
}

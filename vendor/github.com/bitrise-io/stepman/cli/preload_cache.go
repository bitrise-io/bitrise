package cli

import (
	"fmt"

	"github.com/bitrise-io/stepman/stepman"
)

const (
	bitriseStepLibURL = "https://github.com/bitrise-io/bitrise-steplib.git"
	bitriseMaintainer = "bitrise"
)

// PreloadBitriseSteps preloads the cache with Bitrise owned steps
func PreloadBitriseSteps(log stepman.Logger) error {
	// Check if setup was done for collection
	if exist, err := stepman.RootExistForLibrary(bitriseStepLibURL); err != nil {
		return err
	} else if !exist {
		if err := stepman.SetupLibrary(bitriseStepLibURL, log); err != nil {
			return fmt.Errorf("Failed to setup steplib: %w", err)
		}
	}

	stepLib, err := stepman.ReadStepSpec(bitriseStepLibURL)
	if err != nil {
		return err
	}

	for stepID, step := range stepLib.Steps {
		if step.Info.Maintainer != bitriseMaintainer {
			log.Warnf("Skipping step %s as it is not maintained by Bitrise", step)
		}

		latestVersion, found := step.LatestVersion()
		if !found {
			log.Warnf("Failed to find latest version for step %s", stepID)
		}

		_, _, err := downloadStep(stepLib, bitriseStepLibURL, stepID, step.LatestVersionNumber, latestVersion, log)
		if err != nil {
			log.Warnf("Failed to download step %s: %w", stepID, err)
		}
	}

	return nil
}

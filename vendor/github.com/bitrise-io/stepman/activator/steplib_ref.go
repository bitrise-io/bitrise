package activator

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/stepman/activator/steplib"
	"github.com/bitrise-io/stepman/cli"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/stepman"
)

func ActivateSteplibRefStep(
	log stepman.Logger,
	id stepid.CanonicalID,
	activatedStepDir string,
	workDir string,
	didStepLibUpdateInWorkflow bool,
	isOfflineMode bool,
	stepInfoPtr *models.StepInfoModel,
) (ActivatedStep, error) {
	stepYMLPath := filepath.Join(workDir, "current_step.yml")
	activationResult := ActivatedStep{
		StepYMLPath:      stepYMLPath,
		DidStepLibUpdate: false,
	}

	stepInfo, didUpdate, err := prepareStepLibForActivation(log, id, didStepLibUpdateInWorkflow, isOfflineMode)
	activationResult.DidStepLibUpdate = didUpdate
	if err != nil {
		return activationResult, err
	}

	err = steplib.ActivateStep(id.SteplibSource, id.IDorURI, stepInfo.Version, activatedStepDir, stepYMLPath, log, isOfflineMode)
	if err != nil {
		return activationResult, err
	}

	// TODO: this is sketchy, we should clean this up, but this pointer originates in the CLI codebase
	stepInfoPtr.ID = stepInfo.ID
	if stepInfoPtr.Step.Title == nil || *stepInfoPtr.Step.Title == "" {
		stepInfoPtr.Step.Title = pointers.NewStringPtr(stepInfo.ID)
	}
	stepInfoPtr.Version = stepInfo.Version
	stepInfoPtr.LatestVersion = stepInfo.LatestVersion
	stepInfoPtr.OriginalVersion = stepInfo.OriginalVersion
	stepInfoPtr.GroupInfo = stepInfo.GroupInfo

	return activationResult, nil
}

func prepareStepLibForActivation(
	log stepman.Logger,
	id stepid.CanonicalID,
	didStepLibUpdateInWorkflow bool,
	isOfflineMode bool,
) (stepInfo models.StepInfoModel, didUpdate bool, err error) {
	err = cli.Setup(id.SteplibSource, "", log)
	if err != nil {
		return models.StepInfoModel{}, false, fmt.Errorf("setup %s: %s", id.SteplibSource, err)
	}

	versionConstraint, err := models.ParseRequiredVersion(id.Version)
	if err != nil {
		return models.StepInfoModel{}, false, err
	}
	if versionConstraint.VersionLockType == models.InvalidVersionConstraint {
		return models.StepInfoModel{}, false, fmt.Errorf("version constraint is invalid: %s %s", id.IDorURI, id.Version)
	}

	if shouldUpdateStepLibForStep(versionConstraint, isOfflineMode, didStepLibUpdateInWorkflow) {
		log.Infof("Step uses latest version, updating StepLib...")
		_, err = stepman.UpdateLibrary(id.SteplibSource, log)
		if err != nil {
			log.Warnf("Step version constraint is latest or version locked, but failed to update StepLib, err: %s", err)
		} else {
			didUpdate = true
		}
	}

	stepInfo, err = cli.QueryStepInfoFromLibrary(id.SteplibSource, id.IDorURI, id.Version, log)
	if err != nil {
		if !canUpdateStepLib(isOfflineMode, didStepLibUpdateInWorkflow) {
			return stepInfo, didUpdate, err
		}

		log.Infof("Failed to query step info from library, trying to update StepLib...")
		_, err = stepman.UpdateLibrary(id.SteplibSource, log)
		if err != nil {
			return stepInfo, didUpdate, err
		} else {
			didUpdate = true
		}

		stepInfo, err = cli.QueryStepInfoFromLibrary(id.SteplibSource, id.IDorURI, id.Version, log)
		if err != nil {
			return stepInfo, didUpdate, err
		}
	}

	if stepInfo.Step.Title == nil || *stepInfo.Step.Title == "" {
		stepInfo.Step.Title = pointers.NewStringPtr(stepInfo.ID)
	}
	stepInfo.OriginalVersion = id.Version

	return stepInfo, didUpdate, nil
}

func shouldUpdateStepLibForStep(constraint models.VersionConstraint, isOfflineMode bool, didStepLibUpdateInWorkflow bool) bool {
	if !canUpdateStepLib(isOfflineMode, didStepLibUpdateInWorkflow) {
		return false
	}

	return (constraint.VersionLockType == models.Latest) ||
		(constraint.VersionLockType == models.MinorLocked) ||
		(constraint.VersionLockType == models.MajorLocked)
}

func canUpdateStepLib(isOfflineMode bool, didStepLibUpdateInWorkflow bool) bool {
	if isOfflineMode {
		return false
	}

	if didStepLibUpdateInWorkflow {
		return false
	}

	return true
}

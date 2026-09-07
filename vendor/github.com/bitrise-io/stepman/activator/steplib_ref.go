package activator

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/stepman/activator/steplib"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/steplibrary"
	"github.com/bitrise-io/stepman/stepman"
)

// ActivateSteplibRefStep activates a step referenced through a StepLib.
func (a *Activator) ActivateSteplibRefStep(
	id stepid.CanonicalID,
	activatedStepDir string,
	workDir string,
	didStepLibUpdateInWorkflow bool,
) (ActivatedStep, error) {
	log := a.log
	stepYMLPath := filepath.Join(workDir, "current_step.yml")
	//nolint:exhaustruct // missing fields are added down below based on activation result
	activationResult := ActivatedStep{
		StepYMLPath:      stepYMLPath,
		DidStepLibUpdate: false,
	}

	var libraryAPI *steplibrary.Client
	if a.useSteplibAPIFor(id.SteplibSource) {
		libraryAPI = a.library
	}

	// The inventory source is set here, on the same branch that dispatches, and before
	// any return: the caller keeps the partial result on error, so a failed activation
	// is still attributable to the inventory that served it.
	if libraryAPI == nil {
		activationResult.ActivationInventorySource = ActivationInventorySourceSteplib

		// Old stepman preparation codepath
		stepInfo, didUpdate, err := prepareStepLibForActivation(log, id, didStepLibUpdateInWorkflow)
		activationResult.StepInfo = stepInfo
		activationResult.DidStepLibUpdate = didUpdate
		if err != nil {
			return activationResult, err
		}
	} else {
		activationResult.ActivationInventorySource = ActivationInventorySourceSteplibAPI
	}

	// ActivateStep dispatches to the v2 or legacy codepath.
	activateOpts := steplib.Options{
		UsePrecompiled: a.opts.UsePrecompiled,
		StorageURLs:    a.opts.PrecompiledStorageURLs,
	}
	resolvedStep, err := steplib.ActivateStep(id, activatedStepDir, stepYMLPath, log, activateOpts, libraryAPI, a.fetcher)
	activationResult.StepInfo = resolvedStep.StepInfo
	activationResult.ExecutablePath = resolvedStep.ExecPath
	if resolvedStep.ExecPath != "" {
		activationResult.ActivationType = ActivationTypeSteplibExecutable
	} else {
		activationResult.ActivationType = ActivationTypeSteplibSource
	}
	if err != nil {
		return activationResult, err
	}

	return activationResult, nil
}

func prepareStepLibForActivation(
	log stepman.Logger,
	id stepid.CanonicalID,
	didStepLibUpdateInWorkflow bool,
) (stepInfo models.StepInfoModel, didUpdate bool, err error) {
	err = stepman.SetupLibrary(id.SteplibSource, log)
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

	if shouldUpdateStepLibForStep(versionConstraint, didStepLibUpdateInWorkflow) {
		log.Infof("Step uses latest version, updating StepLib...")
		_, err = stepman.UpdateLibrary(id.SteplibSource, log)
		if err != nil {
			log.Warnf("Step version constraint is latest or version locked, but failed to update StepLib, err: %s", err)
		} else {
			didUpdate = true
		}
	}

	stepInfo, err = stepman.QueryStepInfoFromLibrary(id.SteplibSource, id.IDorURI, id.Version, log)
	if err != nil {
		if didStepLibUpdateInWorkflow {
			return stepInfo, didUpdate, err
		}

		log.Infof("Step not found in local StepLib cache, trying to update StepLib...")
		_, err = stepman.UpdateLibrary(id.SteplibSource, log)
		if err != nil {
			return stepInfo, didUpdate, err
		} else {
			didUpdate = true
		}

		stepInfo, err = stepman.QueryStepInfoFromLibrary(id.SteplibSource, id.IDorURI, id.Version, log)
		if err != nil {
			return stepInfo, didUpdate, err
		}
	}

	return stepInfo, didUpdate, nil
}

func shouldUpdateStepLibForStep(constraint models.VersionConstraint, didStepLibUpdateInWorkflow bool) bool {
	if didStepLibUpdateInWorkflow {
		return false
	}

	return (constraint.VersionLockType == models.Latest) ||
		(constraint.VersionLockType == models.MinorLocked) ||
		(constraint.VersionLockType == models.MajorLocked)
}

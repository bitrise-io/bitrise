package activator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/stepman/activator/result"
	"github.com/bitrise-io/stepman/activator/steplib"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/steplibrary"
	"github.com/bitrise-io/stepman/stepman"
)

const useSteplibV2 = "BITRISE_EXPERIMENT_STEPLIB_V2"

const (
	// bitriseV1SteplibURL is the canonical git URL of the official Bitrise
	// steplib, as users reference it in bitrise.yml.
	bitriseV1SteplibURL = "https://github.com/bitrise-io/bitrise-steplib.git"
	// bitriseV2SteplibURL is the compiled-in base URL of the official V2
	// inventory. The official V1 URL is rewritten to this when V2 is enabled.
	bitriseV2SteplibURL = "https://steplibrary.bitrise.io"
)

// resolveSteplibV2URL decides whether to activate via the V2 inventory when the
// BITRISE_EXPERIMENT_STEPLIB_V2 flag is enabled, returning the V2 base URL to use:
//
//   - the official V1 steplib URL is rewritten to the compiled-in V2 URL;
//   - any non-".git" URL is treated as a V2 inventory base URL directly;
//   - any other ".git" URL (an alt-steplib) keeps the legacy V1 path.
func resolveSteplibV2URL(steplibURI string) (v2URL string, useV2 bool) {
	if os.Getenv(useSteplibV2) != "true" && os.Getenv(useSteplibV2) != "1" {
		return "", false
	}
	switch {
	case steplibURI == bitriseV1SteplibURL:
		return bitriseV2SteplibURL, true
	case strings.HasSuffix(steplibURI, ".git"):
		return "", false
	default:
		return steplibURI, true
	}
}

func ActivateSteplibRefStep(
	log stepman.Logger,
	id stepid.CanonicalID,
	activatedStepDir string,
	workDir string,
	didStepLibUpdateInWorkflow bool,
	isOfflineMode bool,
	stepInfoPtr *models.StepInfoModel,
) (result.ActivatedStep, error) {
	stepYMLPath := filepath.Join(workDir, "current_step.yml")
	//nolint:exhaustruct // missing fields are added down below based on activation result
	activationResult := result.ActivatedStep{
		StepYMLPath:      stepYMLPath,
		DidStepLibUpdate: false,
	}

	if inventoryURL, useV2 := resolveSteplibV2URL(id.SteplibSource); useV2 {
		// id.SteplibSource is the identity (keys the V1 cache / source fallback);
		// inventoryURL is where the V2 inventory JSON is fetched from.
		v2 := steplibrary.New(log, id.SteplibSource, inventoryURL, isOfflineMode, fileutil.NewFileManager())
		// TODO: thread context.Context through ActivateSteplibRefStep when callers can supply one.
		activated, err := v2.Activate(context.Background(), id.IDorURI, id.Version, steplibrary.ActivateOutputPaths{
			YMLPath:  stepYMLPath,
			CodePath: activatedStepDir,
		})
		if err != nil {
			return activationResult, err
		}

		populateStepInfo(stepInfoPtr, activated.StepInfo)
		return activated, nil
	}

	stepInfo, didUpdate, err := prepareStepLibForActivation(log, id, didStepLibUpdateInWorkflow, isOfflineMode)
	activationResult.DidStepLibUpdate = didUpdate
	if err != nil {
		return activationResult, err
	}

	execPath, err := steplib.ActivateStep(id.SteplibSource, id.IDorURI, stepInfo.Version, activatedStepDir, stepYMLPath, log, isOfflineMode)
	activationResult.ExecutablePath = execPath
	if execPath != "" {
		activationResult.ActivationType = result.ActivationTypeSteplibExecutable
	} else {
		activationResult.ActivationType = result.ActivationTypeSteplibSource
	}
	if err != nil {
		return activationResult, err
	}

	// TODO: this is sketchy, we should clean this up, but this pointer originates in the CLI codebase
	populateStepInfo(stepInfoPtr, stepInfo)

	return activationResult, nil
}

// populateStepInfo copies the resolved fields from source onto target.
// The Step.Title on target is preserved if it's a non-empty string; otherwise
// it falls back to source.ID so callers always see a usable title.
func populateStepInfo(target *models.StepInfoModel, source models.StepInfoModel) {
	target.ID = source.ID
	if target.Step.Title == nil || *target.Step.Title == "" {
		target.Step.Title = pointers.NewStringPtr(source.ID)
	}
	target.Version = source.Version
	target.LatestVersion = source.LatestVersion
	target.OriginalVersion = source.OriginalVersion
	target.GroupInfo = source.GroupInfo
}

func prepareStepLibForActivation(
	log stepman.Logger,
	id stepid.CanonicalID,
	didStepLibUpdateInWorkflow bool,
	isOfflineMode bool,
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

	if shouldUpdateStepLibForStep(versionConstraint, isOfflineMode, didStepLibUpdateInWorkflow) {
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
		if !canUpdateStepLib(isOfflineMode, didStepLibUpdateInWorkflow) {
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

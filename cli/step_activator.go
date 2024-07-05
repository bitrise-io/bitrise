package cli

import (
	"fmt"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/stepman/activator"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
)

type stepActivator struct {
}

func newStepActivator() stepActivator {
	return stepActivator{}
}

func (a stepActivator) activateStep(
	stepIDData stepid.CanonicalID,
	isStepLibUpdated bool,
	stepDir string, // $TMPDIR/bitrise/step_src
	workDir string, // $TMPDIR/bitrise
	stepInfoPtr *stepmanModels.StepInfoModel,
	isSteplibOfflineMode bool,
) (stepYMLPth string, origStepYMLPth string, didStepLibUpdate bool, err error) {
	stepmanLogger := log.NewLogger(log.GetGlobalLoggerOpts())

	if stepIDData.SteplibSource == "path" {
		log.Debugf("[BITRISE_CLI] - Local step found: (path:%s)", stepIDData.IDorURI)

		activatedStep, err := activator.ActivatePathRefStep(
			stepmanLogger,
			stepIDData,
			stepDir,
			workDir,
		)
		if err != nil {
			return "", "", false, fmt.Errorf("activate local step: %w", err)
		}

		stepYMLPth = activatedStep.StepYMLPath
		origStepYMLPth = activatedStep.OrigStepYMLPath
	} else if stepIDData.SteplibSource == "git" {
		log.Debugf("[BITRISE_CLI] - Remote step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)

		activatedStep, err := activator.ActivateGitRefStep(
			stepmanLogger,
			stepIDData,
			stepDir,
			workDir,
		)
		if err != nil {
			return "", "", false, fmt.Errorf("activate git step reference: %w", err)
		}

		stepYMLPth = activatedStep.StepYMLPath
		origStepYMLPth = activatedStep.OrigStepYMLPath
	} else if stepIDData.SteplibSource != "" {
		activatedStep, err := activator.ActivateSteplibRefStep(
			stepmanLogger,
			stepIDData,
			stepDir,
			workDir,
			isStepLibUpdated,
			isSteplibOfflineMode,
			stepInfoPtr,
		)
		didStepLibUpdate = activatedStep.DidStepLibUpdate
		if err != nil {
			return "", "", didStepLibUpdate, fmt.Errorf("activate steplib step: %w", err)
		}

		stepYMLPth = activatedStep.StepYMLPath
		origStepYMLPth = activatedStep.OrigStepYMLPath
	} else {
		return "", "", didStepLibUpdate, fmt.Errorf("invalid stepIDData: no SteplibSource or LocalPath defined (%v)", stepIDData)
	}

	return stepYMLPth, origStepYMLPth, didStepLibUpdate, nil
}

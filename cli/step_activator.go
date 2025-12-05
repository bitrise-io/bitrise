package cli

import (
	"fmt"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/stepman/activator"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
)

type stepActivator struct {
}

func newStepActivator() stepActivator {
	return stepActivator{}
}

// Note: even when err != nil, the ActivatedStep struct will be returned with a valid DidStepLibUpdate value
func (a stepActivator) activateStep(
	stepIDData stepid.CanonicalID,
	isStepLibUpdated bool,
	stepDir string, // $TMPDIR/bitrise/step_src
	workDir string, // $TMPDIR/bitrise
	stepInfoPtr *stepmanModels.StepInfoModel,
	isSteplibOfflineMode bool,
) (activator.ActivatedStep, error) {
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
			return activator.ActivatedStep{}, fmt.Errorf("activate local step: %w", err)
		}
		return activatedStep, nil
	} else if stepIDData.SteplibSource == "git" {
		log.Debugf("[BITRISE_CLI] - Remote step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)

		activatedStep, err := activator.ActivateGitRefStep(
			stepmanLogger,
			stepIDData,
			stepDir,
			workDir,
		)
		if err != nil {
			return activator.ActivatedStep{}, fmt.Errorf("activate git step reference: %w", err)
		}
		return activatedStep, nil
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
		if err != nil {
			// Note: we return the partial result on purpose because DidStepLibUpdate is important 
			// even in case of an error
			return activatedStep, fmt.Errorf("activate steplib step: %w", err)
		}
		return activatedStep, nil
	} else {
		return activator.ActivatedStep{}, fmt.Errorf("invalid stepIDData: no SteplibSource or LocalPath defined (%v)", stepIDData)
	}
}

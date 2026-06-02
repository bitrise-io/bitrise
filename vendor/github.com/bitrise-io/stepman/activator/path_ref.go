package activator

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/activator/result"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/stepman"
)

func ActivatePathRefStep(
	log stepman.Logger,
	id stepid.CanonicalID,
	activatedStepDir string,
	workDir string,
) (result.ActivatedStep, error) {
	log.Debugf("Local step found: (path:%s)", id.IDorURI)
	// id.IDorURI is a path to the step dir in this case
	stepAbsLocalPth, err := pathutil.AbsPath(id.IDorURI)
	if err != nil {
		return result.ActivatedStep{}, err
	}

	exist, err := pathutil.IsDirExists(stepAbsLocalPth)
	if err != nil {
		return result.ActivatedStep{}, fmt.Errorf("check if a directory exists at %s: %w", stepAbsLocalPth, err)
	} else if !exist {
		return result.ActivatedStep{}, fmt.Errorf("the provided directory doesn't exist: %s", stepAbsLocalPth)
	}

	log.Debugf("stepAbsLocalPth:", stepAbsLocalPth, "|stepDir:", activatedStepDir)

	origStepYMLPth := filepath.Join(stepAbsLocalPth, "step.yml")
	exist, err = pathutil.IsPathExists(origStepYMLPth)
	if err != nil {
		return result.ActivatedStep{}, fmt.Errorf("check if step.yml exists at %s: %w", origStepYMLPth, err)
	} else if !exist {
		return result.ActivatedStep{}, fmt.Errorf("step.yml doesn't exist at %s", origStepYMLPth)
	}

	activatedStepYMLPath := filepath.Join(workDir, "current_step.yml")
	if err := command.CopyFile(origStepYMLPth, activatedStepYMLPath); err != nil {
		return result.ActivatedStep{}, err
	}

	if err := command.CopyDir(stepAbsLocalPth, activatedStepDir, true); err != nil {
		return result.ActivatedStep{}, err
	}

	//nolint:exhaustruct // StepInfo isn't populated by path-ref activation
	return result.ActivatedStep{
		StepYMLPath:      activatedStepYMLPath,
		DidStepLibUpdate: false,
		ActivationType:   result.ActivationTypePathRef,
		ExecutablePath:   "",
	}, nil
}

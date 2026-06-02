package steplibrary

import (
	"context"
	"fmt"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/stepman/activator/result"
	"github.com/bitrise-io/stepman/models"
	"gopkg.in/yaml.v2"
)

func (s *Steplib) Activate(ctx context.Context, stepID, version string, outputPaths ActivateOutputPaths) (result.ActivatedStep, error) {
	stepInfo, resolved, err := s.getStepVersionInfo(ctx, stepID, version)

	var stepModel models.StepModel
	var execPath string
	if err == nil {
		stepModel, err = s.api.GetStepModel(ctx, resolved)
	}

	// Prefer the precompiled binary for the current platform when the step
	// publishes one; transparently fall back to source on any failure so an
	// individual broken executable can't block activation.
	if err == nil {
		if executable, ok := resolveExecutable(stepModel); ok {
			path, perr := s.downloadPrecompiled(ctx, stepID, executable, outputPaths.CodePath)
			if perr == nil {
				execPath = path
			} else {
				s.log.Warnf("Failed to download precompiled binary for %s, falling back to source: %s", currentPlatform(), perr)
			}
		}
	}

	if err == nil && execPath == "" {
		var srcDir string
		srcDir, err = s.fetchSourceDirFn(ctx, resolved)
		if err == nil {
			if cerr := s.fileManager.CopyDir(srcDir, outputPaths.CodePath, &fileutil.CopyOptions{Overwrite: true}); cerr != nil {
				err = fmt.Errorf("copy step source %s to %s: %w", srcDir, outputPaths.CodePath, cerr)
			}
		}
	}
	var stepYML []byte
	if err == nil {
		stepYML, err = yaml.Marshal(stepModel)
		if err != nil {
			err = fmt.Errorf("marshal step model to YAML: %w", err)
		}
	}
	if err == nil {
		err = s.fileManager.WriteBytes(outputPaths.YMLPath, stepYML)
	}
	if err != nil {
		return result.ActivatedStep{}, err
	}

	activationType := result.ActivationTypeSteplibSource
	if execPath != "" {
		activationType = result.ActivationTypeSteplibExecutable
	}
	return result.ActivatedStep{
		StepInfo:         stepInfo,
		StepYMLPath:      outputPaths.YMLPath,
		ExecutablePath:   execPath,
		ActivationType:   activationType,
		DidStepLibUpdate: false, // deprecated
	}, nil
}

package toolkits

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/utils"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// SwiftToolkit ...
type SwiftToolkit struct {
}

// Bootstrap ...
func (toolkit SwiftToolkit) Bootstrap() error {
	return nil
}

// Install ...
func (toolkit SwiftToolkit) Install() error {
	return nil
}

// ToolkitName ...
func (toolkit SwiftToolkit) ToolkitName() string {
	return "swift"
}

// Check ...
func (toolkit SwiftToolkit) Check() (bool, ToolkitCheckResult, error) {
	return false, ToolkitCheckResult{}, nil
}

// IsToolAvailableInPATH ...
func (toolkit SwiftToolkit) IsToolAvailableInPATH() bool {
	binPath, err := utils.CheckProgramInstalledPath("swift")
	if err != nil {
		return false
	}
	return len(binPath) > 0
}

func (toolkit SwiftToolkit) CompileStepExecutable(activatedStep stepmanModels.ActivatedStep, packageName string, targetExecutablePath string) (stepmanModels.ActivatedStep, error) {
	return activatedStep, nil
}

// PrepareForStepRun ...
func (toolkit SwiftToolkit) PrepareForStepRun(step stepmanModels.StepModel, _ models.StepIDData, activatedStep stepmanModels.ActivatedStep) (stepmanModels.ActivatedStep, error) {
	if activatedStep.Type != stepmanModels.ActivatedStepTypeSourceDir || activatedStep.SourceAbsDirPath == "" {
		return activatedStep, fmt.Errorf("invalid activated Swift step, missing source dir path")
	}

	binaryLocation := step.Toolkit.Swift.BinaryLocation
	if binaryLocation == "" {
		return activatedStep, nil
	}

	resp, err := http.Get(binaryLocation)
	if err != nil {
		return activatedStep, err
	}

	executablePath := filepath.Join(activatedStep.SourceAbsDirPath, step.Toolkit.Swift.ExecutableName)
	out, err := os.Create(executablePath)
	if err != nil {
		return activatedStep, err
	}

	_, err = io.Copy(out, resp.Body)
	err = os.Chmod(executablePath, 0755)

	err = resp.Body.Close()
	err = out.Close()
	return activatedStep, err
}

// StepRunCommandArguments ...
func (toolkit SwiftToolkit) StepRunCommandArguments(step stepmanModels.StepModel, sIDData models.StepIDData, activatedStep stepmanModels.ActivatedStep) ([]string, error) {
	if activatedStep.Type != stepmanModels.ActivatedStepTypeSourceDir || activatedStep.SourceAbsDirPath == "" {
		return []string{}, fmt.Errorf("invalid activated Swift step, missing source dir path")
	}

	binaryLocation := step.Toolkit.Swift.BinaryLocation
	if binaryLocation == "" {
		return []string{"swift", "run", "--package-path", activatedStep.SourceAbsDirPath, "-c", "release"}, nil
	}
	executablePath := filepath.Join(activatedStep.SourceAbsDirPath, step.Toolkit.Swift.ExecutableName)
	return []string{executablePath}, nil
}

package toolkits

import (
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/utils"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

// PrepareForStepRun ...
func (toolkit SwiftToolkit) PrepareForStepRun(step stepmanModels.StepModel, sIDData models.StepIDData, stepAbsDirPath string) error {
	binaryLocation := step.Toolkit.Swift.BinaryLocation
	if binaryLocation == "" {
		return nil
	}

	resp, err := http.Get(binaryLocation)
	if err != nil {
		return err
	}

	executablePath := filepath.Join(stepAbsDirPath, step.Toolkit.Swift.ExecutableName)
	out, err := os.Create(executablePath)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, resp.Body)
	err = os.Chmod(executablePath, 0777)

	err = resp.Body.Close()
	err = out.Close()
	return err
}

// StepRunCommandArguments ...
func (toolkit SwiftToolkit) StepRunCommandArguments(step stepmanModels.StepModel, sIDData models.StepIDData, stepAbsDirPath string) ([]string, error) {
	binaryLocation := step.Toolkit.Swift.BinaryLocation
	if binaryLocation == "" {
		return []string{"swift", "run", "--package-path", stepAbsDirPath, "-c", "release"}, nil
	}
	executablePath := filepath.Join(stepAbsDirPath, step.Toolkit.Swift.ExecutableName)
	return []string{executablePath}, nil
}

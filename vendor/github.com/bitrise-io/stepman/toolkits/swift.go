package toolkits

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bitrise-io/stepman/models"
)

type SwiftToolkit struct {
}

func (toolkit SwiftToolkit) Bootstrap() error {
	return nil
}

func (toolkit SwiftToolkit) Install() error {
	return nil
}

func (toolkit SwiftToolkit) ToolkitName() string {
	return "swift"
}

func (toolkit SwiftToolkit) Check() (bool, ToolkitCheckResult, error) {
	return false, ToolkitCheckResult{}, nil
}

func (toolkit SwiftToolkit) IsToolAvailableInPATH() bool {
	binPath, err := exec.LookPath("swift")
	if err != nil {
		return false
	}
	return len(binPath) > 0
}

func (toolkit SwiftToolkit) PrepareForStepRun(step models.StepModel, _ models.StepIDData, stepAbsDirPath string) error {
	binaryLocation := step.Toolkit.Swift.BinaryLocation
	if binaryLocation == "" {
		return nil
	}

	executablePath := filepath.Join(stepAbsDirPath, step.Toolkit.Swift.ExecutableName)
	
	err := downloadFile(binaryLocation, executablePath)
	if err != nil {
		return fmt.Errorf("download precompiled step binary: %s", err)
	}

	err = os.Chmod(executablePath, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (toolkit SwiftToolkit) StepRunCommandArguments(step models.StepModel, sIDData models.StepIDData, stepAbsDirPath string) ([]string, error) {
	binaryLocation := step.Toolkit.Swift.BinaryLocation
	if binaryLocation == "" {
		return []string{"swift", "run", "--package-path", stepAbsDirPath, "-c", "release"}, nil
	}
	executablePath := filepath.Join(stepAbsDirPath, step.Toolkit.Swift.ExecutableName)
	return []string{executablePath}, nil
}

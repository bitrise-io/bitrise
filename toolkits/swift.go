package toolkits

import (
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

func (toolkit SwiftToolkit) IsToolAvailableInPATH() bool {
	binPath, err := utils.CheckProgramInstalledPath("swift")
	if err != nil {
		return false
	}
	return len(binPath) > 0
}

func (toolkit SwiftToolkit) PrepareForStepRun(step stepmanModels.StepModel, sIDData models.StepIDData, stepAbsDirPath string) error {
	return nil
}

func (toolkit SwiftToolkit) StepRunCommandArguments(step stepmanModels.StepModel, sIDData models.StepIDData, stepAbsDirPath string) ([]string, error) {
	return []string{}, nil
}

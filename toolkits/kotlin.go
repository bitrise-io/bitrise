package toolkits

import (
	"fmt"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/utils"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// KotlinToolkit ...
type KotlinToolkit struct {
}

// Bootstrap ...
func (toolkit KotlinToolkit) Bootstrap() error {
	fmt.Println("Hello kotlin: Bootstrap")
	return nil
}

// Install ...
func (toolkit KotlinToolkit) Install() error {
	fmt.Println("Hello kotlin: Install")
	return nil
}

// ToolkitName ...
func (toolkit KotlinToolkit) ToolkitName() string {
	return "kotlin"
}

// Check ...
func (toolkit KotlinToolkit) Check() (bool, ToolkitCheckResult, error) {
	fmt.Println("Hello kotlin: Check")
	return false, ToolkitCheckResult{}, nil
}

// IsToolAvailableInPATH ...
func (toolkit KotlinToolkit) IsToolAvailableInPATH() bool {
	binPath, err := utils.CheckProgramInstalledPath("kotlin")
	if err != nil {
		return false
	}
	return len(binPath) > 0
}

// PrepareForStepRun ...
func (toolkit KotlinToolkit) PrepareForStepRun(step stepmanModels.StepModel, sIDData models.StepIDData, stepAbsDirPath string) error {
	return nil
}

// StepRunCommandArguments ...
func (toolkit KotlinToolkit) StepRunCommandArguments(step stepmanModels.StepModel, sIDData models.StepIDData, stepAbsDirPath string) ([]string, error) {
	gradlewPath := stepAbsDirPath + "/gradlew"
	return []string{gradlewPath, "-p", stepAbsDirPath, "run"}, nil
}

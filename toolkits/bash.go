package toolkits

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/utils"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/stringutil"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// BashToolkit ...
type BashToolkit struct {
}

// Check ...
func (toolkit BashToolkit) Check() (bool, ToolkitCheckResult, error) {
	binPath, err := utils.CheckProgramInstalledPath("bash")
	if err != nil {
		return false, ToolkitCheckResult{}, fmt.Errorf("Failed to get bash binary path, error: %s", err)
	}

	verOut, err := command.RunCommandAndReturnStdout("bash", "--version")
	if err != nil {
		return false, ToolkitCheckResult{}, fmt.Errorf("Failed to check bash version, error: %s", err)
	}

	verStr := stringutil.ReadFirstLine(verOut, true)

	return false, ToolkitCheckResult{
		Path:    binPath,
		Version: verStr,
	}, nil
}

// IsToolAvailableInPATH ...
func (toolkit BashToolkit) IsToolAvailableInPATH() bool {
	binPath, err := utils.CheckProgramInstalledPath("bash")
	if err != nil {
		return false
	}
	return len(binPath) > 0
}

// Bootstrap ...
func (toolkit BashToolkit) Bootstrap() error {
	return nil
}

// Install ...
func (toolkit BashToolkit) Install() error {
	return nil
}

// ToolkitName ...
func (toolkit BashToolkit) ToolkitName() string {
	return "bash"
}

// PrepareForStepRun ...
func (toolkit BashToolkit) PrepareForStepRun(_ stepmanModels.StepModel, _ models.StepIDData, _ string) error {
	return nil
}

// StepRunCommandArguments ...
func (toolkit BashToolkit) StepRunCommandArguments(step stepmanModels.StepModel, sIDData models.StepIDData, stepAbsDirPath string) ([]string, error) {
	entryFile := "step.sh"
	if step.Toolkit != nil && step.Toolkit.Bash != nil && step.Toolkit.Bash.EntryFile != "" {
		entryFile = step.Toolkit.Bash.EntryFile
	}

	stepFilePath := filepath.Join(stepAbsDirPath, entryFile)
	cmd := []string{"bash", stepFilePath}
	return cmd, nil
}

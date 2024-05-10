package toolkits

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/bitrise-io/go-utils/stringutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
)

type BashToolkit struct {
}

func (toolkit BashToolkit) Check() (bool, ToolkitCheckResult, error) {
	bashPath, err := exec.LookPath("bash")
	if err != nil {
		return false, ToolkitCheckResult{}, fmt.Errorf("get bash binary path: %s", err)
	}

	verOut, err := exec.Command("bash", "--version").Output()
	if err != nil {
		return false, ToolkitCheckResult{}, fmt.Errorf("check bash version: %w", err)
	}

	verStr := stringutil.ReadFirstLine(string(verOut), true)

	return false, ToolkitCheckResult{
		Path:    bashPath,
		Version: verStr,
	}, nil
}

func (toolkit BashToolkit) IsToolAvailableInPATH() bool {
	_, err := exec.LookPath("bash")
	return err != nil
}

func (toolkit BashToolkit) Bootstrap() error {
	return nil
}

func (toolkit BashToolkit) Install() error {
	return nil
}

func (toolkit BashToolkit) ToolkitName() string {
	return "bash"
}

func (toolkit BashToolkit) PrepareForStepRun(_ models.StepModel, _ stepid.CanonicalID, _ string) error {
	return nil
}

func (toolkit BashToolkit) StepRunCommandArguments(step models.StepModel, sIDData stepid.CanonicalID, stepAbsDirPath string) ([]string, error) {
	entryFile := "step.sh"
	if step.Toolkit != nil && step.Toolkit.Bash != nil && step.Toolkit.Bash.EntryFile != "" {
		entryFile = step.Toolkit.Bash.EntryFile
	}

	stepFilePath := filepath.Join(stepAbsDirPath, entryFile)
	cmd := []string{"bash", stepFilePath}
	return cmd, nil
}

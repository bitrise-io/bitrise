package toolkits

import (
	"fmt"
	"os"
	"path/filepath"

	cliModels "github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
)

type StepExecutor interface {
	// GetStepRunCommand  ...
	GetStepRunCommand(step models.StepModel) ([]string, error)
}

type ScriptStepExecutor struct {
	sourceAbsDirPath string
}

func NewScriptStepExecutor(sourceAbsDirPath string, destinationDir string) (*ScriptStepExecutor, error) {
	if sourceAbsDirPath != destinationDir {
		if err := copyStep(sourceAbsDirPath, destinationDir); err != nil {
			return nil, fmt.Errorf("copy step failed: %s", err)
		}
	}

	return &ScriptStepExecutor{
		sourceAbsDirPath: destinationDir,
	}, nil
}

func (s *ScriptStepExecutor) GetStepRunCommand(step models.StepModel) ([]string, error) {
	entryFile := "step.sh"
	if step.Toolkit != nil && step.Toolkit.Bash != nil && step.Toolkit.Bash.EntryFile != "" {
		entryFile = step.Toolkit.Bash.EntryFile
	}

	stepFilePath := filepath.Join(s.sourceAbsDirPath, entryFile)
	cmd := []string{"bash", stepFilePath}
	return cmd, nil
}

type GoStepExecutor struct {
	executablePath string
}

func NewExecutableStepExecutor(executablePath string) *GoStepExecutor {
	return &GoStepExecutor{
		executablePath: executablePath,
	}
}

func (g *GoStepExecutor) GetStepRunCommand(step models.StepModel) ([]string, error) {
	return []string{g.executablePath}, nil
}

func NewLocalStepExecutor(sourceAbsDirPath string, step models.StepModel, sIDData cliModels.StepIDData, destinationDir string) (StepExecutor, error) {
	toolkitForStep := ToolkitForStep(step)
	toolkitName := toolkitForStep.ToolkitName()

	switch toolkitName {
	case "bash":
		return NewScriptStepExecutor(sourceAbsDirPath, destinationDir)
	case "go":
		return toolkitForStep.PrepareForStepRun(step, sIDData, sourceAbsDirPath)
	default:
		return nil, fmt.Errorf("Unsupported toolkit: %s", toolkitName)
	}
}

func NewSteplibStepExecutor(sourceAbsDirPath string, step models.StepModel, executablePath string, destinationDir string) (StepExecutor, error) {
	toolkitForStep := ToolkitForStep(step)
	toolkitName := toolkitForStep.ToolkitName()

	switch toolkitName {
	case "bash":
		return NewScriptStepExecutor(sourceAbsDirPath, destinationDir)
	case "go":
		packageName := step.Toolkit.Go.PackageName
		return toolkitForStep.CompileStepExecutable(sourceAbsDirPath, packageName, executablePath)
	default:
		return nil, fmt.Errorf("Unsupported toolkit: %s", toolkitName)
	}
}

type ActivatedStep struct {
	StepExecutor StepExecutor
	StepYMLPath  string
}

func copyStep(src, dst string) error {
	if exist, err := pathutil.IsPathExists(dst); err != nil {
		return fmt.Errorf("failed to check if %s path exist: %s", dst, err)
	} else if !exist {
		if err := os.MkdirAll(dst, 0777); err != nil {
			return fmt.Errorf("failed to create dir for %s path: %s", dst, err)
		}
	}

	if err := command.CopyDir(src+"/", dst, true); err != nil {
		return fmt.Errorf("copy command failed: %s", err)
	}
	return nil
}

package toolkits

import "path/filepath"

// BashToolkit ...
type BashToolkit struct {
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

// StepRunCommandArguments ...
func (toolkit BashToolkit) StepRunCommandArguments(stepDirPath string) ([]string, error) {
	stepFilePath := filepath.Join(stepDirPath, "step.sh")
	cmd := []string{"bash", stepFilePath}
	return cmd, nil
}

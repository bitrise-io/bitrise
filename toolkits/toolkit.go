package toolkits

import stepmanModels "github.com/bitrise-io/stepman/models"

// ToolkitCheckResult ...
type ToolkitCheckResult struct {
	Path    string
	Version string
}

// Toolkit ...
type Toolkit interface {
	// ToolkitName : a one liner name/id of the toolkit, for logging purposes
	ToolkitName() string
	// Check the toolkit - first returned value (bool) indicates
	// whether the toolkit is "operational", or have to be installed.
	// "Have to be installed" can be true if the toolkit is not installed,
	// or if an older version is installed, and an update/newer version is required.
	Check() (bool, ToolkitCheckResult, error)
	// Install the toolkit
	Install() error
	// PrepareForStepRun can be used to pre-compile or otherwise
	// prepare for the step's execution
	PrepareForStepRun(step stepmanModels.StepModel, stepAbsDirPath string) error
	// Bootstrap : initialize the toolkit for use,
	// e.g. setting Env Vars
	Bootstrap() error
	// StepRunCommandArguments ...
	StepRunCommandArguments(stepDirPath string) ([]string, error)
}

//
// === Utils ===

// ToolkitForStep ...
func ToolkitForStep(step stepmanModels.StepModel) Toolkit {
	if step.Toolkit != nil {
		stepToolkit := step.Toolkit
		if stepToolkit.Go != nil {
			return GoToolkit{}
		} else if stepToolkit.Bash != nil {
			return BashToolkit{}
		}
	}

	// default
	return BashToolkit{}
}

// AllSupportedToolkits ...
func AllSupportedToolkits() []Toolkit {
	return []Toolkit{GoToolkit{}, BashToolkit{}}
}

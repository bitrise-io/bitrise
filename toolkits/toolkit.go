package toolkits

import (
	"github.com/bitrise-io/bitrise/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

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
	// whether the toolkit have to be installed (true=install required | false=no install required).
	// "Have to be installed" can be true if the toolkit is not installed,
	// or if an older version is installed, and an update/newer version is required.
	Check() (bool, ToolkitCheckResult, error)
	// Install the toolkit (if required / system installed version is not sufficient)
	Install() error
	// Bootstrap : initialize the toolkit for use,
	// e.g. setting Env Vars. Will run only once, before the build would actually start,
	// so that it can set e.g. a default Go or other language version,
	// but then the user should be able to choose another one when/if needed!
	// Bootstrap should only set a sensible default, but should not enforce it!
	// The `PrepareForStepRun` function will be called for every step,
	// the toolkit should be "enforced" there, BUT ONLY FOR THE STEP (e.g. don't call os.Setenv there!)
	Bootstrap() error
	// PrepareForStepRun can be used to pre-compile or otherwise
	// prepare for the step's execution.
	// Important: do NOT enforce the toolkit for subsequent / unrelated steps,
	// the toolkit can be "enforced" here (e.g. during the compilation),
	// BUT ONLY FOR THE STEP (e.g. don't call `os.Setenv` !! Instead add
	// the required envs to the command's `.Env`!)
	PrepareForStepRun(step stepmanModels.StepModel, sIDData models.StepIDData, stepAbsDirPath string) error
	// StepRunCommandArguments ...
	StepRunCommandArguments(stepDirPath string, sIDData models.StepIDData) ([]string, error)
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

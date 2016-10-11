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

	// Install the toolkit
	Install() error

	// Check whether the toolkit's tool (e.g. Go, Ruby, Bash, ...) is available
	// and "usable"" without any bootstrapping.
	// Return true even if the version is older than the required version for this toolkit,
	// you can pick / init the right version later. This function only checks
	// whether the system has a pre-installed version of the toolkit's tool or not,
	// no compability check is required!
	IsToolAvailableInPATH() bool

	// Bootstrap : initialize the toolkit for use, ONLY IF THERE'S NO SYSTEM INSTALLED VERSION!
	// If there's any version of the tool (e.g. Go) installed, Bootstrap should not overwrite it,
	// so that the non toolkit steps will still use the system installed version!
	//
	// Will run only once, before the build would actually start,
	// so that it can set e.g. a default Go or other language version, if there's no System Installed version!
	//
	// Bootstrap should only set a sensible default if there's no System Installed version of the tool,
	// but should not enforce the toolkit's version of the tool!
	// The `PrepareForStepRun` function will be called for every step,
	// the toolkit should be "enforced" there, BUT ONLY FOR THAT FUNCTION (e.g. don't call os.Setenv there!)
	Bootstrap() error

	// PrepareForStepRun can be used to pre-compile or otherwise prepare for the step's execution.
	//
	// Important: do NOT enforce the toolkit for subsequent / unrelated steps or functions,
	// the toolkit should/can be "enforced" here (e.g. during the compilation),
	// BUT ONLY for this function! E.g. don't call `os.Setenv` or something similar
	// which would affect other functions, just pass the required envs to the compilation command!
	PrepareForStepRun(step stepmanModels.StepModel, sIDData models.StepIDData, stepAbsDirPath string) error

	// StepRunCommandArguments ...
	StepRunCommandArguments(step stepmanModels.StepModel, sIDData models.StepIDData, stepAbsDirPath string) ([]string, error)
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

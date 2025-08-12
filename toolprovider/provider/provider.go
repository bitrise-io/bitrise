package provider

import "fmt"

type ToolProvider interface {
	ID() string

	Bootstrap() error

	InstallTool(tool ToolRequest) (ToolInstallResult, error)

	ActivateEnv(result ToolInstallResult) (EnvironmentActivation, error)
}

type ToolID string

type ToolRequest struct {
	ToolName ToolID

	// UnparsedVersion is the version string as provided by the user.
	// It may or may not be a valid semantic version.
	UnparsedVersion    string
	ResolutionStrategy ResolutionStrategy

	// PluginIdentifier is an optional user-defined identifier for the tool-plugin.
	PluginIdentifier *string
}

type ResolutionStrategy int

const (
	ResolutionStrategyStrict ResolutionStrategy = iota
	ResolutionStrategyLatestInstalled
	ResolutionStrategyLatestReleased
)

type ToolInstallResult struct {
	ToolName           ToolID
	IsAlreadyInstalled bool

	// ConcreteVersion is the version that was actually installed and we resolved to.
	// It may differ from the requested version if the requested version was not a concrete version.
	// This value may or may not be a valid semantic version.
	ConcreteVersion string
}

// TODO: Mise merges envs and $PATH changes into one output, maybe we should do the same for asdf?
// It would simplify the activation process.
type EnvironmentActivation struct {
	ContributedEnvVars map[string]string
	ContributedPaths   []string
}

type ToolInstallError struct {
	ToolName         ToolID
	RequestedVersion string

	// Optional fields
	RawOutput      string
	Cause          string
	Recommendation string
}

func (e ToolInstallError) Error() string {
	msg := fmt.Sprintf("Error: failed to install %s %s", e.ToolName, e.RequestedVersion)

	if e.Cause != "" {
		msg += "\nCause: " + e.Cause
	}

	if e.Recommendation != "" {
		msg += "\nRecommendation: " + e.Recommendation
	}

	if e.RawOutput != "" {
		msg += "\nAdditional info: " + e.RawOutput
	}

	return msg
}

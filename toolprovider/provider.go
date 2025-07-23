package toolprovider

var SupportedProviders = []string{ "asdf", "mise" }

type ToolID string

type ToolRequest struct {
	ToolName ToolID

	// UnparsedVersion is the version string as provided by the user.
	// It may or may not be a valid semantic version.
	UnparsedVersion string

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

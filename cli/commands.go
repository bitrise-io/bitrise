package cli

import "github.com/spf13/cobra"

// Flags constants used across commands
const (
	JSONParamsKey       = "json-params"
	JSONParamsBase64Key = "json-params-base64"

	WorkflowKey = "workflow"

	PatternKey        = "pattern"
	PushBranchKey     = "push-branch"
	PRSourceBranchKey = "pr-source-branch"
	PRTargetBranchKey = "pr-target-branch"
	PRReadyStateKey   = "pr-ready-state"

	ConfigKey      = "config"
	InventoryKey   = "inventory"
	OuputFormatKey = "format"
)

func registerCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(
		initCmd,
		setupCmd,
		stepsCmd,
		toolsCmd,
		versionCmd,
		validateCmd,
		updateCmd,
		runCmd,
		triggerCheckCmd,
		triggerCmd,
		workflowsCmd,
		shareCmd,
		pluginCmd,
		envmanCmd,
		mergeCmd,
	)
}

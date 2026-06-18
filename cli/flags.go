package cli

import (
	"github.com/spf13/pflag"
)

// Flags ...
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

const (
	CollectionPathEnvKey = "STEPMAN_COLLECTION"

	CIKey        = "ci"
	PRKey        = "pr"
	DebugModeKey = "debug"

	CollectionKey = "collection"

	inventoryShortKey = "i"

	InventoryBase64Key = "inventory-base64"

	configShortKey = "c"

	ConfigBase64Key = "config-base64"

	TagKey    = "tag"
	GitKey    = "git"
	StepIDKey = "stepid"
)

func addConfigAndInventoryFlags(fs *pflag.FlagSet) {
	fs.StringP(ConfigKey, configShortKey, "", "Path where the workflow config file is located.")
	fs.String(ConfigBase64Key, "", "base64 encoded config data.")
	fs.StringP(InventoryKey, inventoryShortKey, "", "Path of the inventory file.")
	fs.String(InventoryBase64Key, "", "base64 encoded inventory data.")
}

func addJSONParamsFlags(fs *pflag.FlagSet) {
	fs.String(JSONParamsKey, "", "Specify command flags with json string-string hash.")
	fs.String(JSONParamsBase64Key, "", "Specify command flags with base64 encoded json string-string hash.")
}

func addTriggerFilterFlags(fs *pflag.FlagSet) {
	fs.String(PatternKey, "", "trigger pattern.")
	fs.String(PushBranchKey, "", "Git push branch name.")
	fs.String(PRSourceBranchKey, "", "Git pull request source branch name.")
	fs.String(PRTargetBranchKey, "", "Git pull request target branch name.")
	fs.String(PRReadyStateKey, "", "Git pull request ready state. Options: ready_for_review draft converted_to_ready_for_review")
	fs.String(TagKey, "", "Git tag name.")
}

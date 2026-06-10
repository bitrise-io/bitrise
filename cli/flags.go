package cli

const (
	CollectionPathEnvKey = "STEPMAN_COLLECTION"
	CIKey                = "ci"
	PRKey                = "pr"
	DebugModeKey         = "debug"

	VersionKey = "version"

	CollectionKey = "collection"

	inventoryShortKey  = "i"
	InventoryBase64Key = "inventory-base64"

	configShortKey  = "c"
	ConfigBase64Key = "config-base64"

	HelpKey = "help"

	MinimalModeKey = "minimal"
	FullModeKey    = "full"

	OuputPathKey = "outpath"
	PrettyFormatKey     = "pretty"

	IDKey    = "id"
	ShortKey = "short"

	StepYMLKey = "step-yml"

	TagKey    = "tag"
	GitKey    = "git"
	StepIDKey = "stepid"
)

// Global persistent flag values — bound to root command's persistent flags in cli.go
var (
	debugMode bool
	ciMode    bool
	prMode    bool
)

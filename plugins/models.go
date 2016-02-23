package plugins

const (
	// TypeGeneric ...
	TypeGeneric = "_"
	// TypeInit ...
	TypeInit = "init"
	// TypeRun ....
	TypeRun = "run"
)

// PluginRoute ...
type PluginRoute struct {
	Name       string `yaml:"name"`
	Source     string `yaml:"source"`
	Version    string `yaml:"version"`
	CommitHash string `yaml:"commit_hash"`
	Executable string `yaml:"executable"`
}

// PluginRouting ...
type PluginRouting struct {
	RouteMap map[string]PluginRoute `yaml:"route_map"`
}

// Plugin ...
type Plugin struct {
	Name         string        `yaml:"name"`
	Description  string        `yaml:"description"`
	Executable   string        `yaml:"executable"`
	Requirements []Requirement `yaml:"requirements"`
}

// Requirement ...
type Requirement struct {
	Tool       string `yaml:"tool"`
	MinVersion string `yaml:"min_version"`
	MaxVersion string `yaml:"max_version"`
}

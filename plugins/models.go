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
	Name         string `yaml:"name"`
	Source       string `yaml:"source"`
	Version      string `yaml:"version"`
	CommitHash   string `yaml:"commit_hash"`
	Executable   string `yaml:"executable"`
	TriggerEvent string `yaml:"trigger"`
}

// PluginRouting ...
type PluginRouting struct {
	RouteMap map[string]PluginRoute `yaml:"route_map"`
}

// Plugin ...
type Plugin struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Executable  struct {
		Osx   string `yaml:"osx"`
		Linux string `yaml:"linux"`
	}
	TriggerEvent string        `yaml:"trigger"`
	Requirements []Requirement `yaml:"requirements"`
}

// Requirement ...
type Requirement struct {
	Tool       string `yaml:"tool"`
	MinVersion string `yaml:"min_version"`
	MaxVersion string `yaml:"max_version"`
}

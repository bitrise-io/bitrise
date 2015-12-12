package plugins

import (
	"fmt"
	"strings"
)

const (
	// TypeCustom ...
	TypeCustom = "custom"

	// TypeInit ...
	TypeInit = "init"

	// TypeRun ....
	TypeRun = "run"
)

// Plugin ...
type Plugin struct {
	Path string
	Name string
	Type string
}

// PrintableName ...
func (plugin Plugin) PrintableName() string {
	switch plugin.Type {
	case TypeCustom:
		return fmt.Sprintf(":%s", plugin.Name)
	default:
		return fmt.Sprintf("%s:%s", plugin.Type, plugin.Name)
	}
}

// ParsePrintableName ...
func ParsePrintableName(printableName string) (string, string, error) {
	if !strings.Contains(printableName, ":") {
		return "", "", fmt.Errorf("Invalid plugin name: %s", printableName)
	}

	if strings.HasPrefix(printableName, ":") {
		return strings.TrimPrefix(printableName, ":"), TypeCustom, nil
	}

	splits := strings.Split(printableName, ":")
	if len(splits) == 2 {
		return splits[1], splits[0], nil
	}

	return "", "", fmt.Errorf("Invalid plugin name: %s", printableName)
}

// PrintableName ...
func PrintableName(pluginName, pluginType string) string {
	switch pluginType {
	case TypeCustom:
		return fmt.Sprintf(":%s", pluginName)
	default:
		return fmt.Sprintf("%s:%s", pluginType, pluginName)
	}
}

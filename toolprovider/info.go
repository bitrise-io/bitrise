package toolprovider

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
)

// InstalledTool represents an installed tool with its versions.
type InstalledTool struct {
	Name              string   `json:"name"`
	InstalledVersions []string `json:"installed_versions,omitempty"`
	ActiveVersion     string   `json:"active_version,omitempty"`
	Source            string   `json:"source,omitempty"`
}

// ListInstalledTools lists all tools installed by the configured provider.
// If activeOnly is true, only tools that are currently active in the shell context are returned.
func ListInstalledTools(providerName string, activeOnly bool) ([]InstalledTool, error) {
	if providerName == "" {
		providerName = "mise"
	}

	switch providerName {
	case "asdf":
		return listAsdfTools(activeOnly)
	case "mise":
		return listMiseTools(activeOnly)
	default:
		return nil, fmt.Errorf("unsupported tool provider: %s", providerName)
	}
}

func listAsdfTools(activeOnly bool) ([]InstalledTool, error) {
	asdfProvider := &asdf.AsdfToolProvider{
		ExecEnv: execenv.ExecEnv{
			EnvVars:            map[string]string{},
			ShellInit:          "",
			ClearInheritedEnvs: false,
		},
	}

	if activeOnly {
		output, err := asdfProvider.ExecEnv.RunAsdf("current")
		if err != nil {
			return nil, fmt.Errorf("list asdf current tools: %w", err)
		}

		var tools []InstalledTool
		lines := parseLines(output)
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				source := ""
				if len(parts) >= 3 {
					source = parts[2]
				}
				tools = append(tools, InstalledTool{
					Name:          parts[0],
					ActiveVersion: parts[1],
					Source:        source,
				})
			}
		}

		sort.Slice(tools, func(i, j int) bool {
			return tools[i].Name < tools[j].Name
		})

		return tools, nil
	}

	output, err := asdfProvider.ExecEnv.RunAsdf("plugin", "list")
	if err != nil {
		return nil, fmt.Errorf("list asdf plugins: %w", err)
	}

	pluginNames := parseLines(output)
	var tools []InstalledTool

	for _, pluginName := range pluginNames {
		versionsOutput, err := asdfProvider.ExecEnv.RunAsdf("list", pluginName)
		if err != nil {
			// Plugin might be installed but no versions installed yet.
			tools = append(tools, InstalledTool{
				Name:              pluginName,
				InstalledVersions: []string{},
			})
			continue
		}

		versions := parseLines(versionsOutput)
		var activeVersion string
		var installedVersions []string

		for _, v := range versions {
			// asdf marks the current version with an asterisk.
			if len(v) > 0 && v[0] == '*' {
				activeVersion = strings.TrimSpace(v[1:])
				installedVersions = append(installedVersions, activeVersion)
			} else {
				installedVersions = append(installedVersions, strings.TrimSpace(v))
			}
		}

		tools = append(tools, InstalledTool{
			Name:              pluginName,
			InstalledVersions: installedVersions,
			ActiveVersion:     activeVersion,
		})
	}

	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})

	return tools, nil
}

// miseToolEntry represents a single tool version entry from mise ls --json output.
type miseToolEntry struct {
	Version   string `json:"version"`
	Requested string `json:"requested_version"`
	Source    *struct {
		Type string `json:"type"`
		Path string `json:"path"`
	} `json:"source"`
	Installed bool `json:"installed"`
	Active    bool `json:"active"`
}

func listMiseTools(activeOnly bool) ([]InstalledTool, error) {
	miseInstallDir, miseDataDir := mise.Dirs(mise.GetMiseVersion())
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, false)
	if err != nil {
		return nil, fmt.Errorf("create mise provider: %w", err)
	}

	err = miseProvider.Bootstrap()
	if err != nil {
		return nil, fmt.Errorf("bootstrap mise: %w", err)
	}

	// Use mise ls --json to get tools. Add --current flag for active tools only.
	args := []string{"ls", "--json"}
	if activeOnly {
		args = append(args, "--current")
	}
	output, err := miseProvider.ExecEnv.RunMise(args...)
	if err != nil {
		return nil, fmt.Errorf("list mise tools: %w", err)
	}

	// mise ls --json now returns a map with tool names as keys.
	// e.g., {"go": [{"version": "1.21.0", ...}], "node": [{...}]}
	var miseToolsMap map[string][]miseToolEntry
	if err := json.Unmarshal([]byte(output), &miseToolsMap); err != nil {
		return nil, fmt.Errorf("parse mise ls output: %w", err)
	}

	var tools []InstalledTool
	for toolName, entries := range miseToolsMap {
		tool := InstalledTool{
			Name: toolName,
		}

		for _, entry := range entries {
			if activeOnly {
				if entry.Active {
					tool.ActiveVersion = entry.Version
					if entry.Source != nil {
						tool.Source = entry.Source.Path
					}
				}
			} else {
				if entry.Installed {
					tool.InstalledVersions = append(tool.InstalledVersions, entry.Version)
				}
				if entry.Active {
					tool.ActiveVersion = entry.Version
					if entry.Source != nil {
						tool.Source = entry.Source.Path
					}
				}
			}
		}

		// Only include tools that have relevant data.
		if activeOnly {
			if tool.ActiveVersion != "" {
				tools = append(tools, tool)
			}
		} else {
			if len(tool.InstalledVersions) > 0 || tool.ActiveVersion != "" {
				tools = append(tools, tool)
			}
		}
	}

	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})

	return tools, nil
}

func parseLines(output string) []string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

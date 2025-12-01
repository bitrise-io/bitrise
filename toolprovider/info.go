package toolprovider

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
)

// InstalledTool represents an installed tool with its versions.
type InstalledTool struct {
	Name              string   `json:"name"`
	InstalledVersions []string `json:"installed_versions"`
	ActiveVersion     string   `json:"active_version,omitempty"`
}

// ListInstalledTools lists all tools installed by the configured provider.
func ListInstalledTools(providerName string) ([]InstalledTool, error) {
	if providerName == "" {
		providerName = "mise"
	}

	switch providerName {
	case "asdf":
		return listAsdfTools()
	case "mise":
		return listMiseTools()
	default:
		return nil, fmt.Errorf("unsupported tool provider: %s", providerName)
	}
}

func listAsdfTools() ([]InstalledTool, error) {
	asdfProvider := &asdf.AsdfToolProvider{
		ExecEnv: execenv.ExecEnv{
			EnvVars:            map[string]string{},
			ShellInit:          "",
			ClearInheritedEnvs: false,
		},
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
				activeVersion = v[1:]
				activeVersion = trimWhitespace(activeVersion)
				installedVersions = append(installedVersions, activeVersion)
			} else {
				installedVersions = append(installedVersions, trimWhitespace(v))
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

func listMiseTools() ([]InstalledTool, error) {
	miseInstallDir, miseDataDir := mise.Dirs(mise.GetMiseVersion())
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, models.ToolConfigModel{})
	if err != nil {
		return nil, fmt.Errorf("create mise provider: %w", err)
	}

	err = miseProvider.Bootstrap()
	if err != nil {
		return nil, fmt.Errorf("bootstrap mise: %w", err)
	}

	// Use mise ls --json to get all installed tools.
	output, err := miseProvider.ExecEnv.RunMise("ls", "--json")
	if err != nil {
		return nil, fmt.Errorf("list mise tools: %w", err)
	}

	var miseTools []struct {
		Name      string `json:"name"`
		Version   string `json:"version"`
		Requested string `json:"requested_version"`
		Source    struct {
			Type string `json:"type"`
			Path string `json:"path"`
		} `json:"source"`
		Installed bool `json:"installed"`
		Active    bool `json:"active"`
	}

	if err := json.Unmarshal([]byte(output), &miseTools); err != nil {
		return nil, fmt.Errorf("parse mise ls output: %w", err)
	}

	toolMap := make(map[string]*InstalledTool)
	for _, mt := range miseTools {
		if !mt.Installed {
			continue
		}

		tool, exists := toolMap[mt.Name]
		if !exists {
			tool = &InstalledTool{
				Name:              mt.Name,
				InstalledVersions: []string{},
			}
			toolMap[mt.Name] = tool
		}

		versionExists := false
		for _, v := range tool.InstalledVersions {
			if v == mt.Version {
				versionExists = true
				break
			}
		}
		if !versionExists {
			tool.InstalledVersions = append(tool.InstalledVersions, mt.Version)
		}

		if mt.Active {
			tool.ActiveVersion = mt.Version
		}
	}

	var tools []InstalledTool
	for _, tool := range toolMap {
		tools = append(tools, *tool)
	}

	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})

	return tools, nil
}

func parseLines(output string) []string {
	var lines []string
	line := ""
	for _, c := range output {
		if c == '\n' {
			if trimmed := trimWhitespace(line); trimmed != "" {
				lines = append(lines, trimmed)
			}
			line = ""
		} else {
			line += string(c)
		}
	}
	if trimmed := trimWhitespace(line); trimmed != "" {
		lines = append(lines, trimmed)
	}
	return lines
}

func trimWhitespace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

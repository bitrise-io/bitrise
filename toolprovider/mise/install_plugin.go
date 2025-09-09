package mise

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

type PluginSource struct {
	PluginName  provider.ToolID
	GitCloneURL string
}

var pluginSourceMap = map[provider.ToolID]PluginSource{
	"flutter": {PluginName: "flutter", GitCloneURL: "https://github.com/asdf-community/asdf-flutter.git"},
	"tuist":   {PluginName: "tuist", GitCloneURL: "https://github.com/tuist/asdf-tuist.git"},
}

var miseCoreTools = []string{
	"bun",
	"deno",
	"elixir",
	"erlang",
	"go",
	"golang",
	"java",
	"node",
	"nodejs",
	"python",
	"ruby",
	"rust",
	"swift",
	"zig",
}

// InstallPlugin installs a plugin for the specified tool, if needed.
//
// It resolves the plugin source from the tool request or predefined map,
// checks if the plugin is already installed, and if not, installs it using mise.
func (m *MiseToolProvider) InstallPlugin(tool provider.ToolRequest) error {
	if slices.Contains(miseCoreTools, string(tool.ToolName)) {
		// Core tools do not require plugin installation.
		return nil
	}

	plugin := parsePluginSource(tool)
	if plugin == nil {
		return provider.ToolInstallError{
			ToolName:         tool.ToolName,
			RequestedVersion: tool.UnparsedVersion,
			Cause:            fmt.Sprintf("This tool integration (%s) is not tested or vetted by Bitrise.", tool.ToolName),
			Recommendation:   fmt.Sprintf("If you want to use this tool anyway, look up its asdf plugin and set its git clone URL in tool_config.extra_plugins. For example: `%s: https://github/url/to/asdf/plugin/repo.git`", tool.ToolName),
		}
	}
	if plugin.PluginName == "" {
		// Plugin name is required to install the plugin.
		return fmt.Errorf("plugin name for tool %s is not defined", tool.ToolName)
	}

	installed, err := m.isPluginInstalled(*plugin)
	if err != nil {
		log.Warnf("Failed to check if plugin is already installed: %v", err)
	}
	if installed {
		log.Debugf("Tool plugin %s is already installed, skipping installation.", tool.ToolName)
		return nil
	}

	pluginInstallArgs := []string{"install", string(plugin.PluginName)}

	if plugin.GitCloneURL != "" {
		pluginInstallArgs = append(pluginInstallArgs, plugin.GitCloneURL)
	}

	_, err = m.ExecEnv.RunMisePlugin(pluginInstallArgs...)
	if err != nil {
		return err
	}

	// Check if the plugin is found in the list of installed plugins after adding.
	installed, err = m.isPluginInstalled(*plugin)
	if err != nil {
		return fmt.Errorf("check if plugin was installed successfully: %w", err)
	}
	if !installed {
		return fmt.Errorf("%s plugin could not be installed", tool.ToolName)
	}

	return nil
}

func (m *MiseToolProvider) isPluginInstalled(plugin PluginSource) (bool, error) {
	pluginListArgs := []string{"list", "--urls"}
	out, err := m.ExecEnv.RunMisePlugin(pluginListArgs...)
	if err != nil {
		return false, err
	}
	// If no plugins are installed, mise returns exit code 0.

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, string(plugin.PluginName)) {
			if plugin.GitCloneURL != "" && !strings.Contains(line, string(plugin.GitCloneURL)) {
				log.Warnf("installed, but required URL does not match current:\n%s %s\n%s", plugin.PluginName, plugin.GitCloneURL, line)
			}
			return true, nil
		}
	}

	return false, nil
}

func parsePluginSource(toolRequest provider.ToolRequest) *PluginSource {
	if toolRequest.PluginURL != nil {
		url := strings.TrimSpace(*toolRequest.PluginURL)
		if url != "" {
			// User provided a non empty plugin identifier, use it as a git clone URL to the asdf plugin.
			return &PluginSource{
				PluginName:  toolRequest.ToolName,
				GitCloneURL: url,
			}
		}
	}

	// Check if we have a predefined plugin source.
	if toolPlugin, exists := pluginSourceMap[toolRequest.ToolName]; exists {
		return &toolPlugin
	}

	// No predefined plugin source found and no error in plugin identifier parsing,
	// return nil to indicate that no plugin source is defined for this tool.
	return nil
}

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

// Ref: https://mise.jdx.dev/core-tools.html
// Note: we might need to sync this list from time to time
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
	plugin, err := m.pluginToInstall(tool)
	if err != nil {
		return err
	}
	if plugin == nil {
		// No plugin installation needed (either core tool or registry tool).
		log.Debugf("[TOOLPROVIDER] No plugin installation needed for tool %s", tool.ToolName)
		return nil
	}

	installed, err := m.isPluginInstalled(plugin)
	if err != nil {
		log.Warnf("Failed to check if plugin is already installed: %v", err)
	}
	if installed {
		log.Debugf("[TOOLPROVIDER] Tool plugin %s is already installed, skipping installation.", tool.ToolName)
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
	installed, err = m.isPluginInstalled(plugin)
	if err != nil {
		return fmt.Errorf("check if plugin was installed successfully: %w", err)
	}
	if !installed {
		return fmt.Errorf("%s plugin could not be installed", tool.ToolName)
	}

	return nil
}

// RegistryChecker interface required for testing.
type RegistryChecker interface {
	isPluginInRegistry(name string) error
}

// pluginToInstall is a wrapper to call the pure function with the MiseToolProvider as RegistryChecker.
func (m *MiseToolProvider) pluginToInstall(tool provider.ToolRequest) (*PluginSource, error) {
	return pluginToInstall(tool, m)
}

// pluginToInstall is a pure function that determines what plugin needs to be installed for a given tool request.
// It takes a RegistryChecker interface to allow for easy testing with mocks.
func pluginToInstall(tool provider.ToolRequest, registryChecker RegistryChecker) (*PluginSource, error) {
	pluginName := tool.ToolName
	if pluginName == "" {
		// Plugin name is required to install the plugin.
		return nil, fmt.Errorf("tool name is not defined for plugin installation")
	}

	if tool.PluginURL != nil {
		url := strings.TrimSpace(*tool.PluginURL)
		if url != "" {
			// User provided a non empty plugin git clone URL, use it as a git clone URL to the asdf plugin.
			return &PluginSource{
				PluginName:  pluginName,
				GitCloneURL: url,
			}, nil
		}
	}

	if slices.Contains(miseCoreTools, string(pluginName)) {
		// Core tools do not require plugin installation, if user did not specify a custom plugin URL.
		return nil, nil
	}

	if err := registryChecker.isPluginInRegistry(string(pluginName)); err == nil {
		// The tool is found in the registry, no need to install a plugin.
		return nil, nil
	}

	return nil, provider.ToolInstallError{
		ToolName:         pluginName,
		RequestedVersion: tool.UnparsedVersion,
		Cause:            fmt.Sprintf("This tool integration (%s) is not tested or vetted by Bitrise.", pluginName),
		Recommendation:   fmt.Sprintf("If you want to use this tool anyway, look up its asdf plugin and set its git clone URL in tool_config.extra_plugins. For example: `%s: https://github/url/to/asdf/plugin/repo.git`", pluginName),
	}
}

func (m *MiseToolProvider) isPluginInstalled(plugin *PluginSource) (bool, error) {
	pluginListArgs := []string{"list", "--urls", "--quiet"}
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

func (m *MiseToolProvider) isPluginInRegistry(name string) error {
	registryArgs := []string{"registry", name, "--quiet"}
	_, err := m.ExecEnv.RunMise(registryArgs...)
	if err != nil {
		// If the tool is not found in registry, mise returns exit code 1 with error message
		// "tool not found in registry: <toolname>".
		return fmt.Errorf("tool not found in registry: %s", name)
	}

	return nil
}

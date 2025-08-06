package asdf

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/hashicorp/go-version"
)

func (a *AsdfToolProvider) asdfVersion() (*version.Version, error) {
	output, err := a.ExecEnv.RunAsdf("--version")
	if err != nil {
		return nil, err
	}

	versionStr := strings.TrimSpace(string(output))
	pattern := regexp.MustCompile(`(?:asdf version )?v?(\d+\.\d+\.\d+)( \(?:revision .+\))?`)
	matches := pattern.FindStringSubmatch(versionStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("parse version from --version output: %s", versionStr)
	}
	ver, err := version.NewVersion(matches[1])
	if err != nil {
		return nil, fmt.Errorf("parse asdf version: %w", err)
	}
	return ver, nil
}

func (a *AsdfToolProvider) listInstalled(toolName provider.ToolID) ([]string, error) {
	isInstalled, err := a.isPluginInstalled(PluginSource{PluginName: provider.ToolID(toolName)})
	if err != nil {
		return nil, fmt.Errorf("check if %s is installed: %w", toolName, err)
	}
	if !isInstalled {
		return nil, fmt.Errorf("tool plugin %s is not installed", toolName)
	}

	output, err := a.ExecEnv.RunAsdf("list", string(toolName))
	if err != nil {
		// asdf 0.16.0+ returns exit code 1 if no versions are installed
		if strings.Contains(err.Error(), "No compatible versions installed") {
			return []string{}, nil
		}
		return nil, err
	}

	installedVersions := parseAsdfListOutput(output)
	filteredVersions, err := filterAliasVersions(string(toolName), installedVersions)
	if err != nil {
		return nil, fmt.Errorf("filter alias versions: %w", err)
	}
	return filteredVersions, nil
}

func (a *AsdfToolProvider) listReleased(toolName provider.ToolID) ([]string, error) {
	isInstalled, err := a.isPluginInstalled(PluginSource{PluginName: provider.ToolID(toolName)})
	if err != nil {
		return nil, fmt.Errorf("check if %s is installed: %w", toolName, err)
	}
	if !isInstalled {
		return nil, fmt.Errorf("tool plugin %s is not installed", toolName)
	}

	asdfVer, err := a.asdfVersion()
	if err != nil {
		return nil, err
	}
	var subcommands []string
	if asdfVer.GreaterThanOrEqual(version.Must(version.NewVersion("0.16.0"))) {
		subcommands = []string{"list", "all", string(toolName)}
	} else {
		subcommands = []string{"list-all", string(toolName)}
	}

	output, err := a.ExecEnv.RunAsdf(subcommands...)
	if err != nil {
		return nil, err
	}

	releasedVersions := parseAsdfListOutput(output)
	return releasedVersions, nil
}

func parseAsdfListOutput(output string) []string {
	// There is no machine-readable output, we are parsing this:
	//   1.21.0
	//   1.21.11
	//   1.21
	//   1.22.0
	//  *1.22
	//   1.23.5
	//   1.23.7
	//   1.23
	//   1.24.0
	//   1

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var versions = []string{}

	if lines[0] == "No versions installed" || strings.Contains(lines[0], "No compatible versions installed") {
		return versions
	}
	for i := range lines {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		versions = append(versions, strings.TrimSpace(strings.Replace(lines[i], "*", "", 1)))
	}
	return versions
}

func filterAliasVersions(tool string, versions []string) ([]string, error) {
	// Filter out versions that are symlinks created by the asdf-alias plugin.
	var filtered []string
	for _, v := range versions {
		out, err := exec.Command("asdf", "where", tool, v).Output()
		if err != nil {
			return nil, fmt.Errorf("asdf where %s %s: %w", tool, v, err)
		}

		fileInfo, err := os.Lstat(strings.TrimSpace(string(out)))
		if err != nil {
			return nil, fmt.Errorf("lstat %s: %w", strings.TrimSpace(string(out)), err)
		}

		if fileInfo.Mode()&os.ModeSymlink == 0 {
			filtered = append(filtered, v)
		}
	}
	return filtered, nil
}

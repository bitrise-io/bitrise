package asdf

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
)

type ProviderOptions struct {
	AsdfVersion string
}

type AsdfToolProvider struct {
	ExecEnv execenv.ExecEnv
}

func (a AsdfToolProvider) ID() string {
	return "asdf"
}

func (a AsdfToolProvider) Bootstrap() error {
	// TODO:
	// Check if asdf is installed
	// Check if asdf version satisfies the supported version range

	return nil
}

func (a AsdfToolProvider) InstallTool(tool toolprovider.ToolRequest) (toolprovider.ToolInstallResult, error) {
	err := a.InstallPlugin(tool)
	if err != nil {
		return toolprovider.ToolInstallResult{}, fmt.Errorf("install tool plugin %s: %w", tool.ToolName, err)
	}

	installedVersions, err := a.listInstalled(tool.ToolName)
	if err != nil {
		return toolprovider.ToolInstallResult{}, fmt.Errorf("list installed versions: %w", err)
	}

	// Short-circuit for exact version match among installed versions.
	// Fetching released versions is a slow operation that we want to avoid.
	v := strings.TrimSpace(tool.UnparsedVersion)
	if tool.ResolutionStrategy == toolprovider.ResolutionStrategyStrict && slices.Contains(installedVersions, v) {
		return toolprovider.ToolInstallResult{
			ToolName:           tool.ToolName,
			IsAlreadyInstalled: true,
			ConcreteVersion:    v,
		}, nil
	}

	releasedVersions, err := a.listReleased(tool.ToolName)
	if err != nil {
		return toolprovider.ToolInstallResult{}, fmt.Errorf("list released versions: %w", err)
	}

	if len(releasedVersions) == 0 && len(installedVersions) == 0 {
		return toolprovider.ToolInstallResult{}, &ErrNoMatchingVersion{
			RequestedVersion:  tool.UnparsedVersion,
			AvailableVersions: releasedVersions,
		}
	}

	resolution, err := ResolveVersion(tool, releasedVersions, installedVersions)
	if err != nil {
		var nomatchErr *ErrNoMatchingVersion
		if errors.As(err, &nomatchErr) {
			log.Warn("No matching version found, updating asdf-%s plugin and retrying...", tool.ToolName)
			// Some asdf plugins hardcode the list of installable versions and need a new plugin release to support new versions.
			_, err = a.ExecEnv.RunAsdf("plugin", "update", string(tool.ToolName))
			if err != nil {
				return toolprovider.ToolInstallResult{}, fmt.Errorf("update plugin: %w", err)
			}
			releasedVersions, err = a.listReleased(tool.ToolName)
			if err != nil {
				return toolprovider.ToolInstallResult{}, fmt.Errorf("list released versions after plugin update: %w", err)
			}
			resolution, err = ResolveVersion(tool, releasedVersions, installedVersions)
			if err != nil {
				if errors.As(err, &nomatchErr) {
					errorDetails := toolprovider.ToolInstallError{
						ToolName:         string(tool.ToolName),
						RequestedVersion: tool.UnparsedVersion,
						Cause:            nomatchErr.Error(),
						Recommendation:   fmt.Sprintf("You might want to use `%s:installed` or `%s:latest` to install the latest installed or latest released version of %s %s.", tool.UnparsedVersion, tool.UnparsedVersion, tool.ToolName, tool.UnparsedVersion),
					}
					return toolprovider.ToolInstallResult{}, errorDetails
				}
				return toolprovider.ToolInstallResult{}, fmt.Errorf("resolve version: %w", err)
			}
		}

		return toolprovider.ToolInstallResult{}, fmt.Errorf("resolve version: %w", err)
	}

	if resolution.IsInstalled {
		return toolprovider.ToolInstallResult{
			ToolName:           tool.ToolName,
			IsAlreadyInstalled: true,
			ConcreteVersion:    resolution.VersionString,
		}, nil
	} else {
		err = a.installToolVersion(tool.ToolName, resolution.VersionString)
		if err != nil {
			return toolprovider.ToolInstallResult{}, err
		}

		return toolprovider.ToolInstallResult{
			ToolName:           tool.ToolName,
			IsAlreadyInstalled: false,
			ConcreteVersion:    resolution.VersionString,
		}, nil
	}
}

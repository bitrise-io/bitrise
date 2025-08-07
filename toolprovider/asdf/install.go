package asdf

import (
	"fmt"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/workarounds"
)

func (a *AsdfToolProvider) installToolVersion(
	toolName provider.ToolID,
	versionString string,
) error {
	if toolName == "" || versionString == "" {
		return fmt.Errorf("toolName and versionString must not be empty")
	}

	out, err := a.ExecEnv.RunAsdf("install", string(toolName), versionString)
	if err != nil {
		return provider.ToolInstallError{
			ToolName:         toolName,
			RequestedVersion: versionString,
			Cause:            fmt.Sprintf("asdf install %s %s: %s", string(toolName), versionString, err),
			RawOutput:        out,
		}
	}

	if toolName == "nodejs" {
		err = workarounds.SetupCorepack(a.ExecEnv, versionString)
		if err != nil {
			return fmt.Errorf("setup corepack for %s %s: %w", string(toolName), versionString, err)
		}
	}
	return nil
}

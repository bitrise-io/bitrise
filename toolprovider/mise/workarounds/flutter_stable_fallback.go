package workarounds

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// Flutter versions in mise can be inconsistent with the -stable suffix.
// Some versions are available as "3.32.1" but not "3.32.1-stable", even though
// FVM uses the "@stable" notation (which we convert to "-stable").
// This workaround checks if a version without -stable exists when the original fails.

// ShouldTryStableFallback checks if we should retry a Flutter installation without the -stable suffix.
// Returns the fallback version to try, or empty string if no fallback should be attempted.
func ShouldTryStableFallback(
	execEnv execenv.ExecEnv,
	toolErr provider.ToolInstallError,
	silent bool,
) (string, error) {
	if toolErr.ToolName != "flutter" {
		return "", nil
	}

	if !strings.HasSuffix(toolErr.RequestedVersion, "-stable") {
		return "", nil
	}

	if !strings.Contains(toolErr.Cause, "no match for requested version") {
		return "", nil
	}

	fallbackVersion := strings.TrimSuffix(toolErr.RequestedVersion, "-stable")
	exists, err := versionExistsRemote(execEnv, provider.ToolID("flutter"), fallbackVersion)
	if err != nil {
		return "", fmt.Errorf("check if flutter %s exists remotely: %w", fallbackVersion, err)
	}

	if !exists {
		if !silent {
			log.Debugf("[WORKAROUND] Flutter %s does not exist remotely, no fallback available", fallbackVersion)
		}
		return "", nil
	}

	if !silent {
		log.Infof("Flutter %s not found, but %s exists - using fallback", toolErr.RequestedVersion, fallbackVersion)
	}

	return fallbackVersion, nil
}

// versionExistsRemote checks if a version exists in the remote registry using mise ls-remote.
func versionExistsRemote(execEnv execenv.ExecEnv, toolName provider.ToolID, version string) (bool, error) {
	versionString := string(toolName)
	if version != "" && version != "latest" {
		versionString = fmt.Sprintf("%s@%s", toolName, version)
	}

	output, err := execEnv.RunMiseWithTimeout(execenv.DefaultTimeout, "ls-remote", "--quiet", versionString)
	if err != nil {
		return false, fmt.Errorf("mise ls-remote %s: %w", versionString, err)
	}

	return strings.TrimSpace(string(output)) != "", nil
}

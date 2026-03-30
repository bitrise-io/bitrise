package workarounds

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// Flutter versions in mise can be inconsistent with the -stable suffix.
// Some versions are available as "3.32.1" but not "3.32.1-stable", even though
// FVM uses the "@stable" notation (which we convert to "-stable").
// This workaround proactively checks if a Flutter version with -stable suffix exists,
// and if not, attempts to use the version without the suffix.

// AdjustFlutterStableVersion proactively checks and adjusts Flutter versions with -stable suffix.
// If the requested version ends with -stable and doesn't exist remotely, but the version without
// -stable does exist, it returns the adjusted version (without -stable).
// Returns empty string if no adjustment is needed.
func AdjustFlutterStableVersion(
	versionExistsRemote func(provider.ToolID, string) (bool, error),
	toolName provider.ToolID,
	version string,
	silent bool,
) (string, error) {
	if toolName != "flutter" || !strings.HasSuffix(version, "-stable") {
		return "", nil
	}

	exists, err := versionExistsRemote(toolName, version)
	if err != nil {
		return "", fmt.Errorf("check if flutter %s exists remotely: %w", version, err)
	}
	if exists {
		return "", nil
	}

	fallbackVersion := strings.TrimSuffix(version, "-stable")
	fallbackExists, err := versionExistsRemote(toolName, fallbackVersion)
	if err != nil {
		return "", fmt.Errorf("check if flutter %s exists remotely: %w", fallbackVersion, err)
	}
	if !fallbackExists {
		if !silent {
			log.Debugf("Flutter %s not found (with or without -stable suffix)", version)
		}
		return "", nil
	}

	if !silent {
		log.Infof("Flutter %s not found, using %s instead", version, fallbackVersion)
	}

	return fallbackVersion, nil
}

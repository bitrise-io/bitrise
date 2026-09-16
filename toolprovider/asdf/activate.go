package asdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

func (a *AsdfToolProvider) ActivateEnv(result provider.ToolInstallResult) (provider.EnvironmentActivation, error) {
	envKey := fmt.Sprint("ASDF_", strings.ToUpper(string(result.ToolName)), "_VERSION")
	activation := provider.EnvironmentActivation{
		ContributedEnvVars: map[string]string{
			envKey: result.ConcreteVersion,
		},
	}

	// The version env var above is only honored by the asdf shims, so they have to come first in $PATH.
	// Other version managers (such as mise) also provide shims for the same tools and ignore this env var,
	// resolving the tool version on their own instead.
	shimsDir := a.shimsDir()
	if shimsDir != "" {
		if _, err := os.Stat(shimsDir); err == nil {
			activation.ContributedPaths = []string{shimsDir}
		}
	}

	return activation, nil
}

func (a *AsdfToolProvider) shimsDir() string {
	dataDir := a.ExecEnv.EnvVars["ASDF_DATA_DIR"]
	if dataDir == "" {
		dataDir = os.Getenv("ASDF_DATA_DIR")
	}
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dataDir = filepath.Join(home, ".asdf")
	}

	return filepath.Join(dataDir, "shims")
}

package workarounds

import (
	"fmt"

	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
)

// When installing a new Node.js version, the `corepack` executable is missing until we reshim the installed version.
// https://github.com/asdf-vm/asdf-nodejs/blob/90b8ecaa556916daba983a7b01869a9ea682f285/README.md#corepack
func SetupCorepack(execEnv execenv.ExecEnv, nodeVersion string) error {
	extraEnvs := map[string]string{
		// Simulate the activated environment
		"ASDF_NODEJS_VERSION": nodeVersion,
	}

	out, err := execEnv.RunCommand(extraEnvs, "corepack", "enable")
	if err != nil {
		return fmt.Errorf("enable corepack: %w\n\nOutput:\n%s", err, out)
	}

	out, err = execEnv.RunAsdf("reshim", "nodejs", nodeVersion)
	if err != nil {
		return fmt.Errorf("reshim nodejs after corepack setup: %w\n\nOutput:\n%s", err, out)
	}

	return nil
}

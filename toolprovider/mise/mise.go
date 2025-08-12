package mise

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// We pin one Mise version because:
// - Mise doesn't follow SemVer, there are breaking changes in regular releases sometimes
// - We depend on the exact layout of the release .tar.gz archive in Bootstrap(), this is probably not stable
const miseVersion = "v2025.8.7"

type MiseToolProvider struct {
	ExecEnv execenv.ExecEnv
}

func NewToolProvider(installDir string, dataDir string) (*MiseToolProvider, error) {
	if installDir == "" {
		return nil, errors.New("install directory must be provided")
	}
	if dataDir == "" {
		return nil, errors.New("data directory must be provided")
	}

	return &MiseToolProvider{
		ExecEnv: execenv.ExecEnv{
			InstallDir: installDir,

			// https://mise.jdx.dev/configuration.html#environment-variables
			ExtraEnvs: map[string]string{
				"MISE_DATA_DIR": dataDir,

				// Isolate this mise instance's "global" config from system-wide config
				"MISE_CONFIG_DIR":         filepath.Join(dataDir),
				"MISE_GLOBAL_CONFIG_FILE": filepath.Join(dataDir, "config.toml"),
				"MISE_GLOBAL_CONFIG_ROOT": dataDir,
			},
		},
	}, nil
}

func (m *MiseToolProvider) ID() string {
	return "mise"
}

func (m *MiseToolProvider) Bootstrap() error {
	fmt.Printf("Installing Mise %s...", miseVersion)
	fmt.Println()

	err := installReleaseBinary(miseVersion, m.ExecEnv.InstallDir)
	if err != nil {
		return fmt.Errorf("bootstrap mise: %w", err)
	}

	return nil
}

func (m *MiseToolProvider) InstallTool(tool provider.ToolRequest) (provider.ToolInstallResult, error) {
	isAlreadyInstalled, err := isAlreadyInstalled(tool, m.resolveToLatestInstalled)
	if err != nil {
		return provider.ToolInstallResult{}, err
	}

	versionExists, err := m.versionExists(tool.ToolName, tool.UnparsedVersion)
	if err != nil {
		return provider.ToolInstallResult{}, fmt.Errorf("check if version exists: %w", err)
	}
	if !versionExists {
		return provider.ToolInstallResult{}, provider.ToolInstallError{
			ToolName:         tool.ToolName,
			RequestedVersion: tool.UnparsedVersion,
			Cause:            fmt.Sprintf("no match for requested version %s", tool.UnparsedVersion),
		}
	}

	err = m.installToolVersion(tool)
	if err != nil {
		return provider.ToolInstallResult{}, err
	}

	concreteVersion, err := m.resolveToConcreteVersionAfterInstall(tool)
	if err != nil {
		return provider.ToolInstallResult{}, fmt.Errorf("resolve exact version after install: %w", err)
	}

	return provider.ToolInstallResult{
		ToolName:           tool.ToolName,
		IsAlreadyInstalled: isAlreadyInstalled,
		ConcreteVersion:    concreteVersion,
	}, nil
}

func (m *MiseToolProvider) ActivateEnv(result provider.ToolInstallResult) (provider.EnvironmentActivation, error) {
	envs, err := m.envVarsForTool(result)
	if err != nil {
		return provider.EnvironmentActivation{}, fmt.Errorf("get mise env: %w", err)
	}

	return processEnvOutput(envs), nil
}

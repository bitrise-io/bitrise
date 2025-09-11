package mise

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// We pin one Mise version because:
// - Mise doesn't follow SemVer, there are breaking changes in regular releases sometimes
// - We depend on the exact layout of the release .tar.gz archive in Bootstrap(), this is probably not stable

// UPDATE PROCESS:
// 1. Pick a new version, review changelog between the two releases
// 2. Download release artifacts: $ gh release download --repo jdx/mise v2025.8.7 --pattern 'mise-v*-*-*.tar.gz'
// 3. Verify checksums
// 4. Update version string and checksums below
// 5. IMPORTANT, DO NOT FORGET: Mirror artifacts to GCS bucket (see bootstrap.go) in case github.com goes down
const MiseVersion = "v2025.8.7"

var MiseChecksums = map[string]string{
	"linux-x64":   "c2d67d52880373931166343ef9a3b97665175ac2796dc95b9310179d341b2713",
	"linux-arm64": "d8dfa34d55762125e90b56ce8c9aaa037f7890fd00ac0c9cd8a097cc8530b126",
	"macos-x64":   "2b685b3507339f07d0da97b7dcf99354a3b14a16e8767af73057711e0ddce72f",
	"macos-arm64": "0b5893de7c8c274736867b7c4c7ed565b4429f4d6272521ace802f8a21422319",
}

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

	err := os.MkdirAll(installDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("create install dir at %s: %w", installDir, err)
	}

	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("create data dir at %s: %w", dataDir, err)
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
	if isMiseInstalled(m.ExecEnv.InstallDir) {
		return nil
	}

	err := installReleaseBinary(MiseVersion, MiseChecksums, m.ExecEnv.InstallDir)
	if err != nil {
		return fmt.Errorf("bootstrap mise: %w", err)
	}

	return nil
}

func (m *MiseToolProvider) InstallTool(tool provider.ToolRequest) (provider.ToolInstallResult, error) {
	err := m.InstallPlugin(tool)
	if err != nil {
		return provider.ToolInstallResult{}, fmt.Errorf("install tool plugin %s: %w", tool.ToolName, err)
	}

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

package mise

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
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
const misePreviewVersion = "v2025.10.8"

var misePreviewChecksums = map[string]string{
	"linux-x64":   "895db0eb777b90c449c4c79a36bd5f749fd614749876782ea32ede02c45e6bc2",
	"linux-arm64": "c949d574a46b68bf8d5834d099614818d6774935d908f53051f47d24ac0601c8",
	"macos-x64":   "422260046b8a24f0c72bfad60ac94837f834c83b5e7823e79f997ae7ff660de2",
	"macos-arm64": "bc7c40c48a43dfd80537e7ca5e55a2cf7dd37924bf7595d74b29848a6ab0e2ea",
}

const miseStableVersion = "v2025.10.8"

var miseStableChecksums = map[string]string{
	"linux-x64":   "895db0eb777b90c449c4c79a36bd5f749fd614749876782ea32ede02c45e6bc2",
	"linux-arm64": "c949d574a46b68bf8d5834d099614818d6774935d908f53051f47d24ac0601c8",
	"macos-x64":   "422260046b8a24f0c72bfad60ac94837f834c83b5e7823e79f997ae7ff660de2",
	"macos-arm64": "bc7c40c48a43dfd80537e7ca5e55a2cf7dd37924bf7595d74b29848a6ab0e2ea",
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

				// Enable corepack by default for Node.js installations. This mirrors the preinstalled Node versions on Bitrise stacks.
				// https://mise.jdx.dev/lang/node.html#environment-variables
				"MISE_NODE_COREPACK": "1",
			},
		},
	}, nil
}

func (m *MiseToolProvider) ID() string {
	return "mise"
}

func (m *MiseToolProvider) Bootstrap() error {
	if isMiseInstalled(m.ExecEnv.InstallDir) {
		log.Debugf("[TOOLPROVIDER] Mise already installed in %s, skipping bootstrap", m.ExecEnv.InstallDir)
		return nil
	}

	err := installReleaseBinary(GetMiseVersion(), GetMiseChecksums(), m.ExecEnv.InstallDir)
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

	activationResult := processEnvOutput(envs)
	// Some core plugins create shims to executables (e.g. npm). These shims call `mise reshim` and require the `mise` binary to be in $PATH.
	miseExecPath := filepath.Join(m.ExecEnv.InstallDir, "bin")
	activationResult.ContributedPaths = append(activationResult.ContributedPaths, miseExecPath)
	return activationResult, nil
}

func isEdgeStack() (isEdge bool) {
	if stack, variablePresent := os.LookupEnv("BITRISEIO_STACK_ID"); variablePresent && strings.Contains(stack, "edge") {
		isEdge = true
	} else {
		isEdge = false
	}
	log.Debugf("Mise: Stack is edge: %s", isEdge)
	return
}

func GetMiseVersion() string {
	if isEdgeStack() {
		return misePreviewVersion
	}
	// Fallback to stable version for non-edge stacks
	return miseStableVersion
}

func GetMiseChecksums() map[string]string {
	if isEdgeStack() {
		return misePreviewChecksums
	}
	// Fallback to stable version for non-edge stacks
	return miseStableChecksums
}

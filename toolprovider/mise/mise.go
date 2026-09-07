package mise

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/nixpkgs"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/workarounds"
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
const misePreviewVersion = "v2026.9.1"

var misePreviewChecksums = map[string]string{
	"linux-x64":   "063dda9149ab6be53da877c2d176afe0eac68e64cf8ca295bd0528720701c65d",
	"linux-arm64": "98d2ea7b82dd966afdb8a9f4e9edbca771acf2a30d2842bfc0efdb7b61c886a3",
	"macos-x64":   "eed76838c68aa49b7bf07c468dd4993855bbb342a4442f67355b6ffbe746e4d4",
	"macos-arm64": "bfea0ab417b48c1e8b99412fcaf20ce17424a3286a8766d7d2b0051fe321d565",
}

const miseStableVersion = "v2026.5.12"

var miseStableChecksums = map[string]string{
	"linux-x64":   "bd0930c0b619f51ddb60e32e5cce18a5533567b2f1ba9fc4875b9f39a2bb3ed8",
	"linux-arm64": "67c2bd96da9c6da030db4174b2dd0f8e6636c25519b23a15f0b734556e6e5ee0",
	"macos-x64":   "dcab53de40bbd42c10607d64081e9df328c4885db30b41c4421f27e18b8f7efa",
	"macos-arm64": "5b883c868a0748dd0c595d30fd000ec5138dfabdeef2c30222866ebf34af1ae3",
}

var KnownGitHubTokenEnvVars = []string{
	"GITHUB_TOKEN",
	"GH_TOKEN",
	"GITHUB_API_TOKEN",
	"MISE_GITHUB_TOKEN",
	"MISE_GITHUB_ENTERPRISE_TOKEN",
}

type MiseToolProvider struct {
	ExecEnv        execenv.ExecEnv
	UseFastInstall bool
	Silent         bool
}

func NewToolProvider(installDir string, dataDir string, useFastInstall, silent bool, extraEnvs map[string]string) (*MiseToolProvider, error) {
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

	// https://mise.jdx.dev/configuration.html#environment-variables
	miseEnvs := map[string]string{
		"MISE_DATA_DIR": dataDir,

		// Isolate this mise instance's "global" config from system-wide config.
		"MISE_CONFIG_DIR":         filepath.Join(dataDir),
		"MISE_GLOBAL_CONFIG_FILE": filepath.Join(dataDir, "config.toml"),
		"MISE_GLOBAL_CONFIG_ROOT": dataDir,

		// Enable corepack by default for Node.js installations. This mirrors the preinstalled Node versions on Bitrise stacks.
		// https://mise.jdx.dev/lang/node.html#environment-variables
		"MISE_NODE_COREPACK": "1",

		// Older Python versions predate attestation support in astral-sh/python-build-standalone,
		// so verification fails for any such version (e.g. 3.11.4).
		"MISE_PYTHON_GITHUB_ATTESTATIONS": "false",
	}
	maps.Copy(miseEnvs, extraEnvs)

	return &MiseToolProvider{
		ExecEnv:        execenv.NewMiseExecEnv(installDir, miseEnvs),
		UseFastInstall: useFastInstall,
		Silent:         silent,
	}, nil
}

func (m *MiseToolProvider) ID() string {
	return "mise"
}

func (m *MiseToolProvider) Bootstrap() error {
	installDir := m.ExecEnv.InstallDir()
	if isMiseInstalled(installDir) {
		if !m.Silent {
			log.Debugf("[TOOLPROVIDER] Mise already installed in %s, skipping bootstrap", installDir)
		}
		return nil
	}

	err := installReleaseBinary(GetMiseVersion(), GetMiseChecksums(), installDir)
	if err != nil {
		return fmt.Errorf("bootstrap mise: %w", err)
	}

	return nil
}

func (m *MiseToolProvider) InstallTool(tool provider.ToolRequest) (provider.ToolInstallResult, error) {
	useNix := canBeInstalledWithNix(tool, m.ExecEnv, m.UseFastInstall, nixpkgs.ShouldUseBackend, m.Silent)
	if !useNix {
		err := m.InstallPlugin(tool)
		if err != nil {
			return provider.ToolInstallResult{}, fmt.Errorf("install tool plugin %s: %w", tool.ToolName, err)
		}
	} // else: nixpkgs plugin is already installed in canBeInstalledWithNix()

	installRequest := installRequest(tool, useNix)

	normalizedRequest, err := normalizeRequest(m.ExecEnv, installRequest, m.Silent)
	if err != nil {
		return provider.ToolInstallResult{}, err
	}

	// Flutter stable suffix workaround.
	adjustedVersion, err := workarounds.AdjustFlutterStableVersion(
		func(toolName provider.ToolID, version string) (bool, error) {
			return versionExistsRemote(m.ExecEnv, toolName, version)
		},
		normalizedRequest.ToolName,
		normalizedRequest.UnparsedVersion,
		m.Silent,
	)
	if err != nil {
		return provider.ToolInstallResult{}, fmt.Errorf("adjust flutter version: %w", err)
	}
	if adjustedVersion != "" {
		normalizedRequest.UnparsedVersion = adjustedVersion
	}

	concreteVersion, err := resolveToConcreteVersion(
		m.ExecEnv,
		normalizedRequest.ToolName,
		normalizedRequest.UnparsedVersion,
		normalizedRequest.ResolutionStrategy,
	)
	if err != nil {
		if errors.Is(err, errNoMatchingVersion) {
			return provider.ToolInstallResult{}, provider.ToolInstallError{
				ToolName:         installRequest.ToolName,
				RequestedVersion: installRequest.UnparsedVersion,
				Cause:            fmt.Sprintf("no match for requested version %s", installRequest.UnparsedVersion),
			}
		}
		return provider.ToolInstallResult{}, fmt.Errorf("resolve %s@%s: %w", installRequest.ToolName, installRequest.UnparsedVersion, err)
	}
	if !m.Silent {
		log.Debugf("[TOOLPROVIDER] Resolved %s@%s to concrete version: %s",
			installRequest.ToolName, installRequest.UnparsedVersion, concreteVersion)
	}
	isAlreadyInstalled, err := m.isAlreadyInstalled(installRequest.ToolName, concreteVersion)
	if err != nil {
		return provider.ToolInstallResult{}, err
	}

	if !isAlreadyInstalled {
		err = m.installToolVersion(installRequest.ToolName, concreteVersion)
		if err != nil {
			return provider.ToolInstallResult{}, err
		}
	}

	installedVersion, err := resolveToLatestInstalled(m.ExecEnv, installRequest.ToolName, concreteVersion)
	if err != nil || installedVersion != concreteVersion {
		return provider.ToolInstallResult{}, fmt.Errorf(
			"install verification failed: expected %s, got %s", concreteVersion, installedVersion)
	}

	return provider.ToolInstallResult{
		// Note: we return installRequest.ToolName instead of the original tool.ToolName.
		// This is because installRequest might use a custom backend plugin and the value returned here
		// is what gets used in ActivateEnv(), the two should be consistent.
		ToolName:           installRequest.ToolName,
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
	miseExecPath := filepath.Join(m.ExecEnv.InstallDir(), "bin")
	activationResult.ContributedPaths = append(activationResult.ContributedPaths, miseExecPath)
	return activationResult, nil
}

// ResolveLatestVersion resolves a tool to its latest version without installing it.
func (m *MiseToolProvider) ResolveLatestVersion(tool provider.ToolRequest) (string, error) {
	useNix := canBeInstalledWithNix(tool, m.ExecEnv, m.UseFastInstall, nixpkgs.ShouldUseBackend, m.Silent)
	if !useNix {
		err := m.InstallPlugin(tool)
		if err != nil {
			return "", fmt.Errorf("install tool plugin %s: %w", tool.ToolName, err)
		}
	}

	installRequest := installRequest(tool, useNix)

	normalizedRequest, err := normalizeRequest(m.ExecEnv, installRequest, m.Silent)
	if err != nil {
		return "", err
	}

	concreteVersion, err := resolveToConcreteVersion(
		m.ExecEnv,
		normalizedRequest.ToolName,
		normalizedRequest.UnparsedVersion,
		normalizedRequest.ResolutionStrategy,
	)
	if err != nil {
		if errors.Is(err, errNoMatchingVersion) {
			return "", provider.ToolInstallError{
				ToolName:         tool.ToolName,
				RequestedVersion: tool.UnparsedVersion,
				Cause:            fmt.Sprintf("no match for requested version %s", tool.UnparsedVersion),
			}
		}
		return "", fmt.Errorf("resolve %s@%s: %w", tool.ToolName, tool.UnparsedVersion, err)
	}

	return concreteVersion, nil
}

func (m *MiseToolProvider) ListReleasedVersions(toolName provider.ToolID) ([]string, error) {
	err := m.InstallPlugin(provider.ToolRequest{ToolName: toolName})
	if err != nil {
		return nil, fmt.Errorf("install tool plugin %s: %w", toolName, err)
	}
	return listRemoteVersions(m.ExecEnv, toolName)
}

func GetMiseVersion() string {
	isEdge := configs.IsEdgeStack()
	if isEdge {
		return misePreviewVersion
	}
	// Fallback to stable version for non-edge stacks
	return miseStableVersion
}

func GetMiseChecksums() map[string]string {
	if configs.IsEdgeStack() {
		return misePreviewChecksums
	}
	// Fallback to stable version for non-edge stacks
	return miseStableChecksums
}

package mise

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/nixpkgs"
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
const misePreviewVersion = "v2025.12.1"

var misePreviewChecksums = map[string]string{
	"linux-x64":   "0e62b1a0a8b87329d0cf24fc6af5d1c3aae0819194bea2f43fcf3f556edc9c29",
	"linux-arm64": "35573ccc8f13895884b8e7a3365736c2942ad531ce24fc420ba0a941dbb57ce5",
	"macos-x64":   "68b250632b1f1f29f6116ca513d1641097dfdc2cf05520ee0ca23907962b3d6f",
	"macos-arm64": "94659ac9b7b30d149464ef4a76498182b0c5cadeccef1811ab9e75ff3d1ad159",
}

const miseStableVersion = "v2025.10.8"

var miseStableChecksums = map[string]string{
	"linux-x64":   "895db0eb777b90c449c4c79a36bd5f749fd614749876782ea32ede02c45e6bc2",
	"linux-arm64": "c949d574a46b68bf8d5834d099614818d6774935d908f53051f47d24ac0601c8",
	"macos-x64":   "422260046b8a24f0c72bfad60ac94837f834c83b5e7823e79f997ae7ff660de2",
	"macos-arm64": "bc7c40c48a43dfd80537e7ca5e55a2cf7dd37924bf7595d74b29848a6ab0e2ea",
}

type MiseToolProvider struct {
	ExecEnv        execenv.ExecEnv
	UseFastInstall bool
}

func NewToolProvider(installDir string, dataDir string, useFastInstall bool) (*MiseToolProvider, error) {
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
		ExecEnv: execenv.NewMiseExecEnv(installDir, map[string]string{
			// https://mise.jdx.dev/configuration.html#environment-variables
			"MISE_DATA_DIR": dataDir,

			// Isolate this mise instance's "global" config from system-wide config.
			"MISE_CONFIG_DIR":         filepath.Join(dataDir),
			"MISE_GLOBAL_CONFIG_FILE": filepath.Join(dataDir, "config.toml"),
			"MISE_GLOBAL_CONFIG_ROOT": dataDir,

			// Enable corepack by default for Node.js installations. This mirrors the preinstalled Node versions on Bitrise stacks.
			// https://mise.jdx.dev/lang/node.html#environment-variables
			"MISE_NODE_COREPACK": "1",
		},
		),
		UseFastInstall: useFastInstall,
	}, nil
}

func (m *MiseToolProvider) ID() string {
	return "mise"
}

func (m *MiseToolProvider) Bootstrap() error {
	installDir := m.ExecEnv.InstallDir()
	if isMiseInstalled(installDir) {
		log.Debugf("[TOOLPROVIDER] Mise already installed in %s, skipping bootstrap", installDir)
		return nil
	}

	err := installReleaseBinary(GetMiseVersion(), GetMiseChecksums(), installDir)
	if err != nil {
		return fmt.Errorf("bootstrap mise: %w", err)
	}

	return nil
}

func (m *MiseToolProvider) InstallTool(tool provider.ToolRequest) (provider.ToolInstallResult, error) {
	// TODO: disable Nix-based install on Linux until we solve the dynamic linking issues
	useNix := runtime.GOOS == "darwin" && canBeInstalledWithNix(tool, m.ExecEnv, m.UseFastInstall, nixpkgs.ShouldUseBackend)
	if !useNix {
		err := m.InstallPlugin(tool)
		if err != nil {
			return provider.ToolInstallResult{}, fmt.Errorf("install tool plugin %s: %w", tool.ToolName, err)
		}
	} // else: nixpkgs plugin is already installed in canBeInstalledWithNix()

	installRequest := installRequest(tool, useNix)

	normalizedRequest, err := normalizeRequest(m.ExecEnv, installRequest)
	if err != nil {
		return provider.ToolInstallResult{}, err
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
	log.Debugf("[TOOLPROVIDER] Resolved %s@%s to concrete version: %s",
		installRequest.ToolName, installRequest.UnparsedVersion, concreteVersion)

	if !useNix {
		versionExists, err := versionExistsRemote(m.ExecEnv, installRequest.ToolName, concreteVersion)
		if err != nil {
			return provider.ToolInstallResult{}, fmt.Errorf("check if version exists for %s@%s: %w", installRequest.ToolName, concreteVersion, err)
		}
		if !versionExists {
			return provider.ToolInstallResult{}, provider.ToolInstallError{
				ToolName:         installRequest.ToolName,
				RequestedVersion: installRequest.UnparsedVersion,
				Cause:            fmt.Sprintf("no match for requested version %s", installRequest.UnparsedVersion),
			}
		}
	} // else: canBeInstalledWithNix() already verified version existence

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
// If checkInstalled is true, returns the latest installed version. Otherwise, returns the latest released version.
func (m *MiseToolProvider) ResolveLatestVersion(tool provider.ToolRequest, checkInstalled bool) (string, error) {
	// TODO: disable Nix-based install on Linux until we solve the dynamic linking issues
	useNix := runtime.GOOS == "darwin" && canBeInstalledWithNix(tool, m.ExecEnv, m.UseFastInstall, nixpkgs.ShouldUseBackend)
	if !useNix {
		err := m.InstallPlugin(tool)
		if err != nil {
			return "", fmt.Errorf("install tool plugin %s: %w", tool.ToolName, err)
		}
	}

	installRequest := installRequest(tool, useNix)

	if checkInstalled {
		installRequest.ResolutionStrategy = provider.ResolutionStrategyLatestInstalled
	} else {
		installRequest.ResolutionStrategy = provider.ResolutionStrategyLatestReleased
	}

	normalizedRequest, err := normalizeRequest(m.ExecEnv, installRequest)
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

func GetMiseVersion() string {
	isEdge := configs.IsEdgeStack()
	log.Debugf("[TOOLPROVIDER] Stack is edge: %t", isEdge)
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

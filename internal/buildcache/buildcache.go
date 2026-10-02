// Package buildcache installs the bitrise-build-cache CLI once and runs the build-start activations with the build's envs, which plugins never see.
// Which tools to activate, and whether the workspace is enabled at all, are the CLI's decisions.
package buildcache

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v3/analytics"
	"github.com/bitrise-io/bitrise/v3/configs"
	"github.com/bitrise-io/bitrise/v3/log"
	"github.com/bitrise-io/bitrise/v3/log/logwriter"
	envmanModels "github.com/bitrise-io/envman/v2/models"
)

const (
	// EnvActivateAll opts a build in; the DEN agent passes BITRISE_* envs through, so preboot can set it.
	EnvActivateAll = "BITRISE_BUILD_CACHE_ACTIVATE_ALL"
	// EnvActivateGradleMirrors opts a build in to the Gradle repository mirrors, independently of the activation above.
	EnvActivateGradleMirrors = "BITRISE_BUILD_CACHE_ACTIVATE_GRADLE_MIRRORS"
	// EnvHostCacheURL serves release tarballs by bare name, e.g. http://192.168.64.1:59020/build-cache-cli-releases.
	EnvHostCacheURL = "BITRISE_BUILD_CACHE_CLI_HOST_CACHE_URL"
	// EnvCLIVersion and EnvCLISHA256 carry the VM setup's own pin, so there is no second one to bump here.
	EnvCLIVersion = "BITRISE_BUILD_CACHE_CLI_VERSION"
	EnvCLISHA256  = "BITRISE_BUILD_CACHE_CLI_SHA256"

	servicesTokenKey      = "BITRISEIO_BITRISE_SERVICES_ACCESS_TOKEN"
	buildHubVMTokenKey    = "BITRISEIO_BUILD_HUB_VM_TOKEN"
	buildHubVMTokenURLKey = "BITRISEIO_BUILD_HUB_VM_TOKEN_URL"
	buildCacheEnvPrefix   = "BITRISE_BUILD_CACHE_"

	activationTimeout = 2 * time.Minute

	envDisableHostsOverride = "BITRISE_DEN_DISABLE_HOSTS_OVERRIDE"
	envMavenCentralProxy    = "BITRISE_MAVENCENTRAL_PROXY_ENABLED"
)

// The version ends up in a path and a URL.
var (
	versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+[0-9A-Za-z.+-]*$`)
	sha256Pattern  = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// ActivateIfEnabled reports whether it ran; it never fails the build, problems are logged as warnings.
func ActivateIfEnabled(logger log.Logger, buildEnvs []envmanModels.EnvironmentItemModel) bool {
	// A nested `bitrise run` inherits the step execution ID; the outer run already activated.
	if os.Getenv(analytics.StepExecutionIDEnvKey) != "" {
		return false
	}

	all := enabled(buildEnvs, EnvActivateAll)
	mirrors := enabled(buildEnvs, EnvActivateGradleMirrors)
	if mirrors && os.Getenv(envDisableHostsOverride) == "true" {
		logger.Infof("Skipping the Gradle mirrors because %s is true", envDisableHostsOverride)
		mirrors = false
	}
	if !all && !mirrors {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), activationTimeout)
	defer cancel()

	logger.Infof("Activating Bitrise Build Cache")
	bin, err := installCLI(ctx, logger)
	if err != nil {
		logger.Warnf("Bitrise Build Cache activation failed, continuing without it: %s", err)

		return true
	}

	if mirrors {
		// The mirrors' init script re-reads this at build time, so a build can still switch them off.
		env := append(os.Environ(), envMavenCentralProxy+"=true")
		if err := run(ctx, logger, bin, env, "activate", "gradle-mirrors", "-d"); err != nil {
			logger.Warnf("activating bitrise-build-cache gradle-mirrors failed: %s", err)
		}
	}
	if all {
		// exec keeps the last value of a duplicated key, so build envs win.
		env := append(os.Environ(), cacheEnvs(buildEnvs)...)
		if err := run(ctx, logger, bin, env, "activate", "all", "--auto"); err != nil {
			logger.Warnf("Bitrise Build Cache activation failed, continuing without it: activate all: %s", err)
		}
	}

	return true
}

func enabled(buildEnvs []envmanModels.EnvironmentItemModel, name string) bool {
	if os.Getenv(name) == "true" {
		return true
	}
	for _, env := range buildEnvs {
		if key, value, err := env.GetKeyValuePair(); err == nil && key == name && value == "true" {
			return true
		}
	}

	return false
}

// installCLI puts the pinned CLI on the tools PATH and returns the linked binary.
func installCLI(ctx context.Context, logger log.Logger) (string, error) {
	version, checksum := os.Getenv(EnvCLIVersion), os.Getenv(EnvCLISHA256)
	if !versionPattern.MatchString(version) || !sha256Pattern.MatchString(checksum) {
		return "", fmt.Errorf("%s (%q) and %s (%q) must be set by the VM setup", EnvCLIVersion, version, EnvCLISHA256, checksum)
	}

	installDir := filepath.Join(configs.GetBitriseHomeDirPath(), "build-cache", version)
	versioned, err := newInstaller(logger, os.Getenv(EnvHostCacheURL), version, checksum).install(ctx, installDir)
	if err != nil {
		return "", fmt.Errorf("install bitrise-build-cache %s: %w", version, err)
	}
	// Generated configs name the CLI bare when it is on PATH, which survives upgrades.
	bin := filepath.Join(configs.GetBitriseToolsDirPath(), binaryName)
	if err := link(versioned, bin); err != nil {
		return "", fmt.Errorf("link bitrise-build-cache onto PATH: %w", err)
	}

	return bin, nil
}

// cacheEnvs keeps the build's other secrets out of the CLI's environment.
func cacheEnvs(buildEnvs []envmanModels.EnvironmentItemModel) []string {
	var envs []string
	for _, env := range buildEnvs {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			continue
		}
		if isCacheCredential(key) || strings.HasPrefix(key, buildCacheEnvPrefix) {
			envs = append(envs, key+"="+value)
		}
	}
	return envs
}

func isCacheCredential(key string) bool {
	return key == servicesTokenKey || key == buildHubVMTokenKey || key == buildHubVMTokenURLKey
}

func run(ctx context.Context, logger log.Logger, bin string, env []string, args ...string) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = env
	writer := logwriter.NewLogWriter(logger)
	cmd.Stdout = writer
	cmd.Stderr = writer
	err := cmd.Run()
	if closeErr := writer.Close(); closeErr != nil {
		logger.Warnf("Failed to flush bitrise-build-cache output: %s", closeErr)
	}
	return err
}

// Package buildcache installs the bitrise-build-cache CLI and activates it with the build's envs, which plugins never see.
package buildcache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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
	// EnvHostCacheURL serves release tarballs by bare name, e.g. http://192.168.64.1:59020/build-cache-cli-releases.
	EnvHostCacheURL = "BITRISE_BUILD_CACHE_CLI_HOST_CACHE_URL"
	// EnvCLIVersion and EnvCLISHA256 carry the VM setup's own pin, so there is no second one to bump here.
	EnvCLIVersion = "BITRISE_BUILD_CACHE_CLI_VERSION"
	EnvCLISHA256  = "BITRISE_BUILD_CACHE_CLI_SHA256"

	servicesTokenKey    = "BITRISEIO_BITRISE_SERVICES_ACCESS_TOKEN"
	buildCacheEnvPrefix = "BITRISE_BUILD_CACHE_"

	activationTimeout = 2 * time.Minute
)

// The version ends up in a path and a URL.
var (
	versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+[0-9A-Za-z.+-]*$`)
	sha256Pattern  = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// ActivateIfEnabled reports whether it ran; it never fails the build, problems are logged as warnings.
func ActivateIfEnabled(logger log.Logger, buildEnvs []envmanModels.EnvironmentItemModel) bool {
	// A nested `bitrise run` inherits the step execution ID; the outer run already activated.
	if os.Getenv(analytics.StepExecutionIDEnvKey) != "" || !enabled(buildEnvs) {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), activationTimeout)
	defer cancel()

	logger.Infof("Activating Bitrise Build Cache")
	if err := activateAll(ctx, logger, buildEnvs); err != nil {
		logger.Warnf("Bitrise Build Cache activation failed, continuing without it: %s", err)
	}
	return true
}

func enabled(buildEnvs []envmanModels.EnvironmentItemModel) bool {
	if os.Getenv(EnvActivateAll) == "true" {
		return true
	}
	for _, env := range buildEnvs {
		if key, value, err := env.GetKeyValuePair(); err == nil && key == EnvActivateAll && value == "true" {
			return true
		}
	}
	return false
}

func activateAll(ctx context.Context, logger log.Logger, buildEnvs []envmanModels.EnvironmentItemModel) error {
	version, checksum := os.Getenv(EnvCLIVersion), os.Getenv(EnvCLISHA256)
	if !versionPattern.MatchString(version) || !sha256Pattern.MatchString(checksum) {
		return fmt.Errorf("%s (%q) and %s (%q) must be set by the VM setup", EnvCLIVersion, version, EnvCLISHA256, checksum)
	}

	installDir := filepath.Join(configs.GetBitriseHomeDirPath(), "build-cache", version)
	versioned, err := newInstaller(logger, os.Getenv(EnvHostCacheURL), version, checksum).install(ctx, installDir)
	if err != nil {
		return fmt.Errorf("install bitrise-build-cache %s: %w", version, err)
	}
	// Generated configs name the CLI bare when it is on PATH, which survives upgrades.
	bin := filepath.Join(configs.GetBitriseToolsDirPath(), binaryName)
	if err := link(versioned, bin); err != nil {
		return fmt.Errorf("link bitrise-build-cache onto PATH: %w", err)
	}

	// exec keeps the last value of a duplicated key, so build envs win.
	env := append(os.Environ(), cacheEnvs(buildEnvs)...)

	var errs []error
	for _, tool := range tools(runtime.GOOS) {
		if err := run(ctx, logger, bin, env, "activate", tool); err != nil {
			errs = append(errs, fmt.Errorf("activate %s: %w", tool, err))
		}
	}
	return errors.Join(errs...)
}

// cacheEnvs keeps the build's other secrets out of the CLI's environment.
func cacheEnvs(buildEnvs []envmanModels.EnvironmentItemModel) []string {
	var envs []string
	for _, env := range buildEnvs {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			continue
		}
		if key == servicesTokenKey || strings.HasPrefix(key, buildCacheEnvPrefix) {
			envs = append(envs, key+"="+value)
		}
	}
	return envs
}

// tools stands in for `activate all` until the bitrise-build-cache CLI ships it.
func tools(goos string) []string {
	tools := []string{"gradle", "bazel"}
	if goos == "darwin" {
		tools = append(tools, "xcode")
	}
	return append(tools, "react-native")
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

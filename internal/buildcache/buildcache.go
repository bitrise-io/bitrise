// Package buildcache installs the bitrise-build-cache CLI once and runs the build-start activations with the build's envs, which plugins never see.
// Which tools to activate, and whether the workspace is enabled at all, are the CLI's decisions.
package buildcache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v3/analytics"
	"github.com/bitrise-io/bitrise/v3/configs"
	"github.com/bitrise-io/bitrise/v3/log"
	"github.com/bitrise-io/bitrise/v3/log/logwriter"
	"github.com/bitrise-io/bitrise/v3/tools"
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
	gradleUserHomeKey     = "GRADLE_USER_HOME"

	activationTimeout = 3 * time.Minute

	envDisableHostsOverride = "BITRISE_DEN_DISABLE_HOSTS_OVERRIDE"
	envMavenCentralProxy    = "BITRISE_MAVENCENTRAL_PROXY_ENABLED"
)

var (
	commandTimeout = 45 * time.Second
	// Bounds how long a descendant that keeps the output pipe open can block a finished command.
	commandWaitDelay = 5 * time.Second
)

// The version ends up in a path and a URL.
var (
	versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+[0-9A-Za-z.+-]*$`)
	sha256Pattern  = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// ActivateIfEnabled reports whether activation was attempted; it never fails the build, problems are logged as warnings.
func ActivateIfEnabled(logger log.Logger, buildEnvs []envmanModels.EnvironmentItemModel) bool {
	// A nested `bitrise run` inherits the step execution ID; the outer run already activated.
	if os.Getenv(analytics.StepExecutionIDEnvKey) != "" {
		return false
	}

	values, err := evaluate(buildEnvs)
	if err != nil {
		// Only a build that opted in at the VM level has reason to see this.
		if os.Getenv(EnvActivateAll) == "true" || os.Getenv(EnvActivateGradleMirrors) == "true" {
			logger.Warnf("Bitrise Build Cache activation skipped, could not evaluate the build envs: %s", err)
		}

		return false
	}

	all := enabled(values, EnvActivateAll)
	mirrors := enabled(values, EnvActivateGradleMirrors)
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
		if home, ok := values[gradleUserHomeKey]; ok {
			env = append(env, gradleUserHomeKey+"="+home)
		}
		if err := run(ctx, logger, bin, env, "activate", "gradle-mirrors", "-d"); err != nil {
			logger.Warnf("activating bitrise-build-cache gradle-mirrors failed: %s", err)
		}
	}
	if all {
		// exec keeps the last value of a duplicated key, so build envs win.
		env := append(os.Environ(), cacheEnvs(values)...)
		if err := run(ctx, logger, bin, env, "activate", "all", "--auto"); err != nil {
			logger.Warnf("Bitrise Build Cache activation failed, continuing without it: activate all: %s", err)
		}
	}

	return true
}

// enabled lets the evaluated build envs win (the last declaration, as a step would see it), then the process env.
func enabled(values map[string]string, name string) bool {
	if value, ok := values[name]; ok {
		return value == "true"
	}

	return os.Getenv(name) == "true"
}

// evaluate resolves the ordered build envs the way a step sees them: expansion per is_expand, skip_if_empty, and unset.
func evaluate(buildEnvs []envmanModels.EnvironmentItemModel) (map[string]string, error) {
	// Defaults (is_expand is true) go on copies, so the caller's items stay as declared.
	items := make([]envmanModels.EnvironmentItemModel, 0, len(buildEnvs))
	for _, env := range buildEnvs {
		item := envmanModels.EnvironmentItemModel{}
		for key, value := range env {
			item[key] = value
		}
		if err := item.FillMissingDefaults(); err != nil {
			return nil, fmt.Errorf("apply the env defaults: %w", err)
		}
		if err := blankIfUnset(item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	values, err := tools.ExpandEnvItems(items, os.Environ())
	if err != nil {
		return nil, fmt.Errorf("expand the build envs: %w", err)
	}

	return values, nil
}

// blankIfUnset turns an unset declaration into an empty value that is not skipped, so the helper's ordered
// expansion sees the variable as gone for every later declaration, as a step would.
func blankIfUnset(item envmanModels.EnvironmentItemModel) error {
	key, _, err := item.GetKeyValuePair()
	if err != nil {
		return fmt.Errorf("read a build env: %w", err)
	}
	opts, err := item.GetOptions()
	if err != nil {
		return fmt.Errorf("read the options of %s: %w", key, err)
	}
	if opts.Unset == nil || !*opts.Unset {
		return nil
	}

	keep := false
	opts.SkipIfEmpty = &keep
	item[key] = ""
	item[envmanModels.OptionsKey] = opts

	return nil
}

// installCLI puts the pinned CLI on the tools PATH and returns its versioned path, which the activations run from,
// so another run relinking the shared name cannot swap the binary that was verified.
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

	return versioned, nil
}

// cacheEnvs selects the evaluated build envs the CLI needs; the rest of the build envs are not added to its environment.
// GRADLE_USER_HOME decides where the CLI writes the Gradle init scripts.
func cacheEnvs(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		if isCacheCredential(key) || key == gradleUserHomeKey || strings.HasPrefix(key, buildCacheEnvPrefix) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	envs := make([]string, 0, len(keys))
	for _, key := range keys {
		envs = append(envs, key+"="+values[key])
	}

	return envs
}

func isCacheCredential(key string) bool {
	return key == servicesTokenKey || key == buildHubVMTokenKey || key == buildHubVMTokenURLKey
}

func run(ctx context.Context, logger log.Logger, bin string, env []string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = env
	cmd.WaitDelay = commandWaitDelay
	writer := logwriter.NewLogWriter(logger)
	cmd.Stdout = writer
	cmd.Stderr = writer
	err := cmd.Run()
	// Run reports this only for a command that exited cleanly but left a daemon on the output pipe.
	if errors.Is(err, exec.ErrWaitDelay) {
		err = nil
	}
	if closeErr := writer.Close(); closeErr != nil {
		logger.Warnf("Failed to flush bitrise-build-cache output: %s", closeErr)
	}
	return err
}

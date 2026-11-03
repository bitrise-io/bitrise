// Package buildcache puts the bitrise-build-cache CLI the VM setup installed on the tools PATH and runs the build-start activations with the build's envs, which plugins never see.
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
	envmanModels "github.com/bitrise-io/bitrise/v3/envman/models"
	"github.com/bitrise-io/bitrise/v3/log"
	"github.com/bitrise-io/bitrise/v3/log/logwriter"
	"github.com/bitrise-io/bitrise/v3/tools"
)

const (
	// EnvActivateAll opts a build in; the DEN agent passes BITRISE_* envs through, so preboot can set it.
	EnvActivateAll = "BITRISE_BUILD_CACHE_ACTIVATE_ALL"
	// EnvCLIPath points at the CLI the VM setup installed; the hook never downloads one itself.
	EnvCLIPath = "BITRISE_BUILD_CACHE_CLI_PATH"

	servicesTokenKey      = "BITRISEIO_BITRISE_SERVICES_ACCESS_TOKEN"
	buildHubVMTokenKey    = "BITRISEIO_BUILD_HUB_VM_TOKEN"
	buildHubVMTokenURLKey = "BITRISEIO_BUILD_HUB_VM_TOKEN_URL"
	buildCacheEnvPrefix   = "BITRISE_BUILD_CACHE_"
	gradleUserHomeKey     = "GRADLE_USER_HOME"

	activationTimeout = 3 * time.Minute
)

var (
	commandTimeout = 45 * time.Second
	versionTimeout = 5 * time.Second
	// Bounds how long a descendant that keeps the output pipe open can block a finished command.
	commandWaitDelay = 5 * time.Second
)

// The reported version ends up in a log line, so anything else a binary prints is rejected.
var versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+[0-9A-Za-z.+-]*$`)

// ActivateIfEnabled reports whether activation was attempted; it never fails the build, problems are logged as warnings.
func ActivateIfEnabled(logger log.Logger, buildEnvs []envmanModels.EnvironmentItemModel) bool {
	// A nested `bitrise run` inherits the step execution ID; the outer run already activated.
	if os.Getenv(analytics.StepExecutionIDEnvKey) != "" {
		return false
	}

	values, err := evaluate(buildEnvs)
	if err != nil {
		// Only a build that opted in at the VM level has reason to see this.
		if os.Getenv(EnvActivateAll) == "true" {
			logger.Warnf("Bitrise Build Cache activation skipped, could not evaluate the build envs: %s", err)
		}

		return false
	}

	if !enabled(values, EnvActivateAll) {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), activationTimeout)
	defer cancel()

	logger.Infof("Activating Bitrise Build Cache")
	bin, err := resolveCLI(ctx, logger)
	if err != nil {
		logger.Warnf("Bitrise Build Cache activation failed, continuing without it: %s", err)

		return true
	}

	// exec keeps the last value of a duplicated key, so build envs win.
	env := append(os.Environ(), cacheEnvs(values)...)
	if err := run(ctx, logger, bin, env, "activate", "all", "--auto"); err != nil {
		logger.Warnf("Bitrise Build Cache activation failed, continuing without it: activate all: %s", err)
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

// resolveCLI puts the CLI the VM setup installed on the tools PATH and returns the path the activations run from.
// The VM setup owns the version, so any CLI that runs and reports one is used; a missing or broken one means no activation.
// It returns a concrete file, not the shared link, so relinking cannot swap the binary that was verified.
func resolveCLI(ctx context.Context, logger log.Logger) (string, error) {
	path := os.Getenv(EnvCLIPath)
	if path == "" {
		return "", fmt.Errorf("%s is not set, the VM setup has to install the CLI", EnvCLIPath)
	}

	bin, version, err := usePreinstalled(ctx, path)
	if err != nil {
		return "", fmt.Errorf("%s=%s is not usable: %w", EnvCLIPath, path, err)
	}
	logger.Infof("Using bitrise-build-cache %s from %s", version, bin)

	return bin, nil
}

func toolsLinkPath() string {
	return filepath.Join(configs.GetBitriseToolsDirPath(), binaryName)
}

// usePreinstalled ensures the tools entry resolves to the CLI and returns its absolute, symlink-free path and the
// version it reports, so a relative value cannot leave a dangling link and the entry never points at itself.
func usePreinstalled(ctx context.Context, path string) (string, string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("resolve the path: %w", err)
	}
	bin, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", "", fmt.Errorf("resolve the path: %w", err)
	}
	version, err := reportedVersion(ctx, bin)
	if err != nil {
		return "", "", err
	}

	if tools := toolsLinkPath(); !sameFile(tools, bin) {
		if err := link(bin, tools); err != nil {
			return "", "", fmt.Errorf("link onto PATH: %w", err)
		}
	}

	return bin, version, nil
}

func sameFile(a, b string) bool {
	infoA, errA := os.Stat(a)
	infoB, errB := os.Stat(b)

	return errA == nil && errB == nil && os.SameFile(infoA, infoB)
}

func reportedVersion(ctx context.Context, bin string) (string, error) {
	if !isExecutable(bin) {
		return "", errors.New("not an executable file")
	}

	ctx, cancel := context.WithTimeout(ctx, versionTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "version")
	cmd.WaitDelay = commandWaitDelay
	out, err := cmd.Output()
	if err != nil && !errors.Is(err, exec.ErrWaitDelay) {
		return "", fmt.Errorf("`version` failed: %w", err)
	}
	version := strings.TrimPrefix(strings.TrimSpace(string(out)), "v")
	if !versionPattern.MatchString(version) {
		return "", fmt.Errorf("reports %q, which is not a version", version)
	}

	return version, nil
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

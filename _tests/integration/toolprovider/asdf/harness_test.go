//go:build linux_and_mac
// +build linux_and_mac

package asdf

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"al.essio.dev/pkg/shellescape"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
)

type flavor int

const (
	flavorAsdfClassic flavor = iota
	flavorAsdfRewrite
)

const cacheDir = "bitrise-asdf-test-cache"

type asdfInstallation struct {
	flavor  flavor
	version string
	plugins []string
}

type testEnv struct {
	envVars   map[string]string
	shellInit string
}

// createTestEnv creates an isolated installation of a given asdf version for testing.
func createTestEnv(t *testing.T, installRequest asdfInstallation) (testEnv, error) {
	homeDir := t.TempDir()
	dataDir := t.TempDir()
	shimsDir := filepath.Join(dataDir, "shims")

	installDir, err := install(t, installRequest)
	if err != nil {
		return testEnv{}, fmt.Errorf("install asdf: %w", err)
	}

	testingEnv := testEnv{
		envVars: map[string]string{
			// We intentionally clear up $PATH to avoid the system-wide path influencing tests
			"PATH": fmt.Sprintf("%s:%s:/bin:/usr/bin", installDir, shimsDir),
			// ASDF_DATA_DIR is where plugins and tool versions are installed (not to be confused with ASDF_DIR)
			"ASDF_DATA_DIR":    dataDir,
			"ASDF_CONFIG_FILE": filepath.Join(dataDir, ".asdfrc"),
			// Avoid conflicts with other asdf installations (global .tool-versions file is in $HOME)
			"HOME": homeDir,
			"PWD":  homeDir,
		},
	}

	if installRequest.flavor == flavorAsdfClassic {
		// https://github.com/asdf-vm/asdf/blob/v0.14.1/docs/guide/getting-started.md
		testingEnv.shellInit = fmt.Sprintf(". %s", filepath.Join(installDir, "asdf.sh"))
		// ASDF_DIR is where asdf itself is installed (unlike ASDF_DATA_DIR)
		testingEnv.envVars["ASDF_DIR"] = installDir
	}

	for _, plugin := range installRequest.plugins {
		out, err := testingEnv.runAsdf("plugin", "add", plugin)
		if err != nil {
			return testingEnv, fmt.Errorf("install asdf plugin %s: %w\n\nOutput:\n%s", plugin, err, out)
		}
	}

	return testingEnv, nil
}

func (te *testEnv) toExecEnv() execenv.ExecEnv {
	return execenv.ExecEnv{
		EnvVars:            te.envVars,
		ClearInheritedEnvs: true,
		ShellInit:          te.shellInit,
	}
}

func (te *testEnv) runAsdf(args ...string) (string, error) {
	cmdWithArgs := append([]string{"asdf"}, args...)
	return te.runCommand(nil, cmdWithArgs...)
}

func (te *testEnv) runCommand(extraEnvs map[string]string, args ...string) (string, error) {
	innerShellCmd := []string{}
	if te.shellInit != "" {
		innerShellCmd = append(innerShellCmd, te.shellInit+" &&")
	}
	innerShellCmd = append(innerShellCmd, shellescape.QuoteCommand(args))
	bashArgs := []string{"-c", strings.Join(innerShellCmd, " ")}
	bashCmd := exec.Command("bash", bashArgs...)
	bashCmd.Env = os.Environ()
	for k, v := range te.envVars {
		bashCmd.Env = append(bashCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range extraEnvs {
		bashCmd.Env = append(bashCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	output, err := bashCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %v: %w\n\nOutput:\n%s", "bash", bashArgs, err, output)
	}

	return string(output), nil
}

func install(t *testing.T, install asdfInstallation) (string, error) {
	installDir := filepath.Join(os.TempDir(), cacheDir, fmt.Sprintf("asdf-v%s", install.version))
	_, err := os.Stat(installDir)
	if err == nil {
		// Already installed and cached
		return installDir, nil
	}

	if err != nil && !os.IsNotExist(err) {
		// Unexpected error
		return "", fmt.Errorf("check cache directory %s: %w", installDir, err)
	}

	// Not found in cache, installing now
	if install.flavor == flavorAsdfRewrite {
		if err := downloadReleaseBinary(t, install.version, installDir); err != nil {
			return "", fmt.Errorf("download asdf binary: %w", err)
		}
	} else if install.flavor == flavorAsdfClassic {
		if err := gitCheckout(t, install.version, installDir); err != nil {
			return "", fmt.Errorf("checkout asdf version %s: %w", install.version, err)
		}
	}
	return installDir, nil
}

func gitCheckout(t *testing.T, version string, targetDir string) error {
	gitCmd := exec.Command(
		"git",
		"clone",
		"--depth=1",
		"https://github.com/asdf-vm/asdf.git",
		targetDir,
		"--branch", "v"+version,
	)
	return gitCmd.Run()
}

func downloadReleaseBinary(t *testing.T, version string, targetDir string) error {
	url := fmt.Sprintf("https://github.com/asdf-vm/asdf/releases/download/v%s/asdf-v%s-%s-%s.tar.gz", version, version, runtime.GOOS, runtime.GOARCH)
	downloadDir := t.TempDir()
	tarballPath := filepath.Join(downloadDir, fmt.Sprintf("asdf-v%s.tar.gz", version))

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download from %s: received status code %d", url, resp.StatusCode)
	}

	tarballFile, err := os.Create(tarballPath)
	if err != nil {
		return fmt.Errorf("create temporary file %s: %w", tarballPath, err)
	}
	defer tarballFile.Close()

	_, err = io.Copy(tarballFile, resp.Body)
	if err != nil {
		return fmt.Errorf("write to temporary file %s: %w", tarballPath, err)
	}
	tarballFile.Close()

	fileReader, err := os.Open(tarballPath)
	if err != nil {
		return fmt.Errorf("open tarball %s: %w", tarballPath, err)
	}
	defer fileReader.Close()

	gzipReader, err := gzip.NewReader(fileReader)
	if err != nil {
		return fmt.Errorf("create gzip reader for %s: %w", tarballPath, err)
	}
	defer gzipReader.Close()
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create directory for %s: %w", targetDir, err)
	}

	tarReader := tar.NewReader(gzipReader)
	header, err := tarReader.Next()
	if err == io.EOF {
		return fmt.Errorf("empty tarball %s", tarballPath)
	}
	if err != nil {
		return fmt.Errorf("read tar header: %w", err)
	}

	if header.Typeflag != tar.TypeReg {
		return fmt.Errorf("first entry is not a regular file in %s", tarballPath)
	}

	outFile, err := os.OpenFile(filepath.Join(targetDir, "asdf"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create output file %s: %w", filepath.Join(targetDir, "asdf"), err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, tarReader); err != nil {
		return fmt.Errorf("extract asdf to %s: %w", targetDir, err)
	}

	return nil
}

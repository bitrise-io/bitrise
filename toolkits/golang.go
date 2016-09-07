package toolkits

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/bitrise-io/bitrise/utils"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/progress"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-utils/stringutil"
	"github.com/bitrise-io/go-utils/versions"
)

const (
	minGoVersionForToolkit = "1.7"
)

// GoToolkit ...
type GoToolkit struct {
}

// ToolkitName ...
func (toolkit GoToolkit) ToolkitName() string {
	return "go"
}

// Check ...
func (toolkit GoToolkit) Check() (bool, ToolkitCheckResult, error) {
	binPath, err := utils.CheckProgramInstalledPath("go")
	if err != nil {
		return true, ToolkitCheckResult{}, nil
	}

	verOut, err := cmdex.RunCommandAndReturnStdout("go", "version")
	if err != nil {
		return false, ToolkitCheckResult{}, fmt.Errorf("Failed to check go version, error: %s", err)
	}

	verStr := stringutil.ReadFirstLine(verOut, true)

	return false, ToolkitCheckResult{
		Path:    binPath,
		Version: verStr,
	}, nil
}

func goToolkitRootPath() string {
	return filepath.Join(configs.GetBitriseToolkitsDirPath(), "go")
}

func goToolkitInstallRootPath() string {
	return filepath.Join(goToolkitRootPath(), "go")
}

func goToolkitBinsPath() string {
	return filepath.Join(goToolkitInstallRootPath(), "bin")
}

func parseGoVersionFromGoVersionOutput(goVersionCallOutput string) (string, error) {
	origGoVersionCallOutput := goVersionCallOutput
	goVersionCallOutput = strings.TrimSpace(goVersionCallOutput)
	if goVersionCallOutput == "" {
		return "", errors.New("Failed to parse Go version, error: version call output was empty")
	}

	// example goVersionCallOutput: go version go1.7 darwin/amd64
	goVerExp := regexp.MustCompile(`go version go(?P<goVersionNumber>[0-9.]+) (?P<platform>[a-zA-Z0-9]+/[a-zA-Z0-9]+)`)
	expRes := goVerExp.FindStringSubmatch(goVersionCallOutput)
	if expRes == nil {
		return "", fmt.Errorf("Failed to parse Go version, error: failed to find version in input: %s", origGoVersionCallOutput)
	}
	verStr := expRes[1]

	return verStr, nil
}

func isGoInPATHSufficient() bool {
	if configs.IsDebugUseSystemTools() {
		log.Warn("[BitriseDebug] Using system tools (system installed Go), instead of the ones in BITRISE_HOME")
		return true
	}

	if _, err := utils.CheckProgramInstalledPath("go"); err != nil {
		return false
	}

	verOut, err := cmdex.RunCommandAndReturnStdout("go", "version")
	if err != nil {
		return false
	}

	verStr, err := parseGoVersionFromGoVersionOutput(verOut)
	if err != nil {
		return false
	}

	// version check
	isVersionOk, err := versions.IsVersionGreaterOrEqual(verStr, minGoVersionForToolkit)
	if err != nil {
		return false
	}
	if !isVersionOk {
		return false
	}

	return true
}

// Bootstrap ...
func (toolkit GoToolkit) Bootstrap() error {
	if isGoInPATHSufficient() {
		return nil
	}

	pthWithGoBins := configs.GeneratePATHEnvString(os.Getenv("PATH"), goToolkitBinsPath())
	if err := os.Setenv("PATH", pthWithGoBins); err != nil {
		return fmt.Errorf("Failed to set PATH to include the Go toolkit bins, error: %s", err)
	}

	if err := os.Setenv("GOROOT", goToolkitInstallRootPath()); err != nil {
		return fmt.Errorf("Failed to set GOROOT to Go toolkit root, error: %s", err)
	}

	return nil
}

func installGoTar(goTarGzPath string) error {
	installToPath := goToolkitRootPath()

	if err := os.RemoveAll(installToPath); err != nil {
		return fmt.Errorf("Failed to remove previous Go toolkit install (path: %s), error: %s", installToPath, err)
	}
	if err := pathutil.EnsureDirExist(installToPath); err != nil {
		return fmt.Errorf("Failed create Go toolkit directory (path: %s), error: %s", installToPath, err)
	}

	cmd := cmdex.NewCommand("tar", "-C", installToPath, "-xzf", goTarGzPath)
	if combinedOut, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		log.Errorln(" [!] Failed to uncompress Go toolkit, output:")
		log.Errorln(combinedOut)
		return fmt.Errorf("Failed to uncompress Go toolkit, error: %s", err)
	}
	return nil
}

// Install ...
func (toolkit GoToolkit) Install() error {
	if isGoInPATHSufficient() {
		fmt.Print("System Installed Go is sufficient, no need to install it for the toolkit")
		return nil
	}

	versionStr := minGoVersionForToolkit
	osStr := runtime.GOOS
	archStr := runtime.GOARCH
	extentionStr := "tar.gz"
	if osStr == "windows" {
		extentionStr = "zip"
	}
	downloadURL := fmt.Sprintf("https://storage.googleapis.com/golang/go%s.%s-%s.%s",
		versionStr, osStr, archStr, extentionStr)
	log.Infoln("downloadURL: ", downloadURL)

	// bitriseToolkitsDirPath := configs.GetBitriseToolkitsDirPath()
	toolkitsTmpDirPath := getBitriseToolkitsTmpDirPath()
	if err := pathutil.EnsureDirExist(toolkitsTmpDirPath); err != nil {
		return fmt.Errorf("Failed to create Toolkits TMP directory, error: %s", err)
	}

	localFileName := "go." + extentionStr
	destinationPth := filepath.Join(toolkitsTmpDirPath, localFileName)

	var downloadErr error
	fmt.Print("=> Downloading ...")
	progress.SimpleProgress(".", 2*time.Second, func() {
		downloadErr = retry.Times(2).Wait(5 * time.Second).Try(func(attempt uint) error {
			if attempt > 0 {
				fmt.Println()
				fmt.Println("==> Download failed, retrying ...")
				fmt.Println()
			}
			return tools.DownloadFile(downloadURL, destinationPth)
		})
	})
	if downloadErr != nil {
		return fmt.Errorf("Failed to download toolkit (%s), error: %s", downloadURL, downloadErr)
	}
	log.Infoln("Toolkit downloaded to: ", destinationPth)

	fmt.Println("=> Installing ...")
	if err := installGoTar(destinationPth); err != nil {
		return fmt.Errorf("Failed to install Go toolkit, error: %s", err)
	}
	fmt.Println("=> Installing [DONE]")

	return nil
}

// StepRunCommandArguments ...
func (toolkit GoToolkit) StepRunCommandArguments(stepDirPath string) ([]string, error) {
	stepFilePath := filepath.Join(stepDirPath, "main.go")
	cmd := []string{"go", "run", stepFilePath}
	return cmd, nil
}

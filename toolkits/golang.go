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
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/bitrise-io/bitrise/utils"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/progress"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-utils/versions"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/bitrise-tools/gows/gows"
)

const (
	minGoVersionForToolkit = "1.7"
)

// === Base Toolkit struct ===

// GoToolkit ...
type GoToolkit struct {
}

// ToolkitName ...
func (toolkit GoToolkit) ToolkitName() string {
	return "go"
}

// === Toolkit: Check ===

// GoConfigurationModel ...
type GoConfigurationModel struct {
	// full path of the go binary to use
	GoBinaryPath string
	// GOROOT env var value to set (unless empty)
	GOROOT string
}

func checkGoConfiguration(goConfig GoConfigurationModel) (bool, ToolkitCheckResult, error) {
	cmdEnvs := os.Environ()
	if len(goConfig.GOROOT) > 0 {
		cmdEnvs = append(cmdEnvs, "GOROOT="+goConfig.GOROOT)
	}
	verOut, err := cmdex.NewCommand(goConfig.GoBinaryPath, "version").SetEnvs(cmdEnvs).RunAndReturnTrimmedOutput()
	if err != nil {
		return false, ToolkitCheckResult{}, fmt.Errorf("Failed to check go version, error: %s", err)
	}

	verStr, err := parseGoVersionFromGoVersionOutput(verOut)
	if err != nil {
		return false, ToolkitCheckResult{}, fmt.Errorf("Failed to parse go version, error: %s", err)
	}

	checkRes := ToolkitCheckResult{
		Path:    goConfig.GoBinaryPath,
		Version: verStr,
	}

	// version check
	isVersionOk, err := versions.IsVersionGreaterOrEqual(verStr, minGoVersionForToolkit)
	if err != nil {
		return false, checkRes, fmt.Errorf("Failed to validate installed go version, error: %s", err)
	}
	if !isVersionOk {
		return true, checkRes, nil
	}

	return false, checkRes, nil
}

func selectGoConfiguration() (bool, ToolkitCheckResult, GoConfigurationModel, error) {
	potentialGoConfigurations := []GoConfigurationModel{}
	// from PATH
	{
		binPath, err := utils.CheckProgramInstalledPath("go")
		if err == nil {
			potentialGoConfigurations = append(potentialGoConfigurations, GoConfigurationModel{GoBinaryPath: binPath})
		}
	}
	// from Bitrise Toolkits
	{
		binPath := goBinaryInToolkitFullPath()
		if isExist, err := pathutil.IsPathExists(binPath); err != nil {
			log.Warnf("Failed to check the status of the 'go' binary inside the Bitrise Toolkit dir, error: %s", err)
		} else if isExist {
			potentialGoConfigurations = append(potentialGoConfigurations, GoConfigurationModel{
				GoBinaryPath: binPath,
				GOROOT:       goToolkitInstallRootPath(),
			})
		}
	}

	isRequireInstall := true
	checkResult := ToolkitCheckResult{}
	goConfig := GoConfigurationModel{}
	var checkError error
	for _, aPotentialGoInfoToUse := range potentialGoConfigurations {
		isInstReq, chkRes, err := checkGoConfiguration(aPotentialGoInfoToUse)
		checkResult = chkRes
		checkError = err
		if !isInstReq {
			// select this one
			goConfig = aPotentialGoInfoToUse
			isRequireInstall = false
			break
		}
	}

	if len(potentialGoConfigurations) > 0 && isRequireInstall {
		log.Warnf("Installed go found (path: %s), but not a supported version: %s", checkResult.Path, checkResult.Version)
	}

	return isRequireInstall, checkResult, goConfig, checkError
}

// Check ...
func (toolkit GoToolkit) Check() (bool, ToolkitCheckResult, error) {
	isInstallRequired, checkResult, _, err := selectGoConfiguration()
	return isInstallRequired, checkResult, err
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

// === Toolkit: Bootstrap ===

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

// === Toolkit: Install ===

func installGoTar(goTarGzPath string) error {
	installToPath := goToolkitInstallToPath()

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
	log.Debugln("downloadURL: ", downloadURL)

	goTmpDirPath := goToolkitTmpDirPath()
	if err := pathutil.EnsureDirExist(goTmpDirPath); err != nil {
		return fmt.Errorf("Failed to create Toolkits TMP directory, error: %s", err)
	}

	localFileName := "go." + extentionStr
	goArchiveDownloadPath := filepath.Join(goTmpDirPath, localFileName)

	var downloadErr error
	fmt.Print("=> Downloading ...")
	progress.SimpleProgress(".", 2*time.Second, func() {
		downloadErr = retry.Times(2).Wait(5 * time.Second).Try(func(attempt uint) error {
			if attempt > 0 {
				fmt.Println()
				fmt.Println("==> Download failed, retrying ...")
				fmt.Println()
			}
			return tools.DownloadFile(downloadURL, goArchiveDownloadPath)
		})
	})
	if downloadErr != nil {
		return fmt.Errorf("Failed to download toolkit (%s), error: %s", downloadURL, downloadErr)
	}
	log.Debugln("Toolkit downloaded to: ", goArchiveDownloadPath)

	fmt.Println("=> Installing ...")
	if err := installGoTar(goArchiveDownloadPath); err != nil {
		return fmt.Errorf("Failed to install Go toolkit, error: %s", err)
	}
	if err := os.Remove(goArchiveDownloadPath); err != nil {
		return fmt.Errorf("Failed to remove the downloaded Go archive (path: %s), error: %s", goArchiveDownloadPath, err)
	}
	fmt.Println("=> Installing [DONE]")

	return nil
}

// === Toolkit: Prepare for Step Run ===

func goBuildInIsolation(packageName, srcPath, outputBinPath string) error {
	log.Debugf("=> Installing package (%s) to path (%s) ...", packageName, srcPath)
	workspaceRootPath, err := pathutil.NormalizedOSTempDirPath("bitrise-go-toolkit")
	if err != nil {
		return fmt.Errorf("Failed to create root directory of isolated workspace, error: %s", err)
	}
	log.Debugln("=> Using sandboxed workspace:", workspaceRootPath)

	// origGOPATH := os.Getenv("GOPATH")
	// if origGOPATH == "" {
	// 	return fmt.Errorf("You don't have a GOPATH environment - please set it; GOPATH/bin will be symlinked")
	// }

	// log.Debugln("=> Symlink GOPATH/bin into sandbox ...")
	// if err := gows.CreateGopathBinSymlink(origGOPATH, workspaceRootPath); err != nil {
	// 	return fmt.Errorf("Failed to create GOPATH/bin symlink, error: %s", err)
	// }
	// log.Debugln("   [DONE]")

	fullPackageWorkspacePath := filepath.Join(workspaceRootPath, "src", packageName)
	log.Debugf("=> Creating Symlink: (%s) -> (%s)", srcPath, fullPackageWorkspacePath)
	if err := gows.CreateOrUpdateSymlink(srcPath, fullPackageWorkspacePath); err != nil {
		return fmt.Errorf("Failed to create Project->Workspace symlink, error: %s", err)
	}
	log.Debugf(" [DONE] Symlink is in place")

	log.Debugln("=> Building package " + packageName + " ...")
	{
		isInstallRequired, _, goConfig, err := selectGoConfiguration()
		if err != nil {
			return fmt.Errorf("Failed to select an appropriate Go installation for compiling the step, error: %s", err)
		}
		if isInstallRequired {
			return fmt.Errorf("Failed to select an appropriate Go installation for compiling the step, error: %s",
				"Found Go version is older than required. Please run 'bitrise setup' to check and install the required version")
		}

		cmd := gows.CreateCommand(workspaceRootPath, workspaceRootPath,
			goConfig.GoBinaryPath, "build", "-o", outputBinPath, packageName)
		cmd.Env = append(os.Environ(), "GOROOT="+goConfig.GOROOT)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to install package, error: %s", err)
		}
	}
	log.Debugln("   [DONE] Package successfully installed")

	log.Debugln("=> Delete isolated workspace ...")
	{
		if err := os.RemoveAll(workspaceRootPath); err != nil {
			return fmt.Errorf("Failed to delete temporary isolated workspace, error: %s", err)
		}
	}
	log.Debugln("   [DONE]")

	return nil
}

// stepIDorURI : doesn't work for "path::./" yet!!
func stepBinaryFilename(sIDData models.StepIDData) string {
	//
	replaceRexp, err := regexp.Compile("[^A-Za-z0-9.-]")
	if err != nil {
		log.Warn("Invalid regex, error: %s", err)
		return ""
	}

	compositeStepID := fmt.Sprintf("%s-%s-%s",
		sIDData.SteplibSource, sIDData.IDorURI, sIDData.Version)

	safeStepID := replaceRexp.ReplaceAllString(compositeStepID, "_")
	fmt.Println(" (debug) safeStepID: ", safeStepID)
	//
	return safeStepID
}

func stepBinaryCacheFullPath(sIDData models.StepIDData) string {
	return filepath.Join(goToolkitCacheRootPath(), stepBinaryFilename(sIDData))
}

// PrepareForStepRun ...
func (toolkit GoToolkit) PrepareForStepRun(step stepmanModels.StepModel, sIDData models.StepIDData, stepAbsDirPath string) error {
	fullStepBinPath := stepBinaryCacheFullPath(sIDData)

	// try to use cached binary, if possible
	if sIDData.IsUniqueResourceID() {
		if exists, err := pathutil.IsPathExists(fullStepBinPath); err != nil {
			log.Warn("Failed to check cached binary for step, error: %s", err)
		} else if exists {
			log.Debugln("No need to compile, binary already exists")
			return nil
		}
	}

	// it's not cached, so compile it

	if step.Toolkit == nil {
		return errors.New("No Toolkit information specified in step!")
	}
	if step.Toolkit.Go == nil {
		return errors.New("No Toolkit.Go information specified in step!")
	}
	packageName := step.Toolkit.Go.PackageName

	return goBuildInIsolation(packageName, stepAbsDirPath, fullStepBinPath)
}

// === Toolkit: Step Run ===

// StepRunCommandArguments ...
func (toolkit GoToolkit) StepRunCommandArguments(stepDirPath string, sIDData models.StepIDData) ([]string, error) {
	// stepFilePath := filepath.Join(stepDirPath, "main.go")
	// cmd := []string{"go", "run", stepFilePath}

	fullStepBinPath := stepBinaryCacheFullPath(sIDData)
	return []string{fullStepBinPath}, nil
}

// === Toolkit path utility function ===

func goToolkitRootPath() string {
	return filepath.Join(configs.GetBitriseToolkitsDirPath(), "go")
}

func goToolkitTmpDirPath() string {
	return filepath.Join(goToolkitRootPath(), "tmp")
}

func goToolkitInstallToPath() string {
	return filepath.Join(goToolkitRootPath(), "inst")
}
func goToolkitCacheRootPath() string {
	return filepath.Join(goToolkitRootPath(), "cache")
}

func goToolkitInstallRootPath() string {
	return filepath.Join(goToolkitInstallToPath(), "go")
}

func goToolkitBinsPath() string {
	return filepath.Join(goToolkitInstallRootPath(), "bin")
}

func goBinaryInToolkitFullPath() string {
	return filepath.Join(goToolkitBinsPath(), "go")
}

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

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/progress"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/bitrise-io/bitrise/utils"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-utils/versions"
	"github.com/bitrise-io/gows/gows"
	stepmanModels "github.com/bitrise-io/stepman/models"
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
	verOut, err := command.New(goConfig.GoBinaryPath, "version").SetEnvs(cmdEnvs...).RunAndReturnTrimmedOutput()
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
	goVerExp := regexp.MustCompile(`go version go(?P<goVersionNumber>[0-9.]+)[a-zA-Z0-9]* (?P<platform>[a-zA-Z0-9]+/[a-zA-Z0-9]+)`)
	expRes := goVerExp.FindStringSubmatch(goVersionCallOutput)
	if expRes == nil {
		return "", fmt.Errorf("Failed to parse Go version, error: failed to find version in input: %s", origGoVersionCallOutput)
	}
	verStr := expRes[1]

	return verStr, nil
}

// IsToolAvailableInPATH ...
func (toolkit GoToolkit) IsToolAvailableInPATH() bool {
	if configs.IsDebugUseSystemTools() {
		log.Warnf("[BitriseDebug] Using system tools (system installed Go), instead of the ones in BITRISE_HOME")
		return true
	}

	if _, err := utils.CheckProgramInstalledPath("go"); err != nil {
		return false
	}

	if _, err := command.RunCommandAndReturnStdout("go", "version"); err != nil {
		return false
	}

	return true
}

// === Toolkit: Bootstrap ===

// Bootstrap ...
func (toolkit GoToolkit) Bootstrap() error {
	if toolkit.IsToolAvailableInPATH() {
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

	cmd := command.New("tar", "-C", installToPath, "-xzf", goTarGzPath)
	if combinedOut, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		log.Errorf(" [!] Failed to uncompress Go toolkit, output:")
		log.Errorf(combinedOut)
		return fmt.Errorf("Failed to uncompress Go toolkit, error: %s", err)
	}
	return nil
}

// Install ...
func (toolkit GoToolkit) Install() error {
	versionStr := minGoVersionForToolkit
	osStr := runtime.GOOS
	archStr := runtime.GOARCH
	extentionStr := "tar.gz"
	if osStr == "windows" {
		extentionStr = "zip"
	}
	downloadURL := fmt.Sprintf("https://storage.googleapis.com/golang/go%s.%s-%s.%s", versionStr, osStr, archStr, extentionStr)

	goTmpDirPath := goToolkitTmpDirPath()
	if err := pathutil.EnsureDirExist(goTmpDirPath); err != nil {
		return fmt.Errorf("Failed to create Toolkits TMP directory, error: %s", err)
	}

	localFileName := "go." + extentionStr
	goArchiveDownloadPath := filepath.Join(goTmpDirPath, localFileName)

	var downloadErr error
	progress.ShowIndicator("Downloading", func() {
		downloadErr = retry.Times(2).Wait(5 * time.Second).Try(func(attempt uint) error {
			if attempt > 0 {
				log.Warnf("==> Download failed, retrying ...")
			}
			return tools.DownloadFile(downloadURL, goArchiveDownloadPath)
		})
	})
	if downloadErr != nil {
		return fmt.Errorf("Failed to download toolkit (%s), error: %s", downloadURL, downloadErr)
	}

	log.Print("=> Installing ...")
	if err := installGoTar(goArchiveDownloadPath); err != nil {
		return fmt.Errorf("Failed to install Go toolkit, error: %s", err)
	}
	if err := os.Remove(goArchiveDownloadPath); err != nil {
		return fmt.Errorf("Failed to remove the downloaded Go archive (path: %s), error: %s", goArchiveDownloadPath, err)
	}
	log.Print("=> Installing [DONE]")

	return nil
}

// === Toolkit: Prepare for Step Run ===

func goBuildStep(cmdRunner commandRunner, goConfig GoConfigurationModel, packageName, stepAbsDirPath, outputBinPath string) error {
	cmdBuilder := goCmdBuilder{goConfig: goConfig}

	if isGoPathModeStep(stepAbsDirPath) {
		log.Debugf("[Go deps] Step requires GOPATH mode")
		// Go 1.17 will ignore GO111MODULE (https://blog.golang.org/go116-module-changes)
		// GO111MODULE needs to be set to "on" when GOPATH is no longer supported.
		// If GO111MODULE is not set, will be handled as it was "on".
		mode, err := getGoEnv(cmdRunner, goConfig.GoBinaryPath, "GO111MODULE")
		if err != nil {
			log.Warnf("[Go deps] Could not determine if GOPATH mode is supported: %v", err)
		}

		log.Debugf("[Go deps] GO111MODULE='%s'", mode)
		if isGoPathModeSupported(mode) {
			return goBuildInGoPathMode(cmdRunner, goConfig, packageName, stepAbsDirPath, outputBinPath)
		}

		log.Debugf("[Go deps] Migrating Step to Go modules as Go installation does not support GOPATH mode")
		if err := migrateToGoModules(stepAbsDirPath, packageName); err != nil {
			return fmt.Errorf("failed to migrate to go modules: %v", err)
		}

		buildCmd := cmdBuilder.goBuildMigratedModules(stepAbsDirPath, outputBinPath)
		if _, err := cmdRunner.runForOutput(buildCmd); err != nil {
			return fmt.Errorf("failed to build Step in directory (%s): %v", stepAbsDirPath, err)
		}

		return nil
	}

	log.Debugf("[Go deps] Step requires Go modules mode")
	buildCmd := cmdBuilder.goBuildInModuleMode(stepAbsDirPath, outputBinPath)
	if _, err := cmdRunner.runForOutput(buildCmd); err != nil {
		return fmt.Errorf("failed to build Step in directory (%s): %v", stepAbsDirPath, err)
	}

	return nil
}

func goBuildInGoPathMode(cmdRunner commandRunner, goConfig GoConfigurationModel, packageName, srcPath, outputBinPath string) error {
	workspaceRootPath, err := pathutil.NormalizedOSTempDirPath("bitrise-go-toolkit")
	if err != nil {
		return fmt.Errorf("Failed to create root directory of isolated workspace, error: %s", err)
	}

	fullPackageWorkspacePath := filepath.Join(workspaceRootPath, "src", packageName)
	if err := gows.CreateOrUpdateSymlink(srcPath, fullPackageWorkspacePath); err != nil {
		return fmt.Errorf("Failed to create Project->Workspace symlink, error: %s", err)
	}

	cmd := gows.CreateCommand(workspaceRootPath, workspaceRootPath, goConfig.GoBinaryPath, "build", "-o", outputBinPath, packageName)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.Env = append(cmd.Env, "GOROOT="+goConfig.GOROOT)

	buildCmd := command.NewWithCmd(cmd)

	if _, err := cmdRunner.runForOutput(buildCmd); err != nil {
		return fmt.Errorf("Failed to install package, error: %s", err)
	}

	if err := os.RemoveAll(workspaceRootPath); err != nil {
		return fmt.Errorf("Failed to delete temporary isolated workspace, error: %s", err)
	}

	return nil
}

// stepIDorURI : doesn't work for "path::./" yet!!
func stepBinaryFilename(sIDData models.StepIDData) string {
	//
	replaceRexp, err := regexp.Compile("[^A-Za-z0-9.-]")
	if err != nil {
		log.Warnf("Invalid regex, error: %s", err)
		return ""
	}

	compositeStepID := fmt.Sprintf("%s-%s", sIDData.SteplibSource, sIDData.IDorURI)
	if sIDData.Version != "" {
		compositeStepID += "-" + sIDData.Version
	}

	safeStepID := replaceRexp.ReplaceAllString(compositeStepID, "_")
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
			log.Warnf("Failed to check cached binary for step, error: %s", err)
		} else if exists {
			return nil
		}
	}

	// it's not cached, so compile it
	if step.Toolkit == nil {
		return errors.New("No Toolkit information specified in step")
	}
	if step.Toolkit.Go == nil {
		return errors.New("No Toolkit.Go information specified in step")
	}

	isInstallRequired, _, goConfig, err := selectGoConfiguration()
	if err != nil {
		return fmt.Errorf("Failed to select an appropriate Go installation for compiling the Step: %s", err)
	}
	if isInstallRequired {
		return fmt.Errorf("Failed to select an appropriate Go installation for compiling the Step: %s",
			"Found Go version is older than required. Please run 'bitrise setup' to check and install the required version")
	}

	return goBuildStep(&defaultRunner{}, goConfig, step.Toolkit.Go.PackageName, stepAbsDirPath, fullStepBinPath)
}

// === Toolkit: Step Run ===

// StepRunCommandArguments ...
func (toolkit GoToolkit) StepRunCommandArguments(step stepmanModels.StepModel, sIDData models.StepIDData, stepAbsDirPath string) ([]string, error) {
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

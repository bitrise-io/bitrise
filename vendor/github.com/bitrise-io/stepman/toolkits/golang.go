package toolkits

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-utils/versions"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/stepman"
)

type GoToolkit struct {
	logger stepman.Logger
}

func NewGoToolkit(logger stepman.Logger) GoToolkit {
	return GoToolkit{
		logger: logger,
	}
}

func (toolkit GoToolkit) ToolkitName() string {
	return "go"
}

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
		return false, ToolkitCheckResult{}, fmt.Errorf("check go version: %s", err)
	}

	verStr, err := parseGoVersionFromGoVersionOutput(verOut)
	if err != nil {
		return false, ToolkitCheckResult{}, fmt.Errorf("parse go version: %s", err)
	}

	checkRes := ToolkitCheckResult{
		Path:    goConfig.GoBinaryPath,
		Version: verStr,
	}

	// version check
	isVersionOk, err := versions.IsVersionGreaterOrEqual(verStr, minGoVersionForToolkit)
	if err != nil {
		return false, checkRes, fmt.Errorf("validate installed go version: %s", err)
	}
	if !isVersionOk {
		return true, checkRes, nil
	}

	return false, checkRes, nil
}

func selectGoConfiguration(logger stepman.Logger) (bool, ToolkitCheckResult, GoConfigurationModel, error) {
	potentialGoConfigurations := []GoConfigurationModel{}
	// from PATH
	{
		binPath, err := exec.LookPath("go")
		if err == nil {
			potentialGoConfigurations = append(potentialGoConfigurations, GoConfigurationModel{GoBinaryPath: binPath})
		}
	}
	// from Bitrise Toolkits
	{
		binPath := goBinaryInToolkitFullPath()
		if isExist, err := pathutil.IsPathExists(binPath); err != nil {
			logger.Warnf("Failed to check the status of the 'go' binary inside the Bitrise Toolkit dir, error: %s", err)
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
		logger.Warnf("Installed go found (path: %s), but not a supported version: %s", checkResult.Path, checkResult.Version)
	}

	return isRequireInstall, checkResult, goConfig, checkError
}

func (toolkit GoToolkit) Check() (bool, ToolkitCheckResult, error) {
	isInstallRequired, checkResult, _, err := selectGoConfiguration(toolkit.logger)
	return isInstallRequired, checkResult, err
}

func parseGoVersionFromGoVersionOutput(goVersionCallOutput string) (string, error) {
	origGoVersionCallOutput := goVersionCallOutput
	goVersionCallOutput = strings.TrimSpace(goVersionCallOutput)
	if goVersionCallOutput == "" {
		return "", errors.New("parse Go version: version call output was empty")
	}

	// example goVersionCallOutput: go version go1.7 darwin/amd64
	goVerExp := regexp.MustCompile(`go version go(?P<goVersionNumber>[0-9.]+)[a-zA-Z0-9]* (?P<platform>[a-zA-Z0-9]+/[a-zA-Z0-9]+)`)
	expRes := goVerExp.FindStringSubmatch(goVersionCallOutput)
	if expRes == nil {
		return "", fmt.Errorf("parse Go version, error: failed to find version in input: %s", origGoVersionCallOutput)
	}
	verStr := expRes[1]

	return verStr, nil
}

func (toolkit GoToolkit) IsToolAvailableInPATH() bool {
	if _, err := exec.LookPath("go"); err != nil {
		return false
	}

	if _, err := command.RunCommandAndReturnStdout("go", "version"); err != nil {
		return false
	}

	return true
}

func (toolkit GoToolkit) Bootstrap() error {
	if toolkit.IsToolAvailableInPATH() {
		return nil
	}

	pathWithGoBins := fmt.Sprintf("%s:%s", goToolkitBinsPath(), os.Getenv("PATH"))
	if err := os.Setenv("PATH", pathWithGoBins); err != nil {
		return fmt.Errorf("set PATH to include the Go toolkit bins, error: %s", err)
	}

	if err := os.Setenv("GOROOT", goToolkitInstallRootPath()); err != nil {
		return fmt.Errorf("set GOROOT to Go toolkit root, error: %s", err)
	}

	return nil
}

func installGoTar(logger stepman.Logger, goTarGzPath string) error {
	installToPath := goToolkitInstallToPath()

	if err := os.RemoveAll(installToPath); err != nil {
		return fmt.Errorf("remove previous Go toolkit install (path: %s): %s", installToPath, err)
	}
	if err := pathutil.EnsureDirExist(installToPath); err != nil {
		return fmt.Errorf("create Go toolkit directory (path: %s): %s", installToPath, err)
	}

	cmd := command.New("tar", "-C", installToPath, "-xzf", goTarGzPath)
	if combinedOut, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		logger.Errorf(" [!] Failed to uncompress Go toolkit, output:")
		logger.Errorf(combinedOut)
		return fmt.Errorf("uncompress Go toolkit: %s", err)
	}
	return nil
}

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
		return fmt.Errorf("create Toolkits TMP directory: %s", err)
	}

	localFileName := "go." + extentionStr
	goArchiveDownloadPath := filepath.Join(goTmpDirPath, localFileName)

	var downloadErr error

	toolkit.logger.Infof("=> Downloading ...")
	downloadErr = retry.Times(2).Wait(5 * time.Second).Try(func(attempt uint) error {
		if attempt > 0 {
			toolkit.logger.Warnf("==> Download failed, retrying ...")
		}
		return downloadFile(downloadURL, goArchiveDownloadPath)
	})
	if downloadErr != nil {
		return fmt.Errorf("download Go toolkit: %s", downloadErr)
	}

	toolkit.logger.Infof("=> Installing ...")
	if err := installGoTar(toolkit.logger, goArchiveDownloadPath); err != nil {
		return fmt.Errorf("install Go toolkit: %s", err)
	}
	if err := os.Remove(goArchiveDownloadPath); err != nil {
		return fmt.Errorf("remove the downloaded Go archive at %s: %s", goArchiveDownloadPath, err)
	}
	toolkit.logger.Infof("=> Installing DONE")

	return nil
}

func goBuildStep(logger stepman.Logger, cmdRunner commandRunner, goConfig GoConfigurationModel, packageName, stepAbsDirPath, outputBinPath string) error {
	cmdBuilder := goCmdBuilder{goConfig: goConfig}

	if isGoPathModeStep(stepAbsDirPath) {
		logger.Debugf("[Go deps] Step requires GOPATH mode")

		logger.Debugf("[Go deps] Migrating Step to Go modules as Go installation does not support GOPATH mode")
		if err := migrateToGoModules(stepAbsDirPath, packageName); err != nil {
			return fmt.Errorf("failed to migrate to go modules: %v", err)
		}

		buildCmd := cmdBuilder.goBuildMigratedModules(stepAbsDirPath, outputBinPath)
		if _, err := cmdRunner.runForOutput(buildCmd); err != nil {
			return fmt.Errorf("failed to build Step in directory (%s): %v", stepAbsDirPath, err)
		}

		return nil
	}

	logger.Debugf("[Go deps] Step requires Go modules mode")
	buildCmd := cmdBuilder.goBuildInModuleMode(stepAbsDirPath, outputBinPath)
	if _, err := cmdRunner.runForOutput(buildCmd); err != nil {
		return fmt.Errorf("failed to build Step in directory (%s): %v", stepAbsDirPath, err)
	}

	return nil
}

// stepIDorURI : doesn't work for "path::./" yet!!
func stepBinaryFilename(sIDData stepid.CanonicalID) string {
	//
	replaceRexp := regexp.MustCompile("[^A-Za-z0-9.-]")
	compositeStepID := fmt.Sprintf("%s-%s", sIDData.SteplibSource, sIDData.IDorURI)
	if sIDData.Version != "" {
		compositeStepID += "-" + sIDData.Version
	}

	safeStepID := replaceRexp.ReplaceAllString(compositeStepID, "_")
	//
	return safeStepID
}

func stepBinaryCacheFullPath(sIDData stepid.CanonicalID) string {
	return filepath.Join(goToolkitCacheRootPath(), stepBinaryFilename(sIDData))
}

// PrepareForStepRun ...
func (toolkit GoToolkit) PrepareForStepRun(step models.StepModel, sIDData stepid.CanonicalID, stepAbsDirPath string) error {
	fullStepBinPath := stepBinaryCacheFullPath(sIDData)

	// try to use cached binary, if possible
	if sIDData.IsUniqueResourceID() {
		if exists, err := pathutil.IsPathExists(fullStepBinPath); err != nil {
			toolkit.logger.Warnf("Failed to check cached binary for step, error: %s", err)
		} else if exists {
			return nil
		}
	}

	// it's not cached, so compile it
	if step.Toolkit == nil {
		return errors.New("no toolkit information specified in step")
	}
	if step.Toolkit.Go == nil {
		return errors.New("no toolkit.go information specified in step")
	}

	isInstallRequired, _, goConfig, err := selectGoConfiguration(toolkit.logger)
	if err != nil {
		return fmt.Errorf("select an appropriate Go installation for compiling the Step: %s", err)
	}
	if isInstallRequired {
		return fmt.Errorf("select an appropriate Go installation for compiling the Step: %s",
			"found Go version is older than required. Please run 'bitrise setup' to check and install the required version")
	}

	return goBuildStep(toolkit.logger, newDefaultRunner(toolkit.logger), goConfig, step.Toolkit.Go.PackageName, stepAbsDirPath, fullStepBinPath)
}

// === Toolkit: Step Run ===

// StepRunCommandArguments ...
func (toolkit GoToolkit) StepRunCommandArguments(_ models.StepModel, sIDData stepid.CanonicalID, stepAbsDirPath string) ([]string, error) {
	fullStepBinPath := stepBinaryCacheFullPath(sIDData)
	return []string{fullStepBinPath}, nil
}

// === Toolkit path utility function ===

func goToolkitRootPath() string {
	return toolkitDir("go")
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

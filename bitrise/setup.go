package bitrise

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/toolkits"
)

type SetupMode string

const (
	SetupModeDefault SetupMode = "default"
	SetupModeMinimal SetupMode = "minimal"
)

const (
	minEnvmanVersion  = "2.4.3"
	minStepmanVersion = "0.16.3"
)

type PluginDependency struct {
	Source     string
	MinVersion string
}

var PluginDependencyMap = map[string]PluginDependency{
	"init": {
		Source:     "https://github.com/bitrise-io/bitrise-plugins-init.git",
		MinVersion: "1.10.0",
	},
	"step": {
		Source:     "https://github.com/bitrise-io/bitrise-plugins-step.git",
		MinVersion: "0.10.4",
	},
	"workflow-editor": {
		Source:     "https://github.com/bitrise-io/bitrise-workflow-editor.git",
		MinVersion: "1.3.305",
	},
}

func RunSetupIfNeeded() error {
	if !configs.CheckIsSetupWasDoneForVersion(version.VERSION) {
		log.Warnf("Setup was not performed for this version of bitrise, doing it now...")
		return RunSetup(version.VERSION, SetupModeDefault, false)
	}
	return nil
}

func RunSetup(appVersion string, setupMode SetupMode, doCleanSetup bool) error {
	log.Infof("Setup Bitrise tools...")
	log.Printf("Clean before setup: %v", doCleanSetup)
	log.Printf("Setup mode: %s", setupMode)
	log.Printf("Is CI mode: %v", configs.IsCIMode)
	log.Printf("Is offline Steplib mode: %v", configs.IsSteplibOfflineMode)
	log.Printf("System: %s/%s", runtime.GOOS, runtime.GOARCH)

	if doCleanSetup {
		if err := configs.DeleteBitriseConfigDir(); err != nil {
			return err
		}

		if err := configs.InitPaths(); err != nil {
			return err
		}

		if err := plugins.InitPaths(); err != nil {
			return err
		}
	}

	if err := doSetupBitriseCoreTools(); err != nil {
		return fmt.Errorf("Failed to do common/platform independent setup, error: %s", err)
	}

	deps := []string{}
	if configs.IsSteplibOfflineMode {
		deps = append(deps, "zstd")
	}

	switch runtime.GOOS {
	case "darwin":
		if err := doSetupOnOSX(false, deps); err != nil {
			return fmt.Errorf("Failed to do macOS-specific setup, error: %s", err)
		}
	case "linux":
		if err := doSetupOnLinux(false, deps); err != nil {
			return fmt.Errorf("Failed to do Linux-specific setup: %s", err)
		}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if setupMode != SetupModeMinimal {
		if err := doSetupPlugins(); err != nil {
			return fmt.Errorf("Failed to do Plugins setup, error: %s", err)
		}
	}

	if err := doSetupToolkits(); err != nil {
		return fmt.Errorf("Failed to do Toolkits setup, error: %s", err)
	}

	log.Print()
	log.Donef("All the required tools are installed! We're ready to rock!!")

	if err := configs.SaveSetupSuccessForVersion(appVersion); err != nil {
		return fmt.Errorf("failed to save setup-success into config file, error: %s", err)
	}

	return nil
}

func doSetupToolkits() error {
	log.Print()
	log.Infof("Checking Bitrise Toolkits...")

	coreToolkits := toolkits.AllSupportedToolkits()

	for _, aCoreTK := range coreToolkits {
		toolkitName := aCoreTK.ToolkitName()
		isInstallRequired, checkResult, err := aCoreTK.Check()
		if err != nil {
			return fmt.Errorf("Failed to perform toolkit check (%s), error: %s", toolkitName, err)
		}

		if isInstallRequired {
			log.Warnf("No installed/suitable %s found, installing toolkit ...", toolkitName)
			if err := aCoreTK.Install(); err != nil {
				return fmt.Errorf("Failed to install toolkit (%s), error: %s", toolkitName, err)
			}

			isInstallRequired, checkResult, err = aCoreTK.Check()
			if err != nil {
				return fmt.Errorf("Failed to perform toolkit check (%s), error: %s", toolkitName, err)
			}
		}
		if isInstallRequired {
			return fmt.Errorf("Toolkit (%s) still reports that it isn't (properly) installed", toolkitName)
		}

		log.Printf("%s %s (%s): %s", colorstring.Green("[OK]"), toolkitName, checkResult.Version, checkResult.Path)
	}

	return nil
}

func doSetupPlugins() error {
	log.Print()
	log.Infof("Checking Bitrise Plugins...")

	for pluginName, pluginDependency := range PluginDependencyMap {
		if err := CheckIsPluginInstalled(pluginName, pluginDependency); err != nil {
			return fmt.Errorf("Plugin (%s) failed to install: %s", pluginName, err)
		}
	}

	return nil
}

func doSetupBitriseCoreTools() error {
	log.Print()
	log.Infof("Checking Bitrise Core tools...")

	if err := CheckIsEnvmanInstalled(minEnvmanVersion); err != nil {
		return fmt.Errorf("Envman failed to install: %s", err)
	}

	if err := CheckIsStepmanInstalled(minStepmanVersion); err != nil {
		return fmt.Errorf("Stepman failed to install: %s", err)
	}

	return nil
}

func doSetupOnOSX(isCIMode bool, brewDeps []string) error {
	log.Print()
	log.Infof("Doing macOS-specific setup")
	log.Printf("Checking required tools...")

	if err := CheckIsHomebrewInstalled(); err != nil {
		return errors.New(fmt.Sprint("Homebrew not installed or has some issues. Please fix these before calling setup again. Err:", err))
	}

	for _, dep := range brewDeps {
		log.Printf("Checking brew dependency %s", dep)
		if err := InstallWithBrewIfNeeded(models.BrewDepModel{Name: dep}, isCIMode); err != nil {
			return fmt.Errorf("Failed to install package (%s): %w", dep, err)
		}

		log.Printf("%s %s", colorstring.Green("[OK]"), dep)
	}

	return nil
}

func doSetupOnLinux(isCIMode bool, deps []string) error {
	log.Print()
	log.Infof("Doing Linux-specific setup")
	log.Printf("Checking required tools...")

	for _, dep := range deps {
		log.Printf("Checking APT dependency %s", dep)
		if err := InstallWithAptGetIfNeeded(models.AptGetDepModel{Name: dep}, isCIMode); err != nil {
			return fmt.Errorf("Failed to install package (%s): %w", dep, err)
		}

		log.Printf("%s %s", colorstring.Green("[OK]"), dep)
	}

	return nil
}

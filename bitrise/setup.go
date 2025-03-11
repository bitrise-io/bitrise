package bitrise

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/stepman/toolkits"
)

type SetupMode string

const (
	SetupModeDefault SetupMode = "default"
	SetupModeMinimal SetupMode = "minimal"
)

const (
	minEnvmanVersion  = "2.5.3"
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

func RunSetupIfNeeded(logger log.Logger) error {
	versionMatch, setupVersion := configs.CheckIsSetupWasDoneForVersion(version.VERSION)
	if setupVersion == "" {
		log.Infof("No setup was done yet, running setup now...")
		return RunSetup(logger, version.VERSION, SetupModeMinimal, false)
	}
	if !versionMatch {
		log.Infof("Setup was last performed for version %s, current version is %s. Re-running setup now...", setupVersion, version.VERSION)
		return RunSetup(logger, version.VERSION, SetupModeMinimal, false)
	}
	return nil
}

func RunSetup(logger log.Logger, appVersion string, setupMode SetupMode, doCleanSetup bool) error {
	log.Infof("Setup Bitrise tools...")
	log.Printf("Clean before setup: %v", doCleanSetup)
	log.Printf("Setup mode: %s", setupMode)
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

	switch runtime.GOOS {
	case "darwin":
		if err := doSetupOnOSX(); err != nil {
			return fmt.Errorf("Failed to do macOS-specific setup, error: %s", err)
		}
	case "linux":
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if setupMode != SetupModeMinimal {
		if err := doSetupPlugins(); err != nil {
			return fmt.Errorf("Failed to do Plugins setup, error: %s", err)
		}
	}

	if err := doSetupToolkits(logger); err != nil {
		return fmt.Errorf("Failed to do Toolkits setup, error: %s", err)
	}

	log.Print()
	log.Donef("All the required tools are installed! We're ready to rock!!")

	if err := configs.SaveSetupSuccessForVersion(appVersion); err != nil {
		return fmt.Errorf("failed to save setup-success into config file, error: %s", err)
	}

	return nil
}

func doSetupToolkits(logger log.Logger) error {
	log.Print()
	log.Infof("Checking Bitrise Toolkits...")

	coreToolkits := toolkits.AllSupportedToolkits(logger)

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
		return fmt.Errorf("failed to install envman: %s", err)
	}

	if err := CheckIsStepmanInstalled(minStepmanVersion); err != nil {
		return fmt.Errorf("failed to install stepman: %s", err)
	}

	return nil
}

func doSetupOnOSX() error {
	log.Print()
	log.Infof("Doing macOS-specific setup")
	log.Printf("Checking required tools...")

	if err := CheckIsHomebrewInstalled(); err != nil {
		return errors.New(fmt.Sprint("Homebrew not installed or has some issues. Please fix these before calling setup again. Err:", err))
	}

	return nil
}

package bitrise

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/bitrise/toolkits"
	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
)

const (
	minEnvmanVersion  = "2.3.0"
	minStepmanVersion = "0.13.0"
)

// PluginDependency ..
type PluginDependency struct {
	Source     string
	MinVersion string
}

// PluginDependencyMap ...
var PluginDependencyMap = map[string]PluginDependency{
	"init": {
		Source:     "https://github.com/bitrise-io/bitrise-plugins-init.git",
		MinVersion: "1.3.3",
	},
	"step": {
		Source:     "https://github.com/bitrise-io/bitrise-plugins-step.git",
		MinVersion: "0.9.10",
	},
	"workflow-editor": {
		Source:     "https://github.com/bitrise-io/bitrise-workflow-editor.git",
		MinVersion: "1.3.65",
	},
	"analytics": {
		Source:     "https://github.com/bitrise-io/bitrise-plugins-analytics.git",
		MinVersion: "0.12.3",
	},
}

// RunSetupIfNeeded ...
func RunSetupIfNeeded(appVersion string, isFullSetupMode bool) error {
	if !configs.CheckIsSetupWasDoneForVersion(version.VERSION) {
		log.Warnf(colorstring.Yellow("Setup was not performed for this version of bitrise, doing it now..."))
		return RunSetup(version.VERSION, false, false)
	}
	return nil
}

// RunSetup ...
func RunSetup(appVersion string, isFullSetupMode bool, isCleanSetupMode bool) error {
	log.Infof("Setup")
	log.Printf("Full setup: %v", isFullSetupMode)
	log.Printf("Clean setup: %v", isCleanSetupMode)
	log.Printf("Detected OS: %s", runtime.GOOS)

	if isCleanSetupMode {
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
		if err := doSetupOnOSX(isFullSetupMode); err != nil {
			return fmt.Errorf("Failed to do MacOS specific setup, error: %s", err)
		}
	case "linux":
	default:
		return errors.New("unsupported platform :(")
	}

	if err := doSetupPlugins(); err != nil {
		return fmt.Errorf("Failed to do Plugins setup, error: %s", err)
	}

	if err := doSetupToolkits(); err != nil {
		return fmt.Errorf("Failed to do Toolkits setup, error: %s", err)
	}

	fmt.Println()
	log.Donef("All the required tools are installed! We're ready to rock!!")

	if err := configs.SaveSetupSuccessForVersion(appVersion); err != nil {
		return fmt.Errorf("failed to save setup-success into config file, error: %s", err)
	}

	return nil
}

func doSetupToolkits() error {
	fmt.Println()
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
	fmt.Println()
	log.Infof("Checking Bitrise Plugins...")

	for pluginName, pluginDependency := range PluginDependencyMap {
		if err := CheckIsPluginInstalled(pluginName, pluginDependency); err != nil {
			return fmt.Errorf("Plugin (%s) failed to install: %s", pluginName, err)
		}
	}

	return nil
}

func doSetupBitriseCoreTools() error {
	fmt.Println()
	log.Infof("Checking Bitrise Core tools...")

	if err := CheckIsEnvmanInstalled(minEnvmanVersion); err != nil {
		return fmt.Errorf("Envman failed to install: %s", err)
	}

	if err := CheckIsStepmanInstalled(minStepmanVersion); err != nil {
		return fmt.Errorf("Stepman failed to install: %s", err)
	}

	return nil
}

func doSetupOnOSX(isMinimalSetupMode bool) error {
	fmt.Println()
	log.Infof("Doing OS X specific setup")
	log.Printf("Checking required tools...")

	if err := CheckIsHomebrewInstalled(isMinimalSetupMode); err != nil {
		return errors.New(fmt.Sprint("Homebrew not installed or has some issues. Please fix these before calling setup again. Err:", err))
	}

	if err := PrintInstalledXcodeInfos(); err != nil {
		return errors.New(fmt.Sprint("Failed to detect installed Xcode and Xcode Command Line Tools infos. Err:", err))
	}
	return nil
}

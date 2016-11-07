package bitrise

import (
	"errors"
	"fmt"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/toolkits"
	"github.com/bitrise-io/go-utils/colorstring"
)

const (
	minEnvmanVersion  = "1.1.1"
	minStepmanVersion = "0.9.25"
)

// PluginDependency ..
type PluginDependency struct {
	Source     string
	MinVersion string
}

// PluginDependencyMap ...
var PluginDependencyMap = map[string]PluginDependency{
	"analytics": PluginDependency{
		Source:     "https://github.com/bitrise-core/bitrise-plugins-analytics.git",
		MinVersion: "0.9.5",
	},
}

// RunSetup ...
func RunSetup(appVersion string, isFullSetupMode bool) error {
	log.Infoln("Setup")
	log.Infof("Full setup: %v", isFullSetupMode)
	log.Infoln("Detected OS:", runtime.GOOS)

	if err := doSetupBitriseCoreTools(); err != nil {
		return fmt.Errorf("Failed to do common/platform independent setup, error: %s", err)
	}

	switch runtime.GOOS {
	case "darwin":
		if err := doSetupOnOSX(isFullSetupMode); err != nil {
			return fmt.Errorf("Failed to do MacOS specific setup, error: %s", err)
		}
	default:
		return errors.New("unsupported platform :(")
	}

	if err := doSetupPlugins(); err != nil {
		return fmt.Errorf("Failed to do Plugins setup, error: %s", err)
	}

	if err := doSetupToolkits(); err != nil {
		return fmt.Errorf("Failed to do Toolkits setup, error: %s", err)
	}

	log.Infoln("All the required tools are installed!")

	if err := configs.SaveSetupSuccessForVersion(appVersion); err != nil {
		return fmt.Errorf("failed to save setup-success into config file, error: %s", err)
	}

	// guide
	log.Infoln("We're ready to rock!!")
	fmt.Println()

	return nil
}

func doSetupToolkits() error {
	log.Infoln("Checking Bitrise Toolkits...")

	coreToolkits := toolkits.AllSupportedToolkits()

	for _, aCoreTK := range coreToolkits {
		toolkitName := aCoreTK.ToolkitName()
		isInstallRequired, checkResult, err := aCoreTK.Check()
		if err != nil {
			return fmt.Errorf("Failed to perform toolkit check (%s), error: %s", toolkitName, err)
		}

		if isInstallRequired {
			log.Infoln("No installed/suitable '" + toolkitName + "' found, installing toolkit ...")
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

		log.Infoln(" * "+colorstring.Green("[OK]")+" "+toolkitName+" :", checkResult.Path)
		log.Infoln("        version :", checkResult.Version)
	}

	return nil
}

func doSetupPlugins() error {
	log.Infoln("Checking Bitrise Plugins...")

	for pluginName, pluginDependency := range PluginDependencyMap {
		if err := CheckIsPluginInstalled(pluginName, pluginDependency); err != nil {
			return fmt.Errorf("Plugin (%s) failed to install: %s", pluginName, err)
		}
	}

	return nil
}

func doSetupBitriseCoreTools() error {
	log.Infoln("Checking Bitrise Core tools...")

	if err := CheckIsEnvmanInstalled(minEnvmanVersion); err != nil {
		return fmt.Errorf("Envman failed to install: %s", err)
	}

	if err := CheckIsStepmanInstalled(minStepmanVersion); err != nil {
		return fmt.Errorf("Stepman failed to install: %s", err)
	}

	return nil
}

func doSetupOnOSX(isMinimalSetupMode bool) error {
	log.Infoln("Doing OS X specific setup")
	log.Infoln("Checking required tools...")

	if err := CheckIsHomebrewInstalled(isMinimalSetupMode); err != nil {
		return errors.New(fmt.Sprint("Homebrew not installed or has some issues. Please fix these before calling setup again. Err:", err))
	}

	if err := PrintInstalledXcodeInfos(); err != nil {
		return errors.New(fmt.Sprint("Failed to detect installed Xcode and Xcode Command Line Tools infos. Err:", err))
	}
	return nil
}

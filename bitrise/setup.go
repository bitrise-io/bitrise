package bitrise

import (
	"errors"
	"fmt"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/configs"
)

const (
	minEnvmanVersion  = "1.1.0"
	minStepmanVersion = "0.9.21"
)

// PluginDependency ..
type PluginDependency struct {
	Source     string
	Binary     string
	MinVersion string
}

// OSXPluginDependencyMap ...
var OSXPluginDependencyMap = map[string]PluginDependency{
	"analytics": PluginDependency{
		Source:     "https://github.com/bitrise-core/bitrise-plugins-analytics.git",
		Binary:     "https://github.com/bitrise-core/bitrise-plugins-analytics/releases/download/0.9.4/analytics-Darwin-x86_64",
		MinVersion: "0.9.4",
	},
}

// LinuxPluginDependencyMap ...
var LinuxPluginDependencyMap = map[string]PluginDependency{
	"analytics": PluginDependency{
		Source:     "https://github.com/bitrise-core/bitrise-plugins-analytics.git",
		Binary:     "https://github.com/bitrise-core/bitrise-plugins-analytics/releases/download/0.9.4/analytics-Linux-x86_64",
		MinVersion: "0.9.4",
	},
}

// RunSetup ...
func RunSetup(appVersion string, isFullSetupMode bool) error {
	log.Infoln("Setup")
	log.Infof("Full setup: %v", isFullSetupMode)
	log.Infoln("Detected OS:", runtime.GOOS)
	switch runtime.GOOS {
	case "darwin":
		if err := doSetupOnOSX(isFullSetupMode); err != nil {
			return err
		}
	case "linux":
		if err := doSetupOnLinux(); err != nil {
			return err
		}
	default:
		return errors.New("unsupported platform :(")
	}

	if err := configs.SaveSetupSuccessForVersion(appVersion); err != nil {
		return fmt.Errorf("failed to save setup-success into config file, error: %s", err)
	}

	// guide
	log.Infoln("We're ready to rock!!")
	fmt.Println()

	return nil
}

//
// install with brew example
//
// func checkIsAnsibleInstalled() error {
// 	progInstallPth, err := checkProgramInstalledPath("ansible")
// 	if err != nil {
// 		officialSiteURL := "http://www.ansible.com/home"
// 		officialGitHubURL := "https://github.com/ansible/ansible"
// 		log.Infoln("")
// 		log.Infoln("Ansible was not found.")
// 		log.Infoln("Ansible is used for system provisioning.")
// 		log.Infoln("You can find more information on Ansible's official website:", officialSiteURL)
// 		log.Infoln(" or on it's GitHub page:", officialGitHubURL)
// 		log.Infoln("You can install Ansible through brew:")
// 		log.Infoln("$ brew update && brew install ansible")
// 		isInstall, err := goinp.AskForBool("Would you like to install Ansible right now?")
// 		if err != nil {
// 			return err
// 		}
// 		if !isInstall {
// 			return errors.New("Ansible not found and install was not initiated.")
// 		}
//
// 		// Install
// 		log.Infoln("$ brew update --verbose")
// 		if err := RunCommand("brew", "update", "--verbose"); err != nil {
// 			return err
// 		}
// 		log.Infoln("$ brew install ansible")
// 		if err := RunCommand("brew", "install", "ansible"); err != nil {
// 			return err
// 		}
//
// 		// just check again
// 		return checkIsAnsibleInstalled()
// 	}
// 	log.Infoln(" * [OK] Ansible :", progInstallPth)
// 	return nil
// }

func doSetupOnOSX(isMinimalSetupMode bool) error {
	log.Infoln("Doing OS X specific setup")
	log.Infoln("Checking required tools...")
	if err := CheckIsHomebrewInstalled(isMinimalSetupMode); err != nil {
		return errors.New(fmt.Sprint("Homebrew not installed or has some issues. Please fix these before calling setup again. Err:", err))
	}

	if err := PrintInstalledXcodeInfos(); err != nil {
		return errors.New(fmt.Sprint("Failed to detect installed Xcode and Xcode Command Line Tools infos. Err:", err))
	}
	// if err := CheckIsXcodeCLTInstalled(); err != nil {
	// 	return errors.New(fmt.Sprint("Xcode Command Line Tools not installed. Err:", err))
	// }
	// if err := checkIsAnsibleInstalled(); err != nil {
	// 	return errors.New("Ansible failed to install")
	// }

	if err := CheckIsEnvmanInstalled(minEnvmanVersion); err != nil {
		return fmt.Errorf("Envman failed to install: %s", err)
	}
	if err := CheckIsStepmanInstalled(minStepmanVersion); err != nil {
		return fmt.Errorf("Stepman failed to install: %s", err)
	}
	for pluginName, pluginDependency := range OSXPluginDependencyMap {
		if err := CheckIsPluginInstalled(pluginName, pluginDependency); err != nil {
			return fmt.Errorf("Plugin (%s) failed to install: %s", pluginName, err)
		}
	}

	log.Infoln("All the required tools are installed!")

	return nil
}

func doSetupOnLinux() error {
	log.Infoln("Doing Linux specific setup")
	log.Infoln("Checking required tools...")

	if err := CheckIsEnvmanInstalled(minEnvmanVersion); err != nil {
		return fmt.Errorf("Envman failed to install: %s", err)
	}
	if err := CheckIsStepmanInstalled(minStepmanVersion); err != nil {
		return fmt.Errorf("Stepman failed to install: %s", err)
	}
	for pluginName, pluginDependency := range LinuxPluginDependencyMap {
		if err := CheckIsPluginInstalled(pluginName, pluginDependency); err != nil {
			return fmt.Errorf("Plugin (%s) failed to install: %s", pluginName, err)
		}
	}

	log.Infoln("All the required tools are installed!")

	return nil
}

package bitrise

import (
	"errors"
	"fmt"
	"runtime"

	log "github.com/Sirupsen/logrus"
)

const (
	minEnvmanVersion  = "0.9.7"
	minStepmanVersion = "0.9.12"
)

// RunSetup ...
func RunSetup(appVersion string, isMinimalSetupMode bool) error {
	log.Infoln("[BITRISE_CLI] - Setup")
	log.Infoln("Detected OS:", runtime.GOOS)
	switch runtime.GOOS {
	case "darwin":
		if err := doSetupOnOSX(isMinimalSetupMode); err != nil {
			return err
		}
	case "linux":
		if err := doSetupOnLinux(); err != nil {
			return err
		}
	default:
		return errors.New("Sorry, unsupported platform :(")
	}

	if err := SaveSetupSuccessForVersion(appVersion); err != nil {
		return fmt.Errorf("Failed to save setup-success into config file: %s", err)
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
		return errors.New(fmt.Sprint("Envman failed to install:", err))
	}
	if err := CheckIsStepmanInstalled(minStepmanVersion); err != nil {
		return errors.New(fmt.Sprint("Stepman failed to install:", err))
	}
	log.Infoln("All the required tools are installed!")

	return nil
}

func doSetupOnLinux() error {
	return errors.New("doSetupOnLinux -- Coming soon")
}

package cli

import (
	"errors"
	"fmt"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/dependencies"
	"github.com/codegangsta/cli"
)

const (
	minEnvmanVersion  = "0.9.1"
	minStepmanVersion = "0.9.5"
)

// PrintBitriseHeaderASCIIArt ...
func PrintBitriseHeaderASCIIArt() {
	// generated here: http://patorjk.com/software/taag/#p=display&f=ANSI%20Shadow&t=Bitrise
	fmt.Println(`
  ██████╗ ██╗████████╗██████╗ ██╗███████╗███████╗
  ██╔══██╗██║╚══██╔══╝██╔══██╗██║██╔════╝██╔════╝
  ██████╔╝██║   ██║   ██████╔╝██║███████╗█████╗
  ██╔══██╗██║   ██║   ██╔══██╗██║╚════██║██╔══╝
  ██████╔╝██║   ██║   ██║  ██║██║███████║███████╗
  ╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚══════╝╚══════╝`)
	fmt.Println()
}

func doSetup(c *cli.Context) {
	PrintBitriseHeaderASCIIArt()

	log.Infoln("[BITRISE_CLI] - Setup")
	log.Infoln("Detected OS:", runtime.GOOS)
	switch runtime.GOOS {
	case "darwin":
		if err := doSetupOnOSX(); err != nil {
			log.Fatalln("Setup failed:", err)
		}
	case "linux":
		if err := doSetupOnLinux(); err != nil {
			log.Fatalln("Setup failed:", err)
		}
	default:
		log.Fatalln("Sorry, unsupported platform :(")
	}

	// guide
	fmt.Println()
	log.Infoln("We're ready to rock!!")
	fmt.Println()
	log.Infoln("To start using bitrise-cli:")
	log.Infoln("* cd into your project's directory (if you're not there already)")
	log.Infoln("* call: bitrise-cli init")
	log.Infoln("* follow the guide")
	fmt.Println()
	log.Infoln("That's all :)")
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
// 		if err := bitrise.RunCommand("brew", "update", "--verbose"); err != nil {
// 			return err
// 		}
// 		log.Infoln("$ brew install ansible")
// 		if err := bitrise.RunCommand("brew", "install", "ansible"); err != nil {
// 			return err
// 		}
//
// 		// just check again
// 		return checkIsAnsibleInstalled()
// 	}
// 	log.Infoln(" * [OK] Ansible :", progInstallPth)
// 	return nil
// }

func doSetupOnOSX() error {
	log.Infoln("Doing OS X specific setup")
	log.Infoln("Checking required tools...")
	if err := dependencies.CheckIsXcodeCLTInstalled(); err != nil {
		return errors.New(fmt.Sprint("Xcode Command Line Tools not installed. Err:", err))
	}
	if err := dependencies.CheckIsHomebrewInstalled(); err != nil {
		return errors.New(fmt.Sprint("Homebrew not installed. Err:", err))
	}
	// if err := checkIsAnsibleInstalled(); err != nil {
	// 	return errors.New("Ansible failed to install")
	// }
	if err := dependencies.CheckIsEnvmanInstalled(minEnvmanVersion); err != nil {
		return errors.New(fmt.Sprint("Envman failed to install:", err))
	}
	if err := dependencies.CheckIsStepmanInstalled(minStepmanVersion); err != nil {
		return errors.New(fmt.Sprint("Stepman failed to install:", err))
	}
	log.Infoln("All the required tools are installed!")
	return nil
}

func doSetupOnLinux() error {
	return errors.New("doSetupOnLinux -- Coming soon")
}

package cli

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/bitrise"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/codegangsta/cli"
)

func doSetup(c *cli.Context) {
	log.Infoln("[BITRISE_CLI] - Setup - Still work-in-progress")
	log.Infoln("Detected OS:", runtime.GOOS)
	switch runtime.GOOS {
	case "darwin":
		doSetupOnOSX()
	case "linux":
		doSetupOnLinux()
	default:
		log.Fatalln("Sorry, unsupported platform :(")
	}
	os.Exit(1)
}

func checkProgramInstalledPath(clcommand string) (string, error) {
	cmd := exec.Command("which", clcommand)
	cmd.Stderr = os.Stderr
	outBytes, err := cmd.Output()
	outStr := string(outBytes)
	return strings.TrimSpace(outStr), err

	// var stdoutBuff bytes.Buffer
	// cmd := exec.Command("which", clcommand)
	// cmd.Stdin = os.Stdin
	// cmd.Stdout = &stdoutBuff
	// cmd.Stderr = os.Stderr
	// cmdErr := cmd.Run()
	// return string(stdoutBuff.Bytes()), cmdErr
}

func checkIsHomebrewInstalled() error {
	brewRubyInstallCmdString := `$ ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"`
	officialSiteURL := "http://brew.sh/"

	progInstallPth, err := checkProgramInstalledPath("brew")
	if err != nil {
		log.Infoln("")
		log.Infoln("It seems that Homebrew is not installed on your system.")
		log.Infoln("Homebrew (short: brew) is required in order to be able to auto-install all the bitrise dependencies.")
		log.Infoln("You should be able to install brew by copying this command and running it in your Terminal:")
		log.Infoln(brewRubyInstallCmdString)
		log.Infoln("You can find more information about Homebrew on it's official site at:", officialSiteURL)
		log.Infoln("Once the installation of brew is finished you can call the bitrise setup again.")
		return err
	}
	log.Infoln(" * [OK] Homebrew :", progInstallPth)
	return nil
}

func checkIsAnsibleInstalled() error {
	progInstallPth, err := checkProgramInstalledPath("ansible")
	if err != nil {
		officialSiteURL := "http://www.ansible.com/home"
		officialGitHubURL := "https://github.com/ansible/ansible"
		log.Infoln("")
		log.Infoln("Ansible was not found.")
		log.Infoln("Ansible is used for system provisioning.")
		log.Infoln("You can find more information on Ansible's official website:", officialSiteURL)
		log.Infoln(" or on it's GitHub page:", officialGitHubURL)
		log.Infoln("You can install Ansible through brew:")
		log.Infoln("$ brew update && brew install ansible")
		isInstall, err := goinp.AskForBool("Would you like to install Ansible right now?")
		if err != nil {
			return err
		}
		if !isInstall {
			return errors.New("Ansible not found and install was not initiated.")
		}

		log.Infoln("$ brew update --verbose")
		if err := bitrise.RunCommand("brew", "update", "--verbose"); err != nil {
			return err
		}
		log.Infoln("$ brew install ansible")
		if err := bitrise.RunCommand("brew", "install", "ansible"); err != nil {
			return err
		}

		// just check again
		return checkIsAnsibleInstalled()
	}
	log.Infoln(" * [OK] Ansible :", progInstallPth)
	return nil
}

func checkIsEnvmanInstalled() error {
	progInstallPth, err := checkProgramInstalledPath("envman")
	if err != nil {
		log.Infoln("")
		log.Infoln("envman was not found.")
		return errors.New("envman was not found")
	}
	log.Infoln(" * [OK] envman :", progInstallPth)
	return nil
}

func checkIsStepmanInstalled() error {
	progInstallPth, err := checkProgramInstalledPath("stepman")
	if err != nil {
		log.Infoln("")
		log.Infoln("stepman was not found.")
		return errors.New("stepman was not found")
	}
	log.Infoln(" * [OK] stepman :", progInstallPth)
	return nil
}

func doSetupOnOSX() {
	log.Infoln("Doing OS X specific setup")
	log.Infoln("Checking required tools...")
	if err := checkIsHomebrewInstalled(); err != nil {
		log.Fatalln("Failed:", err)
	}
	if err := checkIsAnsibleInstalled(); err != nil {
		log.Fatalln("Failed:", err)
	}
	if err := checkIsEnvmanInstalled(); err != nil {
		log.Fatalln("Failed:", err)
	}
	if err := checkIsStepmanInstalled(); err != nil {
		log.Fatalln("Failed:", err)
	}
	log.Infoln("All the required tools are installed!")
	log.Infoln("We're ready to rock!!")
}

func doSetupOnLinux() {
	log.Infoln("[BITRISE_CLI] - doSetupOnLinux -- Coming soon!")
}

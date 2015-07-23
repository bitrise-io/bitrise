package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/bitrise"
	"github.com/codegangsta/cli"
)

const (
	minEnvmanVersion  = "0.9.1"
	minStepmanVersion = "0.9.2"
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

func checkProgramInstalledPath(clcommand string) (string, error) {
	cmd := exec.Command("which", clcommand)
	cmd.Stderr = os.Stderr
	outBytes, err := cmd.Output()
	outStr := string(outBytes)
	return strings.TrimSpace(outStr), err
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
	verStr, err := bitrise.RunCommandAndReturnStdout("brew", "--version")
	if err != nil {
		log.Infoln("")
		return errors.New("Failed to get version")
	}
	log.Infoln(" * [OK] Homebrew :", progInstallPth)
	log.Infoln("        version :", verStr)
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

func checkIsBitriseToolInstalled(toolname, minVersion string, isInstall bool) error {
	doInstall := func() error {
		installCmdLines := []string{
			"curl -L https://github.com/bitrise-io/" + toolname + "/releases/download/" + minVersion + "/" + toolname + "-$(uname -s)-$(uname -m) > /usr/local/bin/" + toolname,
			"chmod +x /usr/local/bin/" + toolname,
		}
		officialGithub := "https://github.com/bitrise-io/" + toolname
		fmt.Println()
		log.Warnln("No supported " + toolname + " version found.")
		log.Infoln("You can find more information about "+toolname+" on it's official GitHub page:", officialGithub)
		fmt.Println()

		// Install
		log.Infoln("Installing...")
		fmt.Println(strings.Join(installCmdLines, "\n"))
		if err := bitrise.RunBashCommandLines(installCmdLines); err != nil {
			return err
		}

		// check again
		return checkIsBitriseToolInstalled(toolname, minVersion, false)
	}

	// check whether installed
	progInstallPth, err := checkProgramInstalledPath(toolname)
	if err != nil {
		if !isInstall {
			return err
		}

		return doInstall()
	}
	verStr, err := bitrise.RunCommandAndReturnStdout(toolname, "-version")
	if err != nil {
		log.Infoln("")
		return errors.New("Failed to get version")
	}

	// version check
	isVersionOk, err := bitrise.IsVersionGreaterOrEqual(verStr, minVersion)
	if err != nil {
		log.Error("Failed to validate installed version")
		return err
	}
	if !isVersionOk {
		log.Warn("Installed "+toolname+" found, but not a supported version: ", verStr)
		if !isInstall {
			return errors.New("Failed to install required version.")
		}
		log.Warn("Updating...")
		return doInstall()
	}

	log.Infoln(" * [OK] "+toolname+" :", progInstallPth)
	log.Infoln("        version :", verStr)
	return nil
}

func checkIsEnvmanInstalled() error {
	toolname := "envman"
	minVersion := minEnvmanVersion
	if err := checkIsBitriseToolInstalled(toolname, minVersion, true); err != nil {
		return err
	}
	return nil
}

func checkIsStepmanInstalled() error {
	toolname := "stepman"
	minVersion := minStepmanVersion
	if err := checkIsBitriseToolInstalled(toolname, minVersion, true); err != nil {
		return err
	}
	return nil
}

func doSetupOnOSX() error {
	log.Infoln("Doing OS X specific setup")
	log.Infoln("Checking required tools...")
	if err := checkIsHomebrewInstalled(); err != nil {
		return errors.New("Homebrew failed to install")
	}
	// if err := checkIsAnsibleInstalled(); err != nil {
	// 	return errors.New("Ansible failed to install")
	// }
	if err := checkIsEnvmanInstalled(); err != nil {
		return errors.New(fmt.Sprint("Envman failed to install:", err))
	}
	if err := checkIsStepmanInstalled(); err != nil {
		return errors.New(fmt.Sprint("Stepman failed to install:", err))
	}
	log.Infoln("All the required tools are installed!")
	return nil
}

func doSetupOnLinux() error {
	return errors.New("doSetupOnLinux -- Coming soon")
}

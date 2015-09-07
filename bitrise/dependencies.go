package bitrise

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/versions"
	"github.com/bitrise-io/goinp/goinp"
)

// CheckProgramInstalledPath ...
func CheckProgramInstalledPath(clcommand string) (string, error) {
	cmd := exec.Command("which", clcommand)
	cmd.Stderr = os.Stderr
	outBytes, err := cmd.Output()
	outStr := string(outBytes)
	return strings.TrimSpace(outStr), err
}

// CheckIsRubyGemsInstalled ...
func CheckIsRubyGemsInstalled() error {
	officialSiteURL := "https://rubygems.org"

	progInstallPth, err := CheckProgramInstalledPath("gem")
	if err != nil {
		fmt.Println()
		log.Warn("It seems that RubyGems is not installed on your system.")
		log.Infoln("RubyGems is required in order to be able to auto-install all the bitrise dependencies.")
		// log.Infoln("You should be able to install brew by copying this command and running it in your Terminal:")
		// log.Infoln(brewRubyInstallCmdString)
		log.Infoln("You can find more information about RubyGems on it's official site at:", officialSiteURL)
		log.Warn("Once the installation of RubyGems is finished you should call the bitrise setup again.")
		return err
	}
	verStr, err := cmdex.RunCommandAndReturnStdout("gem", "--version")
	if err != nil {
		log.Infoln("")
		return errors.New("Failed to get version")
	}
	log.Debugln(" * [OK] RubyGems :", progInstallPth)
	log.Debugln("        version :", verStr)
	return nil
}

// CheckIsHomebrewInstalled ...
func CheckIsHomebrewInstalled(isMinimalSetupMode bool) error {
	brewRubyInstallCmdString := `$ ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"`
	officialSiteURL := "http://brew.sh/"

	progInstallPth, err := CheckProgramInstalledPath("brew")
	if err != nil {
		fmt.Println()
		log.Warn("It seems that Homebrew is not installed on your system.")
		log.Infoln("Homebrew (short: brew) is required in order to be able to auto-install all the bitrise dependencies.")
		log.Infoln("You should be able to install brew by copying this command and running it in your Terminal:")
		log.Infoln(brewRubyInstallCmdString)
		log.Infoln("You can find more information about Homebrew on it's official site at:", officialSiteURL)
		log.Warn("Once the installation of brew is finished you should call the bitrise setup again.")
		return err
	}
	verStr, err := cmdex.RunCommandAndReturnStdout("brew", "--version")
	if err != nil {
		log.Infoln("")
		return errors.New("Failed to get version")
	}

	if !isMinimalSetupMode {
		// brew doctor
		doctorOutput, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("brew", "doctor")
		if err != nil {
			fmt.Println("")
			log.Warn("brew doctor returned an error:")
			log.Warnf("%s", doctorOutput)
			return errors.New("Failed to: brew doctor")
		}
	}

	log.Infoln(" * "+colorstring.Green("[OK]")+" Homebrew :", progInstallPth)
	log.Infoln("        version :", verStr)
	return nil
}

// PrintInstalledXcodeInfos ...
func PrintInstalledXcodeInfos() error {
	xcodeSelectPth, err := cmdex.RunCommandAndReturnStdout("xcode-select", "--print-path")
	if err != nil {
		xcodeSelectPth = "xcode-select --print-path failed to detect the location of activate Xcode Command Line Tools path"
	}

	progInstallPth, err := CheckProgramInstalledPath("xcodebuild")
	if err != nil {
		return errors.New("xcodebuild is not installed")
	}

	isFullXcodeAvailable := false
	verStr, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("xcodebuild", "-version")
	if err != nil {
		// No full Xcode available, only the Command Line Tools
		// verStr is something like "xcode-select: error: tool 'xcodebuild' requires Xcode, but active developer directory '/Library/Developer/CommandLineTools' is a command line tools instance"
		isFullXcodeAvailable = false
	} else {
		// version OK - full Xcode available
		//  we'll just format it a bit to fit into one line
		isFullXcodeAvailable = true
		verStr = strings.Join(strings.Split(verStr, "\n"), " | ")
	}

	log.Infoln(" * "+colorstring.Green("[OK]")+" xcodebuild path :", progInstallPth)
	if !isFullXcodeAvailable {
		log.Infoln("        version (xcodebuild) :", colorstring.Yellowf("%s", verStr))
	} else {
		log.Infoln("        version (xcodebuild) :", verStr)
	}
	log.Infoln("        active Xcode (Command Line Tools) path (xcode-select --print-path) :", xcodeSelectPth)
	if !isFullXcodeAvailable {
		log.Warn(colorstring.Yellowf("%s", "No Xcode found, only the Xcode Command Line Tools are available!"))
		log.Warn(colorstring.Yellowf("%s", "Full Xcode is required to build, test and archive iOS apps!"))
	}

	return nil
}

// // CheckIsXcodeCLTInstalled ...
// func CheckIsXcodeCLTInstalled() error {
// 	xcodeSelectPth, err := cmdex.RunCommandAndReturnStdout("xcode-select", "--print-path")
// 	if err != nil {
// 		fmt.Println()
// 		log.Warn("It seems that the Xcode Command Line Tools are not installed on your system.")
// 		fmt.Println()
// 		log.Infoln("If you use OS X Mavericks or a more recent OS X version")
// 		log.Infoln(" you can install it by running the following command in your Terminal:")
// 		log.Infoln("xcode-select --install")
// 		fmt.Println()
// 		log.Infoln("If you use OS X Mountain Lion or an earlier version of OS X then you'll")
// 		log.Infoln(" have to download and install the Xcode Command Line Tools")
// 		log.Infoln(" directly from the Apple Developer / Downloads Portal: https://developer.apple.com/downloads/")
// 		fmt.Println()
// 		log.Warn("Once the installation is finished you should call the bitrise setup again.")
// 		return errors.New("Failed to get Xcode Command Line Tools path")
// 	}
// 	progInstallPth, err := CheckProgramInstalledPath("xcodebuild")
// 	if err != nil {
// 		return errors.New("xcodebuild is not installed")
// 	}
//
// 	isFullXcodeAvailable := false
// 	verStr, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("xcodebuild", "-version")
// 	if err != nil {
// 		// No full Xcode available, only the Command Line Tools
// 		// verStr is something like "xcode-select: error: tool 'xcodebuild' requires Xcode, but active developer directory '/Library/Developer/CommandLineTools' is a command line tools instance"
// 		isFullXcodeAvailable = false
// 	} else {
// 		// version OK - full Xcode available
// 		//  we'll just format it a bit to fit into one line
// 		isFullXcodeAvailable = true
// 		verStr = strings.Join(strings.Split(verStr, "\n"), " | ")
// 	}
//
// 	log.Infoln(" * "+colorstring.Green("[OK]")+" xcodebuild path :", progInstallPth)
// 	if !isFullXcodeAvailable {
// 		log.Infoln("        version (xcodebuild) :", colorstring.Yellowf("%s", verStr))
// 	} else {
// 		log.Infoln("        version (xcodebuild) :", verStr)
// 	}
// 	log.Infoln("        active Xcode (Command Line Tools) path (xcode-select --print-path) :", xcodeSelectPth)
// 	if !isFullXcodeAvailable {
// 		log.Warn(colorstring.Yellowf("%s", "No Xcode found, only the Xcode Command Line Tools are available!"))
// 		log.Warn(colorstring.Yellowf("%s", "Full Xcode is required to build, test and archive iOS apps!"))
// 	}
//
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
		if err := cmdex.RunBashCommandLines(installCmdLines); err != nil {
			return err
		}

		// check again
		return checkIsBitriseToolInstalled(toolname, minVersion, false)
	}

	// check whether installed
	progInstallPth, err := CheckProgramInstalledPath(toolname)
	if err != nil {
		if !isInstall {
			return err
		}

		return doInstall()
	}
	verStr, err := cmdex.RunCommandAndReturnStdout(toolname, "-version")
	if err != nil {
		log.Infoln("")
		return errors.New("Failed to get version")
	}

	// version check
	isVersionOk, err := versions.IsVersionGreaterOrEqual(verStr, minVersion)
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

	log.Infoln(" * "+colorstring.Green("[OK]")+" "+toolname+" :", progInstallPth)
	log.Infoln("        version :", verStr)
	return nil
}

// CheckIsEnvmanInstalled ...
func CheckIsEnvmanInstalled(minEnvmanVersion string) error {
	toolname := "envman"
	minVersion := minEnvmanVersion
	if err := checkIsBitriseToolInstalled(toolname, minVersion, true); err != nil {
		return err
	}
	return nil
}

// CheckIsStepmanInstalled ...
func CheckIsStepmanInstalled(minStepmanVersion string) error {
	toolname := "stepman"
	minVersion := minStepmanVersion
	if err := checkIsBitriseToolInstalled(toolname, minVersion, true); err != nil {
		return err
	}
	return nil
}

func checkWithBrewProgramInstalled(tool string) error {
	args := []string{"list", tool}
	cmd := exec.Command("brew", args...)

	if outBytes, err := cmd.CombinedOutput(); err != nil {
		log.Debugf("%s", outBytes)
		return err
	}

	return nil
}

// InstallWithBrewIfNeeded ...
func InstallWithBrewIfNeeded(tool string, isCIMode bool) error {
	if err := checkWithBrewProgramInstalled(tool); err != nil {
		if !isCIMode {
			log.Infof("This step requires %s, which is not installed", tool)
			allow, err := goinp.AskForBool("Would you like to install (" + tool + ") with brew ? [yes/no]")
			if err != nil {
				return err
			}
			if !allow {
				return errors.New("(" + tool + ") is required for step")
			}
		}
		args := []string{"install", tool}
		log.Infof("Installing required dependency (%s) with brew ...", tool)
		if out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("brew", args...); err != nil {
			log.Errorf("Failed to install tool (%s)", tool)
			log.Errorf("Output was: %s", out)
			return err
		}
		log.Infof(" * "+colorstring.Green("[OK]")+" %s installed", tool)
		return nil
	}

	return nil
}

// DependencyTryCheckTool ...
func DependencyTryCheckTool(tool string) error {
	var cmd *exec.Cmd
	errMsg := ""

	switch tool {
	case "xcode":
		cmd = exec.Command("xcodebuild", "-version")
		errMsg = "The full Xcode app is not installed, required for this step. You can install it from the App Store."
		break
	default:
		cmdFields := strings.Fields(tool)
		if len(cmdFields) >= 2 {
			cmd = exec.Command(cmdFields[0], cmdFields[1:]...)
		} else if len(cmdFields) == 1 {
			cmd = exec.Command(cmdFields[0])
		} else {
			return fmt.Errorf("Invalid tool name (%s)", tool)
		}
	}

	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		if errMsg != "" {
			return errors.New(errMsg)
		}
		log.Infof("Output was: %s", outBytes)
		return fmt.Errorf("Dependency check failed for: %s", tool)
	}

	return nil
}

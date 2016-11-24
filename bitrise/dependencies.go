package bitrise

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/bitrise-io/bitrise/utils"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/progress"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-utils/versions"
	"github.com/bitrise-io/goinp/goinp"
	stepmanModels "github.com/bitrise-io/stepman/models"
	ver "github.com/hashicorp/go-version"
)

// CheckIsPluginInstalled ...
func CheckIsPluginInstalled(name string, dependency PluginDependency) error {
	_, found, err := plugins.LoadPlugin(name)
	if err != nil {
		return err
	}

	currentVersion := ""
	installOrUpdate := false

	if !found {
		fmt.Println()
		log.Warnf("Default plugin (%s) NOT found.", name)
		fmt.Println()

		fmt.Print("Installing...")
		installOrUpdate = true
		currentVersion = dependency.MinVersion
	} else {
		installedVersion, err := plugins.GetPluginVersion(name)
		if err != nil {
			return err
		}

		if installedVersion == nil {
			fmt.Println()
			log.Warnf("Default plugin (%s) is not installed from git, no version info available.", name)
			fmt.Println()

			currentVersion = "local"
		} else {
			currentVersion = installedVersion.String()

			minVersion, err := ver.NewVersion(dependency.MinVersion)
			if err != nil {
				return err
			}

			if installedVersion.LessThan(minVersion) {
				fmt.Println()
				log.Warnf("Default plugin (%s) version (%s) is lower than required (%s).", name, installedVersion.String(), minVersion.String())
				fmt.Println()

				log.Infoln("Updating...")
				installOrUpdate = true
				currentVersion = dependency.MinVersion
			}
		}
	}

	if installOrUpdate {
		var plugin plugins.Plugin
		err := progress.SimpleProgressE(".", 2*time.Second, func() error {
			return retry.Times(2).Wait(5 * time.Second).Try(func(attempt uint) error {
				if attempt > 0 {
					fmt.Println()
					fmt.Print("==> Download failed, retrying ...")
				}
				p, _, err := plugins.InstallPlugin(dependency.Source, dependency.MinVersion)
				plugin = p
				return err
			})
		})
		fmt.Println()

		if err != nil {
			return fmt.Errorf("Failed to install plugin, error: %s", err)
		}

		if len(plugin.Description) > 0 {
			fmt.Println()
			fmt.Println(plugin.Description)
			fmt.Println()
		}
	}

	pluginDir := plugins.GetPluginDir(name)

	log.Infof(" * %s Plugin (%s) : %s", colorstring.Green("[OK]"), name, pluginDir)
	log.Infof("        version : %s", currentVersion)

	return nil
}

// CheckIsRubyGemsInstalled ...
func CheckIsRubyGemsInstalled() error {
	officialSiteURL := "https://rubygems.org"

	progInstallPth, err := utils.CheckProgramInstalledPath("gem")
	if err != nil {
		fmt.Println()
		log.Warn("It seems that RubyGems is not installed on your system.")
		log.Infoln("RubyGems is required in order to be able to auto-install all the bitrise dependencies.")
		// log.Infoln("You should be able to install brew by copying this command and running it in your Terminal:")
		// log.Infoln(brewRubyInstallCmdString)
		log.Infoln("You can find more information about RubyGems on its official site at:", officialSiteURL)
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
func CheckIsHomebrewInstalled(isFullSetupMode bool) error {
	brewRubyInstallCmdString := `$ ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"`
	officialSiteURL := "http://brew.sh/"

	progInstallPth, err := utils.CheckProgramInstalledPath("brew")
	if err != nil {
		fmt.Println()
		log.Warn("It seems that Homebrew is not installed on your system.")
		log.Infoln("Homebrew (short: brew) is required in order to be able to auto-install all the bitrise dependencies.")
		log.Infoln("You should be able to install brew by copying this command and running it in your Terminal:")
		log.Infoln(brewRubyInstallCmdString)
		log.Infoln("You can find more information about Homebrew on its official site at:", officialSiteURL)
		log.Warn("Once the installation of brew is finished you should call the bitrise setup again.")
		return err
	}
	verStr, err := cmdex.RunCommandAndReturnStdout("brew", "--version")
	if err != nil {
		log.Infoln("")
		return errors.New("Failed to get version")
	}

	if isFullSetupMode {
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

	progInstallPth, err := utils.CheckProgramInstalledPath("xcodebuild")
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

func checkIsBitriseToolInstalled(toolname, minVersion string, isInstall bool) error {
	doInstall := func() error {
		officialGithub := "https://github.com/bitrise-io/" + toolname
		fmt.Println()
		log.Warnln("No supported " + toolname + " version found.")
		log.Infoln("You can find more information about "+toolname+" on its official GitHub page:", officialGithub)

		// Install
		fmt.Print("Installing...")
		err := progress.SimpleProgressE(".", 2*time.Second, func() error {
			return retry.Times(2).Wait(5 * time.Second).Try(func(attempt uint) error {
				if attempt > 0 {
					fmt.Println()
					fmt.Print("==> Download failed, retrying ...")
				}
				return tools.InstallToolFromGitHub(toolname, "bitrise-io", minVersion)
			})
		})
		fmt.Println()
		if err != nil {
			return err
		}

		// check again
		return checkIsBitriseToolInstalled(toolname, minVersion, false)
	}

	// check whether installed
	progInstallPth, err := utils.CheckProgramInstalledPath(toolname)
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
			return errors.New("Failed to install required version")
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

func checkIfBrewPackageInstalled(packageName string) bool {
	cmd := exec.Command("brew", "list", packageName)

	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		log.Debugf("%s", outBytes)
		return false
	}

	return len(outBytes) > 0
}

func checkIfAptPackageInstalled(packageName string) bool {
	cmd := exec.Command("dpkg", "-s", packageName)

	if outBytes, err := cmd.CombinedOutput(); err != nil {
		log.Debugf("%s", outBytes)
		return false
	}

	return true
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

// InstallWithBrewIfNeeded ...
func InstallWithBrewIfNeeded(brewDep stepmanModels.BrewDepModel, isCIMode bool) error {
	isDepInstalled := false
	// First do a "which", to see if the binary is available.
	// Can be available from another source, not just from brew,
	// e.g. it's common to use NVM or similar to install and manage the Node.js version.
	{
		if out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("which", brewDep.GetBinaryName()); err != nil {
			if err.Error() == "exit status 1" && out == "" {
				isDepInstalled = false
			} else {
				// unexpected `which` error
				return fmt.Errorf("which (%s) failed -- out: (%s) err: (%s)", brewDep.Name, out, err)
			}
		} else if out != "" {
			isDepInstalled = true
		} else {
			// no error but which's output was empty
			return fmt.Errorf("which (%s) failed -- no error (exit code 0) but output was empty", brewDep.Name)
		}
	}

	// then do a package manager specific lookup
	{
		if !isDepInstalled {
			// which did not find the binary, also check in brew,
			// whether the package is installed
			isDepInstalled = checkIfBrewPackageInstalled(brewDep.Name)
		}
	}

	if !isDepInstalled {
		// Tool isn't installed -- install it...
		if !isCIMode {
			log.Infof(`This step requires "%s" to be available, but it is not installed.`, brewDep.GetBinaryName())
			allow, err := goinp.AskForBoolWithDefault(`Would you like to install the "`+brewDep.Name+`" package with brew?`, true)
			if err != nil {
				return err
			}
			if !allow {
				return errors.New("(" + brewDep.Name + ") is required for step")
			}
		}

		log.Infof("(%s) isn't installed, installing...", brewDep.Name)
		if cmdOut, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("brew", "install", brewDep.Name); err != nil {
			log.Errorf("brew install %s failed -- out: (%s) err: (%s)", brewDep.Name, cmdOut, err)
			return err
		}
		log.Infof(" * "+colorstring.Green("[OK]")+" %s installed", brewDep.Name)
	}

	return nil
}

// InstallWithAptGetIfNeeded ...
func InstallWithAptGetIfNeeded(aptGetDep stepmanModels.AptGetDepModel, isCIMode bool) error {
	isDepInstalled := false
	// First do a "which", to see if the binary is available.
	// Can be available from another source, not just from brew,
	// e.g. it's common to use NVM or similar to install and manage the Node.js version.
	{
		if out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("which", aptGetDep.GetBinaryName()); err != nil {
			if err.Error() == "exit status 1" && out == "" {
				isDepInstalled = false
			} else {
				// unexpected `which` error
				return fmt.Errorf("which (%s) failed -- out: (%s) err: (%s)", aptGetDep.Name, out, err)
			}
		} else if out != "" {
			isDepInstalled = true
		} else {
			// no error but which's output was empty
			return fmt.Errorf("which (%s) failed -- no error (exit code 0) but output was empty", aptGetDep.Name)
		}
	}

	// then do a package manager specific lookup
	{
		if !isDepInstalled {
			// which did not find the binary, also check in brew,
			// whether the package is installed
			isDepInstalled = checkIfAptPackageInstalled(aptGetDep.Name)
		}
	}

	if !isDepInstalled {
		// Tool isn't installed -- install it...
		if !isCIMode {
			log.Infof(`This step requires "%s" to be available, but it is not installed.`, aptGetDep.GetBinaryName())
			allow, err := goinp.AskForBoolWithDefault(`Would you like to install the "`+aptGetDep.Name+`" package with apt-get?`, true)
			if err != nil {
				return err
			}
			if !allow {
				return errors.New("(" + aptGetDep.Name + ") is required for step")
			}
		}

		log.Infof("(%s) isn't installed, installing...", aptGetDep.Name)
		if cmdOut, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("sudo", "apt-get", "-y", "install", aptGetDep.Name); err != nil {
			log.Errorf("sudo apt-get -y install %s failed -- out: (%s) err: (%s)", aptGetDep.Name, cmdOut, err)
			return err
		}

		log.Infof(" * "+colorstring.Green("[OK]")+" %s installed", aptGetDep.Name)
	}

	return nil
}

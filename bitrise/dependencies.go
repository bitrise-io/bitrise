package bitrise

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/bitrise-io/bitrise/utils"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/progress"
	"github.com/bitrise-io/go-utils/retry"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/bitrise-io/go-utils/versions"
	"github.com/bitrise-io/goinp/goinp"
	stepmanModels "github.com/bitrise-io/stepman/models"
	ver "github.com/hashicorp/go-version"
)

var isAptGetUpdated bool

func removeEmptyNewLines(text string) string {
	split := strings.Split(text, "\n")
	cleanedLines := []string{}
	for _, line := range split {
		if strings.TrimSpace(line) != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}
	return strings.Join(cleanedLines, "\n")
}

// CheckIsPluginInstalled ...
func CheckIsPluginInstalled(name string, dependency PluginDependency) error {
	_, found, err := plugins.LoadPlugin(name)
	if err != nil {
		return err
	}

	currentVersion := ""
	installOrUpdate := false

	if !found {
		log.Warnf("Default plugin (%s) NOT found, installing...", name)
		installOrUpdate = true
		currentVersion = dependency.MinVersion
	} else {
		installedVersion, err := plugins.GetPluginVersion(name)
		if err != nil {
			return err
		}

		if installedVersion == nil {
			log.Warnf("Default plugin (%s) is not installed from git, no version info available.", name)
			currentVersion = ""
		} else {
			currentVersion = installedVersion.String()

			minVersion, err := ver.NewVersion(dependency.MinVersion)
			if err != nil {
				return err
			}

			if installedVersion.LessThan(minVersion) {
				log.Warnf("Default plugin (%s) version (%s) is lower than required (%s), updating...", name, installedVersion.String(), minVersion.String())
				installOrUpdate = true
				currentVersion = dependency.MinVersion
			}
		}
	}

	if installOrUpdate {
		var plugin plugins.Plugin
		err := retry.Times(2).Wait(5 * time.Second).Try(func(attempt uint) error {
			if attempt > 0 {
				log.Warnf("Download failed, retrying ...")
			}
			p, _, err := plugins.InstallPlugin(dependency.Source, dependency.MinVersion)
			plugin = p
			return err
		})
		if err != nil {
			return fmt.Errorf("Failed to install plugin, error: %s", err)
		}

		if len(plugin.Description) > 0 {
			log.Print(removeEmptyNewLines(plugin.Description))
		}
	}

	pluginDir := plugins.GetPluginDir(name)

	log.Printf("%s Plugin %s (%s): %s", colorstring.Green("[OK]"), name, currentVersion, pluginDir)

	return nil
}

// CheckIsHomebrewInstalled ...
func CheckIsHomebrewInstalled(isFullSetupMode bool) error {
	brewRubyInstallCmdString := `$ ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"`
	officialSiteURL := "http://brew.sh/"

	progInstallPth, err := utils.CheckProgramInstalledPath("brew")
	if err != nil {
		log.Print()
		log.Warnf("It seems that Homebrew is not installed on your system.")
		log.Infof("Homebrew (short: brew) is required in order to be able to auto-install all the bitrise dependencies.")
		log.Infof("You should be able to install brew by copying this command and running it in your Terminal:")
		log.Infof(brewRubyInstallCmdString)
		log.Infof("You can find more information about Homebrew on its official site at: %s", officialSiteURL)
		log.Warnf("Once the installation of brew is finished you should call the bitrise setup again.")
		return err
	}
	verStr, err := command.RunCommandAndReturnStdout("brew", "--version")
	if err != nil {
		log.Infof("")
		return errors.New("Failed to get version")
	}

	if isFullSetupMode {
		// brew doctor
		doctorOutput := ""
		var err error
		progress.NewDefaultWrapper("brew doctor").WrapAction(func() {
			doctorOutput, err = command.RunCommandAndReturnCombinedStdoutAndStderr("brew", "doctor")
		})
		if err != nil {
			log.Print()
			log.Warnf("brew doctor returned an error:")
			log.Warnf("%s", doctorOutput)
			return errors.New("command failed: brew doctor")
		}
	}

	verSplit := strings.Split(verStr, "\n")
	if len(verSplit) == 2 {
		log.Printf("%s %s: %s", colorstring.Green("[OK]"), verSplit[0], progInstallPth)
		log.Printf("%s %s", colorstring.Green("[OK]"), verSplit[1])
	} else {
		log.Printf("%s %s: %s", colorstring.Green("[OK]"), verStr, progInstallPth)
	}

	return nil
}

// PrintInstalledXcodeInfos ...
func PrintInstalledXcodeInfos() error {
	xcodeSelectPth, err := command.RunCommandAndReturnStdout("xcode-select", "--print-path")
	if err != nil {
		xcodeSelectPth = "xcode-select --print-path failed to detect the location of activate Xcode Command Line Tools path"
	}

	progInstallPth, err := utils.CheckProgramInstalledPath("xcodebuild")
	if err != nil {
		return errors.New("xcodebuild is not installed")
	}

	isFullXcodeAvailable := false
	verStr, err := command.RunCommandAndReturnCombinedStdoutAndStderr("xcodebuild", "-version")
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

	if !isFullXcodeAvailable {
		log.Printf("%s xcodebuild (%s): %s", colorstring.Green("[OK]"), colorstring.Yellow(verStr), progInstallPth)
	} else {
		log.Printf("%s xcodebuild (%s): %s", colorstring.Green("[OK]"), verStr, progInstallPth)
	}

	log.Printf("%s active Xcode (Command Line Tools) path (xcode-select --print-path): %s", colorstring.Green("[OK]"), xcodeSelectPth)

	if !isFullXcodeAvailable {
		log.Warnf("No Xcode found, only the Xcode Command Line Tools are available!")
		log.Warnf("Full Xcode is required to build, test and archive iOS apps!")
	}

	return nil
}

func checkIsBitriseToolInstalled(toolname, minVersion string, isInstall bool) error {
	doInstall := func() error {
		officialGithub := "https://github.com/bitrise-io/" + toolname
		log.Warnf("No supported %s version found", toolname)
		log.Printf("You can find more information about %s on its official GitHub page: %s", toolname, officialGithub)

		// Install
		var err error
		progress.NewDefaultWrapper("Installing").WrapAction(func() {
			err = retry.Times(2).Wait(5 * time.Second).Try(func(attempt uint) error {
				if attempt > 0 {
					log.Warnf("Download failed, retrying ...")
				}
				return tools.InstallToolFromGitHub(toolname, "bitrise-io", minVersion)
			})
		})

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
	verStr, err := command.RunCommandAndReturnStdout(toolname, "-version")
	if err != nil {
		log.Infof("")
		return errors.New("Failed to get version")
	}

	// version check
	isVersionOk, err := versions.IsVersionGreaterOrEqual(verStr, minVersion)
	if err != nil {
		log.Errorf("Failed to validate installed version")
		return err
	}
	if !isVersionOk {
		if !isInstall {
			log.Warnf("Installed %s found, but not a supported version (%s)", toolname, verStr)
			return errors.New("Failed to install required version")
		}
		return doInstall()
	}

	log.Printf("%s %s (%s): %s", colorstring.Green("[OK]"), toolname, verStr, progInstallPth)
	return nil
}

// CheckIsEnvmanInstalled ...
func CheckIsEnvmanInstalled(minEnvmanVersion string) error {
	toolname := "envman"
	minVersion := minEnvmanVersion
	return checkIsBitriseToolInstalled(toolname, minVersion, true)
}

// CheckIsStepmanInstalled ...
func CheckIsStepmanInstalled(minStepmanVersion string) error {
	toolname := "stepman"
	minVersion := minStepmanVersion
	return checkIsBitriseToolInstalled(toolname, minVersion, true)
}

func checkIfBrewPackageInstalled(packageName string) bool {
	out, err := command.New("brew", "list", packageName).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return false
	}
	return len(out) > 0
}

func checkIfAptPackageInstalled(packageName string) bool {
	err := command.New("dpkg", "-s", packageName).Run()
	return (err == nil)
}

// InstallWithBrewIfNeeded ...
func InstallWithBrewIfNeeded(brewDep stepmanModels.BrewDepModel, isCIMode bool) error {
	isDepInstalled := false
	// First do a "which", to see if the binary is available.
	// Can be available from another source, not just from brew,
	// e.g. it's common to use NVM or similar to install and manage the Node.js version.
	{
		if out, err := command.RunCommandAndReturnCombinedStdoutAndStderr("which", brewDep.GetBinaryName()); err != nil {
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
		if cmdOut, err := command.RunCommandAndReturnCombinedStdoutAndStderr("brew", "install", brewDep.Name); err != nil {
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
		if out, err := command.RunCommandAndReturnCombinedStdoutAndStderr("which", aptGetDep.GetBinaryName()); err != nil {
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

		if !isAptGetUpdated {
			cmd := command.New("sudo", "apt-get", "update")
			log.Infof(cmd.PrintableCommandArgs())
			if err := cmd.SetStdout(os.Stdout).SetStderr(os.Stderr).Run(); err != nil {
				log.Errorf("apt-get update failed, error: %s", err)
			}
			isAptGetUpdated = true
		}

		log.Infof("(%s) isn't installed, installing...", aptGetDep.Name)
		if cmdOut, err := command.RunCommandAndReturnCombinedStdoutAndStderr("sudo", "apt-get", "-y", "install", aptGetDep.Name); err != nil {
			log.Errorf("sudo apt-get -y install %s failed -- out: (%s) err: (%s)", aptGetDep.Name, cmdOut, err)
			return err
		}

		log.Infof(" * "+colorstring.Green("[OK]")+" %s installed", aptGetDep.Name)
	}

	return nil
}

package bitrise

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/stepman/toolkits"
)

type SetupMode string

const (
	SetupModeDefault SetupMode = "default"
	SetupModeMinimal SetupMode = "minimal"
)

const (
	minEnvmanVersion  = "2.5.5"
	minStepmanVersion = "0.18.7"
)

type PluginDependency struct {
	Source     string
	MinVersion string
}

var PluginDependencyMap = map[string]PluginDependency{
	"init": {
		Source:     "https://github.com/bitrise-io/bitrise-plugins-init.git",
		MinVersion: "1.10.0",
	},
	"step": {
		Source:     "https://github.com/bitrise-io/bitrise-plugins-step.git",
		MinVersion: "0.10.4",
	},
	"workflow-editor": {
		Source:     "https://github.com/bitrise-io/bitrise-workflow-editor.git",
		MinVersion: "1.3.305",
	},
}

func RunSetupIfNeeded(logger log.Logger) error {
	versionMatch, setupVersion := configs.CheckIsSetupWasDoneForVersion(version.VERSION)
	if setupVersion == "" {
		log.Infof("No setup was done yet, running setup now...")
		return RunSetup(logger, version.VERSION, SetupModeMinimal, false)
	}
	if !versionMatch {
		log.Infof("Setup was last performed for version %s, current version is %s. Re-running setup now...", setupVersion, version.VERSION)
		return RunSetup(logger, version.VERSION, SetupModeMinimal, false)
	}
	return nil
}

func RunSetup(logger log.Logger, appVersion string, setupMode SetupMode, doCleanSetup bool) error {
	log.Infof("Setup Bitrise tools...")
	log.Printf("Clean before setup: %v", doCleanSetup)
	log.Printf("Setup mode: %s", setupMode)
	log.Printf("System: %s/%s", runtime.GOOS, runtime.GOARCH)

	if doCleanSetup {
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
		return fmt.Errorf("failed to do common/platform independent setup, error: %s", err)
	}

	switch runtime.GOOS {
	case "darwin":
		if err := doSetupOnOSX(); err != nil {
			return fmt.Errorf("failed to do macOS-specific setup, error: %s", err)
		}
	case "linux":
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if setupMode != SetupModeMinimal {
		if err := doSetupPlugins(); err != nil {
			return fmt.Errorf("failed to do Plugins setup, error: %s", err)
		}
	}

	if err := doSetupToolkits(logger); err != nil {
		return fmt.Errorf("failed to do Toolkits setup, error: %s", err)
	}

	log.Print()
	log.Donef("Bitrise tools are installed and ready to use!")

	if err := configs.SaveSetupSuccessForVersion(appVersion); err != nil {
		return fmt.Errorf("failed to save setup-success into config file, error: %s", err)
	}

	return nil
}

func doSetupToolkits(logger log.Logger) error {
	log.Print()
	log.Infof("Checking Bitrise Toolkits...")

	coreToolkits := toolkits.AllSupportedToolkits(logger)

	for _, aCoreTK := range coreToolkits {
		toolkitName := aCoreTK.ToolkitName()
		isInstallRequired, checkResult, err := aCoreTK.Check()
		if err != nil {
			return fmt.Errorf("failed to perform toolkit check (%s), error: %s", toolkitName, err)
		}

		if isInstallRequired {
			log.Warnf("No installed/suitable %s found, installing toolkit ...", toolkitName)
			if err := aCoreTK.Install(); err != nil {
				return fmt.Errorf("failed to install toolkit (%s), error: %s", toolkitName, err)
			}

			isInstallRequired, checkResult, err = aCoreTK.Check()
			if err != nil {
				return fmt.Errorf("failed to perform toolkit check (%s), error: %s", toolkitName, err)
			}
		}
		if isInstallRequired {
			return fmt.Errorf("toolkit (%s) still reports that it isn't (properly) installed", toolkitName)
		}

		log.Printf("%s %s (%s): %s", colorstring.Green("[OK]"), toolkitName, checkResult.Version, checkResult.Path)
	}

	return nil
}

func doSetupPlugins() error {
	log.Print()
	log.Infof("Checking Bitrise Plugins...")

	// Validate currently installed plugins.
	if err := validateInstalledPlugins(); err != nil {
		log.Warnf("Failed to validate installed plugins: %s", err)
	}

	// Check default plugins and install/update if needed.
	for pluginName, pluginDependency := range PluginDependencyMap {
		if err := CheckIsPluginInstalled(pluginName, pluginDependency); err != nil {
			return fmt.Errorf("plugin (%s) failed to install: %s", pluginName, err)
		}
	}

	return nil
}

// validateInstalledPlugins checks all plugins in the routing file and removes or reinstalls broken ones.
func validateInstalledPlugins() error {
	routing, err := plugins.ReadPluginRouting()
	if err != nil {
		return fmt.Errorf("failed to read plugin routing: %s", err)
	}

	if len(routing.RouteMap) == 0 {
		return nil
	}

	var reinstalledCount, removedCount int

	for pluginName, route := range routing.RouteMap {
		_, found, err := plugins.LoadPlugin(pluginName)
		if err != nil || !found {
			if err != nil {
				log.Warnf("Plugin (%s) validation failed: %s", pluginName, err)
			} else {
				log.Warnf("Plugin (%s) found in routing but not installed", pluginName)
			}

			// Reinstall if source is available.
			if route.Source != "" && route.Source != "local" {
				log.Warnf("Attempting to reinstall plugin (%s) from %s", pluginName, route.Source)

				// Clean up broken plugin (directory and routing).
				if err := plugins.DeletePlugin(pluginName); err != nil {
					log.Warnf("Failed to cleanup broken plugin (%s): %s", pluginName, err)
				}

				_, _, err := plugins.InstallPlugin(route.Source, route.Version)
				if err != nil {
					log.Errorf("Failed to reinstall plugin (%s): %s", pluginName, err)
					log.Warnf("You may need to manually reinstall: bitrise plugin install %s", route.Source)
					removedCount++
				} else {
					log.Donef("Successfully reinstalled plugin (%s)", pluginName)
					reinstalledCount++
				}
			} else {
				// Local plugin or no source: remove.
				log.Warnf("Plugin (%s) cannot be automatically reinstalled, removing", pluginName)
				if err := plugins.DeletePlugin(pluginName); err != nil {
					// At least remove from routing, to not block others.
					if err := plugins.DeletePluginRoute(pluginName); err != nil {
						log.Warnf("Failed to remove broken plugin route (%s): %s", pluginName, err)
					}
				}
				removedCount++
			}
		}
	}

	if reinstalledCount > 0 {
		log.Donef("Reinstalled %d broken plugin(s)", reinstalledCount)
	}
	if removedCount > 0 {
		log.Warnf("Removed %d broken plugin(s) that could not be reinstalled", removedCount)
	}

	return nil
}

func doSetupBitriseCoreTools() error {
	log.Print()
	log.Infof("Checking Bitrise Core tools...")

	if err := CheckIsEnvmanInstalled(minEnvmanVersion); err != nil {
		return fmt.Errorf("failed to install envman: %s", err)
	}

	if err := CheckIsStepmanInstalled(minStepmanVersion); err != nil {
		return fmt.Errorf("failed to install stepman: %s", err)
	}

	return nil
}

func doSetupOnOSX() error {
	log.Print()
	log.Infof("Doing macOS-specific setup")
	log.Printf("Checking required tools...")

	if err := CheckIsHomebrewInstalled(); err != nil {
		return errors.New(fmt.Sprint("Homebrew not installed or has some issues. Please fix these before calling setup again. Err:", err))
	}

	return nil
}

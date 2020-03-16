package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/log"
	flog "github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

//=======================================
// Util
//=======================================

func strip(str string) string {
	dirty := true
	strippedStr := str
	for dirty {
		hasWhiteSpacePrefix := false
		if strings.HasPrefix(strippedStr, " ") {
			hasWhiteSpacePrefix = true
			strippedStr = strings.TrimPrefix(strippedStr, " ")
		}

		hasWhiteSpaceSuffix := false
		if strings.HasSuffix(strippedStr, " ") {
			hasWhiteSpaceSuffix = true
			strippedStr = strings.TrimSuffix(strippedStr, " ")
		}

		hasNewlinePrefix := false
		if strings.HasPrefix(strippedStr, "\n") {
			hasNewlinePrefix = true
			strippedStr = strings.TrimPrefix(strippedStr, "\n")
		}

		hasNewlineSuffix := false
		if strings.HasSuffix(strippedStr, "\n") {
			hasNewlinePrefix = true
			strippedStr = strings.TrimSuffix(strippedStr, "\n")
		}

		if !hasWhiteSpacePrefix && !hasWhiteSpaceSuffix && !hasNewlinePrefix && !hasNewlineSuffix {
			dirty = false
		}
	}
	return strippedStr
}

//=======================================
// Main
//=======================================

// RunPluginByEvent ...
func RunPluginByEvent(plugin Plugin, pluginConfig PluginConfig, input []byte) error {
	pluginConfig[PluginConfigPluginModeKey] = string(TriggerMode)

	return runPlugin(plugin, []string{}, pluginConfig, input)
}

// RunPluginByCommand ...
func RunPluginByCommand(plugin Plugin, args []string) error {
	pluginConfig := PluginConfig{
		PluginConfigPluginModeKey: string(CommandMode),
	}

	return runPlugin(plugin, args, pluginConfig, nil)
}

// PrintPluginUpdateInfos ...
func PrintPluginUpdateInfos(newVersion string, plugin Plugin) {
	flog.Warnf("")
	flog.Warnf("New version (%s) of plugin (%s) available", newVersion, plugin.Name)
	flog.Printf("Run command to update plugin:")
	fmt.Println()
	flog.Donef("$ bitrise plugin update %s", plugin.Name)
}

func runPlugin(plugin Plugin, args []string, envs PluginConfig, input []byte) error {
	if !configs.IsCIMode && configs.CheckIsPluginUpdateCheckRequired(plugin.Name) {
		// Check for new version
		log.Infof("Checking for plugin (%s) new version...", plugin.Name)

		if newVersion, err := CheckForNewVersion(plugin); err != nil {
			log.Warnf("")
			log.Warnf("Failed to check for plugin (%s) new version, error: %s", plugin.Name, err)
		} else if newVersion != "" {
			PrintPluginUpdateInfos(newVersion, plugin)
		}

		if err := configs.SavePluginUpdateCheck(plugin.Name); err != nil {
			return err
		}

		fmt.Println()
	}

	// Append common data to plugin iputs
	bitriseVersion, err := version.BitriseCliVersion()
	if err != nil {
		return err
	}
	envs[PluginConfigBitriseVersionKey] = bitriseVersion.String()
	envs[PluginConfigDataDirKey] = GetPluginDataDir(plugin.Name)
	envs[PluginConfigFormatVersionKey] = models.Version

	// Prepare plugin envstore
	pluginWorkDir, err := pathutil.NormalizedOSTempDirPath("plugin-work-dir")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(pluginWorkDir); err != nil {
			log.Warnf("Failed to remove path (%s)", pluginWorkDir)
		}
	}()

	pluginEnvstorePath := filepath.Join(pluginWorkDir, "envstore.yml")

	if err := tools.EnvmanInitAtPath(pluginEnvstorePath); err != nil {
		return err
	}

	if err := tools.EnvmanAdd(pluginEnvstorePath, configs.EnvstorePathEnvKey, pluginEnvstorePath, false, false); err != nil {
		return err
	}

	// Add plugin inputs
	for key, value := range envs {
		if err := tools.EnvmanAdd(pluginEnvstorePath, key, value, false, false); err != nil {
			return err
		}
	}

	// Run plugin executable
	pluginExecutable, isBin, err := GetPluginExecutablePath(plugin.Name)
	if err != nil {
		return err
	}

	cmd := []string{}

	if isBin {
		cmd = append([]string{pluginExecutable}, args...)
	} else {
		cmd = append([]string{"bash", pluginExecutable}, args...)
	}

	if _, err := tools.EnvmanRun(pluginEnvstorePath, "", cmd, -1, nil, input); err != nil {
		return err
	}

	return nil
}

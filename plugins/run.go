package plugins

import (
	"bytes"
	"strings"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/log/logwriter"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/command"
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
	log.Warnf("")
	log.Warnf("New version (%s) of plugin (%s) available", newVersion, plugin.Name)
	log.Printf("Run command to update plugin:")
	log.Print()
	log.Donef("$ bitrise plugin update %s", plugin.Name)
}

func runPlugin(plugin Plugin, args []string, envKeyValues PluginConfig, input []byte) error {
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

		log.Print()
	}

	// Append common data to plugin inputs
	bitriseVersion, err := version.BitriseCliVersion()
	if err != nil {
		return err
	}
	envKeyValues[PluginConfigBitriseVersionKey] = bitriseVersion.String()
	envKeyValues[PluginConfigDataDirKey] = GetPluginDataDir(plugin.Name)
	envKeyValues[PluginConfigFormatVersionKey] = models.FormatVersion

	// Run plugin executable
	pluginExecutable, isBin, err := GetPluginExecutablePath(plugin.Name)
	if err != nil {
		return err
	}

	var cmd *command.Model

	if isBin {
		cmd = command.New(pluginExecutable, args...)
	} else {
		cmd = command.New("bash", append([]string{pluginExecutable}, args...)...)
	}

	cmd.SetStdin(bytes.NewReader(input))

	var envs []string
	for key, value := range envKeyValues {
		envs = append(envs, key+"="+value)
	}

	cmd.AppendEnvs(envs...)

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	logWriter := logwriter.NewLogWriter(logger)

	cmd.SetStdout(logWriter)

	cmdErr := cmd.Run()

	if err := logWriter.Close(); err != nil {
		log.Warnf("Failed to close command output writer: %s", err)
	}

	return cmdErr
}

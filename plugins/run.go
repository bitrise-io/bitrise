package plugins

import (
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	envmanModels "github.com/bitrise-io/envman/models"
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

func commandOutput(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if dir != "" {
		cmd.Dir = dir
	}

	outBytes, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strip(string(outBytes)), nil
}

func command(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Dir = dir

	return cmd.Run()
}

//=======================================
// Main
//=======================================

// RunPlugin ...
func RunPlugin(plugin Plugin, args []string) error {
	// Create plugin input
	bitriseVersionPtr, err := configs.GetBitriseVersion()
	if err != nil {
		return err
	}

	pluginInputBytes, err := json.Marshal(map[string]string{"version": bitriseVersionPtr.String()})
	if err != nil {
		return err
	}
	pluginInputStr := string(pluginInputBytes)

	pluginWorkDir, err := pathutil.NormalizedOSTempDirPath("plugin-work-dir")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(pluginWorkDir); err != nil {
			log.Warnf("Failed to remove path (%s)", pluginWorkDir)
		}
	}()

	pluginEnvstorePath := path.Join(pluginWorkDir, "envstore.yml")

	if err := bitrise.EnvmanInitAtPath(pluginEnvstorePath); err != nil {
		return err
	}

	if err := bitrise.EnvmanAdd(pluginEnvstorePath, bitrise.EnvstorePathEnvKey, pluginEnvstorePath, false); err != nil {
		return err
	}

	if err := bitrise.EnvmanAdd(pluginEnvstorePath, bitrisePluginInputEnvKey, pluginInputStr, false); err != nil {
		return err
	}

	log.Debugf("plugin evstore path (%s)", pluginEnvstorePath)

	// Run plugin executable
	pluginExecutable, isBin, err := GetPluginExecutablePath(plugin.Name)
	if err != nil {
		return err
	}

	cmd := []string{}

	if isBin {
		log.Debugf("Run plugin binary (%s)", pluginExecutable)
		cmd = append([]string{pluginExecutable}, args...)
	} else {
		log.Debugf("Run plugin sh (%s)", pluginExecutable)
		cmd = append([]string{"bash", pluginExecutable}, args...)
	}

	exitCode, err := bitrise.EnvmanRun(pluginEnvstorePath, "", cmd)
	log.Debugf("Plugin run finished with exit code (%d)", exitCode)
	if err != nil {
		return err
	}

	// Read plugin output
	outStr, err := bitrise.EnvmanJSONPrint(pluginEnvstorePath)
	if err != nil {
		return err
	}

	envList, err := envmanModels.EnvsJSONListModel{}.CreateFromJSON(outStr)
	if err != nil {
		return err
	}

	pluginOutputStr, found := envList[bitrisePluginOutputEnvKey]
	if found {
		log.Debugf("Plugin output: %s", pluginOutputStr)
	}

	return nil
}

package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/versions"
)

func getOsAndArch() (string, string, error) {
	osOut, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("uname", "-s")
	if err != nil {
		return "", "", err
	}

	archOut, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("uname", "-m")
	if err != nil {
		return "", "", err
	}

	return osOut, archOut, nil
}

// DownloadPluginFromURL ....
func DownloadPluginFromURL(URL, dst string) error {
	url, err := url.Parse(URL)
	if err != nil {
		return err
	}

	scheme := url.Scheme
	tokens := strings.Split(URL, "/")
	fileName := tokens[len(tokens)-1]

	tmpDstFilePath := ""
	if scheme != "file" {
		OS, arch, err := getOsAndArch()
		if err != nil {
			return err
		}

		urlWithSuffix := URL
		urlSuffix := fmt.Sprintf("-%s-%s", OS, arch)
		if !strings.HasSuffix(URL, urlSuffix) {
			urlWithSuffix = urlWithSuffix + urlSuffix
		}

		urls := []string{urlWithSuffix, URL}

		tmpDir, err := pathutil.NormalizedOSTempDirPath("plugin")
		if err != nil {
			return err
		}
		tmpDst := path.Join(tmpDir, fileName)
		output, err := os.Create(tmpDst)
		if err != nil {
			return err
		}
		defer func() {
			if err := output.Close(); err != nil {
				log.Errorf("Failed to close file, err: %s", err)
			}
		}()

		success := false
		var response *http.Response
		for _, aURL := range urls {

			response, err = http.Get(aURL)
			if response != nil {
				defer func() {
					if err := response.Body.Close(); err != nil {
						log.Errorf("Failed to close response body, err: %s", err)
					}
				}()
			}

			if err != nil {
				log.Errorf("%s", err)
			} else {
				success = true
				break
			}
		}
		if !success {
			return err
		}

		if _, err := io.Copy(output, response.Body); err != nil {
			return err
		}

		tmpDstFilePath = output.Name()
	} else {
		tmpDstFilePath = strings.Replace(URL, scheme+"://", "", -1)
	}

	if err := cmdex.CopyFile(tmpDstFilePath, dst); err != nil {
		return err
	}

	return nil
}

// InstallPlugin ...
func InstallPlugin(bitriseVersion, pluginSource, pluginName, pluginType string) (string, error) {
	checkMinMaxVersion := func(requiredMin, requiredMax, current string) (bool, error) {
		if requiredMin != "" {
			// 1 if version 2 is greater then version 1, -1 if not
			greater, err := versions.CompareVersions(current, requiredMin)
			if err != nil {
				return false, err
			}
			if greater == 1 {
				return false, fmt.Errorf("Required min version (%s) - current (%s)", requiredMin, current)
			}
		}

		if requiredMax != "" {
			greater, err := versions.CompareVersions(requiredMax, current)
			if err != nil {
				return false, err
			}
			if greater == 1 {
				return false, fmt.Errorf("Allowed max version (%s) - current (%s)", requiredMax, current)
			}
		}

		return true, nil
	}

	pluginPath, err := GetPluginPath(pluginName, pluginType)
	if err != nil {
		return "", err
	}

	if err := DownloadPluginFromURL(pluginSource, pluginPath); err != nil {
		return "", err
	}

	if err := os.Chmod(pluginPath, 0777); err != nil {
		return "", err
	}

	plugin, err := GetPlugin(pluginName, pluginType)
	if err != nil {
		return "", err
	}

	messageFromPlugin, err := RunPlugin(bitriseVersion, plugin, []string{"requirements"})
	if err != nil {
		return "", err
	}
	if messageFromPlugin != "" {
		log.Infoln("=> Checking plugin requirements...")
		log.Debugf("requirements messageFromPlugin: %s", messageFromPlugin)

		envmanVersion, err := bitrise.EnvmanVersion()
		if err != nil {
			return "", err
		}

		stepmanVersion, err := bitrise.StepmanVersion()
		if err != nil {
			return "", err
		}

		currentVersionMap := map[string]string{
			"bitrise": bitriseVersion,
			"envman":  envmanVersion,
			"stepman": stepmanVersion,
		}

		type Requirement struct {
			ToolName   string
			MinVersion string
			MaxVersion string
		}

		var requirements []Requirement
		if err := json.Unmarshal([]byte(messageFromPlugin), &requirements); err != nil {
			return "", err
		}

		for _, requirement := range requirements {
			toolName := requirement.ToolName
			minVersion := requirement.MinVersion
			maxVersion := requirement.MaxVersion
			currentVersion := currentVersionMap[toolName]

			ok, err := checkMinMaxVersion(minVersion, maxVersion, currentVersion)
			if err != nil {
				return "", fmt.Errorf("%s requirements failed, err: %s", toolName, err)
			}

			if !ok {
				log.Infof(" (i) %s min version: %s / max version: %s - current  version: %s", toolName, minVersion, maxVersion, bitriseVersion)
			}
		}
	}

	printableName := PrintableName(pluginName, pluginType)
	return printableName, nil
}

// DeletePlugin ...
func DeletePlugin(pluginName, pluginType string) error {
	pluginPath, err := GetPluginPath(pluginName, pluginType)
	if err != nil {
		return err
	}

	if exists, err := pathutil.IsPathExists(pluginPath); err != nil {
		return fmt.Errorf("Failed to check dir (%s), err: %s", pluginPath, err)
	} else if !exists {
		return fmt.Errorf("Plugin (%s) not installed", PrintableName(pluginName, pluginType))
	}
	return os.Remove(pluginPath)
}

// ListPlugins ...
func ListPlugins() (map[string][]Plugin, error) {
	collectPlugin := func(dir, pluginType string) ([]Plugin, error) {
		plugins := []Plugin{}

		pluginsPath, err := GetPluginPath("", pluginType)
		if err != nil {
			return []Plugin{}, err
		}

		files, err := ioutil.ReadDir(pluginsPath)
		if err != nil {
			return []Plugin{}, err
		}
		for _, file := range files {
			if !strings.HasPrefix(file.Name(), ".") {
				plugin, err := GetPlugin(file.Name(), pluginType)
				if err != nil {
					return []Plugin{}, err
				}
				plugins = append(plugins, plugin)
			}
		}
		return plugins, nil
	}

	pluginMap := map[string][]Plugin{}
	pluginsPath, err := GetPluginsDir()
	if err != nil {
		return map[string][]Plugin{}, err
	}

	pluginTypes := []string{TypeGeneric, TypeInit, TypeRun}
	for _, pType := range pluginTypes {
		ps, err := collectPlugin(pluginsPath, pType)
		if err != nil {
			return map[string][]Plugin{}, err
		}
		pluginMap[pType] = ps
	}

	return pluginMap, nil
}

// ParseArgs ...
func ParseArgs(args []string) (string, string, []string, bool) {
	const bitrisePluginPrefix = ":"

	log.Debugf("args: %v", args)

	if len(args) > 0 {
		plugin := ""
		pluginArgs := []string{}
		for idx, arg := range args {
			if strings.Contains(arg, bitrisePluginPrefix) {
				plugin = arg
				pluginArgs = args[idx:len(args)]
			}
		}

		// generic plugins
		if strings.HasPrefix(plugin, bitrisePluginPrefix) {
			pluginName := strings.TrimPrefix(plugin, bitrisePluginPrefix)
			return pluginName, TypeGeneric, pluginArgs, true
		}

		// typed plugins
		if strings.Contains(plugin, ":") {
			pluginSplits := strings.Split(plugin, ":")
			if len(pluginSplits) == 2 {
				pluginType := pluginSplits[0]
				pluginName := pluginSplits[1]
				return pluginName, pluginType, pluginArgs, true
			}
		}
	}

	return "", "", []string{}, false
}

// GetPlugin ...
func GetPlugin(name, pluginType string) (Plugin, error) {
	pluginPath, err := GetPluginPath(name, pluginType)
	if err != nil {
		return Plugin{}, err
	}

	if exists, err := pathutil.IsPathExists(pluginPath); err != nil {
		return Plugin{}, fmt.Errorf("Failed to check dir (%s), err: %s", pluginPath, err)
	} else if !exists {
		return Plugin{}, nil
	}

	plugin := Plugin{
		Name: name,
		Path: pluginPath,
		Type: pluginType,
	}

	return plugin, nil
}

// RunPlugin ...
func RunPlugin(bitriseVersion string, plugin Plugin, args []string) (string, error) {
	var outBuffer bytes.Buffer

	bitriseInfos := map[string]string{
		"version": bitriseVersion,
	}
	bitriseInfosStr, err := json.Marshal(bitriseInfos)
	if err != nil {
		return "", err
	}
	if err := os.Setenv("BITRISE_PLUGINS_MESSAGE", string(bitriseInfosStr)); err != nil {
		return "", err
	}

	err = cmdex.RunCommandWithWriters(io.Writer(&outBuffer), os.Stderr, plugin.Path, args...)
	return outBuffer.String(), err
}

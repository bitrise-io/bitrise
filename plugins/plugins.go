package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/pathutil"
)

// Plugin ...
type Plugin struct {
	Path string
	Name string
	Type string
}

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
func DownloadPluginFromURL(url, dst string) error {
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]

	OS, arch, err := getOsAndArch()
	if err != nil {
		return err
	}

	urlWithSuffix := url
	urlSuffix := fmt.Sprintf("-%s-%s", OS, arch)
	if !strings.HasSuffix(url, urlSuffix) {
		urlWithSuffix = urlWithSuffix + urlSuffix
	}

	urls := []string{urlWithSuffix, url}

	tmpDir, err := pathutil.NormalizedOSTempDirPath("plugin")
	if err != nil {
		return err
	}
	tmpDst := path.Join(tmpDir, fileName)
	output, err := os.Create(tmpDst)
	if err != nil {
		return err
	}
	defer output.Close()

	success := false
	var response *http.Response
	for _, aURL := range urls {
		log.Infof("")
		log.Infof(" => Downloading (%s) to (%s)", aURL, dst)

		response, err = http.Get(aURL)
		if err != nil {
			log.Errorf("%s", err)
		} else {
			success = true
		}
		if response != nil {
			defer response.Body.Close()
		}
		if success {
			break
		}
	}
	if !success {
		return err
	}

	n, err := io.Copy(output, response.Body)
	if err != nil {
		return err
	}
	if err := cmdex.CopyFile(output.Name(), dst); err != nil {
		return err
	}

	log.Infof(" (i) %d bytes downloaded", n)
	return nil
}

// InstallPlugin ...
func InstallPlugin(pluginSource, pluginName, pluginType string) (string, error) {
	pluginPath := GetPluginPath(pluginName, pluginType)
	fmt.Println()
	log.Infoln(" => Download plugin")
	if err := DownloadPluginFromURL(pluginSource, pluginPath); err != nil {
		return "", err
	}

	fmt.Println()
	log.Infoln(" => Change plugin permission")
	if err := os.Chmod(pluginPath, 0777); err != nil {
		return "", err
	}

	printableName := ":" + pluginName
	if pluginType != "custom" {
		printableName = pluginType + printableName
	}
	return printableName, nil
}

// ParsePlugin ...
func ParsePlugin(args []string) (string, string, []string, bool) {
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

		// custom plugins
		if strings.HasPrefix(plugin, bitrisePluginPrefix) {
			pluginName := strings.TrimPrefix(plugin, bitrisePluginPrefix)
			return pluginName, "custom", pluginArgs, true
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

// AllPluginNames ...
func AllPluginNames() ([]string, error) {
	collectPlugin := func(dir, pluginType string) ([]string, error) {
		pluginNames := []string{}

		pluginsPath := path.Join(dir, pluginType)
		files, err := ioutil.ReadDir(pluginsPath)
		if err != nil {
			return []string{}, fmt.Errorf("Failed to read plugins dir (%s), err: %s", pluginsPath, err)
		}
		for _, file := range files {
			if !strings.HasPrefix(file.Name(), ".") {
				pluginName := file.Name()
				switch pluginType {
				case "custom":
					pluginName = ":" + pluginName
				case "init":
					pluginName = "init:" + pluginName
				case "run":
					pluginName = "run:" + pluginName
				}
				pluginNames = append(pluginNames, pluginName)
			}
		}
		return pluginNames, nil
	}

	pluginNames := []string{}
	pluginsPath := GetPluginsPath()
	pluginTypes := []string{"custom", "init", "run"}
	for _, pType := range pluginTypes {
		ps, err := collectPlugin(pluginsPath, pType)
		if err != nil {
			return []string{}, fmt.Errorf("Failed to collect plugins (%s), err: %s", pType, err)
		}
		pluginNames = append(pluginNames, ps...)
	}

	return pluginNames, nil
}

// GetPlugin ...
func GetPlugin(name, pluginType string) (Plugin, error) {
	pluginPath := GetPluginPath(name, pluginType)
	if exists, err := pathutil.IsPathExists(pluginPath); err != nil {
		return Plugin{}, fmt.Errorf("Failed to check dir (%s), err: %s", pluginPath, err)
	} else if !exists {
		return Plugin{}, fmt.Errorf("Plugin executable not found at: %s", pluginPath)
	}

	plugin := Plugin{
		Name: name,
		Path: pluginPath,
		Type: pluginType,
	}

	return plugin, nil
}

// RunPlugin ...
func RunPlugin(plugin Plugin, args []string) (string, string, error) {
	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	bitriseInfos := map[string]string{
		"version": "1.2.4",
	}
	bitriseInfosStr, err := json.Marshal(bitriseInfos)
	if err != nil {
		return "", "", err
	}

	pluginArgs := []string{string(bitriseInfosStr)}
	pluginArgs = append(pluginArgs, args...)

	err = cmdex.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), plugin.Path, pluginArgs...)
	return outBuffer.String(), errBuffer.String(), err
}

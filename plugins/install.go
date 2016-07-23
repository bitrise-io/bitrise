package plugins

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/pathutil"
	ver "github.com/hashicorp/go-version"
)

//=======================================
// Util
//=======================================

func validatePath(pth string) error {
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		return fmt.Errorf("failed to check bitrise-plugin.yml path (%s), error: %s", pth, err)
	} else if !exist {
		return fmt.Errorf("no bitrise-plugin.yml found at (%s)", pth)
	}
	return nil
}

func validateVersion(current, requiredMin ver.Version, requiredMax *ver.Version) error {
	if current.Compare(&requiredMin) == -1 {
		return fmt.Errorf("Current version (%s) is less then min version (%s)  ", current.String(), requiredMin.String())
	}

	if requiredMax != nil && current.Compare(requiredMax) == 1 {
		return fmt.Errorf("Current version (%s) is greater then max version (%s)  ", current.String(), (*requiredMax).String())
	}

	return nil
}

func validateRequirements(requirements []Requirement, currentVersionMap map[string]ver.Version) error {
	var err error

	for _, requirement := range requirements {
		currentVersion := currentVersionMap[requirement.Tool]

		var minVersionPtr *ver.Version
		if requirement.MinVersion == "" {
			return fmt.Errorf("plugin requirement min version is required")
		}

		minVersionPtr, err = ver.NewVersion(requirement.MinVersion)
		if err != nil {
			return fmt.Errorf("failed to parse plugin required min version (%s) for tool (%s), error: %s", requirement.MinVersion, requirement.Tool, err)
		}

		var maxVersionPtr *ver.Version
		if requirement.MaxVersion != "" {
			maxVersionPtr, err = ver.NewVersion(requirement.MaxVersion)
			if err != nil {
				return fmt.Errorf("failed to parse plugin requirement version (%s) for tool (%s), error: %s", requirement.MaxVersion, requirement.Tool, err)
			}
		}

		if err := validateVersion(currentVersion, *minVersionPtr, maxVersionPtr); err != nil {
			return fmt.Errorf("checking plugin tool (%s) requirements failed, error: %s", requirement.Tool, err)
		}
	}

	return nil
}

func clonePluginSrc(sourceURL, versionTag, destinationDir string) (*ver.Version, string, error) {
	url, err := url.Parse(sourceURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse url (%s), error: %s", sourceURL, err)
	}

	// Download local source dir
	if url.Scheme == "file" {
		sourceDir := strings.Replace(sourceURL, url.Scheme+"://", "", -1)

		if err := cmdex.CopyDir(sourceDir, destinationDir, true); err != nil {
			return nil, "", fmt.Errorf("failed to copy (%s) to (%s), error: %s", sourceDir, destinationDir, err)
		}

		return nil, "", nil
	}

	// Download remote source dir
	version, hash, err := GitCloneAndCheckoutVersion(destinationDir, sourceURL, versionTag)
	if err != nil {
		return nil, "", fmt.Errorf("failed to git clone (%s), error: %s", sourceURL, err)
	}

	return version, hash, nil
}

func downloadPluginBin(sourceURL, destinationPth string) error {

	url, err := url.Parse(sourceURL)
	if err != nil {
		return fmt.Errorf("failed to parse url (%s), error: %s", sourceURL, err)
	}

	// Download local binary
	if url.Scheme == "file" {
		src := strings.Replace(sourceURL, url.Scheme+"://", "", -1)

		if err := cmdex.CopyFile(src, destinationPth); err != nil {
			return fmt.Errorf("failed to copy (%s) to (%s)", src, destinationPth)
		}
		return nil
	}

	// Download remote binary
	out, err := os.Create(destinationPth)
	defer func() {
		if err := out.Close(); err != nil {
			log.Warnf("failed to close (%s)", destinationPth)
		}
	}()
	if err != nil {
		return fmt.Errorf("failed to create (%s), error: %s", destinationPth, err)
	}

	resp, err := http.Get(sourceURL)
	if err != nil {
		return fmt.Errorf("failed to download from (%s), error: %s", sourceURL, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf("failed to close (%s) body", sourceURL)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to download from (%s), error: %s", sourceURL, err)
	}

	return nil
}

//=======================================
// Main
//=======================================

// InstallPlugin ...
func InstallPlugin(srcURL, binURL, versionTag string) (Plugin, string, error) {
	//
	// Download plugin src
	pluginSrcTmpDir, err := pathutil.NormalizedOSTempDirPath("plugin-src-tmp")
	if err != nil {
		return Plugin{}, "", fmt.Errorf("failed to create plugin src temp directory, error: %s", err)
	}
	defer func() {
		if err := os.RemoveAll(pluginSrcTmpDir); err != nil {
			log.Warnf("Failed to remove path (%s)", pluginSrcTmpDir)
		}
	}()

	newVersionPtr, newVersinHash, err := clonePluginSrc(srcURL, versionTag, pluginSrcTmpDir)
	if err != nil {
		return Plugin{}, "", fmt.Errorf("failed to download plugin, error: %s", err)
	}

	log.Debugf("Plugin downloaded from (%s) to (%s)", srcURL, pluginSrcTmpDir)

	//
	// Parse and validate plugin.yml

	// Validate bitrise-plugin.yml
	tmpPluginYMLPath := filepath.Join(pluginSrcTmpDir, pluginYMLName)

	if err := validatePath(tmpPluginYMLPath); err != nil {
		return Plugin{}, "", fmt.Errorf("bitrise-plugin.yml validation failed, error: %s", err)
	}

	newPlugin, err := NewPluginFromYML(tmpPluginYMLPath)
	if err != nil {
		return Plugin{}, "", fmt.Errorf("failed to parse bitrise-plugin.yml (%s), error: %s", tmpPluginYMLPath, err)
	}

	// Check if executable exist
	if newPlugin.ExecutableURL() == "" && binURL == "" {
		tmpPluginExecutablePath := filepath.Join(pluginSrcTmpDir, pluginShName)
		if err := validatePath(tmpPluginExecutablePath); err != nil {
			return Plugin{}, "", fmt.Errorf("bitrise-plugin.sh validation failed, error: %s", err)
		}
	}

	// Check tool requirements
	currentVersionMap, err := version.ToolVersionMap()
	if err != nil {
		return Plugin{}, "", fmt.Errorf("failed to get current version map, error: %s", err)
	}

	if err := validateRequirements(newPlugin.Requirements, currentVersionMap); err != nil {
		return Plugin{}, "", fmt.Errorf("requirements validation failed, error: %s", err)
	}

	log.Debugf("Downloaded plugin: %#v validated", newPlugin)

	//
	// Check if plugin already installed
	if route, found, err := ReadPluginRoute(newPlugin.Name); err != nil {
		return Plugin{}, "", fmt.Errorf("failed to check if plugin already installed, error: %s", err)
	} else if found {
		log.Debugf("Plugin already installed with name (%s)", newPlugin.Name)

		if route.Source != srcURL {
			return Plugin{}, "", fmt.Errorf("plugin already installed with name (%s) from different source (%s)", route.Name, route.Source)
		}

		installedPluginVersionPtr, err := GetPluginVersion(route.Name)
		if err != nil {
			return Plugin{}, "", fmt.Errorf("failed to check installed plugin (%s) version, error: %s", route.Name, err)
		}

		if newVersionPtr != nil && installedPluginVersionPtr != nil && installedPluginVersionPtr.GreaterThan(newVersionPtr) {
			return Plugin{}, "", fmt.Errorf("installed plugin version (%s) greater then new plugin version (%s)", installedPluginVersionPtr.String(), (*newVersionPtr).String())
		}

		installedPluginVersion := "local"
		if installedPluginVersionPtr != nil {
			installedPluginVersion = (*installedPluginVersionPtr).String()
		}

		fmt.Println()
		log.Infof("Installed plugin found with version (%s), overriding it...", installedPluginVersion)
	}

	//
	// Intsall plugin into bitrise
	installSuccess := true

	pluginDir := GetPluginDir(newPlugin.Name)

	if err := os.RemoveAll(pluginDir); err != nil {
		return Plugin{}, "", fmt.Errorf("failed to remove plugin dir (%s), error: %s", pluginDir, err)
	}
	defer func() {
		if installSuccess {
			return
		}

		if err := os.RemoveAll(pluginDir); err != nil {
			log.Warnf("Failed to remove path (%s)", pluginDir)
		}
	}()

	// Install plugin src
	plginSrcDir := GetPluginSrcDir(newPlugin.Name)

	if err := os.MkdirAll(plginSrcDir, 0777); err != nil {
		installSuccess = false
		return Plugin{}, "", fmt.Errorf("failed to create plugin src dir (%s), error: %s", plginSrcDir, err)
	}

	if err := cmdex.CopyDir(pluginSrcTmpDir, plginSrcDir, true); err != nil {
		installSuccess = false
		return Plugin{}, "", fmt.Errorf("failed to copy plugin from temp dir (%s) to (%s), error: %s", pluginSrcTmpDir, plginSrcDir, err)
	}

	executableURL := newPlugin.ExecutableURL()
	if binURL != "" {
		executableURL = binURL
	}
	if executableURL != "" {
		// Install plugin bin
		pluginBinTmpDir, err := pathutil.NormalizedOSTempDirPath("plugin-bin-tmp")
		if err != nil {
			installSuccess = false
			return Plugin{}, "", fmt.Errorf("failed to create plugin bin temp directory, error: %s", err)
		}
		defer func() {
			if err := os.RemoveAll(pluginBinTmpDir); err != nil {
				log.Warnf("Failed to remove path (%s)", pluginBinTmpDir)
			}
		}()

		pluginBinTmpFilePath := filepath.Join(pluginBinTmpDir, newPlugin.Name)

		if err := downloadPluginBin(executableURL, pluginBinTmpFilePath); err != nil {
			installSuccess = false
			return Plugin{}, "", fmt.Errorf("failed to download plugin executable from (%s), error: %s", executableURL, err)
		}

		plginBinDir := GetPluginBinDir(newPlugin.Name)

		if err := os.MkdirAll(plginBinDir, 0777); err != nil {
			installSuccess = false
			return Plugin{}, "", fmt.Errorf("failed to create plugin bin dir (%s), error: %s", plginBinDir, err)
		}

		pluginBinFilePath := filepath.Join(plginBinDir, newPlugin.Name)

		if err := cmdex.CopyFile(pluginBinTmpFilePath, pluginBinFilePath); err != nil {
			installSuccess = false
			return Plugin{}, "", fmt.Errorf("failed to copy plugin from temp dir (%s) to (%s), error: %s", pluginBinTmpFilePath, pluginBinFilePath, err)
		}

		if err := os.Chmod(pluginBinFilePath, 0777); err != nil {
			installSuccess = false
			return Plugin{}, "", fmt.Errorf("failed to make plugin bin executable, error: %s", err)
		}
	}

	newVersionStr := ""
	if newVersionPtr != nil {
		newVersionStr = (*newVersionPtr).String()
	}

	pluginDataDir := filepath.Join(pluginDir, "data")
	if err := os.MkdirAll(pluginDataDir, 0777); err != nil {
		installSuccess = false
		return Plugin{}, "", fmt.Errorf("failed to create plugin data dir (%s), error: %s", pluginDataDir, err)
	}

	if err := CreateAndAddPluginRoute(newPlugin.Name, srcURL, executableURL, newVersionStr, newVersinHash, newPlugin.TriggerEvent); err != nil {
		installSuccess = false
		return Plugin{}, "", fmt.Errorf("failed to add plugin route, error: %s", err)
	}

	if newVersionStr == "" {
		newVersionStr = "local"
	}

	return newPlugin, newVersionStr, nil
}

// DeletePlugin ...
func DeletePlugin(name string) error {
	pluginDir := GetPluginDir(name)

	if exists, err := pathutil.IsDirExists(pluginDir); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("Plugin (%s) not installed", name)
	}

	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("failed to delete dir (%s)", pluginDir)
	}

	return DeletePluginRoute(name)
}

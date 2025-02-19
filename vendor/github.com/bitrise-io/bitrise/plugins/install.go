package plugins

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/progress"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/sliceutil"
	ver "github.com/hashicorp/go-version"
)

//=======================================
// Util
//=======================================

func validatePath(pth string) error {
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		return fmt.Errorf("failed to check path (%s), error: %s", pth, err)
	} else if !exist {
		return fmt.Errorf("no file found at (%s)", pth)
	}
	return nil
}

func validateVersion(current, requiredMin ver.Version, requiredMax *ver.Version) error {
	if current.LessThan(&requiredMin) {
		return fmt.Errorf("current version (%s) is less then min version (%s)  ", current.String(), requiredMin.String())
	} else if requiredMax != nil && current.GreaterThan(requiredMax) {
		return fmt.Errorf("current version (%s) is greater than max version (%s)  ", current.String(), (*requiredMax).String())
	}
	return nil
}

func downloadPluginBin(sourceURL, destinationPth string) error {
	url, err := url.Parse(sourceURL)
	if err != nil {
		return fmt.Errorf("failed to parse url (%s), error: %s", sourceURL, err)
	}

	// Download local binary
	if url.Scheme == "file" {
		src := strings.Replace(sourceURL, url.Scheme+"://", "", -1)

		if err := command.CopyFile(src, destinationPth); err != nil {
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non success status code (%d)", resp.StatusCode)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to download from (%s), error: %s", sourceURL, err)
	}

	return nil
}

func cleanupPlugin(name string) error {
	pluginDir := GetPluginDir(name)

	if err := os.RemoveAll(pluginDir); err != nil {
		return err
	}

	return DeletePluginRoute(name)
}

// plugins from bitrise-core to bitrise-io GitHub org have been moved
// so it is better to not to detect this as a different plugin source
func isSourceURIChanged(installed, new string) bool {
	if urlsForOrg := func(org string) []string {
		return []string{
			"https://github.com/" + org + "/bitrise-plugins-init.git",
			"https://github.com/" + org + "/bitrise-plugins-step.git",
			"https://github.com/" + org + "/bitrise-plugins-analytics.git",
		}
	}; (installed == new) || (sliceutil.IsStringInSlice(installed, urlsForOrg("bitrise-core")) &&
		sliceutil.IsStringInSlice(new, urlsForOrg("bitrise-io"))) {
		return false
	}
	return true
}

func installLocalPlugin(pluginSourceURI, pluginLocalPth string) (Plugin, error) {
	// Parse & validate plugin
	tmpPluginYMLPath := filepath.Join(pluginLocalPth, pluginDefinitionFileName)

	if err := validatePath(tmpPluginYMLPath); err != nil {
		return Plugin{}, fmt.Errorf("bitrise-plugin.yml validation failed, error: %s", err)
	}

	newPlugin, err := ParsePluginFromYML(tmpPluginYMLPath)
	if err != nil {
		return Plugin{}, fmt.Errorf("failed to parse bitrise-plugin.yml (%s), error: %s", tmpPluginYMLPath, err)
	}

	if err := validatePlugin(newPlugin, pluginSourceURI, os.Args[0]); err != nil {
		return Plugin{}, fmt.Errorf("plugin validation failed, error: %s", err)
	}
	// ---

	// Check if plugin already installed
	if route, found, err := ReadPluginRoute(newPlugin.Name); err != nil {
		return Plugin{}, fmt.Errorf("failed to check if plugin already installed, error: %s", err)
	} else if found {
		if isSourceURIChanged(route.Source, pluginSourceURI) {
			return Plugin{}, fmt.Errorf("plugin already installed with name (%s) from different source (%s)", route.Name, route.Source)
		}

		installedPluginVersionPtr, err := GetPluginVersion(route.Name)
		if err != nil {
			return Plugin{}, fmt.Errorf("failed to check installed plugin (%s) version, error: %s", route.Name, err)
		}

		if installedPluginVersionPtr != nil {
			log.Warnf("installed plugin found with version %s, upgrading...", (*installedPluginVersionPtr).String())
		} else {
			log.Warnf("installed local plugin found, upgrading...")
		}
	}
	// ---

	tmpPluginDir, err := pathutil.NormalizedOSTempDirPath("__plugin__")
	if err != nil {
		return Plugin{}, fmt.Errorf("failed to create tmp plugin dir, error: %s", err)
	}
	defer func() {
		err = os.RemoveAll(tmpPluginDir)
		if err != nil {
			log.Warnf("Failed to clean up temp dir after plugin installation, error: %s", err)
		}
	}()

	// Install plugin executable
	executableURL := newPlugin.ExecutableURL()
	if executableURL != "" {
		tmpPluginBinDir := filepath.Join(tmpPluginDir, "bin")
		if err := os.MkdirAll(tmpPluginBinDir, 0755); err != nil {
			return Plugin{}, fmt.Errorf("failed to create tmp plugin bin dir, error: %s", err)
		}

		tmpPluginBinPth := filepath.Join(tmpPluginBinDir, newPlugin.Name)

		var err error
		progress.ShowIndicator("Downloading plugin binary", func() {
			err = downloadPluginBin(executableURL, tmpPluginBinPth)
		})
		if err != nil {
			return Plugin{}, fmt.Errorf("failed to download plugin executable from (%s), error: %s", executableURL, err)
		}
	}
	// ---

	// Install plugin source
	tmpPluginSrcDir := filepath.Join(tmpPluginDir, "src")
	if err := os.MkdirAll(tmpPluginSrcDir, 0755); err != nil {
		return Plugin{}, fmt.Errorf("failed to create tmp plugin src dir, error: %s", err)
	}

	if err := command.CopyDir(pluginLocalPth, tmpPluginSrcDir, true); err != nil {
		return Plugin{}, fmt.Errorf("failed to copy plugin from (%s) to (%s), error: %s", pluginLocalPth, tmpPluginSrcDir, err)
	}
	// ---

	// Create plugin work dir
	tmpPluginDataDir := filepath.Join(tmpPluginDir, "data")
	if err := os.MkdirAll(tmpPluginDataDir, 0755); err != nil {
		return Plugin{}, fmt.Errorf("failed to create tmp plugin data dir (%s), error: %s", tmpPluginDataDir, err)
	}
	// ---

	pluginDir := GetPluginDir(newPlugin.Name)
	if err := command.CopyDir(tmpPluginDir, pluginDir, true); err != nil {
		if err := cleanupPlugin(newPlugin.Name); err != nil {
			log.Warnf("Failed to cleanup plugin (%s), error: %s", newPlugin.Name, err)
		}
		return Plugin{}, fmt.Errorf("failed to copy plugin, error: %s", err)
	}

	if executableURL != "" {
		pluginBinDir := GetPluginBinDir(newPlugin.Name)
		pluginBinPth := filepath.Join(pluginBinDir, newPlugin.Name)
		if err := os.Chmod(pluginBinPth, 0755); err != nil {
			if err := cleanupPlugin(newPlugin.Name); err != nil {
				log.Warnf("Failed to cleanup plugin (%s), error: %s", newPlugin.Name, err)
			}
			return Plugin{}, fmt.Errorf("failed to make plugin bin executable, error: %s", err)
		}
	}

	return newPlugin, nil
}

func isLocalURL(urlStr string) bool {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	if parsed == nil {
		return false
	}
	return (parsed.Scheme == "file" || parsed.Scheme == "")
}

//=======================================
// Main
//=======================================

// InstallPlugin ...
func InstallPlugin(pluginSourceURI, versionTag string) (Plugin, string, error) {
	newVersion := ""
	pluginDir := ""

	if !isLocalURL(pluginSourceURI) {
		pluginSrcTmpDir, err := pathutil.NormalizedOSTempDirPath("plugin-src-tmp")
		if err != nil {
			return Plugin{}, "", fmt.Errorf("failed to create plugin src temp directory, error: %s", err)
		}
		defer func() {
			if err := os.RemoveAll(pluginSrcTmpDir); err != nil {
				log.Warnf("Failed to remove path (%s)", pluginSrcTmpDir)
			}
		}()

		version := ""
		err = nil

		progress.ShowIndicator("git clone plugin source", func() {
			version, err = GitCloneAndCheckoutVersionOrLatestVersion(pluginSrcTmpDir, pluginSourceURI, versionTag)
		})

		if err != nil {
			return Plugin{}, "", fmt.Errorf("failed to download plugin, error: %s", err)
		}

		pluginDir = pluginSrcTmpDir
		newVersion = version
	} else {
		pluginSourceURI = strings.TrimPrefix(pluginSourceURI, "file://")
		pluginDir = pluginSourceURI
	}

	newPlugin, err := installLocalPlugin(pluginSourceURI, pluginDir)
	if err != nil {
		return Plugin{}, "", err
	}

	// Register plugin
	if err := CreateAndAddPluginRoute(newPlugin, pluginSourceURI, newVersion); err != nil {
		if err := cleanupPlugin(newPlugin.Name); err != nil {
			log.Warnf("Failed to cleanup plugin (%s), error: %s", newPlugin.Name, err)
		}
		return Plugin{}, "", fmt.Errorf("failed to add plugin route, error: %s", err)
	}
	// ---

	return newPlugin, newVersion, nil
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

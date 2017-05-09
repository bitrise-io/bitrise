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
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
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
	if current.Compare(&requiredMin) == -1 {
		return fmt.Errorf("Current version (%s) is less then min version (%s)  ", current.String(), requiredMin.String())
	}

	if requiredMax != nil && current.Compare(requiredMax) == 1 {
		return fmt.Errorf("Current version (%s) is greater then max version (%s)  ", current.String(), (*requiredMax).String())
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

//=======================================
// Main
//=======================================

func installLocalPlugin(origSrcURI, srcDir string) (Plugin, error) {
	// Parse & validate plugin
	tmpPluginYMLPath := filepath.Join(srcDir, pluginDefinitionFileName)

	if err := validatePath(tmpPluginYMLPath); err != nil {
		return Plugin{}, fmt.Errorf("bitrise-plugin.yml validation failed, error: %s", err)
	}

	newPlugin, err := ParseAndValidatePluginFromYML(tmpPluginYMLPath)
	if err != nil {
		return Plugin{}, fmt.Errorf("failed to parse bitrise-plugin.yml (%s), error: %s", tmpPluginYMLPath, err)
	}
	// ---

	// Check if plugin already installed
	if route, found, err := ReadPluginRoute(newPlugin.Name); err != nil {
		return Plugin{}, fmt.Errorf("failed to check if plugin already installed, error: %s", err)
	} else if found {
		if route.Source != origSrcURI {
			return Plugin{}, fmt.Errorf("plugin already installed with name (%s) from different source (%s)", route.Name, route.Source)
		}

		installedPluginVersionPtr, err := GetPluginVersion(route.Name)
		if err != nil {
			return Plugin{}, fmt.Errorf("failed to check installed plugin (%s) version, error: %s", route.Name, err)
		}

		if installedPluginVersionPtr != nil {
			fmt.Println()
			log.Infof("Installed plugin found with version (%s), overriding it...", (*installedPluginVersionPtr).String())
		} else {
			fmt.Println()
			log.Infof("Installed local plugin found, overriding it...")
		}
	}
	// ---

	// Install plugin source
	installSuccess := false
	pluginDir := GetPluginDir(newPlugin.Name)
	if err := os.RemoveAll(pluginDir); err != nil {
		return Plugin{}, fmt.Errorf("failed to remove plugin dir (%s), error: %s", pluginDir, err)
	}
	defer func() {
		if !installSuccess {
			if err := os.RemoveAll(pluginDir); err != nil {
				log.Warnf("Failed to remove path (%s)", pluginDir)
			}
		}
	}()

	pluginSrcDir := GetPluginSrcDir(newPlugin.Name)
	if err := os.MkdirAll(pluginSrcDir, 0777); err != nil {
		return Plugin{}, fmt.Errorf("failed to create plugin src dir (%s), error: %s", pluginSrcDir, err)
	}
	if err := command.CopyDir(srcDir, pluginSrcDir, true); err != nil {
		return Plugin{}, fmt.Errorf("failed to copy plugin from temp dir (%s) to (%s), error: %s", srcDir, pluginSrcDir, err)
	}
	// ---

	// Install plugin executable
	executableURL := newPlugin.ExecutableURL()
	if executableURL != "" {
		pluginBinTmpDir, err := pathutil.NormalizedOSTempDirPath("plugin-bin-tmp")
		if err != nil {
			return Plugin{}, fmt.Errorf("failed to create plugin bin temp directory, error: %s", err)
		}
		defer func() {
			if err := os.RemoveAll(pluginBinTmpDir); err != nil {
				log.Warnf("Failed to remove path (%s)", pluginBinTmpDir)
			}
		}()

		pluginBinTmpFilePath := filepath.Join(pluginBinTmpDir, newPlugin.Name)

		if err := downloadPluginBin(executableURL, pluginBinTmpFilePath); err != nil {
			return Plugin{}, fmt.Errorf("failed to download plugin executable from (%s), error: %s", executableURL, err)
		}

		plginBinDir := GetPluginBinDir(newPlugin.Name)

		if err := os.MkdirAll(plginBinDir, 0777); err != nil {
			return Plugin{}, fmt.Errorf("failed to create plugin bin dir (%s), error: %s", plginBinDir, err)
		}

		pluginBinFilePath := filepath.Join(plginBinDir, newPlugin.Name)

		if err := command.CopyFile(pluginBinTmpFilePath, pluginBinFilePath); err != nil {
			return Plugin{}, fmt.Errorf("failed to copy plugin from temp dir (%s) to (%s), error: %s", pluginBinTmpFilePath, pluginBinFilePath, err)
		}

		if err := os.Chmod(pluginBinFilePath, 0777); err != nil {
			return Plugin{}, fmt.Errorf("failed to make plugin bin executable, error: %s", err)
		}
	}
	// ---

	// Create plugin work dir
	pluginDataDir := filepath.Join(pluginDir, "data")
	if err := os.MkdirAll(pluginDataDir, 0777); err != nil {
		return Plugin{}, fmt.Errorf("failed to create plugin data dir (%s), error: %s", pluginDataDir, err)
	}
	// ---

	installSuccess = true

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

// InstallPlugin ...
func InstallPlugin(srcURL, versionTag string) (Plugin, string, error) {
	newVersion := ""
	pluginDir := ""

	if !isLocalURL(srcURL) {
		pluginSrcTmpDir, err := pathutil.NormalizedOSTempDirPath("plugin-src-tmp")
		if err != nil {
			return Plugin{}, "", fmt.Errorf("failed to create plugin src temp directory, error: %s", err)
		}
		defer func() {
			if err := os.RemoveAll(pluginSrcTmpDir); err != nil {
				log.Warnf("Failed to remove path (%s)", pluginSrcTmpDir)
			}
		}()

		version, err := GitCloneAndCheckoutVersionOrLatestVersion(pluginSrcTmpDir, srcURL, versionTag)
		if err != nil {
			return Plugin{}, "", fmt.Errorf("failed to download plugin, error: %s", err)
		}

		pluginDir = pluginSrcTmpDir
		newVersion = version
	} else {
		srcURL = strings.TrimPrefix(srcURL, "file://")
		pluginDir = srcURL
	}

	newPlugin, err := installLocalPlugin(srcURL, pluginDir)
	if err != nil {
		return Plugin{}, "", err
	}

	// Register to bitrise
	if err := CreateAndAddPluginRoute(newPlugin, srcURL, newVersion); err != nil {
		return Plugin{}, "", fmt.Errorf("failed to add plugin route, error: %s", err)
	}

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

package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/log"
	ver "github.com/hashicorp/go-version"
	"github.com/urfave/cli"
)

const tagsURL = "https://api.github.com/repos/bitrise-io/bitrise/tags"
const downloadURL = "https://github.com/bitrise-io/bitrise/releases/download/%s/bitrise-%s-x86_64"

func checkUpdate() error {
	if configs.IsCIMode {
		return nil
	}
	if configs.CheckIsCLIUpdateCheckRequired() {
		log.Infof("Checking for new CLI version...")

		newVersion, err := newCLIVersion()
		if err != nil {
			return fmt.Errorf("Failed to check update for CLI, error: %s\n", err)
		}
		if newVersion != "" {
			printCLIUpdateInfos(newVersion)
		}

		if err := configs.SaveCLIUpdateCheck(); err != nil {
			return err
		}

	}

	installedPlugins, err := plugins.InstalledPluginList()
	if err != nil {
		return fmt.Errorf("Failed to list installed plugins: %s\n", err)
	}
	for _, plugin := range installedPlugins {
		if configs.CheckIsPluginUpdateCheckRequired(plugin.Name) {
			log.Infof("\nChecking for plugin (%s) new version...", plugin.Name)

			if newVersion, err := plugins.CheckForNewVersion(plugin); err != nil {
				log.Warnf("\nFailed to check for plugin (%s) new version, error: %s", plugin.Name, err)
			} else if newVersion != "" {
				plugins.PrintPluginUpdateInfos(newVersion, plugin)
			}

			if err := configs.SavePluginUpdateCheck(plugin.Name); err != nil {
				log.Warnf("\nFailed to update last check for plugin (%s), error: %s", plugin.Name, err)
			}
		}
	}
	return nil
}

func printCLIUpdateInfos(newVersion string) {
	log.Warnf("\nNew version (%s) of CLI available", newVersion)
	log.Printf("Run command to update CLI:\n")
	log.Donef("$ bitrise update")
}

func installedWithBrew() (bool, error) {
	if runtime.GOOS != `darwin` {
		return false, nil
	}
	if _, err := exec.LookPath("brew"); err != nil {
		return false, nil
	}

	out, err := exec.Command("brew", "list").Output()
	if err != nil {
		return false, err
	}
	formulas := strings.Split(string(out), "\n")
	for _, f := range formulas {
		if f == "bitrise" {
			return true, nil
		}
	}
	return false, nil
}

func newVersionFromBrew() (string, error) {
	if err := exec.Command("brew", "update").Run(); err != nil {
		return "", err
	}
	out, err := exec.Command("brew", "outdated", "--verbose").Output()
	if err != nil {
		return "", err
	}
	formulas := strings.Split(string(out), "\n")
	for _, f := range formulas {
		if strings.Contains(f, "bitrise") {
			return strings.Split(f, " ")[3], nil
		}
	}
	return "", nil
}

func newCLIVersion() (string, error) {
	withBrew, err := installedWithBrew()
	if err != nil {
		return "", err
	}
	if withBrew {
		return newVersionFromBrew()
	}

	latest, err := latestTag()
	if err != nil {
		return "", err
	}
	current, err := ver.NewVersion(version.VERSION)
	if latest.GreaterThan(current) {
		return latest.String(), nil
	}
	return "", nil
}

func latestTag() (*ver.Version, error) {
	resp, err := http.Get(tagsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return ver.NewVersion(result[0].Name)
}

func download(url, dst string) error {
	f, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Error while downloading url (%s), error: %v\n", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Can't download url (%s), status: %s\n", url, http.StatusText(resp.StatusCode))
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("Error while writing to file (%s), error: %v\n", dst, err)
	}
	return nil
}

func update(c *cli.Context) error {
	log.Infof("Updating Bitrise CLI...")
	withBrew, err := installedWithBrew()
	if err != nil {
		return err
	}
	if withBrew {
		cmd := exec.Command("brew", "upgrade", "bitrise")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	version := c.String(CLIVersionKey)
	if version == "" {
		latest, err := latestTag()
		if err != nil {
			return err
		}
		version = latest.String()
	}

	path, err := exec.LookPath("bitrise")
	if err != nil {
		return err
	}
	os := strings.Title(runtime.GOOS)
	url := fmt.Sprintf(downloadURL, version, os)
	return download(url, path)
}

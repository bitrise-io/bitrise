package cmdutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/bitrise-io/bitrise/v2/version"
	ver "github.com/hashicorp/go-version"
)

const tagsURL = "https://api.github.com/repos/bitrise-io/bitrise/tags"

// CheckUpdate ...
func CheckUpdate() error {
	if configs.IsCIMode {
		return nil
	}
	if configs.CheckIsCLIUpdateCheckRequired() {
		log.Infof("Checking for new CLI version...")

		newVersion, err := NewCLIVersion()
		if err != nil {
			return fmt.Errorf("failed to check update for CLI, error: %s", err)
		}
		if newVersion != "" {
			PrintCLIUpdateInfos(newVersion)
		}

		if err := configs.SaveCLIUpdateCheck(); err != nil {
			return err
		}
	}

	installedPlugins, err := plugins.InstalledPluginList()
	if err != nil {
		return fmt.Errorf("failed to list installed plugins: %s", err)
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

// PrintCLIUpdateInfos ...
func PrintCLIUpdateInfos(newVersion string) {
	log.Warnf("\nNew version (%s) of the Bitrise CLI available", newVersion)
	log.Printf("Run command to update the Bitrise CLI:")
	log.Donef("$ bitrise update")
}

// InstalledWithBrew ...
func InstalledWithBrew() (bool, error) {
	if runtime.GOOS != `darwin` {
		return false, nil
	}
	if _, err := exec.LookPath("brew"); err != nil {
		return false, nil
	}

	out, err := exec.Command("brew", "list", "--formula").Output()
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

// NewVersionFromBrew ...
func NewVersionFromBrew() (string, error) {
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
			// formula (version) < newVersion
			return strings.Split(f, " ")[3], nil
		}
	}
	return "", nil
}

// NewCLIVersion ...
func NewCLIVersion() (string, error) {
	withBrew, err := InstalledWithBrew()
	if err != nil {
		return "", err
	}
	if withBrew {
		return NewVersionFromBrew()
	}

	latest, err := LatestTag()
	if err != nil {
		return "", err
	}
	current, err := ver.NewVersion(version.VERSION)
	if err != nil {
		// Dev builds (no ldflags) have VERSION="dev" which is not valid semver -> skip the update check.
		return "", nil
	}
	if latest.GreaterThan(current) {
		return latest.String(), nil
	}
	return "", nil
}

// LatestTag ...
func LatestTag() (*ver.Version, error) {
	resp, err := http.Get(tagsURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf(err.Error())
		}
	}()

	var result []struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return ver.NewVersion(result[0].Name)
}

package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/bitrise-io/go-utils/command"
	ver "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

const (
	tagsURL         = "https://api.github.com/repos/bitrise-io/bitrise/tags"
	releasesBaseURL = "https://github.com/bitrise-io/bitrise/releases/download"
)

var updateCommand = &cobra.Command{
	Use:   "update",
	Short: "Updates the Bitrise CLI.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		logCommandParameters(cmd)

		if err := update(cmd); err != nil {
			log.Errorf("Update Bitrise CLI failed, error: %s", err)
			os.Exit(1)
		}

		return nil
	},
}

// The endpoints and the install method check are fields so tests can replace
// them with local fakes.
type updater struct {
	tagsURL         string
	releasesBaseURL string
	client          *http.Client
	isBrewInstall   func() (bool, error)
}

func newUpdater() updater {
	return updater{
		tagsURL:         tagsURL,
		releasesBaseURL: releasesBaseURL,
		client:          http.DefaultClient,
		isBrewInstall:   installedWithBrew,
	}
}

func init() {
	updateCommand.Flags().String("version", "", "version to update - only for GitHub release page installations.")
}

func checkUpdate() error {
	if configs.IsCIMode {
		return nil
	}
	if configs.CheckIsCLIUpdateCheckRequired() {
		log.Infof("Checking for new CLI version...")

		newVersion, err := newUpdater().newCLIVersion()
		if err != nil {
			return fmt.Errorf("failed to check update for CLI, error: %s", err)
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

func printCLIUpdateInfos(newVersion string) {
	log.Warnf("\nNew version (%s) of the Bitrise CLI available", newVersion)
	log.Printf("Run command to update the Bitrise CLI:")
	log.Donef("$ bitrise update")
}

func installedWithBrew() (bool, error) {
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
			// formula (version) < newVersion
			return strings.Split(f, " ")[3], nil
		}
	}
	return "", nil
}

func (u updater) newCLIVersion() (string, error) {
	withBrew, err := u.isBrewInstall()
	if err != nil {
		return "", err
	}
	if withBrew {
		return newVersionFromBrew()
	}

	latest, err := u.latestTag()
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

func (u updater) latestTag() (*ver.Version, error) {
	resp, err := u.client.Get(u.tagsURL)
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

func (u updater) download(version string) error {
	path, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}
	url := u.binaryURL(version, runtime.GOOS)

	tmpfile, err := os.CreateTemp("", "bitrise")
	if err != nil {
		return fmt.Errorf("can't create temporary file: %s", err)
	}

	resp, err := u.client.Get(url)
	if err != nil {
		return fmt.Errorf("error while downloading url (%s), error: %v", url, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf(err.Error())
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("can't download url (%s), status: %s", url, http.StatusText(resp.StatusCode))
	}

	_, err = io.Copy(tmpfile, resp.Body)
	if err != nil {
		return fmt.Errorf("error while writing to temp file, error: %v", err)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("can't remove file (%s), error: %s", path, err)
	}

	if err := CopyFile(tmpfile.Name(), path, true); err != nil {
		return err
	}
	log.Donef("Bitrise CLI is successfully updated!")

	return nil
}

func (u updater) binaryURL(version, goos string) string {
	return fmt.Sprintf("%s/v%s/%s", u.releasesBaseURL, version, assetName(goos))
}

func assetName(goos string) string {
	return fmt.Sprintf("bitrise-%s-x86_64", strings.ToUpper(goos[:1])+goos[1:])
}

func update(cmd *cobra.Command) error {
	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	logger.Infof("Updating Bitrise CLI...")

	versionFlag, _ := cmd.Flags().GetString("version")
	logger.Printf("Current version: %s", version.VERSION)

	u := newUpdater()

	withBrew, err := u.isBrewInstall()
	if err != nil {
		return err
	}

	if withBrew {
		logger.Infof("Bitrise CLI installed with homebrew")

		if versionFlag != "" {
			return errors.New("it seems you installed Bitrise CLI with Homebrew. Version flag is only supported for GitHub release page installations")
		}

		cmd := command.New("brew", "upgrade", "bitrise")

		logger.Printf("$ %s", cmd.PrintableCommandArgs())

		var out bytes.Buffer
		cmd.SetStdout(&out)
		cmd.SetStderr(&out)

		if err := cmd.Run(); err != nil {
			output := out.String()
			if strings.Contains(output, "already installed") {
				logger.Donef("Bitrise CLI is already up-to-date")
				return nil
			}

			logger.Printf(output)
			return err
		}

		logger.Printf(out.String())
		return nil
	}

	logger.Infof("Bitrise CLI installed from source")

	if versionFlag == "" {
		latest, err := u.latestTag()
		if err != nil {
			return err
		}
		versionFlag = latest.String()
	}

	if versionFlag == version.VERSION {
		logger.Donef("Bitrise CLI is already up-to-date")
		return nil
	}

	logger.Printf("Updating to version: %s", versionFlag)

	logger.Print("Downloading Bitrise CLI...")
	if err := u.download(versionFlag); err != nil {
		return err
	}

	return bitrise.RunSetup(logger, versionFlag, bitrise.SetupModeDefault, false, false)
}

// CopyFile ...
func CopyFile(src, dst string, remove bool) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := from.Close(); err != nil {
			log.Warnf(err.Error())
		}
	}()

	to, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer func() {
		if err := to.Close(); err != nil {
			log.Warnf(err.Error())
		}
	}()

	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}

	if remove {
		return os.Remove(src)
	}
	return nil
}

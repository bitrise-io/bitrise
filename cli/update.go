package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	ver "github.com/hashicorp/go-version"
	"github.com/urfave/cli"
)

const (
	tagsURL     = "https://api.github.com/repos/bitrise-io/bitrise/tags"
	downloadURL = "https://github.com/bitrise-io/bitrise/releases/download/%s/bitrise-%s-x86_64"
)

var updateCommand = cli.Command{
	Name:  "update",
	Usage: "Updates the Bitrise CLI.",
	Action: func(c *cli.Context) error {
		if err := update(c); err != nil {
			log.Errorf("Update Bitrise CLI failed, error: %s", err)
			os.Exit(1)
		}

		return nil
	},
	Flags: []cli.Flag{
		cli.StringFlag{Name: "version", Usage: "version to update - only for GitHub release page installations."},
	},
}

func checkUpdate() error {
	if configs.IsCIMode {
		return nil
	}
	if configs.CheckIsCLIUpdateCheckRequired() {
		log.Infof("Checking for new CLI version...")

		newVersion, err := newCLIVersion()
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
	log.Printf("Run command to update the Bitrise CLI:\n")
	log.Donef("$ bitrise update")
}

func installedWithBrew() (bool, error) {
	if runtime.GOOS != `darwin` {
		return false, nil
	}
	if _, err := exec.LookPath("brew"); err != nil {
		return false, nil
	}

	out, err := exec.Command("brew", "list").CombinedOutput()
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
	current := ver.Must(ver.NewVersion(version.VERSION))
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

func download(version string) error {
	path, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}
	url := fmt.Sprintf(downloadURL, version, strings.Title(runtime.GOOS))

	tmpfile, err := ioutil.TempFile("", "bitrise")
	if err != nil {
		return fmt.Errorf("can't create temporary file: %s", err)
	}

	resp, err := http.Get(url)
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

func update(c *cli.Context) error {
	log.Infof("Updating Bitrise CLI...")

	versionFlag := c.String("version")
	fmt.Printf("Current version: %s\n", version.VERSION)

	withBrew, err := installedWithBrew()
	if err != nil {
		return err
	}

	if withBrew {
		log.Infof("Bitrise CLI installer with homebrew")

		if versionFlag != "" {
			return errors.New("it seems you installed Bitrise CLI with Homebrew. Version flag is only supported for GitHub release page installations")
		}

		cmd := command.New("brew", "upgrade", "bitrise")

		log.Printf("$ %s", cmd.PrintableCommandArgs())

		var out bytes.Buffer
		cmd.SetStdout(&out)
		cmd.SetStderr(&out)

		if err := cmd.Run(); err != nil {
			output := out.String()
			if strings.Contains(output, "already installed") {
				log.Donef("Bitrise CLI is already up-to-date")
				return nil
			}

			log.Printf(output)
			return err
		}

		log.Printf(out.String())
		return nil
	}

	log.Infof("Bitrise CLI installer from source")

	if versionFlag == "" {
		latest, err := latestTag()
		if err != nil {
			return err
		}
		versionFlag = latest.String()
	}

	if versionFlag == version.VERSION {
		log.Donef("Bitrise CLI is already up-to-date")
		return nil
	}

	fmt.Printf("Updating to version: %s\n", versionFlag)

	fmt.Println("Downloading Bitrise CLI...")
	if err := download(versionFlag); err != nil {
		return err
	}

	return bitrise.RunSetup(versionFlag, false, false)
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

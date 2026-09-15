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
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"

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
	releaseNotesURL = "https://github.com/bitrise-io/bitrise/releases/tag"

	// The GitHub tags API returns the tags unordered, so the whole first page is
	// read and sorted here.
	tagsPerPage = 100
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

// A new major version is reported separately because it is never installed
// automatically: it can contain breaking changes, so the user has to ask for it.
type availableUpdates struct {
	sameMajor *ver.Version
	newMajor  *ver.Version
}

func newUpdater() updater {
	return updater{
		tagsURL:         tagsURL,
		releasesBaseURL: releasesBaseURL,
		client:          http.DefaultClient,
		// Cached: the check shells out to brew, and a single run asks for the
		// install method more than once.
		isBrewInstall: sync.OnceValues(installedWithBrew),
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

		u := newUpdater()

		updates, err := u.availableUpdates()
		if err != nil {
			return fmt.Errorf("failed to check update for CLI, error: %s", err)
		}
		if updates.sameMajor != nil {
			printCLIUpdateInfos(updates.sameMajor.String())
		}
		withBrew, err := u.isBrewInstall()
		if err != nil {
			return fmt.Errorf("failed to check update for CLI, error: %s", err)
		}
		printNewMajorInfos(updates.newMajor, withBrew)

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

func printNewMajorInfos(newMajor *ver.Version, withBrew bool) {
	if newMajor == nil {
		return
	}

	log.Warnf("\nBitrise CLI %s is available (new major version)", newMajor)
	log.Printf("It is not installed automatically, because a major version can contain breaking changes.")
	log.Printf("To install it:")
	log.Donef("$ %s", majorUpdateCommand(newMajor, withBrew))
	log.Printf("Release notes: %s/v%s", releaseNotesURL, newMajor)
}

func majorUpdateCommand(newMajor *ver.Version, withBrew bool) string {
	if withBrew {
		return "brew upgrade bitrise"
	}
	return fmt.Sprintf("bitrise update --version %s", newMajor)
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

func (u updater) availableUpdates() (availableUpdates, error) {
	current, err := ver.NewVersion(version.VERSION)
	if err != nil {
		// Dev builds (no ldflags) have VERSION="dev" which is not valid semver -> skip the update check.
		return availableUpdates{}, nil
	}

	withBrew, err := u.isBrewInstall()
	if err != nil {
		return availableUpdates{}, err
	}

	var published []*ver.Version
	if withBrew {
		published, err = brewVersions()
	} else {
		published, err = u.publishedVersions()
	}
	if err != nil {
		return availableUpdates{}, err
	}

	return newAvailableUpdates(current, published), nil
}

func newAvailableUpdates(current *ver.Version, published []*ver.Version) availableUpdates {
	var updates availableUpdates

	currentMajor := current.Segments()[0]
	for _, candidate := range published {
		switch major := candidate.Segments()[0]; {
		case major == currentMajor && candidate.GreaterThan(current):
			updates.sameMajor = candidate
		case major > currentMajor:
			updates.newMajor = candidate
		}
	}

	return updates
}

func brewVersions() ([]*ver.Version, error) {
	newVersion, err := newVersionFromBrew()
	if err != nil || newVersion == "" {
		return nil, err
	}

	parsed, err := ver.NewVersion(newVersion)
	if err != nil {
		return nil, err
	}
	return []*ver.Version{parsed}, nil
}

// Pre-releases are left out: they are published for opt-in testing, and must
// never be offered to someone who did not ask for them.
func (u updater) publishedVersions() ([]*ver.Version, error) {
	req, err := http.NewRequest(http.MethodGet, u.tagsURL, nil)
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
	query.Set("per_page", strconv.Itoa(tagsPerPage))
	req.URL.RawQuery = query.Encode()

	resp, err := u.client.Do(req)
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

	var versions []*ver.Version
	for _, tag := range result {
		parsed, err := ver.NewVersion(tag.Name)
		if err != nil || parsed.Prerelease() != "" {
			continue
		}
		versions = append(versions, parsed)
	}
	slices.SortFunc(versions, func(a, b *ver.Version) int { return a.Compare(b) })

	return versions, nil
}

func (u updater) download(version string) error {
	path, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}

	url, err := u.binaryURL(version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
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

	if err := replaceBinary(path, resp.Body); err != nil {
		return err
	}
	log.Donef("Bitrise CLI is successfully updated!")

	return nil
}

// The replacement is written next to the binary and renamed over it, so a
// failure at any point leaves the working CLI in place. The temporary file
// cannot live in the system temp dir, because a rename does not cross
// filesystems.
func replaceBinary(path string, contents io.Reader) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(path), ".bitrise-update-")
	if err != nil {
		return fmt.Errorf("can't create temporary file: %s", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			log.Warnf("failed to remove temporary file (%s), error: %s", tmpFile.Name(), err)
		}
	}()

	if _, err := io.Copy(tmpFile, contents); err != nil {
		return fmt.Errorf("error while writing to temp file, error: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("can't close temporary file, error: %s", err)
	}
	if err := os.Chmod(tmpFile.Name(), info.Mode()); err != nil {
		return fmt.Errorf("can't set the permissions of the temporary file, error: %s", err)
	}
	if err := os.Rename(tmpFile.Name(), path); err != nil {
		return fmt.Errorf("can't replace file (%s), error: %s", path, err)
	}

	return nil
}

func (u updater) binaryURL(version, goos, goarch string) (string, error) {
	asset, err := assetName(goos, goarch)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/v%s/%s", u.releasesBaseURL, version, asset), nil
}

func assetName(goos, goarch string) (string, error) {
	var arch string
	switch goarch {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "arm64"
	default:
		return "", fmt.Errorf("no Bitrise CLI release is published for %s/%s", goos, goarch)
	}

	return fmt.Sprintf("bitrise-%s-%s", strings.ToUpper(goos[:1])+goos[1:], arch), nil
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

	var newMajor *ver.Version
	if versionFlag == "" {
		updates, err := u.availableUpdates()
		if err != nil {
			return err
		}
		newMajor = updates.newMajor

		if updates.sameMajor == nil {
			logger.Donef("Bitrise CLI is already up-to-date")
			printNewMajorInfos(newMajor, withBrew)
			return nil
		}
		versionFlag = updates.sameMajor.String()
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

	if err := bitrise.RunSetup(logger, versionFlag, bitrise.SetupModeDefault, false, false); err != nil {
		return err
	}

	printNewMajorInfos(newMajor, withBrew)

	return nil
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

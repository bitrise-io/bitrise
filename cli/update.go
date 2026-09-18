package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/bitrise-io/go-utils/command"
	"github.com/spf13/cobra"
)

const downloadURL = "https://github.com/bitrise-io/bitrise/releases/download/v%s/bitrise-%s-x86_64"

var updateCommand = &cobra.Command{
	Use:   "update",
	Short: "Updates the Bitrise CLI.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cmdutil.LogCommandParameters(cmd)

		if err := update(cmd); err != nil {
			log.Errorf("Update Bitrise CLI failed, error: %s", err)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	updateCommand.Flags().String("version", "", "version to update - only for GitHub release page installations.")
}

func download(version string) error {
	path, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}
	url := fmt.Sprintf(downloadURL, version, strings.ToUpper(runtime.GOOS[:1])+runtime.GOOS[1:])

	tmpfile, err := os.CreateTemp("", "bitrise")
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

func update(cmd *cobra.Command) error {
	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	logger.Infof("Updating Bitrise CLI...")

	versionFlag, _ := cmd.Flags().GetString("version")
	logger.Printf("Current version: %s", version.VERSION)

	withBrew, err := cmdutil.InstalledWithBrew()
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
		latest, err := cmdutil.LatestTag()
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
	if err := download(versionFlag); err != nil {
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

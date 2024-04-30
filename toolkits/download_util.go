package toolkits

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/utils"
	"github.com/bitrise-io/go-utils/pathutil"
	"golang.org/x/sys/unix"
)

// InstallToolFromGitHub ...
func InstallToolFromGitHub(toolname, githubUser, toolVersion string) error {
	unameGOOS, err := utils.UnameGOOS()
	if err != nil {
		return fmt.Errorf("Failed to determine OS: %s", err)
	}
	unameGOARCH, err := utils.UnameGOARCH()
	if err != nil {
		return fmt.Errorf("Failed to determine ARCH: %s", err)
	}
	downloadURL := "https://github.com/" + githubUser + "/" + toolname + "/releases/download/" + toolVersion + "/" + toolname + "-" + unameGOOS + "-" + unameGOARCH

	return InstallFromURL(toolname, downloadURL)
}

// DownloadFile ...
func DownloadFile(downloadURL, targetDirPath string) error {
	outFile, err := os.Create(targetDirPath)
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Warnf("Failed to close (%s)", targetDirPath)
		}
	}()
	if err != nil {
		return fmt.Errorf("failed to create (%s), error: %s", targetDirPath, err)
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download from (%s), error: %s", downloadURL, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf("failed to close (%s) body", downloadURL)
		}
	}()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to download from (%s), error: %s", downloadURL, err)
	}

	return nil
}

// InstallFromURL ...
func InstallFromURL(toolBinName, downloadURL string) error {
	if len(toolBinName) < 1 {
		return fmt.Errorf("no Tool (bin) Name provided! URL was: %s", downloadURL)
	}

	tmpDir, err := pathutil.NormalizedOSTempDirPath("__tmp_download_dest__")
	if err != nil {
		return fmt.Errorf("failed to create tmp dir for download destination")
	}
	tmpDestinationPth := filepath.Join(tmpDir, toolBinName)

	if err := DownloadFile(downloadURL, tmpDestinationPth); err != nil {
		return fmt.Errorf("failed to download, error: %s", err)
	}

	bitriseToolsDirPath := configs.GetBitriseToolsDirPath()
	destinationPth := filepath.Join(bitriseToolsDirPath, toolBinName)

	if exist, err := pathutil.IsPathExists(destinationPth); err != nil {
		return fmt.Errorf("failed to check if file exist (%s), error: %s", destinationPth, err)
	} else if exist {
		if err := os.Remove(destinationPth); err != nil {
			return fmt.Errorf("failed to remove file (%s), error: %s", destinationPth, err)
		}
	}

	if err := MoveFile(tmpDestinationPth, destinationPth); err != nil {
		return fmt.Errorf("failed to copy (%s) to (%s), error: %s", tmpDestinationPth, destinationPth, err)
	}

	if err := os.Chmod(destinationPth, 0755); err != nil {
		return fmt.Errorf("failed to make file (%s) executable, error: %s", destinationPth, err)
	}

	return nil
}

// MoveFile ...
func MoveFile(oldpath, newpath string) error {
	err := os.Rename(oldpath, newpath)
	if err == nil {
		return nil
	}

	if linkErr, ok := err.(*os.LinkError); ok {
		if linkErr.Err == unix.EXDEV {
			info, err := os.Stat(oldpath)
			if err != nil {
				return err
			}

			data, err := os.ReadFile(oldpath)
			if err != nil {
				return err
			}

			err = os.WriteFile(newpath, data, info.Mode())
			if err != nil {
				return err
			}

			return os.Remove(oldpath)
		}
	}

	return err
}

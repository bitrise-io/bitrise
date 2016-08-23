package toolkits

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/progress"
)

// GoToolkit ...
type GoToolkit struct {
}

// Install ...
func (toolkit *GoToolkit) Install() error {
	versionStr := "1.7"
	osStr := runtime.GOOS
	archStr := runtime.GOARCH
	extentionStr := "tar.gz"
	if osStr == "windows" {
		extentionStr = "zip"
	}
	downloadURL := fmt.Sprintf("https://storage.googleapis.com/golang/go%s.%s-%s.%s",
		versionStr, osStr, archStr, extentionStr)
	log.Infoln("downloadURL: ", downloadURL)

	// bitriseToolkitsDirPath := configs.GetBitriseToolkitsDirPath()
	toolkitsTmpDirPath := getBitriseToolkitsTmpDirPath()
	if err := pathutil.EnsureDirExist(toolkitsTmpDirPath); err != nil {
		return fmt.Errorf("Failed to create Toolkits TMP directory, error: %s", err)
	}

	localFileName := "go.tar.gz"
	destinationPth := filepath.Join(toolkitsTmpDirPath, localFileName)

	var downloadErr error
	fmt.Print("=> Downloading ...")
	progress.SimpleProgress(".", 2*time.Second, func() {
		if err := tools.DownloadFile(downloadURL, destinationPth); err != nil {
			downloadErr = err
		}
	})
	if downloadErr != nil {
		return fmt.Errorf("Failed to download toolkit (%s), error: %s", downloadURL, downloadErr)
	}
	log.Infoln("Toolkit downloaded to: ", destinationPth)

	return nil
}

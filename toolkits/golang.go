package toolkits

import (
	"fmt"
	"path/filepath"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/bitrise-io/go-utils/pathutil"
)

// GoToolkit ...
type GoToolkit struct {
}

// Install ...
func (toolkit *GoToolkit) Install() error {
	versionStr := "1.7"
	osStr := runtime.GOOS
	archStr := runtime.GOARCH
	downloadURL := fmt.Sprintf("https://storage.googleapis.com/golang/go%s.%s-%s.tar.gz",
		versionStr, osStr, archStr)
	log.Infoln("downloadURL: ", downloadURL)

	// bitriseToolkitsDirPath := configs.GetBitriseToolkitsDirPath()
	toolkitsTmpDirPath := getBitriseToolkitsTmpDirPath()
	if err := pathutil.EnsureDirExist(toolkitsTmpDirPath); err != nil {
		return fmt.Errorf("Failed to create Toolkits TMP directory, error: %s", err)
	}

	localFileName := "go.tar.gz"
	destinationPth := filepath.Join(toolkitsTmpDirPath, localFileName)

	if err := tools.DownloadFile(downloadURL, destinationPth); err != nil {
		return fmt.Errorf("Failed to download toolkit (%s), error: %s", downloadURL, err)
	}
	log.Infoln("Toolkit downloaded to: ", destinationPth)

	return nil
}

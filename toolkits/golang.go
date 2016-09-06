package toolkits

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/progress"
	"github.com/bitrise-io/go-utils/retry"
)

// GoToolkit ...
type GoToolkit struct {
}

func goToolkitRootPath() string {
	return filepath.Join(configs.GetBitriseToolkitsDirPath(), "go")
}

func goToolkitInstallRootPath() string {
	return filepath.Join(goToolkitRootPath(), "go")
}

func goToolkitBinsPath() string {
	return filepath.Join(goToolkitInstallRootPath(), "bin")
}

// Bootstrap ...
func (toolkit *GoToolkit) Bootstrap() error {
	if configs.IsDebugUseSystemTools() {
		log.Warn("[BitriseDebug] Using system tools (system installed Go), instead of the ones in BITRISE_HOME")
		return nil
	}

	pthWithGoBins := configs.GeneratePATHEnvString(os.Getenv("PATH"), goToolkitBinsPath())
	if err := os.Setenv("PATH", pthWithGoBins); err != nil {
		return fmt.Errorf("Failed to set PATH to include the Go toolkit bins, error: %s", err)
	}

	if err := os.Setenv("GOROOT", goToolkitInstallRootPath()); err != nil {
		return fmt.Errorf("Failed to set GOROOT to Go toolkit root, error: %s", err)
	}

	return nil
}

func installGoTar(goTarGzPath string) error {
	installToPath := goToolkitRootPath()

	if err := os.RemoveAll(installToPath); err != nil {
		return fmt.Errorf("Failed to remove previous Go toolkit install (path: %s), error: %s", installToPath, err)
	}
	if err := pathutil.EnsureDirExist(installToPath); err != nil {
		return fmt.Errorf("Failed create Go toolkit directory (path: %s), error: %s", installToPath, err)
	}

	cmd := cmdex.NewCommand("tar", "-C", installToPath, "-xzf", goTarGzPath)
	if combinedOut, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		log.Errorln(" [!] Failed to uncompress Go toolkit, output:")
		log.Errorln(combinedOut)
		return fmt.Errorf("Failed to uncompress Go toolkit, error: %s", err)
	}
	return nil
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
		downloadErr = retry.Times(2).Wait(5).Try(func(attempt uint) error {
			if attempt > 0 {
				fmt.Println()
				fmt.Println("==> Download failed, retrying ...")
				fmt.Println()
			}
			return tools.DownloadFile(downloadURL, destinationPth)
		})
	})
	if downloadErr != nil {
		return fmt.Errorf("Failed to download toolkit (%s), error: %s", downloadURL, downloadErr)
	}
	log.Infoln("Toolkit downloaded to: ", destinationPth)

	fmt.Println("=> Installing ...")
	if err := installGoTar(destinationPth); err != nil {
		return fmt.Errorf("Failed to install Go toolkit, error: %s", err)
	}
	fmt.Println("=> Installing [DONE]")

	return nil
}

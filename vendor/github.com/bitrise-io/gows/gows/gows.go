package gows

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/bitrise-io/go-utils/pathutil"
)

// CreateGopathBinSymlink ...
func CreateGopathBinSymlink(origGOPATH string, workspaceRootPath string) error {
	fullWorkspaceBinPath := filepath.Join(workspaceRootPath, "bin")
	originalGopathBinPath, err := pathutil.AbsPath(filepath.Join(origGOPATH, "bin"))
	if err != nil {
		return fmt.Errorf("Failed to get the path of 'bin' dir inside your GOPATH (%s), error: %s", origGOPATH, err)
	}

	log.Debugf("=> Creating Symlink: (%s) -> (%s)", originalGopathBinPath, fullWorkspaceBinPath)

	// create symlink for GOPATH/bin, if not yet created
	if err := CreateOrUpdateSymlink(originalGopathBinPath, fullWorkspaceBinPath); err != nil {
		return fmt.Errorf("Failed to create GOPATH/bin symlink, error: %s", err)
	}

	log.Debugf(" [DONE] Symlink is in place")

	return nil
}

// CreateOrUpdateSymlink ...
func CreateOrUpdateSymlink(symlinkTargetPath, symlinkLocationPath string) error {
	fileInfo, isExists, err := pathutil.PathCheckAndInfos(symlinkLocationPath)
	if err != nil {
		return fmt.Errorf("Failed to check Symlink status (at: %s), error: %s", symlinkLocationPath, err)
	}
	isSymlinkAlreadyInPlace := false
	if isExists && fileInfo.Mode()&os.ModeSymlink != 0 {
		log.Debug(" Symlink already exists")
		originPth, err := os.Readlink(symlinkLocationPath)
		if err != nil {
			return fmt.Errorf("Symlink found (at: %s), but failed to open: %s", symlinkLocationPath, err)
		}

		if originPth == symlinkTargetPath {
			isSymlinkAlreadyInPlace = true
		} else {
			// remove
			log.Warningf("Symlink already exists (at: %s), but target (%s) is not the current one (%s)", symlinkLocationPath, originPth, symlinkTargetPath)
			log.Warning("Removing and re-creating the symlink ...")
			if err := os.Remove(symlinkLocationPath); err != nil {
				return fmt.Errorf("Failed to remove Symlink (at: %s), error: %s", symlinkLocationPath, err)
			}
		}
	}

	if !isSymlinkAlreadyInPlace {
		log.Debug(" Creating symlink ...")
		// create the parent directory
		if err := os.MkdirAll(filepath.Dir(symlinkLocationPath), 0777); err != nil {
			return fmt.Errorf("Failed to create base directory for symlink into: %s", symlinkLocationPath)
		}
		// create symlink
		if err := os.Symlink(symlinkTargetPath, symlinkLocationPath); err != nil {
			return fmt.Errorf("Failed to create symlink from project directory (%s) into gows Workspace directory (%s), error: %s", symlinkTargetPath, symlinkLocationPath, err)
		}
	}

	return nil
}

func filteredEnvsList(envsList []string, ignoreEnv string) []string {
	filteredEnvs := []string{}
	for _, envItem := range envsList {
		// an env item is a single string with the syntax: key=the value
		if !strings.HasPrefix(envItem, fmt.Sprintf("%s=", ignoreEnv)) {
			filteredEnvs = append(filteredEnvs, envItem)
		}
	}
	return filteredEnvs
}

// CreateCommand creates a command, prepared to run
// in the isolated workspace environment.
func CreateCommand(cmdWorkdir string, gopath string, cmdName string, cmdArgs ...string) *exec.Cmd {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = cmdWorkdir
	//
	cmdEnvs := os.Environ()
	cmdEnvs = filteredEnvsList(cmdEnvs, "GOPATH")
	cmdEnvs = filteredEnvsList(cmdEnvs, "PWD")
	cmdEnvs = append(cmdEnvs,
		fmt.Sprintf("GOPATH=%s", gopath),
		fmt.Sprintf("PWD=%s", cmdWorkdir),
	)
	cmd.Env = cmdEnvs

	return cmd
}

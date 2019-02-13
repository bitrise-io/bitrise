package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/gows/config"
	"github.com/bitrise-tools/gows/gows"
)

const (
	gowsCopyModeActiveFilePath = "./GOWS-COPY-MODE-ACTIVE"
)

// PrepareEnvironmentAndRunCommand ...
// Returns the exit code of the command and any error occured in the function
func PrepareEnvironmentAndRunCommand(userConfig config.UserConfigModel, cmdName string, cmdArgs ...string) (int, error) {
	projectConfig, err := config.LoadProjectConfigFromFile()
	if err != nil {
		log.Info("Run " + colorstring.Green("gows init") + " to initialize a workspace & gows config for this project")
		return 0, fmt.Errorf("Failed to read Project Config: %s", err)
	}

	gowsConfig, err := config.LoadGOWSConfigFromFile()
	if err != nil {
		return 0, fmt.Errorf("Failed to read gows configs: %s", err)
	}
	currWorkDir, err := os.Getwd()
	if err != nil {
		return 0, fmt.Errorf("[PrepareEnvironmentAndRunCommand] Failed to get current working directory: %s", err)
	}

	wsConfig, isFound := gowsConfig.WorkspaceForProjectLocation(currWorkDir)
	if !isFound {
		log.Debugln("No initialized workspace dir found for this project, initializing one ...")
		if err := initWorkspaceForProjectPath(currWorkDir, false); err != nil {
			return 0, fmt.Errorf("[PrepareEnvironmentAndRunCommand] Failed to initialize Workspace for Project: %s", err)
		}
		log.Debugln("[DONE] workspace dir initialized - continue running ...")

		// reload config
		gowsConfig, err := config.LoadGOWSConfigFromFile()
		if err != nil {
			return 0, fmt.Errorf("Failed to read gows configs: %s", err)
		}
		wsConfig, isFound = gowsConfig.WorkspaceForProjectLocation(currWorkDir)
	}
	if !isFound {
		log.Info("Run " + colorstring.Green("gows init") + " to initialize a workspace & gows config for this project")
		return 0, fmt.Errorf("No Workspace configuration found for the current project / working directory: %s", currWorkDir)
	}

	origGOPATH := os.Getenv("GOPATH")
	if origGOPATH == "" {
		// since Go 1.8 GOPATH is no longer required, it defaults to $HOME/go if not set:
		// https://golang.org/doc/go1.8#gopath
		p, err := pathutil.AbsPath("$HOME/go")
		if err != nil {
			return 0, errors.Wrap(err, "No GOPATH environment variable specified, and failed to get Abs path of default $HOME/go dir")
		}
		origGOPATH = p
	}

	if err := pathutil.EnsureDirExist(origGOPATH); err != nil {
		return 0, errors.Wrapf(err, "Failed to ensure that GOPATH exists at path: %s", origGOPATH)
	}

	if wsConfig.WorkspaceRootPath == "" {
		return 0, fmt.Errorf("No gows Workspace root path found for the current project / working directory: %s", currWorkDir)
	}
	if projectConfig.PackageName == "" {
		return 0, errors.New("No Package Name specified - make sure you initialized the workspace (with: gows init)")
	}

	if err := pathutil.EnsureDirExist(wsConfig.WorkspaceRootPath); err != nil {
		return 0, fmt.Errorf("Failed to create workspace root directory (path: %s), error: %s", wsConfig.WorkspaceRootPath, err)
	}

	if err := gows.CreateGopathBinSymlink(origGOPATH, wsConfig.WorkspaceRootPath); err != nil {
		return 0, fmt.Errorf("Failed to create GOPATH/bin symlink, error: %s", err)
	}

	fullPackageWorkspacePath := filepath.Join(wsConfig.WorkspaceRootPath, "src", projectConfig.PackageName)

	userConfigSyncMode := userConfig.SyncMode
	if userConfigSyncMode == "" {
		userConfigSyncMode = config.DefaultSyncMode
	}
	log.Debug("[PrepareEnvironmentAndRunCommand] specified Sync Mode : ", userConfigSyncMode)

	// --- prepare ---
	{
		fullPackageWorkspacePathFileInfo, fullPackageWorkspaceIsExists, err := pathutil.PathCheckAndInfos(fullPackageWorkspacePath)
		if err != nil {
			return 0, fmt.Errorf("Failed to check Symlink status (at: %s), error: %s", fullPackageWorkspacePath, err)
		}

		switch userConfigSyncMode {
		case config.SyncModeSymlink:
			// create symlink for Project->Workspace
			if fullPackageWorkspaceIsExists && fullPackageWorkspacePathFileInfo.Mode()&os.ModeSymlink == 0 {
				// directory (non symlink) exists - remove it
				log.Warningf("Directory exists (at: %s)", fullPackageWorkspacePath)
				log.Warning("Removing it ...")
				if err := os.RemoveAll(fullPackageWorkspacePath); err != nil {
					return 0, fmt.Errorf("Failed to remove Directory (at: %s), error: %s", fullPackageWorkspacePath, err)
				}
			}

			log.Debugf("=> Creating Symlink: (%s) -> (%s)", currWorkDir, fullPackageWorkspacePath)
			if err := gows.CreateOrUpdateSymlink(currWorkDir, fullPackageWorkspacePath); err != nil {
				return 0, fmt.Errorf("Failed to create Project->Workspace symlink, error: %s", err)
			}
			log.Debugf(" [DONE] Symlink is in place")
		case config.SyncModeCopy:
			// Sync project into workspace
			if fullPackageWorkspaceIsExists && fullPackageWorkspacePathFileInfo.Mode()&os.ModeSymlink != 0 {
				// symlink exists - remove it
				log.Warningf("Symlink exists (at: %s)", fullPackageWorkspacePath)
				log.Warning("Removing it ...")
				if err := os.Remove(fullPackageWorkspacePath); err != nil {
					return 0, fmt.Errorf("Failed to remove Symlink (at: %s), error: %s", fullPackageWorkspacePath, err)
				}
			}

			log.Debugf("=> Sync project content into workspace: (%s) -> (%s)", currWorkDir, fullPackageWorkspacePath)
			if err := syncDirWithDir(currWorkDir, fullPackageWorkspacePath); err != nil {
				return 0, fmt.Errorf("Failed to sync the project path / workdir into the Workspace, error: %s", err)
			}
			if err := writeGowsCopySyncActiveFileToPath(gowsCopyModeActiveFilePath, fullPackageWorkspacePath, currWorkDir); err != nil {
				log.Warningf(" [!] Failed to write gows-copy-mode-active file to path: %s", gowsCopyModeActiveFilePath)
			}
			log.Debugf(" [DONE] Sync project content into workspace")
		default:
			return 0, fmt.Errorf("Unsupported Sync Mode: %s", userConfigSyncMode)
		}
	}

	// Run the command, in the prepared Workspace
	exitCode, cmdErr := runCommand(fullPackageWorkspacePath, wsConfig, cmdName, cmdArgs...)

	// cleanup / finishing
	{
		switch userConfigSyncMode {
		case config.SyncModeSymlink:
			// nothing to do
		case config.SyncModeCopy:
			// Sync back from workspace into project
			log.Debugf("=> Sync workspace content into project: (%s) -> (%s)", fullPackageWorkspacePath, currWorkDir)
			if err := syncDirWithDir(fullPackageWorkspacePath, currWorkDir); err != nil {
				// we should return the command's exit code and error (if any)
				// maybe if the exitCode==0 and cmdErr==nil only then we could return an error here ...
				// for now we'll just print an error log, but it won't change the "output" of this function
				log.Errorf("Failed to sync back the project content from the Workspace, error: %s", err)
			} else {
				log.Debugf(" [DONE] Sync back project content from workspace")
			}
		default:
			return 0, fmt.Errorf("Unsupported Sync Mode: %s", userConfigSyncMode)
		}
	}

	return exitCode, cmdErr
}

func writeGowsCopySyncActiveFileToPath(pth, gowsWorkspacePath, originalProjectPath string) error {
	gowsCopyModeActiveContent := fmt.Sprintf(`gows workspace is active at the path: %s

Changes you do here (%s) WILL NOT SYNC, and WILL BE OVERWRITTEN by the changes done
inside the workspace (at path: %s) when sync/the current command is finished!

This file will be removed after the sync-back. After that it's safe to work
in this directory again.
`,
		gowsWorkspacePath, originalProjectPath, gowsWorkspacePath)

	return fileutil.WriteStringToFile(pth, gowsCopyModeActiveContent)
}

func syncDirWithDir(syncContentOf, syncIntoDir string) error {
	syncContentOf = filepath.Clean(syncContentOf)
	syncIntoDir = filepath.Clean(syncIntoDir)

	if err := pathutil.EnsureDirExist(syncIntoDir); err != nil {
		return fmt.Errorf("Failed to create target (at: %s), error: %s", syncIntoDir, err)
	}

	cmd := exec.Command("rsync", "-avhP", "--delete", syncContentOf+"/", syncIntoDir+"/")
	cmd.Stdin = os.Stdin

	log.Debugf("[syncDirWithDir] Running command: $ %s", command.NewWithCmd(cmd).PrintableCommandArgs())
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Error("[syncDirWithDir] Sync Error")
			log.Errorf("[syncDirWithDir] Output (Stdout) was: %s", out)
			log.Errorf("[syncDirWithDir] Error Output (Stderr) was: %s", exitError.Stderr)
		} else {
			log.Error("[syncDirWithDir] Failed to convert error to ExitError")
		}
		return fmt.Errorf("Failed to rsync between (%s) and (%s), error: %s", syncContentOf, syncIntoDir, err)
	}
	return nil
}

// runCommand runs the command with it's arguments
// Returns the exit code of the command and any error occured in the function
func runCommand(cmdWorkdir string, wsConfig config.WorkspaceConfigModel, cmdName string, cmdArgs ...string) (int, error) {
	log.Debugf("[RunCommand] Command Name: %s", cmdName)
	log.Debugf("[RunCommand] Command Args: %#v", cmdArgs)
	log.Debugf("[RunCommand] Command Work Dir: %#v", cmdWorkdir)

	cmd := gows.CreateCommand(cmdWorkdir, wsConfig.WorkspaceRootPath, cmdName, cmdArgs...)

	cmdExitCode := 0
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus, ok := exitError.Sys().(syscall.WaitStatus)
			if !ok {
				return 0, errors.New("Failed to cast exit status")
			}
			cmdExitCode = waitStatus.ExitStatus()
		}
		return cmdExitCode, err
	}

	return 0, nil
}

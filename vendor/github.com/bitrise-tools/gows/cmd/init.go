package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-tools/gows/config"
	"github.com/bitrise-tools/gows/goutil"
	"gopkg.in/viktorbenei/cobra.v0"
)

var (
	isAllowReset = false
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:           "init",
	Short:         "Initialize gows for your Go project",
	Long:          ``,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("More than one package argument specified")
		}
		packageName := ""
		if len(args) < 1 {
			log.Info("No package name specified, scanning it automatically ...")
			scanRes, err := AutoScanPackageName()
			if err != nil {
				return fmt.Errorf("Failed to auto-scan the package name: %s", err)
			}
			if scanRes == "" {
				return errors.New("Empty package name scanned")
			}
			packageName = scanRes
			log.Infof(" Scanned package name: %s", packageName)
		} else {
			packageName = args[0]
		}

		if isAllowReset {
			log.Warning(colorstring.Red("Will reset the related workspace"))
		}

		if err := InitGOWS(packageName, isAllowReset); err != nil {
			return fmt.Errorf("Failed to initialize: %s", err)
		}

		log.Info("Successful init - " + colorstring.Green("gows is ready for use!"))

		return nil
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&isAllowReset,
		"reset", "",
		false,
		"Delete previous workspace (if any) and initialize a new one")
}

// AutoScanPackageName ...
func AutoScanPackageName() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Errorf("[AutoScanPackageName] (Error) Output was: %s", exitError.Stderr)
		} else {
			log.Error("[AutoScanPackageName] Failed to convert error to ExitError")
		}
		return "", fmt.Errorf("Failed to get git remote url for origin: %s", err)
	}

	outStr := string(out)
	gitRemoteStr := strings.TrimSpace(outStr)
	log.Debugf("Found Git Remote: %s", gitRemoteStr)
	packageName, err := goutil.ParsePackageNameFromURL(gitRemoteStr)
	if err != nil {
		return "", fmt.Errorf("Failed to parse package name from remote URL (%s), error: %s", outStr, err)
	}
	if packageName == "" {
		return "", fmt.Errorf("Failed to parse package name from remote URL (%s), error: empty package name parsed", outStr)
	}

	return packageName, nil
}

func initGoWorkspaceAtPath(wsRootPath string) error {
	if err := os.MkdirAll(filepath.Join(wsRootPath, "src"), 0777); err != nil {
		return fmt.Errorf("Failed to create GOPATH/src directory: %s", err)
	}
	return nil
}

// initWorkspaceForProjectPath ...
// Workspaces are linked to project paths, not to package IDs!
// You can have multiple workspaces for the same package ID, but not for the
// same (project) path.
func initWorkspaceForProjectPath(projectPath string, isAllowReset bool) error {
	log.Debug("[Init] Initializing Workspace & Config ...")

	gowsWorspacesRootDirAbsPath, err := config.GOWSWorspacesRootDirAbsPath()
	if err != nil {
		return fmt.Errorf("Failed to get absolute path for gows workspaces root dir, error: %s", err)
	}

	// Create the Workspace
	gowsConfig, err := config.LoadGOWSConfigFromFile()
	if err != nil {
		return fmt.Errorf("Failed to load gows config: %s", err)
	}

	projectWorkspaceAbsPath := ""
	wsConfig, isFound := gowsConfig.WorkspaceForProjectLocation(projectPath)
	if isFound {
		if wsConfig.WorkspaceRootPath == "" {
			return fmt.Errorf("A workspace is found for this project (path: %s), but the workspace root directory path is not defined!", projectPath)
		}
		projectWorkspaceAbsPath = wsConfig.WorkspaceRootPath

		if isAllowReset {
			if err := os.RemoveAll(projectWorkspaceAbsPath); err != nil {
				return fmt.Errorf("Failed to delete previous workspace at path: %s", projectWorkspaceAbsPath)
			}
			// init a new one
			projectWorkspaceAbsPath = ""
		} else {
			log.Warning(colorstring.Yellow("A workspace already exists for this project") + " (" + projectWorkspaceAbsPath + "), will be reused.")
			log.Warning("If you want to delete the previous workspace of this project and generate a new one you should run: " + colorstring.Green("gows clear"))
		}
	}

	if projectWorkspaceAbsPath == "" {
		// generate one
		projectBaseWorkspaceDirName := fmt.Sprintf("%s-%d", filepath.Base(projectPath), time.Now().Unix())
		projectWorkspaceAbsPath = filepath.Join(gowsWorspacesRootDirAbsPath, projectBaseWorkspaceDirName)
	}

	log.Debugf("  projectWorkspaceAbsPath: %s", projectWorkspaceAbsPath)
	if err := initGoWorkspaceAtPath(projectWorkspaceAbsPath); err != nil {
		return fmt.Errorf("Failed to initialize workspace at path: %s", projectWorkspaceAbsPath)
	}
	log.Debugf("  Workspace successfully created")

	// Save the location into Workspace config
	{
		workspaceConf := config.WorkspaceConfigModel{
			WorkspaceRootPath: projectWorkspaceAbsPath,
		}
		gowsConfig.Workspaces[projectPath] = workspaceConf

		if err := config.SaveGOWSConfigToFile(gowsConfig); err != nil {
			return fmt.Errorf("Failed to save gows config: %s", err)
		}
	}
	log.Debug("[Init] Workspace Config saved")

	return nil
}

// InitGOWS ...
func InitGOWS(packageName string, isAllowReset bool) error {
	log.Infof("[Init] Initializing package: %s", packageName)

	log.Info("[Init] Initializing Project Config ...")
	{
		projectConf := config.ProjectConfigModel{
			PackageName: packageName,
		}

		if err := config.SaveProjectConfigToFile(projectConf); err != nil {
			return fmt.Errorf("Failed to write Project Config into file: %s", err)
		}
		log.Infof("       [OK] Project Config file saved to: %s", colorstring.Green(config.ProjectConfigFilePath))
	}

	log.Info("[Init] Initializing User Config ...")
	{
		if _, err := config.LoadUserConfigFromFile(); err == nil {
			log.Infof("       [OK] User Config file already exists at %s - will not generate a new one", config.UserConfigFilePath)
		} else {
			userConf := config.CreateDefaultUserConfig()

			if err := config.SaveUserConfigToFile(userConf); err != nil {
				return fmt.Errorf("Failed to write User Config into file: %s", err)
			}
			log.Info("       [OK] User Config file saved as " + colorstring.Green(config.UserConfigFilePath) + " - " + colorstring.Yellow("please add it to your .gitignore file!"))
		}
	}

	// init workspace for project (path)
	currWorkDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get current working directory: %s", err)
	}

	if err := initWorkspaceForProjectPath(currWorkDir, isAllowReset); err != nil {
		return fmt.Errorf("Failed to initialize Workspace for Project: %s", err)
	}

	return nil
}

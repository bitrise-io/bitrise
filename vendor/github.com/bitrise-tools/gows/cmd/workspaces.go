package cmd

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-tools/gows/config"
	"gopkg.in/viktorbenei/cobra.v0"
)

// workspacesCmd represents the workspaces command
var workspacesCmd = &cobra.Command{
	Use:           "workspaces",
	Short:         "List registered gows projects -> workspaces path pairs",
	Long:          ``,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		gowsConfig, err := config.LoadGOWSConfigFromFile()
		if err != nil {
			return fmt.Errorf("Failed to load gows config: %s", err)
		}

		currWorkDir, err := os.Getwd()
		if err != nil {
			log.Debugf("Failed to get current working directory: %s", err)
		}

		fmt.Println()
		fmt.Println("=== Registered gows [project -> workspace] path list ===")
		for projectPath, wsConfig := range gowsConfig.Workspaces {
			if projectPath == currWorkDir {
				fmt.Println(colorstring.Greenf(" * %s -> %s", projectPath, wsConfig.WorkspaceRootPath))
			} else {
				fmt.Printf(" * %s -> %s\n", projectPath, wsConfig.WorkspaceRootPath)
			}
		}
		fmt.Println("========================================================")
		fmt.Println()

		return nil
	},
}

func init() {
	RootCmd.AddCommand(workspacesCmd)
}

package cmd

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-tools/gows/config"
	"gopkg.in/viktorbenei/cobra.v0"
)

// wspathCmd represents the wspath command
var wspathCmd = &cobra.Command{
	Use:   "wspath",
	Short: "Prints the current workspace path",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		gowsConfig, err := config.LoadGOWSConfigFromFile()
		if err != nil {
			return fmt.Errorf("Failed to load gows config: %s", err)
		}

		currWorkDir, err := os.Getwd()
		if err != nil {
			log.Debugf("Failed to get current working directory: %s", err)
		}

		wsConfig, isFound := gowsConfig.WorkspaceForProjectLocation(currWorkDir)
		if !isFound {
			return fmt.Errorf("No Workspace configuration found for the current project / working directory: %s", currWorkDir)
		}

		fmt.Println(wsConfig.WorkspaceRootPath)

		return nil
	},
}

func init() {
	RootCmd.AddCommand(wspathCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// wspathCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wspathCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

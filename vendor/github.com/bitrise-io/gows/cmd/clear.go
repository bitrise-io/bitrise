package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-tools/gows/config"
	"gopkg.in/viktorbenei/cobra.v0"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear out the project's workspace",
	Long:  `Clear out the project's workspace`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectConfig, err := config.LoadProjectConfigFromFile()
		if err != nil {
			log.Info("Run " + colorstring.Green("gows init") + " to initialize a workspace & gows config for this project")
			return fmt.Errorf("Failed to read Project Config: %s", err)
		}
		if projectConfig.PackageName == "" {
			log.Info("Run " + colorstring.Green("gows init") + " to initialize a workspace & gows config for this project")
			return fmt.Errorf("Package Name is empty")
		}

		if err := InitGOWS(projectConfig.PackageName, true); err != nil {
			return fmt.Errorf("Failed to initialize: %s", err)
		}

		log.Println("Done, workspace is clean!")

		return nil
	},
}

func init() {
	RootCmd.AddCommand(clearCmd)
}

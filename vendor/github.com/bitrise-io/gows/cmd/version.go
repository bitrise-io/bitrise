package cmd

import (
	"fmt"

	"github.com/bitrise-tools/gows/version"
	"gopkg.in/viktorbenei/cobra.v0"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", version.VERSION)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

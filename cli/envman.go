package cli

import (
	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/spf13/cobra"
)

var envmanCmd = &cobra.Command{
	Use:                "envman",
	Short:              "Runs an envman command.",
	DisableFlagParsing: true,
	RunE:               runEnvman,
}

func runEnvman(cmd *cobra.Command, args []string) error {
	logCommandParameters(cmd)
	cmd2 := command.NewWithStandardOuts("envman", args...).SetStdin(os.Stdin)
	return cmd2.Run()
}

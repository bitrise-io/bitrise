package cli

import (
	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/urfave/cli"
)

var envmanCommand = cli.Command{
	Name:            "envman",
	Usage:           "Runs an envman command.",
	SkipFlagParsing: true,
	Action: func(c *cli.Context) error {
		logCommandParameters(c)

		if err := runCommandWith("envman", c); err != nil {
			failf("Command failed, error: %s", err)
		}
		return nil
	},
}

func runCommandWith(toolName string, c *cli.Context) error {
	args := c.Args()
	cmd := command.NewWithStandardOuts(toolName, args...).SetStdin(os.Stdin)
	return cmd.Run()
}

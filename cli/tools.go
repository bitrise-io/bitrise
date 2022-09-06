package cli

import (
	"os"

	"github.com/bitrise-io/go-utils/command"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/urfave/cli"
)

var stepmanCommand = cli.Command{
	Name:            "stepman",
	Usage:           "Runs a stepman command.",
	SkipFlagParsing: true,
	Action: func(c *cli.Context) error {
		if err := runCommandWith("stepman", c); err != nil {
			log.Fatalf("Command failed, error: %s", err)
		}
		return nil
	},
}

var envmanCommand = cli.Command{
	Name:            "envman",
	Usage:           "Runs an envman command.",
	SkipFlagParsing: true,
	Action: func(c *cli.Context) error {
		if err := runCommandWith("envman", c); err != nil {
			log.Fatalf("Command failed, error: %s", err)
		}
		return nil
	},
}

func runCommandWith(toolName string, c *cli.Context) error {
	args := c.Args()
	cmd := command.NewWithStandardOuts(toolName, args...).SetStdin(os.Stdin)
	return cmd.Run()
}

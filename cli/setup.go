package cli

import (
	"os"

	log "github.com/bitrise-io/bitrise/advancedlog"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/urfave/cli"
)

var setupCommand = cli.Command{
	Name:  "setup",
	Usage: "Setup the current host. Install every required tool to run Workflows.",
	Action: func(c *cli.Context) error {
		if err := setup(c); err != nil {
			log.Errorf("Setup failed, error: %s", err)
			os.Exit(1)
		}
		return nil
	},
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "full",
			Usage: "Also calls 'brew doctor'.",
		},
		cli.BoolFlag{
			Name:  "clean",
			Usage: "Removes bitrise's workdir before setup.",
		},
	},
}

// PrintBitriseHeaderASCIIArt ...
func PrintBitriseHeaderASCIIArt(appVersion string) {
	// generated here: http://patorjk.com/software/taag/#p=display&f=ANSI%20Shadow&t=Bitrise
	log.Print(`
  ██████╗ ██╗████████╗██████╗ ██╗███████╗███████╗
  ██╔══██╗██║╚══██╔══╝██╔══██╗██║██╔════╝██╔════╝
  ██████╔╝██║   ██║   ██████╔╝██║███████╗█████╗
  ██╔══██╗██║   ██║   ██╔══██╗██║╚════██║██╔══╝
  ██████╔╝██║   ██║   ██║  ██║██║███████║███████╗
  ╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚══════╝╚══════╝`)
	log.Print()
	log.Donef("  version: %s", appVersion)
	log.Print()
}

func setup(c *cli.Context) error {
	PrintBitriseHeaderASCIIArt(c.App.Version)

	fullMode := c.Bool("full")
	cleanMode := c.Bool("clean")

	if err := bitrise.RunSetup(c.App.Version, fullMode, cleanMode); err != nil {
		return err
	}

	log.Print()
	log.Infof("To start using bitrise:")
	log.Printf("* cd into your project's directory (if you're not there already)")
	log.Printf("* call: bitrise init")
	log.Printf("* follow the guide")
	log.Print()
	log.Donef("That's all :)")

	return nil
}

package cli

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/go-utils/colorstring"
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
	fmt.Println(`
  ██████╗ ██╗████████╗██████╗ ██╗███████╗███████╗
  ██╔══██╗██║╚══██╔══╝██╔══██╗██║██╔════╝██╔════╝
  ██████╔╝██║   ██║   ██████╔╝██║███████╗█████╗
  ██╔══██╗██║   ██║   ██╔══██╗██║╚════██║██╔══╝
  ██████╔╝██║   ██║   ██║  ██║██║███████║███████╗
  ╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚══════╝╚══════╝`)
	fmt.Println()
	fmt.Println(colorstring.Greenf("Version: %s", appVersion))
	fmt.Println()
}

func setup(c *cli.Context) error {
	PrintBitriseHeaderASCIIArt(c.App.Version)

	fullMode := c.Bool("full")
	cleanMode := c.Bool("clean")

	if cleanMode {
		if err := configs.DeleteBitriseConfigDir(); err != nil {
			return err
		}

		if err := configs.InitPaths(); err != nil {
			return err
		}

		if err := plugins.InitPaths(); err != nil {
			return err
		}
	}

	if err := bitrise.RunSetup(c.App.Version, fullMode); err != nil {
		return err
	}

	log.Infoln("To start using bitrise:")
	log.Infoln("* cd into your project's directory (if you're not there already)")
	log.Infoln("* call: bitrise init")
	log.Infoln("* follow the guide")
	fmt.Println()
	log.Infoln("That's all :)")

	return nil
}

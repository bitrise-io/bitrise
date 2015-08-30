package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/codegangsta/cli"
)

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

func setup(c *cli.Context) {
	PrintBitriseHeaderASCIIArt(c.App.Version)

	if err := bitrise.RunSetup(c.App.Version, c.Bool(MinimalModeKey)); err != nil {
		log.Fatalln("Setup failed:", err)
	}

	log.Infoln("To start using bitrise:")
	log.Infoln("* cd into your project's directory (if you're not there already)")
	log.Infoln("* call: bitrise init")
	log.Infoln("* follow the guide")
	fmt.Println()
	log.Infoln("That's all :)")
}

package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/codegangsta/cli"
)

// PrintBitriseHeaderASCIIArt ...
func PrintBitriseHeaderASCIIArt() {
	// generated here: http://patorjk.com/software/taag/#p=display&f=ANSI%20Shadow&t=Bitrise
	fmt.Println(`
  ██████╗ ██╗████████╗██████╗ ██╗███████╗███████╗
  ██╔══██╗██║╚══██╔══╝██╔══██╗██║██╔════╝██╔════╝
  ██████╔╝██║   ██║   ██████╔╝██║███████╗█████╗
  ██╔══██╗██║   ██║   ██╔══██╗██║╚════██║██╔══╝
  ██████╔╝██║   ██║   ██║  ██║██║███████║███████╗
  ╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚══════╝╚══════╝`)
	fmt.Println()
}

func doSetup(c *cli.Context) {
	PrintBitriseHeaderASCIIArt()

	if err := bitrise.RunSetup(c.App.Version); err != nil {
		log.Fatalln("Setup failed:", err)
	}

	log.Infoln("To start using bitrise:")
	log.Infoln("* cd into your project's directory (if you're not there already)")
	log.Infoln("* call: bitrise init")
	log.Infoln("* follow the guide")
	fmt.Println()
	log.Infoln("That's all :)")
}

package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/urfave/cli"
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

func setup(c *cli.Context) error {
	PrintBitriseHeaderASCIIArt(c.App.Version)

	if c.IsSet(MinimalModeKey) {
		log.Warn("'minimal' flag is deprecated")
		log.Warn("currently setup without any flag does the same as minimal setup in previous versions")
		log.Warn("use 'full' flag to achive the full setup process (which includes the 'brew doctor' call)")
		fmt.Println()
	}

	if err := bitrise.RunSetup(c.App.Version, c.Bool(FullModeKey)); err != nil {
		log.Fatalf("Setup failed, error: %s", err)
	}

	log.Infoln("To start using bitrise:")
	log.Infoln("* cd into your project's directory (if you're not there already)")
	log.Infoln("* call: bitrise init")
	log.Infoln("* follow the guide")
	fmt.Println()
	log.Infoln("That's all :)")

	return nil
}

package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func doSetup(c *cli.Context) {
	log.Info("[BITRISE_CLI] - Setup -- Coming soon!")
	os.Exit(1)
}

func init() {

}

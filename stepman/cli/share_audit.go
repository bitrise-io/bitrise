package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func printFinishAudit(share ShareModel) {
	fmt.Println()
	log.Infof(" * "+colorstring.Green("[OK]")+" Your step (%s) (%s) is valid.", share.StepID, share.StepTag)
	fmt.Println()
	fmt.Println("   " + GuideTextForShareFinish())
}

func shareAudit(c *cli.Context) {
	share, err := ReadShareSteplibFromFile()
	if err != nil {
		log.Error(err)
		log.Fatal("You have to start sharing with `stepman share start`, or you can read instructions with `stepman share`")
	}

	_, found := stepman.ReadRoute(share.Collection)
	if !found {
		log.Fatalln("No route found for collectionURI (%s)", share.Collection)
	}

	if err := auditStepLib(share.Collection); err != nil {
		log.Fatalln("[STEPMAN] - Audit failed %s", err)
	}

	printFinishAudit(share)
}

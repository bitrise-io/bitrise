package cli

import (
	"fmt"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/bitrise-tools/colorstring"
	"github.com/urfave/cli"
)

func printFinishAudit(share ShareModel, toolMode bool) {
	b := colorstring.NewBuilder()
	b.Green("your step (%s@%s) is valid", share.StepID, share.StepTag).NewLine()
	b.NewLine()
	b.Plain(GuideTextForShareFinish(toolMode))
	fmt.Println(b.String())
}

func shareAudit(c *cli.Context) error {
	toolMode := c.Bool(ToolMode)

	log.Infof("Validating Step share params...")
	share, err := ReadShareSteplibFromFile()
	if err != nil {
		log.Errorf(err.Error())
		fail("You have to start sharing with `stepman share start`, or you can read instructions with `stepman share`")
	}
	log.Donef("all inputs are valid")

	fmt.Println()
	log.Infof("Auditing the StepLib...")
	_, found := stepman.ReadRoute(share.Collection)
	if !found {
		fail("No route found for collectionURI (%s)", share.Collection)
	}

	if err := auditStepLibBeforeSharePullRequest(share.Collection); err != nil {
		fail("Audit Step Collection failed, err: %s", err)
	}

	printFinishAudit(share, toolMode)
	fmt.Println()

	return nil
}

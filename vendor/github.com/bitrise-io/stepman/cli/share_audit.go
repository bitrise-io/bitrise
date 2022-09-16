package cli

import (
	"github.com/bitrise-io/colorstring"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func printFinishAudit(share ShareModel, toolMode bool) {
	b := colorstring.NewBuilder()
	b.Green("your step (%s@%s) is valid", share.StepID, share.StepTag).NewLine()
	b.NewLine()
	b.Plain(GuideTextForShareFinish(toolMode))
	log.Println(b.String())
}

func shareAudit(c *cli.Context) error {
	toolMode := c.Bool(ToolMode)

	log.Infof("Validating Step share params...")
	share, err := ReadShareSteplibFromFile()
	if err != nil {
		log.Errorf(err.Error())
		failf("You have to start sharing with `stepman share start`, or you can read instructions with `stepman share`")
	}
	log.Donef("all inputs are valid")

	log.Println()
	log.Infof("Auditing the StepLib...")
	_, found := stepman.ReadRoute(share.Collection)
	if !found {
		failf("No route found for collectionURI (%s)", share.Collection)
	}

	if err := auditStepLibBeforeSharePullRequest(share.Collection); err != nil {
		failf("Audit Step Collection failed, err: %s", err)
	}

	printFinishAudit(share, toolMode)
	log.Println()

	return nil
}

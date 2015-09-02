package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func auditStep(step models.StepModel, stepID, version string) error {
	if err := step.Validate(true); err != nil {
		return err
	}

	pth, err := pathutil.NormalizedOSTempDirPath(stepID + version)
	if err != nil {
		return err
	}
	if err := cmdex.GitCloneTagOrBranchAndValidateCommitHash(step.Source.Git, pth, version, step.Source.Commit); err != nil {
		return err
	}

	return nil
}

func auditStepLib(gitURI string) error {
	if exist, err := stepman.RootExistForCollection(gitURI); err != nil {
		log.Fatal("[STEPMAN] - Failed to check routing:", err)
	} else if !exist {
		log.Fatalf("[STEPMAN] - Missing routing for collection, call 'stepman setup -c %s' before audit.", gitURI)
	}

	collection, err := stepman.ReadStepSpec(gitURI)
	if err != nil {
		log.Fatalln("[STEPMAN] - Failed to read steps spec (spec.json)")
	}

	for stepID, stepGroup := range collection.Steps {
		log.Debugf("Start audit StepGrup, with ID: (%s)", stepID)
		for version, step := range stepGroup.Versions {
			log.Debugf("Start audit Step (%s) (%s)", stepID, version)
			if err := auditStep(step, stepID, version); err != nil {
				log.Errorf(" * "+colorstring.Redf("[FAILED] ")+"Failed audit (%s) (%s)", stepID, version)
				log.Fatalf("   Error: %s", err.Error())
			} else {
				log.Infof(" * "+colorstring.Greenf("[OK] ")+"Success audit (%s) (%s)", stepID, version)
			}
		}
	}
	return nil
}

func audit(c *cli.Context) {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}

	if err := auditStepLib(collectionURI); err != nil {
		log.Fatalln("[STEPMAN] - Audit failed %s", err)
	}
}

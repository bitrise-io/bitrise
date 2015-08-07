package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func finish(c *cli.Context) {
	share, err := stepman.ReadShareSteplibFromFile()
	if err != nil {
		log.Fatal(err)
	}

	route, found := stepman.ReadRoute(share.Collection)
	if !found {
		log.Fatalln("No route found for collectionURI (%s)", share.Collection)
	}

	collectionDir := stepman.GetCollectionBaseDirPath(route)
	log.Info("Collection dir:", collectionDir)
	if err := stepman.DoGitCheckoutBranch(collectionDir, share.StepName); err != nil {
		log.Fatal(err)
	}

	stepDirInSteplib := stepman.GetStepCollectionDirPath(route, share.StepName, share.StepTag)
	stepYMLPathInSteplib := stepDirInSteplib + "/step.yml"
	log.Info("New step.yml:", stepYMLPathInSteplib)
	if err := stepman.DoGitAdd(collectionDir, stepYMLPathInSteplib); err != nil {
		log.Fatal(err)
	}

	if err := stepman.DoGitCommit(collectionDir, share.StepName+share.StepTag); err != nil {
		log.Fatal(err)
	}

	if err := stepman.DoGitPush(collectionDir); err != nil {
		log.Fatal(err)
	}
}

package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func finish(c *cli.Context) {
	share, err := ReadShareSteplibFromFile()
	if err != nil {
		log.Fatal(err)
	}

	route, found := stepman.ReadRoute(share.Collection)
	if !found {
		log.Fatalln("No route found for collectionURI (%s)", share.Collection)
	}

	collectionDir := stepman.GetCollectionBaseDirPath(route)
	log.Info("Checkout to branch:", share.StepID)
	if err := cmdex.GitCheckoutBranch(collectionDir, share.StepID); err != nil {
		log.Fatal(err)
	}

	stepDirInSteplib := stepman.GetStepCollectionDirPath(route, share.StepID, share.StepTag)
	stepYMLPathInSteplib := stepDirInSteplib + "/step.yml"
	log.Info("New step.yml:", stepYMLPathInSteplib)
	if err := cmdex.GitAddFile(collectionDir, stepYMLPathInSteplib); err != nil {
		log.Fatal(err)
	}

	log.Info("Do commit")
	if err := cmdex.GitCommit(collectionDir, share.StepID+share.StepTag); err != nil {
		log.Fatal(err)
	}

	log.Info("Do push")
	if err := cmdex.GitPushToOrigin(collectionDir, share.StepID); err != nil {
		log.Fatal(err)
	}

	if err := DeleteShareSteplibFile(); err != nil {
		log.Fatal(err)
	}
}

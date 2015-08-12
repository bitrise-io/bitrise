package cli

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func createPullRequestURL(gitURI string) (string, error) {
	if !strings.HasSuffix(gitURI, ".git") {
		return "", errors.New("Git URI should have suffix `.git`")
	}
	pullURL := strings.TrimSuffix(gitURI, ".git")
	pullURL = pullURL + "/pulls"
	return pullURL, nil
}

func printFinishShare(pullURL string) {
	fmt.Println()
	log.Info(" * " + colorstring.Green("[OK] ") + "Yeah!! You rock!!")
	fmt.Println()
	fmt.Println("   " + GuideTextForFinish())
}

func finish(c *cli.Context) {
	share, err := ReadShareSteplibFromFile()
	if err != nil {
		log.Error(err)
		log.Fatal("You have to start sharing with `stepman share start`, or you can read instructions with `stepman share`")
	}

	route, found := stepman.ReadRoute(share.Collection)
	if !found {
		log.Fatalln("No route found for collectionURI (%s)", share.Collection)
	}

	collectionDir := stepman.GetCollectionBaseDirPath(route)
	if err := cmdex.GitCheckIsNoChanges(collectionDir); err == nil {
		log.Warn("No git changes!")
		pullURL, err := createPullRequestURL(share.Collection)
		if err != nil {
			log.Fatal(err)
		}
		printFinishShare(pullURL)
		return
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

	log.Info("Pushing to your fork: ", share.Collection)
	if err := cmdex.GitPushToOrigin(collectionDir, share.StepID); err != nil {
		log.Fatal(err)
	}
	pullURL, err := createPullRequestURL(share.Collection)
	if err != nil {
		log.Fatal(err)
	}
	printFinishShare(pullURL)
}

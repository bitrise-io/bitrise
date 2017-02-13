package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func printFinishShare() {
	fmt.Println()
	log.Info(" * " + colorstring.Green("[OK] ") + "Yeah!! You rock!!")
	fmt.Println()
	fmt.Println("   " + GuideTextForFinish())
	fmt.Println()
	msg := `   You can create a pull request in your forked StepLib repository,
   if you used the main StepLib repository then your repository's url looks like: ` + `
   ` + colorstring.Green("https://github.com/[your-username]/bitrise-steplib") + `

   On GitHub you can find a ` + colorstring.Green("'Compare & pull request'") + ` button, in the ` + colorstring.Green("'Your recently pushed branches:'") + ` section,
   which will bring you to the 'Open a pull request' page, where you can review and create your Pull Request.
	`
	fmt.Println(msg)
}

func finish(c *cli.Context) error {
	share, err := ReadShareSteplibFromFile()
	if err != nil {
		log.Error(err)
		log.Fatal("You have to start sharing with `stepman share start`, or you can read instructions with `stepman share`")
	}

	route, found := stepman.ReadRoute(share.Collection)
	if !found {
		log.Fatalln("No route found for collectionURI (%s)", share.Collection)
	}

	collectionDir := stepman.GetLibraryBaseDirPath(route)
	if err := git.CheckIsNoChanges(collectionDir); err == nil {
		log.Warn("No git changes!")
		printFinishShare()
		return nil
	}

	stepDirInSteplib := stepman.GetStepCollectionDirPath(route, share.StepID, share.StepTag)
	stepYMLPathInSteplib := stepDirInSteplib + "/step.yml"
	log.Info("New step.yml:", stepYMLPathInSteplib)
	if err := git.AddFile(collectionDir, stepYMLPathInSteplib); err != nil {
		log.Fatal(err)
	}

	log.Info("Do commit")
	msg := share.StepID + " " + share.StepTag
	if err := git.Commit(collectionDir, msg); err != nil {
		log.Fatal(err)
	}

	log.Info("Pushing to your fork: ", share.Collection)
	if err := git.PushToOrigin(collectionDir, share.ShareBranchName()); err != nil {
		log.Fatal(err)
	}
	printFinishShare()

	return nil
}

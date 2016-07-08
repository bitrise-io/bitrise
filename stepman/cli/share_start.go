package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func printFinishStart(specPth string, toolMode bool) {
	fmt.Println()
	log.Info(" * "+colorstring.Green("[OK]")+" You can find your StepLib repo at: ", specPth)
	fmt.Println()
	fmt.Println("   " + GuideTextForShareCreate(toolMode))
}

func start(c *cli.Context) error {
	// Input validation
	toolMode := c.Bool(ToolMode)

	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}

	if route, found := stepman.ReadRoute(collectionURI); found {
		collLocalPth := stepman.GetCollectionBaseDirPath(route)
		log.Warnf("StepLib found locally at: %s", collLocalPth)
		log.Info("For sharing it's required to work with a clean StepLib repository.")
		if val, err := goinp.AskForBool("Would you like to remove the local version (your forked StepLib repository) and re-clone it?"); err != nil {
			log.Fatalln(err)
		} else {
			if !val {
				log.Errorln("Unfortunately we can't continue with sharing without a clean StepLib repository.")
				log.Fatalln("Please finish your changes, run this command again and allow it to remove the local StepLib folder!")
			}
			if err := stepman.CleanupRoute(route); err != nil {
				log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
			}
		}
	}

	// cleanup
	if err := DeleteShareSteplibFile(); err != nil {
		log.Fatal(err)
	}

	var route stepman.SteplibRoute
	isSuccess := false
	defer func() {
		if !isSuccess {
			if err := stepman.CleanupRoute(route); err != nil {
				log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
			}
			if err := DeleteShareSteplibFile(); err != nil {
				log.Fatal(err)
			}
		}
	}()

	// Preparing steplib
	alias := stepman.GenerateFolderAlias()
	route = stepman.SteplibRoute{
		SteplibURI:  collectionURI,
		FolderAlias: alias,
	}

	pth := stepman.GetCollectionBaseDirPath(route)
	if err := cmdex.GitClone(collectionURI, pth); err != nil {
		log.Fatal("[STEPMAN] - Failed to setup step spec:", err)
	}

	specPth := pth + "/steplib.yml"
	collection, err := stepman.ParseStepCollection(specPth)
	if err != nil {
		log.Fatal("[STEPMAN] - Failed to read step spec:", err)
	}

	if err := stepman.WriteStepSpecToFile(collection, route); err != nil {
		log.Fatal("[STEPMAN] - Failed to save step spec:", err)
	}

	if err := stepman.AddRoute(route); err != nil {
		log.Fatal("[STEPMAN] - Failed to setup routing:", err)
	}

	share := ShareModel{
		Collection: collectionURI,
	}
	if err := WriteShareSteplibToFile(share); err != nil {
		log.Fatal("[STEPMAN] - Failed to save share steplib to file:", err)
	}

	isSuccess = true
	printFinishStart(pth, toolMode)

	return nil
}

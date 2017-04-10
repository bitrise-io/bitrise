package cli

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/retry"
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
		log.Fatalf("No step collection specified")
	}

	if route, found := stepman.ReadRoute(collectionURI); found {
		collLocalPth := stepman.GetLibraryBaseDirPath(route)
		log.Warnf("StepLib found locally at: %s", collLocalPth)
		log.Info("For sharing it's required to work with a clean StepLib repository.")
		if val, err := goinp.AskForBool("Would you like to remove the local version (your forked StepLib repository) and re-clone it?"); err != nil {
			log.Fatalf("Failed to ask for input, error: %s", err)
		} else {
			if !val {
				log.Errorf("Unfortunately we can't continue with sharing without a clean StepLib repository.")
				log.Fatalf("Please finish your changes, run this command again and allow it to remove the local StepLib folder!")
			}
			if err := stepman.CleanupRoute(route); err != nil {
				log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
			}
		}
	}

	// cleanup
	if err := DeleteShareSteplibFile(); err != nil {
		log.Fatalf("Failed to delete share steplib file, error: %s", err)
	}

	var route stepman.SteplibRoute
	isSuccess := false
	defer func() {
		if !isSuccess {
			if err := stepman.CleanupRoute(route); err != nil {
				log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
			}
			if err := DeleteShareSteplibFile(); err != nil {
				log.Fatalf("Failed to delete share steplib file, error: %s", err)
			}
		}
	}()

	// Preparing steplib
	alias := stepman.GenerateFolderAlias()
	route = stepman.SteplibRoute{
		SteplibURI:  collectionURI,
		FolderAlias: alias,
	}

	pth := stepman.GetLibraryBaseDirPath(route)
	if err := retry.Times(2).Wait(3 * time.Second).Try(func(attempt uint) error {
		return git.Clone(collectionURI, pth)
	}); err != nil {
		log.Fatalf("Failed to setup step spec (url: %s) version (%s), error: %s", collectionURI, pth, err)
	}

	specPth := pth + "/steplib.yml"
	collection, err := stepman.ParseStepCollection(specPth)
	if err != nil {
		log.Fatalf("Failed to read step spec, error: %s", err)
	}

	if err := stepman.WriteStepSpecToFile(collection, route); err != nil {
		log.Fatalf("Failed to save step spec, error: %s", err)
	}

	if err := stepman.AddRoute(route); err != nil {
		log.Fatalf("Failed to setup routing, error: %s", err)
	}

	share := ShareModel{
		Collection: collectionURI,
	}
	if err := WriteShareSteplibToFile(share); err != nil {
		log.Fatalf("Failed to save share steplib to file, error: %s", err)
	}

	isSuccess = true
	printFinishStart(pth, toolMode)

	return nil
}

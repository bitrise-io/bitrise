package cli

import (
	"fmt"
	"time"

	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func showSubcommandHelp(c *cli.Context) {
	if err := cli.ShowSubcommandHelp(c); err != nil {
		log.Warnf("Failed to show help, error: %s", err)
	}
}

func start(c *cli.Context) error {
	// Input validation
	log.Infof("Validating Step share params...")

	toolMode := c.Bool(ToolMode)

	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		showSubcommandHelp(c)
		failf("No step collection specified")
	}

	log.Donef("all inputs are valid")

	fmt.Println()
	log.Infof("Preparing StepLib...")

	if route, found := stepman.ReadRoute(collectionURI); found {
		collLocalPth := stepman.GetLibraryBaseDirPath(route)
		log.Printf("StepLib found locally at: %s", collLocalPth)
		log.Warnf("For sharing it's required to work with a clean StepLib repository.")
		if val, err := goinp.AskForBool("Would you like to remove the local version (your forked StepLib repository) and re-clone it?"); err != nil {
			failf("Failed to ask for input, error: %s", err)
		} else {
			if !val {
				log.Errorf("Unfortunately we can't continue with sharing without a clean StepLib repository.")
				failf("Please finish your changes, run this command again and allow it to remove the local StepLib folder!")
			}
			if err := stepman.CleanupRoute(route); err != nil {
				log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
			}
		}
	}

	// cleanup
	if err := DeleteShareSteplibFile(); err != nil {
		failf("Failed to delete share steplib file, error: %s", err)
	}

	var route stepman.SteplibRoute
	isSuccess := false
	defer func() {
		if !isSuccess {
			if err := stepman.CleanupRoute(route); err != nil {
				log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
			}
			if err := DeleteShareSteplibFile(); err != nil {
				failf("Failed to delete share steplib file, error: %s", err)
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
		repo, err := git.New(pth)
		if err != nil {
			return err
		}
		return repo.Clone(collectionURI).Run()
	}); err != nil {
		failf("Failed to setup step spec (url: %s) version (%s), error: %s", collectionURI, pth, err)
	}

	specPth := stepman.GetStepCollectionSpecPath(route)
	collection, err := stepman.ParseStepCollection(specPth)
	if err != nil {
		failf("Failed to read step spec, error: %s", err)
	}

	if err := stepman.WriteStepSpecToFile(collection, route); err != nil {
		failf("Failed to save step spec, error: %s", err)
	}

	if err := stepman.AddRoute(route); err != nil {
		failf("Failed to setup routing, error: %s", err)
	}

	log.Donef("StepLib prepared at: %s", pth)

	share := ShareModel{
		Collection: collectionURI,
	}
	if err := WriteShareSteplibToFile(share); err != nil {
		failf("Failed to save share steplib to file, error: %s", err)
	}

	isSuccess = true

	fmt.Println()
	fmt.Println(GuideTextForShareCreate(toolMode))
	fmt.Println()

	return nil
}

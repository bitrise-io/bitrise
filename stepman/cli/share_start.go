package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/colorstring"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func start(c *cli.Context) {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}

	if route, found := stepman.ReadRoute(collectionURI); found {
		log.Warn("[STEPMAN] - Steplib found localy.")
		if val, err := goinp.AskForBool("Would you like to remove local version of steplib and re-clone it? [yes/no]"); err != nil {
			log.Fatalln("Error:", err)
		} else {
			if !val {
				return
			}
			if err := stepman.CleanupRoute(route); err != nil {
				log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
			}
		}
	}

	// Preparing steplib
	alias := stepman.GenerateFolderAlias()
	route := stepman.SteplibRoute{
		SteplibURI:  collectionURI,
		FolderAlias: alias,
	}

	pth := stepman.GetCollectionBaseDirPath(route)
	if err := stepman.DoGitClone(collectionURI, pth); err != nil {
		if err := stepman.CleanupRoute(route); err != nil {
			log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
		}
		log.Fatal("[STEPMAN] - Failed to setup step spec:", err)
	}

	specPth := pth + "/steplib.yml"
	collection, err := stepman.ParseStepCollection(specPth)
	if err != nil {
		if err := stepman.CleanupRoute(route); err != nil {
			log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
		}
		log.Fatal("[STEPMAN] - Failed to read step spec:", err)
	}

	if err := stepman.WriteStepSpecToFile(collection, route); err != nil {
		if err := stepman.CleanupRoute(route); err != nil {
			log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
		}
		log.Fatal("[STEPMAN] - Failed to save step spec:", err)
	}

	if err := stepman.AddRoute(route); err != nil {
		if err := stepman.CleanupRoute(route); err != nil {
			log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
		}
		log.Fatal("[STEPMAN] - Failed to setup routing:", err)
	}

	fmt.Println()
	log.Info(" * "+colorstring.Green("[OK]")+" You can find your step lib repo at:", specPth)
	fmt.Println()
}

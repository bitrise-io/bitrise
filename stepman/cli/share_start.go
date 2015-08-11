package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
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

	// cleanup
	var route stepman.SteplibRoute
	isSuccess := false
	defer func() {
		if !isSuccess {
			if err := stepman.CleanupRoute(route); err != nil {
				log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
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

	fmt.Println()
	log.Info(" * "+colorstring.Green("[OK]")+" You can find your step lib repo at:", specPth)
	log.Info("   Next call `stepman share create --tag VERSION_TAG --git GIT_URI` to move your step into your steplib fork.")
	fmt.Println()
}

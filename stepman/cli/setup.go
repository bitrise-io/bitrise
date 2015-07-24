package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func setup(c *cli.Context) {
	log.Debugln("[STEPMAN] - Setup")

	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		collectionURI = os.Getenv(CollectionPathEnvKey)
	}
	if collectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}

	if exist, err := stepman.RootExistForCollection(collectionURI); err != nil {
		log.Fatal("[STEPMAN] - Failed to check routing:", err)
	} else if exist {
		log.Debugln("[STEPMAN] - Nothing to setup, everything's ready.")
		return
	}

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
		log.Fatal("[STEPMAN] - Failed to get step spec path:", err)
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

	log.Info("[STEPMAN] - Initialized")
}

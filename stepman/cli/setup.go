package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func setup(c *cli.Context) {
	log.Debugln("[STEPMAN] - Setup")

	// Input validation
	collectionURI := c.String(CollectionKey)
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

	// Cleanup
	isSuccess := false
	defer func() {
		if !isSuccess {
			if err := stepman.CleanupRoute(route); err != nil {
				log.Errorf("Failed to cleanup route for uri: %s", collectionURI)
			}
		}
	}()

	pth := stepman.GetCollectionBaseDirPath(route)
	if !c.Bool(LocalCollectionKey) {
		if err := cmdex.GitClone(collectionURI, pth); err != nil {
			log.Fatal("[STEPMAN] - Failed to setup step spec:", err)
		}
	} else {
		log.Warn("Using local step lib")
		// Local spec path
		log.Info("Creating collection dir: ", pth)
		if err := os.MkdirAll(pth, 0777); err != nil {
			log.Fatal("[STEPMAN] - Failed to create collection dir: ", pth, "| error: ", err)
		}
		log.Info("Collection dir created - OK.")
		if err := cmdex.CopyDir(collectionURI, pth, true); err != nil {
			log.Fatal("[STEPMAN] - Failed to setup local step spec:", err)
		}
	}

	if err := stepman.ReGenerateStepSpec(route); err != nil {
		log.Fatal(err)
	}

	if copySpecJSONPath := c.String(CopySpecJSONKey); copySpecJSONPath != "" {
		log.Info("Copying spec YML to path: ", copySpecJSONPath)

		sourceSpecJSONPth := stepman.GetStepSpecPath(route)
		if err := cmdex.CopyFile(sourceSpecJSONPth, copySpecJSONPath); err != nil {
			log.Fatalf("Failed to copy spec.json from (%s) to (%s)", sourceSpecJSONPth, copySpecJSONPath)
		}
	}

	if err := stepman.AddRoute(route); err != nil {
		log.Fatal("[STEPMAN] - Failed to setup routing:", err)
	}

	isSuccess = true
}

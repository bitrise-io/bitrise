package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func update(c *cli.Context) {
	log.Info("[STEPMAN] - Update")

	collectionURIs := []string{}

	// StepSpec collection path
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Info("[STEPMAN] - No step collection specified, update all")
		collectionURIs = stepman.GetAllStepCollectionPath()
	} else {
		collectionURIs = []string{collectionURI}
	}

	for _, URI := range collectionURIs {
		route, found := stepman.ReadRoute(URI)
		if !found {
			log.Fatal("No route found for lib: " + collectionURI)
		}

		pth := stepman.GetCollectionBaseDirPath(route)
		if exists, err := pathutil.IsPathExists(pth); err != nil {
			log.Fatal(err)
		} else if !exists {
			log.Fatal("[STEPMAN] - Not initialized")
		}

		if err := stepman.DoGitPull(pth); err != nil {
			log.Fatal(err)
		}

		if err := stepman.ReGenerateStepSpec(route); err != nil {
			log.Fatalf("Failed to update collection:%s error:%v", URI, err)
		}
	}

	log.Info("[STEPMAN] - Updated")
}

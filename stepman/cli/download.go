package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func download(c *cli.Context) error {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}
	route, found := stepman.ReadRoute(collectionURI)
	if !found {
		log.Fatal("No route found for lib: " + collectionURI)
	}

	id := c.String(IDKey)
	if id == "" {
		log.Fatal("[STEPMAN] - Missing step id")
	}

	collection, err := stepman.ReadStepSpec(collectionURI)
	if err != nil {
		log.Fatal("[STEPMAN] - Failed to read step spec:", err)
	}

	version := c.String(VersionKey)
	if version == "" {
		log.Debug("[STEPMAN] - Missing step version -- Use latest version")

		latest, err := collection.GetLatestStepVersion(id)
		if err != nil {
			log.Fatal("[STEPMAN] - Failed to get step latest version: ", err)
		}
		log.Debug("[STEPMAN] - Latest version of step: ", latest)
		version = latest
	}

	update := c.Bool(UpdateKey)

	// Check step exist in collection
	step, found := collection.GetStep(id, version)
	if !found {
		if update {
			log.Infof("[STEPMAN] - Collection doesn't contain step (id:%s) (version:%s) -- Updating collection", id, version)
			if err := stepman.ReGenerateStepSpec(route); err != nil {
				log.Fatalf("[STEPMAN] - Failed to update collection:%s error:%v", collectionURI, err)
			}

			if _, found := collection.GetStep(id, version); !found {
				log.Fatalf("[STEPMAN] - Even the updated collection doesn't contain step (id:%s) (version:%s)", id, version)
			}
		} else {
			log.Fatalf("[STEPMAN] - Collection doesn't contain step (id:%s) (version:%s)", id, version)
		}
	}

	if err := stepman.DownloadStep(collectionURI, collection, id, version, step.Source.Commit); err != nil {
		log.Fatal("[STEPMAN] - Failed to download step")
	}

	return nil
}

package cli

import (
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func download(c *cli.Context) error {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		failf("No step collection specified")
	}
	route, found := stepman.ReadRoute(collectionURI)
	if !found {
		failf("No route found for lib: %s", collectionURI)
	}

	id := c.String(IDKey)
	if id == "" {
		failf("Missing step id")
	}

	collection, err := stepman.ReadStepSpec(collectionURI)
	if err != nil {
		failf("Failed to read step spec, error: %s", err)
	}

	version := c.String(VersionKey)
	if version == "" {
		latest, err := collection.GetLatestStepVersion(id)
		if err != nil {
			failf("Failed to get step latest version, error: %s", err)
		}
		version = latest
	}

	update := c.Bool(UpdateKey)

	// Check step exist in collection
	step, stepFound, versionFound := collection.GetStep(id, version)
	if !stepFound || !versionFound {
		if update {
			if !stepFound {
				log.Infof("Collection doesn't contain step with id: %s -- Updating StepLib", id)
			} else if !versionFound {
				log.Infof("Collection doesn't contain step (%s) with version: %s -- Updating StepLib", id, version)
			}

			if err := stepman.ReGenerateLibrarySpec(route); err != nil {
				failf("Failed to update collection:%s error:%v", collectionURI, err)
			}

			if _, stepFound, versionFound := collection.GetStep(id, version); !stepFound || !versionFound {
				if !stepFound {
					failf("Even the updated collection doesn't contain step with id: %s", id)
				} else if !versionFound {
					failf("Even the updated collection doesn't contain step (%s) with version: %s", id, version)
				}
			}
		} else {
			if !stepFound {
				failf("Collection doesn't contain step with id: %s -- Updating StepLib", id)
			} else if !versionFound {
				failf("Collection doesn't contain step (%s) with version: %s -- Updating StepLib", id, version)
			}
		}
	}

	if step.Source == nil {
		failf("Missing step's (%s) Source property", id)
	}

	if err := stepman.DownloadStep(collectionURI, collection, id, version, step.Source.Commit, log.NewDefaultLogger(false)); err != nil {
		failf("Failed to download step, error: %s", err)
	}

	return nil
}

package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func activate(c *cli.Context) error {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Fatalf("No step collection specified")
	}

	id := c.String(IDKey)
	if id == "" {
		log.Fatalf("Missing step id")
	}

	path := c.String(PathKey)
	if path == "" {
		log.Fatalf("Missing destination path")
	}

	version := c.String(VersionKey)
	copyYML := c.String(CopyYMLKey)
	update := c.Bool(UpdateKey)

	// Check if step exist in collection
	collection, err := stepman.ReadStepSpec(collectionURI)
	if err != nil {
		log.Fatalf("Failed to read steps spec (spec.json), error: %s", err)
	}

	_, stepFound, versionFound := collection.GetStep(id, version)
	if !stepFound || !versionFound {
		if !update {
			if !stepFound {
				log.Fatalf("Collection doesn't contain step with id: %s", id)
			} else if !versionFound {
				log.Fatalf("Collection doesn't contain step (%s) with version: %s", id, version)
			}
		}

		if !stepFound {
			log.Infof("Collection doesn't contain step with id: %s -- Updating StepLib", id)
		} else if !versionFound {
			log.Infof("Collection doesn't contain step (%s) with version: %s -- Updating StepLib", id, version)
		}

		collection, err = stepman.UpdateLibrary(collectionURI)
		if err != nil {
			log.Fatalf("Failed to update collection (%s), err: %s", collectionURI, err)
		}

		_, stepFound, versionFound := collection.GetStep(id, version)
		if !stepFound {
			if !stepFound {
				log.Fatalf("Collection doesn't contain step with id: %s", id)
			} else if !versionFound {
				log.Fatalf("Collection doesn't contain step (%s) with version: %s", id, version)
			}
		}
	}

	// If version doesn't provided use latest
	if version == "" {
		latest, err := collection.GetLatestStepVersion(id)
		if err != nil {
			log.Fatalf("Failed to get step latest version, error: %s", err)
		}
		version = latest
	}

	// Check step exist in local cache
	step, stepFound, versionFound := collection.GetStep(id, version)
	if !stepFound {
		log.Fatalf("Collection doesn't contain step with id: %s", id)
	} else if !versionFound {
		log.Fatalf("Collection doesn't contain step (%s) with version: %s", id, version)
	}

	if step.Source == nil {
		log.Fatalf("Invalid step, missing Source property")
	}

	route, found := stepman.ReadRoute(collectionURI)
	if !found {
		log.Fatalf("No route found for lib: %s", collectionURI)
	}

	stepCacheDir := stepman.GetStepCacheDirPath(route, id, version)
	if exist, err := pathutil.IsPathExists(stepCacheDir); err != nil {
		log.Fatalf("Failed to check path, error: %s", err)
	} else if !exist {
		if err := stepman.DownloadStep(collectionURI, collection, id, version, step.Source.Commit); err != nil {
			log.Fatalf("Failed to download step, error: %s", err)
		}
	}

	// Copy to specified path
	srcFolder := stepCacheDir
	destFolder := path

	if exist, err := pathutil.IsPathExists(destFolder); err != nil {
		log.Fatalf("Failed to check path, error: %s", err)
	} else if !exist {
		if err := os.MkdirAll(destFolder, 0777); err != nil {
			log.Fatalf("Failed to create path, error: %s", err)
		}
	}

	if err = command.CopyDir(srcFolder+"/", destFolder, true); err != nil {
		log.Fatalf("Failed to copy step, error: %s", err)
	}

	// Copy step.yml to specified path
	if copyYML != "" {
		if exist, err := pathutil.IsPathExists(copyYML); err != nil {
			log.Fatalf("Failed to check path, error: %s", err)
		} else if exist {
			log.Fatalf("Failed to copy step.yml, error: destination path exists")
		}

		stepCollectionDir := stepman.GetStepCollectionDirPath(route, id, version)
		stepYMLSrc := stepCollectionDir + "/step.yml"
		if err = command.CopyFile(stepYMLSrc, copyYML); err != nil {
			log.Fatalf("Failed to copy step.yml, error: %s", err)
		}
	}

	return nil
}

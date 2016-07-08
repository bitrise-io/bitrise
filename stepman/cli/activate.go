package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func activate(c *cli.Context) error {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}

	id := c.String(IDKey)
	if id == "" {
		log.Fatal("[STEPMAN] - Missing step id")
	}

	path := c.String(PathKey)
	if path == "" {
		log.Fatal("[STEPMAN] - Missing destination path")
	}

	version := c.String(VersionKey)
	copyYML := c.String(CopyYMLKey)
	update := c.Bool(UpdateKey)

	// Check if step exist in collection
	collection, err := stepman.ReadStepSpec(collectionURI)
	if err != nil {
		log.Fatalln("[STEPMAN] - Failed to read steps spec (spec.json)")
	}

	_, stepFound := collection.GetStep(id, version)
	if !stepFound {
		if !update {
			if version == "" {
				log.Fatalf("[STEPMAN] - Collection doesn't contain any version of step (id:%s)", id)
			} else {
				log.Fatalf("[STEPMAN] - Collection doesn't contain step (id:%s) (version:%s)", id, version)
			}
		}

		if version == "" {
			log.Infof("[STEPMAN] - Collection doesn't contain any version of step (id:%s) -- Updating StepLib", id)
		} else {
			log.Infof("[STEPMAN] - Collection doesn't contain step (id:%s) (version:%s) -- Updating StepLib", id, version)
		}

		collection, err = updateCollection(collectionURI)
		if err != nil {
			log.Fatalf("Failed to update collection (%s), err: %s", collectionURI, err)
		}

		_, stepFound := collection.GetStep(id, version)
		if !stepFound {
			if version != "" {
				log.Fatalf("[STEPMAN] - Even the updated collection doesn't contain step (id:%s) (version:%s)", id, version)
			} else {
				log.Fatalf("[STEPMAN] - Even the updated collection doesn't contain any version of step (id:%s)", id)
			}
		}
	}

	// If version doesn't provided use latest
	if version == "" {
		log.Debug("[STEPMAN] - Missing step version -- Use latest version")

		latest, err := collection.GetLatestStepVersion(id)
		if err != nil {
			log.Fatal("[STEPMAN] - Failed to get step latest version: ", err)
		}
		log.Debug("[STEPMAN] - Latest version of step: ", latest)
		version = latest
	}

	// Check step exist in local cache
	step, found := collection.GetStep(id, version)
	if !found {
		log.Fatalf("[STEPMAN] - Collection doesn't contain step (id:%s) (version:%s)", id, version)
	}

	route, found := stepman.ReadRoute(collectionURI)
	if !found {
		log.Fatalf("No route found for lib: %s", collectionURI)
	}

	stepCacheDir := stepman.GetStepCacheDirPath(route, id, version)
	if exist, err := pathutil.IsPathExists(stepCacheDir); err != nil {
		log.Fatal("[STEPMAN] - Failed to check path:", err)
	} else if !exist {
		log.Debug("[STEPMAN] - Step does not exist, download it")
		if err := stepman.DownloadStep(collectionURI, collection, id, version, step.Source.Commit); err != nil {
			log.Fatal("[STEPMAN] - Failed to download step:", err)
		}
	}

	// Copy to specified path
	srcFolder := stepCacheDir
	destFolder := path

	if exist, err := pathutil.IsPathExists(destFolder); err != nil {
		log.Fatalln("[STEPMAN] - Failed to check path:", err)
	} else if !exist {
		if err := os.MkdirAll(destFolder, 0777); err != nil {
			log.Fatalln("[STEPMAN] - Failed to create path:", err)
		}
	}

	if err = cmdex.CopyDir(srcFolder+"/", destFolder, true); err != nil {
		log.Fatalln("[STEPMAN] - Failed to copy step:", err)
	}

	// Copy step.yml to specified path
	if copyYML != "" {
		if exist, err := pathutil.IsPathExists(copyYML); err != nil {
			log.Fatalln("[STEPMAN] - Failed to check path:", err)
		} else if exist {
			log.Fatalln("[STEPMAN] - Copy yml destination path exist")
		}

		stepCollectionDir := stepman.GetStepCollectionDirPath(route, id, version)
		stepYMLSrc := stepCollectionDir + "/step.yml"
		if err = cmdex.CopyFile(stepYMLSrc, copyYML); err != nil {
			log.Fatalln("[STEPMAN] - Failed to copy step.yml:", err)
		}
	}

	return nil
}

package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func activate(c *cli.Context) {
	log.Debugln("[STEPMAN] - Activate")

	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}

	id := c.String(IDKey)
	if id == "" {
		log.Fatal("[STEPMAN] - Missing step id")
	}

	copyYML := c.String(CopyYMLKey)
	update := c.Bool(UpdateKey)

	// Get step
	collection, err := stepman.ReadStepSpec(collectionURI)
	if err != nil {
		log.Fatalln("[STEPMAN] - Failed to read steps spec (spec.json)")
	}

	version := c.String(VersionKey)
	if version == "" {
		log.Debug("[STEPMAN] - Missing step version -- Use latest version")

		latest, err := collection.GetLatestStepVersion(id)
		if err != nil {
			log.Fatal("[STEPMAN] - Failed to get step latest version:", err)
		}
		log.Debug("[STEPMAN] - Latest version of step:", latest)
		version = latest
	}

	path := c.String(PathKey)
	if path == "" {
		log.Fatal("[STEPMAN] - Missing destination path")
	}

	route, found := stepman.ReadRoute(collection.SteplibSource)
	if !found {
		log.Fatalf("No route found for lib: %s", collection.SteplibSource)
	}

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

	// Check step exist in local cache
	stepCacheDir := stepman.GetStepCacheDirPath(route, id, version)
	if exist, err := pathutil.IsPathExists(stepCacheDir); err != nil {
		log.Fatal("[STEPMAN] - Failed to check path:", err)
	} else if !exist {
		log.Debug("[STEPMAN] - Step does not exist, download it")
		if err := stepman.DownloadStep(collection, id, version, step.Source.Commit); err != nil {
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

	if err = stepman.RunCopyDir(srcFolder+"/", destFolder, true); err != nil {
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
		if err = stepman.RunCopyFile(stepYMLSrc, copyYML); err != nil {
			log.Fatalln("[STEPMAN] - Failed to copy step.yml:", err)
		}
	}
}

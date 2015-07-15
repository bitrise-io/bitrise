package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func activate(c *cli.Context) {
	log.Info("[STEPMAN] - Activate")

	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		collectionURI = os.Getenv(CollectionPathEnvKey)
	}
	if collectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}

	id := c.String(IDKey)
	if id == "" {
		log.Fatal("[STEPMAN] - Missing step id")
	}

	version := c.String(VersionKey)
	if version == "" {
		log.Fatal("[STEPMAN] - Missing step version")
	}

	path := c.String(PathKey)
	if path == "" {
		log.Fatal("[STEPMAN] - Missing destination path")
	}

	// Get step
	collection, err := stepman.ReadStepSpec(collectionURI)
	if err != nil {
		log.Fatalln("[STEPMAN] - Failed to read steps spec")
	}

	exist, step := collection.GetStep(id, version)
	if !exist {
		log.Fatalf("[STEPMAN] - Failed to activate Step: %s (v%s), does not exist in local cache.", id, version)
	}

	pth := stepman.GetStepPath(collectionURI, step)
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		log.Fatal("[STEPMAN] - Failed to check path:", err)
	} else if !exist {
		log.Info("[STEPMAN] - Step does not exist, download it")
		if err := stepman.DownloadStep(step, collection); err != nil {
			log.Fatal("[STEPMAN] - Failed to download step:", err)
		}
	}

	// Copy to specified path
	srcFolder := pth
	destFolder := path

	if exist, err = pathutil.IsPathExists(destFolder); err != nil {
		log.Fatal("[STEPMAN] - Failed to check path:", err)
	} else if !exist {
		if err := os.MkdirAll(destFolder, 0777); err != nil {
			log.Fatal("[STEPMAN] - Failed to create path:", err)
		}
	}

	if err = stepman.RunCommand("cp", []string{"-rf", srcFolder, destFolder}...); err != nil {
		log.Fatal("[STEPMAN] - Failed to copy step:", err)
	}
}

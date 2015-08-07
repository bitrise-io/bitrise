package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/colorstring"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func getStepIDFromGit(git string) string {
	splits := strings.Split(git, "/")
	lastPart := splits[len(splits)-1]
	splits = strings.Split(lastPart, ".")
	return splits[0]
}

func create(c *cli.Context) {
	// Input validation
	tag := c.String(TagKey)
	if tag == "" {
		log.Fatalln("[STEPMAN] - No step tag specified")
	}

	URL := c.String(URLKey)
	if URL == "" {
		log.Fatalln("[STEPMAN] - No step url specified")
	}

	// Clone step to tmp dir
	tmp := os.TempDir()
	if err := stepman.RemoveDir(tmp); err != nil {
		log.Fatal(err)
	}
	log.Infof("Cloning step from (%s) with tah (%s) to temporary path (%s)", URL, tag, tmp)
	if err := stepman.DoGitCloneWithVersion(URL, tmp, tag); err != nil {
		log.Fatal(err)
	}

	// Update step.yml
	bytes, err := ioutil.ReadFile(tmp + "step.yml")
	if err != nil {
		log.Fatal(err)
	}

	commit, err := stepman.DoGitGetCommit(tmp)
	if err != nil {
		log.Fatal(err)
	}

	var stepModel models.StepModel
	if err := yaml.Unmarshal(bytes, &stepModel); err != nil {
		log.Fatal(err)
	}
	stepModel.Source = models.StepSourceModel{
		Git:    &URL,
		Commit: &commit,
	}

	stepBytes, err := yaml.Marshal(stepModel)
	if err != nil {
		log.Fatal(err)
	}

	// Copy step.yml to steplib
	share, err := stepman.ReadShareSteplibFromFile()
	if err != nil {
		log.Fatal(err)
	}

	route, found := stepman.ReadRoute(share.Collection)
	if !found {
		log.Fatalln("No route found for collectionURI (%s)", share.Collection)
	}

	ID := getStepIDFromGit(URL)
	log.Infof("Step id from URL:", ID)
	share.StepName = ID
	share.StepTag = tag
	stepman.WriteShareSteplibToFile(share)

	stepDirInSteplib := stepman.GetStepCollectionDirPath(route, ID, tag)
	log.Infof("Step dir in collection:", stepDirInSteplib)
	if exist, err := pathutil.IsPathExists(stepDirInSteplib); err != nil {
		log.Fatal(err)
	} else if !exist {
		if err := os.MkdirAll(stepDirInSteplib, 0777); err != nil {
			log.Fatal(err)
		}
	}

	stepYMLPathInSteplib := stepDirInSteplib + "/step.yml"
	log.Info("Temporary step.yml path in collection:", stepYMLPathInSteplib)

	file, err := os.OpenFile(stepYMLPathInSteplib, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error("[STEPMAN] - Failed to close file:", err)
		}
	}()

	if _, err := file.Write([]byte(stepBytes)); err != nil {
		log.Fatal(err)
	}

	if err := stepman.RemoveDir(tmp); err != nil {
		log.Fatal(err)
	}

	// Update spec.json
	steplibBaseDir := stepman.GetCollectionBaseDirPath(route)
	specPth := steplibBaseDir + "/steplib.yml"
	collection, err := stepman.ParseStepCollection(specPth)
	if err != nil {
		if err := stepman.CleanupRoute(route); err != nil {
			log.Errorf("Failed to cleanup route for uri: %s", share.Collection)
		}
		log.Fatal("[STEPMAN] - Failed to read step spec:", err)
	}

	if err := stepman.WriteStepSpecToFile(collection, route); err != nil {
		if err := stepman.CleanupRoute(route); err != nil {
			log.Errorf("Failed to cleanup route for uri: %s", share.Collection)
		}
		log.Fatal("[STEPMAN] - Failed to save step spec:", err)
	}

	fmt.Println()
	log.Info(" * "+colorstring.Green("[OK]")+" Your step added to local steplib.", specPth)
	fmt.Println()
}

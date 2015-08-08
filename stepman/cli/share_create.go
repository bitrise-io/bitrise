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

	gitURI := c.String(GitKey)
	if gitURI == "" {
		log.Fatalln("[STEPMAN] - No step url specified")
	}

	// Clone step to tmp dir
	tmp, err := pathutil.NormalizedOSTempDirPath("")
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Cloning step from (%s) with tag (%s) to temporary path (%s)", gitURI, tag, tmp)
	if err := stepman.DoGitCloneVersion(gitURI, tmp, tag); err != nil {
		log.Fatal(err)
	}

	// Update step.yml
	bytes, err := ioutil.ReadFile(tmp + "/step.yml")
	if err != nil {
		log.Fatal(err)
	}
	var stepModel models.StepModel
	if err := yaml.Unmarshal(bytes, &stepModel); err != nil {
		log.Fatal(err)
	}
	commit, err := stepman.DoGitGetCommit(tmp)
	if err != nil {
		log.Fatal(err)
	}
	stepModel.Source = models.StepSourceModel{
		Git:    gitURI,
		Commit: commit,
	}

	// Copy step.yml to steplib
	share, err := ReadShareSteplibFromFile()
	if err != nil {
		log.Fatal(err)
	}

	route, found := stepman.ReadRoute(share.Collection)
	if !found {
		log.Fatalln("No route found for collectionURI (%s)", share.Collection)
	}

	ID := getStepIDFromGit(gitURI)
	share.StepID = ID
	share.StepTag = tag
	if err := WriteShareSteplibToFile(share); err != nil {
		log.Fatal("[STEPMAN] - Failed to save share steplib to file:", err)
	}

	stepDirInSteplib := stepman.GetStepCollectionDirPath(route, ID, tag)
	log.Info("Step dir in collection:", stepDirInSteplib)
	if exist, err := pathutil.IsPathExists(stepDirInSteplib); err != nil {
		log.Fatal(err)
	} else if !exist {
		if err := os.MkdirAll(stepDirInSteplib, 0777); err != nil {
			log.Fatal(err)
		}
	}

	stepYMLPathInSteplib := stepDirInSteplib + "/step.yml"
	file, err := os.OpenFile(stepYMLPathInSteplib, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error("[STEPMAN] - Failed to close file:", err)
		}
	}()

	stepBytes, err := yaml.Marshal(stepModel)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := file.Write([]byte(stepBytes)); err != nil {
		log.Fatal(err)
	}

	// Update spec.json
	if err := stepman.ReGenerateStepSpec(route.SteplibURI); err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	log.Infof(" * "+colorstring.Green("[OK]")+" Your step (%s) added to local steplib (%s).", share.StepID, share.Collection)
	log.Info("   Next call `stepman share finish` to commit your changes.")
	fmt.Println()
}

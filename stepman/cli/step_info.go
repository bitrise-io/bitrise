package cli

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

// StepInfoModel ...
type StepInfoModel struct {
	StepID        string `json:"step_id,omitempty" yaml:"step_id,omitempty"`
	StepVersion   string `json:"step_version,omitempty" yaml:"step_version,omitempty"`
	LatestVersion string `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`
}

func getStepInfoString(id, version, latest string) (string, error) {
	stepInfo := StepInfoModel{
		StepID:        id,
		StepVersion:   version,
		LatestVersion: latest,
	}

	bytes, err := json.Marshal(stepInfo)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func stepInfo(c *cli.Context) {
	// Input validation
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}

	id := c.String(IDKey)
	if id == "" {
		log.Fatal("[STEPMAN] - Missing step id")
	}

	version := c.String(VersionKey)

	// Check if step exist in collection
	collection, err := stepman.ReadStepSpec(collectionURI)
	if err != nil {
		log.Fatalln("[STEPMAN] - Failed to read steps spec (spec.json)")
	}

	stepFound := false
	if version == "" {
		stepFound = collection.IsStepExist(id)
	} else {
		_, stepFound = collection.GetStep(id, version)
	}

	if !stepFound {
		if version == "" {
			log.Fatalf("[STEPMAN] - Collection doesn't contain any version of step (id:%s)", id)
		} else {
			log.Fatalf("[STEPMAN] - Collection doesn't contain step (id:%s) (version:%s)", id, version)
		}
	}

	latest, err := collection.GetLatestStepVersion(id)
	if err != nil {
		log.Fatalf("[STEPMAN] - Failed to get latest version of step (id:%s)", id)
	}

	if version == "" {
		version = latest
	}

	stepInfoString, err := getStepInfoString(id, version, latest)
	if err != nil {
		log.Fatal("Failed to generate step info, err:", err)
	}

	fmt.Println(stepInfoString)
}

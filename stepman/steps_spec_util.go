package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bitrise-io/go-pathutil"
	"gopkg.in/yaml.v2"
)

const (
	STEPLIB_SOURCE string = "https://github.com/steplib/steplib"
	FORMAT_VERSION string = "0.9.0"
)

func parseStepYml(pth, id, version string) (StepJsonStruct, error) {
	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return StepJsonStruct{}, err
	}

	var stepYml StepYmlStruct
	err = yaml.Unmarshal(bytes, &stepYml)
	if err != nil {
		return StepJsonStruct{}, err
	}

	stepJson := convertToStepJsonStruct(stepYml)
	stepJson.Id = id
	stepJson.VersionTag = version
	stepJson.StepLibSource = STEPLIB_SOURCE

	return stepJson, nil
}

// version is string like x.y.z
// true if version2 is greater then version1
func isVersionGrater(version1, version2 string) bool {
	version1Slice := strings.Split(version1, ".")
	version2Slice := strings.Split(version2, ".")

	for i, num := range version1Slice {
		num1, _ := strconv.ParseInt(num, 0, 64)
		num2, _ := strconv.ParseInt(version2Slice[i], 0, 64)
		if num2 > num1 {
			return true
		}
	}
	return false
}

func addStepToStepGroup(step StepJsonStruct, stepGroup StepGroupJsonStruct) StepGroupJsonStruct {
	var newStepGroup StepGroupJsonStruct
	if len(stepGroup.Versions) > 0 {
		// Step Group already created -> new version of step
		newStepGroup = stepGroup

		if isVersionGrater(newStepGroup.Latest.VersionTag, step.VersionTag) {
			newStepGroup.Latest = step
		}
	} else {
		// Create Step Group
		newStepGroup = StepGroupJsonStruct{}
		newStepGroup.Latest = step
	}

	versions := make([]StepJsonStruct, len(newStepGroup.Versions))
	for idx, step := range newStepGroup.Versions {
		versions[idx] = step
	}
	versions = append(versions, step)
	newStepGroup.Versions = versions

	newStepGroup.Id = step.Id
	return newStepGroup
}

func generateFormattedJSONForStepsSpec() ([]byte, error) {
	collection := StepCollectionJsonStruct{
		FormatVersion:        FORMAT_VERSION,
		GeneratedAtTimeStamp: time.Now().Unix(),
		SteplibSource:        STEPLIB_SOURCE,
	}

	stepHash := StepJsonHash{}

	stepsSpecDir := pathutil.UserHomeDir() + STEPS_DIR
	err := filepath.Walk(stepsSpecDir, func(path string, f os.FileInfo, err error) error {
		truncatedPath := strings.Replace(path, stepsSpecDir, "", -1)
		match, _ := regexp.MatchString("([a-z]+).yml", truncatedPath)
		if match {
			components := strings.Split(truncatedPath, "/")
			if len(components) == 4 {
				name := components[1]
				version := components[2]

				currentStep, _ := parseStepYml(path, name, version)
				currentStepGroup := addStepToStepGroup(currentStep, stepHash[name])

				stepHash[name] = currentStepGroup
			} else {
				fmt.Println("Path:", truncatedPath)
				fmt.Println("Legth:", len(components))
			}
		}

		return err
	})

	collection.Steps = stepHash

	b, err := json.Marshal(collection)
	if err != nil {
		fmt.Println("error:", err)
		return []byte{}, err
	}

	return b, nil
}

func writeStepSpecToFile() error {
	pth := pathutil.UserHomeDir() + STEP_SPEC_DIR

	if exist, _ := pathutil.IsPathExists(pth); exist == false {
		dir, _ := path.Split(pth)
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}
	} else {
		err := os.Remove(pth)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(pth, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonContBytes, err := generateFormattedJSONForStepsSpec()
	if err != nil {
		return err
	}

	_, err = file.Write(jsonContBytes)
	if err != nil {
		return err
	}

	return nil
}

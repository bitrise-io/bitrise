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

func parseStepYml(pth string) (StepJsonStruct, error) {
	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return StepJsonStruct{}, err
	}

	var stepYml StepYmlStruct
	err = yaml.Unmarshal(bytes, &stepYml)
	if err != nil {
		return StepJsonStruct{}, err
	}

	var inputsJson []*InputJsonStruct
	if len(stepYml.Inputs) > 0 {
		inputsJson = make([]*InputJsonStruct, len(stepYml.Inputs))
		for i, inputYml := range stepYml.Inputs {
			inputJson := InputJsonStruct{
				MappedTo:          inputYml.MappedTo,
				Title:             inputYml.Title,
				Description:       inputYml.Description,
				Value:             inputYml.Value,
				ValueOptions:      inputYml.ValueOptions,
				IsRequired:        inputYml.IsRequired,
				IsExpand:          inputYml.IsExpand,
				IsDontChangeValue: inputYml.IsDontChangeValue,
			}
			inputsJson[i] = &inputJson
		}
	}

	var outputsJson []*OutputJsonStruct
	if len(stepYml.Outputs) > 0 {
		outputsJson = make([]*OutputJsonStruct, len(stepYml.Outputs))
		for i, outputYml := range stepYml.Outputs {
			outputJson := OutputJsonStruct{
				MappedTo:    outputYml.MappedTo,
				Title:       outputYml.Title,
				Description: outputYml.Description,
			}
			outputsJson[i] = &outputJson
		}
	}

	stepJson := StepJsonStruct{
		Name:                stepYml.Name,
		Description:         stepYml.Description,
		Website:             stepYml.Website,
		ForkUrl:             stepYml.ForkUrl,
		Source:              stepYml.Source,
		HostOsTags:          stepYml.HostOsTags,
		ProjectTypeTags:     stepYml.ProjectTypeTags,
		TypeTags:            stepYml.TypeTags,
		IsRequiresAdminUser: stepYml.IsRequiresAdminUser,
		Inputs:              inputsJson,
		Outputs:             outputsJson,
	}

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

				currentStep, _ := parseStepYml(path)
				currentStep.Id = name
				currentStep.StepLibSource = STEPLIB_SOURCE
				currentStep.VersionTag = version

				var currentStepGroup StepGroupJsonStruct
				if len(stepHash[name].Versions) > 0 {
					// Step Group already created -> new version of step
					currentStepGroup = stepHash[name]

					versions := make([]StepJsonStruct, len(currentStepGroup.Versions))
					for idx, step := range currentStepGroup.Versions {
						versions[idx] = step
					}
					versions = append(versions, currentStep)
					currentStepGroup.Versions = versions

					// TODO! decide if latest
					if isVersionGrater(currentStepGroup.Latest.VersionTag, currentStep.VersionTag) {
						currentStepGroup.Latest = currentStep
					}
				} else {
					// Create Step Group
					currentStepGroup = StepGroupJsonStruct{}

					versions := make([]StepJsonStruct, 1)
					versions[0] = currentStep
					currentStepGroup.Versions = versions
					currentStepGroup.Latest = currentStep
				}

				currentStepGroup.Id = name

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

	//file, err := os.Create(pth)
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

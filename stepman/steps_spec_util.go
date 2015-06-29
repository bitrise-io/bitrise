package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
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

	inputsJson := make([]InputJsonStruct, len(stepYml.Inputs))
	for _, inputYml := range stepYml.Inputs {
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
		inputsJson = append(inputsJson, inputJson)
	}

	outputsJson := make([]OutputJsonStruct, len(stepYml.Outputs))
	for _, outputYml := range stepYml.Outputs {
		outputJson := OutputJsonStruct{
			MappedTo:    outputYml.MappedTo,
			Title:       outputYml.Title,
			Description: outputYml.Description,
		}
		outputsJson = append(outputsJson, outputJson)
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

func generateFormattedJSONForStepsSpec() ([]byte, error) {
	collection := StepCollectionJsonStruct{
		FormatVersion:        FORMAT_VERSION,
		GeneratedAtTimeStamp: time.Now().String(),
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
				if name != "activate-ssh-key" {
					return nil
				}

				version := components[2]
				//yml := components[3]

				currentStep, _ := parseStepYml(path)
				currentStep.Id = name
				currentStep.StepLibSource = STEPLIB_SOURCE
				currentStep.VersionTag = version

				fmt.Println(fmt.Sprintf("Adding step:%s version:%s", currentStep.Id, currentStep.VersionTag))

				var currentStepGroup StepGroupJsonStruct

				fmt.Println(fmt.Sprintf("Versions count:", len(stepHash[name].Versions)))

				if len(stepHash[name].Versions) > 0 {
					fmt.Println(fmt.Sprintf("Versions count > 0, appending step"))

					// Step Group already created -> new version of step
					currentStepGroup = stepHash[name]

					versions := make([]StepJsonStruct, len(currentStepGroup.Versions))
					for idx, step := range currentStepGroup.Versions {
						versions[idx] = step
					}
					versions = append(versions, currentStep)
					currentStepGroup.Versions = versions

					// TODO! decide if latest
					currentStepGroup.Latest = currentStep
				} else {
					fmt.Println(fmt.Sprintf("Versions count == 0, creating step group"))
					// Create Step Group
					currentStepGroup = StepGroupJsonStruct{}

					versions := make([]StepJsonStruct, 1)
					versions[0] = currentStep
					currentStepGroup.Versions = versions
					currentStepGroup.Latest = currentStep
				}

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

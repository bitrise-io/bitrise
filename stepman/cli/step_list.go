package cli

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

// StepListModel ...
type StepListModel struct {
	StepLib string   `json:"steplib,omitempty" yaml:"steplib,omitempty"`
	Steps   []string `json:"steps,omitempty" yaml:"steps,omitempty"`
}

func printRawStepList(stepList StepListModel, isShort bool) {
	fmt.Println(colorstring.Bluef("Step in StepLib (%s):", stepList.StepLib))
	for _, stepID := range stepList.Steps {
		fmt.Printf("%s\n", stepID)
	}
	fmt.Println()
	fmt.Println()
}

func printJSONStepList(stepList StepListModel, isShort bool) error {
	bytes, err := json.Marshal(stepList)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))
	return nil
}

func listSteps(stepLibURI, format string) error {
	if exist, err := stepman.RootExistForCollection(stepLibURI); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf("Missing routing for stepLib, call 'stepman setup -c %s' before audit", stepLibURI)
	}

	stepLib, err := stepman.ReadStepSpec(stepLibURI)
	if err != nil {
		return err
	}

	// List
	stepList := StepListModel{
		StepLib: stepLibURI,
	}
	for stepID := range stepLib.Steps {
		stepList.Steps = append(stepList.Steps, stepID)
	}

	switch format {
	case OutputFormatRaw:
		printRawStepList(stepList, false)
		break
	case OutputFormatJSON:
		if err := printJSONStepList(stepList, false); err != nil {
			return err
		}
		break
	default:
		return fmt.Errorf("Invalid format: %s", format)
	}
	return nil
}

func list(c *cli.Context) {
	// Input validation
	stepLibURIs := []string{}
	stepLibURI := c.String(CollectionKey)
	if stepLibURI == "" {
		stepLibURIs = stepman.GetAllStepCollectionPath()
	} else {
		stepLibURIs = []string{stepLibURI}
	}

	format := c.String(FormatKey)
	if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		log.Fatalf("[STEPMAN] - Invalid format: %s", format)
	}

	for _, URI := range stepLibURIs {
		if err := listSteps(URI, format); err != nil {
			log.Errorf("Failed to list steps in StepLib (%s), err: %s", URI, err)
		}
	}
}

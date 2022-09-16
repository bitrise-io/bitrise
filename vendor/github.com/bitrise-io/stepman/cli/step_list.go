package cli

import (
	"encoding/json"
	"fmt"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/stringutil"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func printRawStepList(stepLibURI string, stepLib models.StepCollectionModel, isShort bool) {
	log.Println(colorstring.Bluef("Steps in StepLib (%s):", stepLibURI))
	log.Println()
	for stepID, stepGroupInfo := range stepLib.Steps {
		if isShort {
			// print only step IDs
			log.Printf("%s\n", stepID)
			continue
		}

		latestStepVerInfos, isFound := stepGroupInfo.LatestVersion()
		if !isFound {
			log.Errorf("No version found for step: %s", stepID)
			continue
		}
		log.Printf(" * %s\n", pointers.String(latestStepVerInfos.Title))
		log.Printf("   ID: %s\n", stepID)
		log.Printf("   Latest Version: %s\n", stepGroupInfo.LatestVersionNumber)
		summaryText := "no summary specified"
		if latestStepVerInfos.Summary != nil {
			stepSumText := *latestStepVerInfos.Summary
			// stepSumText = strings.Replace(stepSumText, "\n", " ", -1)
			summaryText = stringutil.IndentTextWithMaxLength(stepSumText, "            ", 130, false)
		}
		log.Printf("   Summary: %s\n", summaryText)
		log.Println()
	}
	log.Println()
}

func printJSONStepList(stepLibURI string, stepLib models.StepCollectionModel, isShort bool) error {
	stepList := models.StepListModel{
		StepLib: stepLibURI,
	}
	for stepID := range stepLib.Steps {
		stepList.Steps = append(stepList.Steps, stepID)
	}

	bytes, err := json.Marshal(stepList)
	if err != nil {
		return err
	}

	log.Println(string(bytes))
	return nil
}

func listSteps(stepLibURI, format string) error {
	// Check if setup was done for collection
	if exist, err := stepman.RootExistForLibrary(stepLibURI); err != nil {
		return err
	} else if !exist {
		if err := stepman.SetupLibrary(stepLibURI); err != nil {
			fail("Failed to setup steplib")
		}
	}

	stepLib, err := stepman.ReadStepSpec(stepLibURI)
	if err != nil {
		return err
	}

	switch format {
	case OutputFormatRaw:
		printRawStepList(stepLibURI, stepLib, false)
		break
	case OutputFormatJSON:
		if err := printJSONStepList(stepLibURI, stepLib, false); err != nil {
			return err
		}
		break
	default:
		return fmt.Errorf("Invalid format: %s", format)
	}
	return nil
}

func stepList(c *cli.Context) error {
	// Input validation
	stepLibURIs := []string{}
	stepLibURI := c.String(CollectionKey)
	if stepLibURI == "" {
		stepLibURIs = stepman.GetAllStepCollectionPath()
	} else {
		stepLibURIs = []string{stepLibURI}
	}

	format := c.String(FormatKey)
	if format == "" {
		format = OutputFormatRaw
	} else if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		failf("Invalid format: %s", format)
	}

	for _, URI := range stepLibURIs {
		if err := listSteps(URI, format); err != nil {
			log.Errorf("Failed to list steps in StepLib (%s), err: %s", URI, err)
		}
	}

	return nil
}

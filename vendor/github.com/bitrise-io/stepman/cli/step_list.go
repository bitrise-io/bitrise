package cli

import (
	"encoding/json"
	"fmt"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/stringutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func ListCachedSteps(log stepman.Logger) error {
	// Check if setup was done for collection
	if exist, err := stepman.RootExistForLibrary(bitriseStepLibURL); err != nil {
		return err
	} else if !exist {
		if err := stepman.SetupLibrary(bitriseStepLibURL, log); err != nil {
			failf("Failed to setup steplib")
		}
	}

	listSteplibURIs(log, []string{bitriseStepLibURL}, bitriseMaintainer, OutputFormatRaw, true)

	return nil
}

func printRawStepList(log stepman.Logger, stepLibURI string, maintaner string, stepLib models.StepCollectionModel, isShort bool) {
	fmt.Println(colorstring.Bluef("Steps in StepLib (%s):", stepLibURI))
	fmt.Println()
	for stepID, stepGroupInfo := range stepLib.Steps {
		if maintaner != "" && stepGroupInfo.Info.Maintainer != maintaner {
			continue
		}

		if isShort { // print only step IDs and version
			versions := ""
			cachedVersions := listCachedStepVersion(log, stepLib, stepLibURI, stepID)
			for _, version := range cachedVersions {
				versions += fmt.Sprintf("%s, ", version)
			}

			id := fmt.Sprintf("%s (%s)", stepID, stepGroupInfo.Info.Maintainer)
			buf := make([]rune, 50)
			for i := range buf {
				if i < len(id) {
					buf[i] = rune(id[i])
				} else {
					buf[i] = ' '
				}
			}

			fmt.Printf("%s cached versions:  %s\n", string(buf), versions)

			continue
		}

		latestStepVerInfos, isFound := stepGroupInfo.LatestVersion()
		if !isFound {
			log.Errorf("No version found for step: %s", stepID)
			continue
		}

		fmt.Printf(" * %s\n", pointers.String(latestStepVerInfos.Title))
		fmt.Printf("   ID: %s\n", stepID)
		fmt.Printf("   Latest Version: %s\n", stepGroupInfo.LatestVersionNumber)
		summaryText := "no summary specified"
		if latestStepVerInfos.Summary != nil {
			stepSumText := *latestStepVerInfos.Summary
			// stepSumText = strings.Replace(stepSumText, "\n", " ", -1)
			summaryText = stringutil.IndentTextWithMaxLength(stepSumText, "            ", 130, false)
		}
		fmt.Printf("   Summary: %s\n", summaryText)
		fmt.Println()
	}
	fmt.Println()
}

func printJSONStepList(stepLibURI string, stepLib models.StepCollectionModel, _ bool) error {
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

	fmt.Println(string(bytes))
	return nil
}

func listSteps(stepLibURI, maintaner string, format string, log stepman.Logger, isShort bool) error {
	// Check if setup was done for collection
	if exist, err := stepman.RootExistForLibrary(stepLibURI); err != nil {
		return err
	} else if !exist {
		if err := stepman.SetupLibrary(stepLibURI, log); err != nil {
			failf("Failed to setup steplib")
		}
	}

	stepLib, err := stepman.ReadStepSpec(stepLibURI)
	if err != nil {
		return err
	}

	switch format {
	case OutputFormatRaw:
		printRawStepList(log, stepLibURI, maintaner, stepLib, isShort)
	case OutputFormatJSON:
		if err := printJSONStepList(stepLibURI, stepLib, false); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Invalid format: %s", format)
	}
	return nil
}

func listSteplibURIs(log stepman.Logger, stepLibURIs []string, maintaner string, format string, isShort bool) {
	for _, URI := range stepLibURIs {
		if err := listSteps(URI, maintaner, format, log, isShort); err != nil {
			log.Errorf("Failed to list steps in StepLib (%s): %s", URI, err)
		}
	}
}

func stepList(c *cli.Context) error {
	logger := log.NewDefaultLogger(false)
	// Input validation
	var stepLibURIs []string
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

	listSteplibURIs(logger, stepLibURIs, "", format, false)

	return nil
}

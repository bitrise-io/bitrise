package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/stringutil"
	"github.com/bitrise-io/stepman/activator/steplib"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func ListCachedSteps(steplibURI, maintaner string, log stepman.Logger) error {
	// Check if setup was done for collection
	if exist, err := stepman.RootExistForLibrary(steplibURI); err != nil {
		return err
	} else if !exist {
		if err := stepman.SetupLibrary(steplibURI, log); err != nil {
			failf("Failed to setup steplib")
		}
	}

	listSteplibURIs(log, []string{steplibURI}, maintaner, OutputFormatRaw, true)

	return nil
}

func printInMaxNChars(text string, maxChars int) string {
	buf := make([]rune, maxChars)
	for i := range buf {
		if i < len(text) {
			buf[i] = rune(text[i])
		} else {
			buf[i] = ' '
		}
	}

	return string(buf)
}

func printRawStepList(log stepman.Logger, stepLibURI string, maintaner string, stepLib models.StepCollectionModel, isShort bool) {
	fmt.Println(colorstring.Bluef("Steps in StepLib (%s):", stepLibURI))
	fmt.Println()

	stepIDs := []string{}
	for stepID := range stepLib.Steps {
		stepIDs = append(stepIDs, stepID)
	}
	sort.Strings(stepIDs)

	skipped := []string{}
	for _, stepID := range stepIDs {
		stepGroupInfo := stepLib.Steps[stepID]
		if maintaner != "" && stepGroupInfo.Info.Maintainer != maintaner {
			skipped = append(skipped, stepID)
			continue
		}

		if isShort { // print only step IDs and cached versions
			cachedVersions := steplib.ListCachedStepVersions(log, stepLib, stepLibURI, stepID)
			id := fmt.Sprintf("%s (%s)", stepID, stepGroupInfo.Info.Maintainer)
			fmt.Printf("%s cached versions:  %s\n", printInMaxNChars(id, 55), strings.Join(cachedVersions, ", "))

			continue
		}

		latestStepVerInfos, isFound := stepGroupInfo.LatestVersion()
		if !isFound {
			log.Errorf("no version found for step: %s", stepID)
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

	fmt.Printf("\nSkipped steps (maintainer filter): %s\n", strings.Join(skipped, ", "))
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
		return fmt.Errorf("invalid format: %s", format)
	}
	return nil
}

func listSteplibURIs(log stepman.Logger, stepLibURIs []string, maintaner string, format string, isShort bool) {
	for _, URI := range stepLibURIs {
		if err := listSteps(URI, maintaner, format, log, isShort); err != nil {
			log.Errorf("failed to list steps in StepLib (%s): %s", URI, err)
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
	} else if format != OutputFormatRaw && format != OutputFormatJSON {
		failf("Invalid format: %s", format)
	}

	listSteplibURIs(logger, stepLibURIs, "", format, false)

	return nil
}

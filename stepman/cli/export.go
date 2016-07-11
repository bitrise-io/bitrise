package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

const (
	fullSpecBasename    = "spec"
	latestSpecBasename  = "latest-" + fullSpecBasename
	minimalSpecBasename = "minimal-" + fullSpecBasename

	jsonExt = ".json"
)

// ExportType ...
type ExportType int8

const (
	exportTypeFull ExportType = iota
	exportTypeLatest
	exportTypeMinimal
)

func parseExportType(exportTypeStr string) (ExportType, error) {
	switch exportTypeStr {
	case "full":
		return exportTypeFull, nil
	case "latest":
		return exportTypeLatest, nil
	case "minimal":
		return exportTypeMinimal, nil
	}

	var exportType ExportType
	return exportType, fmt.Errorf("Invalid export type (%s), available: [full, latest, minimal]", exportTypeStr)
}

func convertToMinimalSpec(stepLib models.StepCollectionModel) models.StepCollectionModel {
	steps := stepLib.Steps

	minimalSteps := models.StepHash{}
	for stepID := range steps {
		minimalSteps[stepID] = models.StepGroupModel{}
	}

	stepLib.Steps = minimalSteps
	return stepLib
}

func convertToLatestSpec(stepLib models.StepCollectionModel) models.StepCollectionModel {
	steps := stepLib.Steps

	latestSteps := models.StepHash{}
	for stepID, stepGroup := range steps {
		groupInfo := stepGroup.Info
		versions := stepGroup.Versions
		latestVersionStr := stepGroup.LatestVersionNumber
		latestStep := versions[latestVersionStr]

		latestSteps[stepID] = models.StepGroupModel{
			Versions: map[string]models.StepModel{
				latestVersionStr: latestStep,
			},
			Info: groupInfo,
		}
	}

	stepLib.Steps = latestSteps
	return stepLib
}

func export(c *cli.Context) error {
	// Input validation
	steplibURI := c.String("steplib")
	outputDirPth := c.String("output")
	exportTypeStr := c.String("export-type")

	if steplibURI == "" {
		return fmt.Errorf("Missing required input: steplib")
	}

	if outputDirPth == "" {
		currentDirPth, err := pathutil.CurrentWorkingDirectoryAbsolutePath()
		if err != nil {
			return fmt.Errorf("Failed to get current dir, error: %s", err)
		}
		outputDirPth = currentDirPth
	}

	exportType := exportTypeFull
	if exportTypeStr != "" {
		var err error
		exportType, err = parseExportType(exportTypeStr)
		if err != nil {
			return err
		}
	}

	log.Infof("Exporting StepLib (%s) spec, export-type: %s, output dir: %s", steplibURI, exportTypeStr, outputDirPth)

	// Setup StepLib
	if exist, err := stepman.RootExistForCollection(steplibURI); err != nil {
		return fmt.Errorf("Failed to check if setup was done for StepLib, error: %s", err)
	} else if !exist {
		log.Infof("StepLib does not exist, setup...")
		if err := setupSteplib(steplibURI, false); err != nil {
			return fmt.Errorf("Failed to setup StepLib, error: %s", err)
		}
	}

	// Prepare spec
	stepLibSpec, err := stepman.ReadStepSpec(steplibURI)
	if err != nil {
		log.Fatalln("Failed to read StepLib spec, error: %s", err)
	}

	outputBasename := fullSpecBasename
	switch exportType {
	case exportTypeMinimal:
		outputBasename = minimalSpecBasename
		stepLibSpec = convertToMinimalSpec(stepLibSpec)
	case exportTypeLatest:
		outputBasename = latestSpecBasename
		stepLibSpec = convertToLatestSpec(stepLibSpec)
	}

	stepLibSpecBytes, err := json.Marshal(stepLibSpec)
	if err != nil {
		return fmt.Errorf("Failed to marshal StepLib, error: %s", err)
	}

	// Export spec
	exist, err := pathutil.IsDirExists(outputDirPth)
	if err != nil {
		return fmt.Errorf("Failed to check if dir (%s) exist, error: %s", outputDirPth, err)
	}
	if !exist {
		if err := os.MkdirAll(outputDirPth, 0777); err != nil {
			return fmt.Errorf("Failed to create dir (%s), error: %s", outputDirPth, err)
		}
	}

	outputPth := filepath.Join(outputDirPth, outputBasename+jsonExt)
	if err := fileutil.WriteBytesToFile(outputPth, stepLibSpecBytes); err != nil {
		return fmt.Errorf("Failed to write StepLib spec to: %s, error: %s", outputPth, err)
	}

	log.Infof("StepLib spec exported to: %s", outputPth)

	return nil
}

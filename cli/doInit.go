package cli

import (
	"encoding/json"
	"errors"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/models"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/codegangsta/cli"
)

func doInit(c *cli.Context) {
	log.Info("[BITRISE_CLI] - Init -- Work-in-progress!")

	projectName, err := goinp.AskForString("Enter the PROJECT_NAME")
	if err != nil {
		log.Fatalln(err)
	}

	workflowModel := models.WorkflowModel{
		FormatVersion: "0.9.0",
		Environments: []models.InputModel{
			models.InputModel{MappedTo: "PROJECT_NAME", Value: projectName},
		},
		Steps: []models.StepModel{},
	}

	if err := SaveToFile("./bitrise.json", workflowModel); err != nil {
		log.Fatalln("Failed to init:", err)
	}
	os.Exit(1)
}

// SaveToFile ...
func SaveToFile(pth string, workflowModel models.WorkflowModel) error {
	if pth == "" {
		return errors.New("No path provided")
	}

	file, err := os.Create(pth)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalln("[BITRISE] - Failed to close file:", err)
		}
	}()

	jsonContBytes, err := generateNonFormattedJSON(workflowModel)
	if err != nil {
		return err
	}

	if _, err := file.Write(jsonContBytes); err != nil {
		return err
	}
	log.Println()
	log.Infoln("=> Init success!")
	log.Infoln("File created at path:", pth)
	log.Infoln("With the content:")
	log.Infoln(string(jsonContBytes))

	return nil
}

func generateNonFormattedJSON(v interface{}) ([]byte, error) {
	jsonContBytes, err := json.Marshal(v)
	if err != nil {
		return []byte{}, err
	}
	return jsonContBytes, nil
}

package cli

import (
	"encoding/json"
	"errors"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/models"
	"github.com/codegangsta/cli"
)

func doInit(c *cli.Context) {
	log.Info("[BITRISE_CLI] - Init -- Work-in-progress!")
	workflowModel := models.WorkflowModel{
		FormatVersion: "0.9.0",
		Environments:  []string{},
		Steps:         []models.StepModel{},
	}

	if err := SaveToFile("./bitrise.json", workflowModel); err != nil {
		log.Fatalln("Failed to init:", err)
	}
	os.Exit(1)
}

func SaveToFile(pth string, workflowModel models.WorkflowModel) error {
	if pth == "" {
		return errors.New("No path provided")
	}

	if file, err := os.Create(pth); err != nil {
		return err
	} else {
		defer func() {
			if err := file.Close(); err != nil {
				log.Fatalln("[BITRISE] - Failed to close file:", err)
			}
		}()

		if jsonContBytes, err := GenerateNonFormattedJSON(workflowModel); err != nil {
			return err
		} else if _, err := file.Write(jsonContBytes); err != nil {
			return err
		}
		return nil
	}
}

func GenerateNonFormattedJSON(v interface{}) ([]byte, error) {
	jsonContBytes, err := json.Marshal(v)
	if err != nil {
		return []byte{}, err
	}
	return jsonContBytes, nil
}

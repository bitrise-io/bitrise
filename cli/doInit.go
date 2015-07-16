package cli

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	models "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/codegangsta/cli"
)

func doInit(c *cli.Context) {
	log.Info("[BITRISE_CLI] - Init -- Work-in-progress!")
	bitriseConfigFileRelPath := "./bitrise.yml"

	if exists, err := pathutil.IsPathExists(bitriseConfigFileRelPath); err != nil {
		log.Fatalln("Error:", err)
	} else if exists {
		ask := fmt.Sprintf("A config file already exists at %s - do you want to overwrite it?", bitriseConfigFileRelPath)
		if val, err := goinp.AskForBool(ask); err != nil {
			log.Fatalln("Error:", err)
		} else if !val {
			log.Infoln("Init canceled, existing file won't be overwritten.")
			os.Exit(0)
		}
	}

	projectSettingsEnvs := []models.InputModel{}
	if val, err := goinp.AskForString("What's the BITRISE_PROJECT_TITLE?"); err != nil {
		log.Fatalln(err)
	} else {
		projectTitleEnv := models.InputModel{MappedTo: "BITRISE_PROJECT_TITLE", Value: val}
		*projectTitleEnv.IsExpand = false
		projectSettingsEnvs = append(projectSettingsEnvs, projectTitleEnv)
	}
	if val, err := goinp.AskForString("What's your primary development branch's name?"); err != nil {
		log.Fatalln(err)
	} else {
		devBranchEnv := models.InputModel{MappedTo: "BITRISE_DEV_BRANCH", Value: val}
		*devBranchEnv.IsExpand = false
		projectSettingsEnvs = append(projectSettingsEnvs, devBranchEnv)
	}

	// TODO:
	//  generate a couple of base steps
	//  * timestamp gen
	//  * bash script

	bitriseConf := models.BitriseConfigModel{
		FormatVersion: "1.0.0", // TODO: move this into a project config file!
		App: models.AppModel{
			Environments: projectSettingsEnvs,
		},
		Workflows: map[string]models.WorkflowModel{
			"primary": models.WorkflowModel{},
		},
	}

	if err := SaveToFile(bitriseConfigFileRelPath, bitriseConf); err != nil {
		log.Fatalln("Failed to init:", err)
	}
	os.Exit(1)
}

// SaveToFile ...
func SaveToFile(pth string, bitriseConf models.BitriseConfigModel) error {
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

	contBytes, err := generateYAML(bitriseConf)
	if err != nil {
		return err
	}

	if _, err := file.Write(contBytes); err != nil {
		return err
	}
	log.Println()
	log.Infoln("=> Init success!")
	log.Infoln("File created at path:", pth)
	log.Infoln("With the content:")
	log.Infoln(string(contBytes))

	return nil
}

func generateYAML(v interface{}) ([]byte, error) {
	bytes, err := yaml.Marshal(v)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

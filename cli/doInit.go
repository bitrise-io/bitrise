package cli

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/colorstring"
	models "github.com/bitrise-io/bitrise/models/models_1_0_0"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/bitrise-io/goinp/goinp"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/codegangsta/cli"
)

const (
	defaultStepLibSource = "https://github.com/bitrise-io/bitrise-steplib"
	//
	defaultSecretsContent = `envs:
- MY_HOME: $HOME
- MY_SECRET_PASSWORD: XyZ
  opts:
    # You can include some options as well if you
    #  want to change how the value is passed to a command.
    is_expand: no
    # For example you can use is_expand: no
    #  if you want to make it sure that
    #  the value is preserved as-it-is, and won't be
    #  expanded before use.
    # For example if your password contains the dollar sign ($)
    #  it would (by default) be expanded as an environment variable,
    #  just like $HOME would be expanded/replaced with your home
    #  directory path.
    # You can prevent this with is_expand: no`
)

func doInit(c *cli.Context) {
	PrintBitriseHeaderASCIIArt()

	bitriseConfigFileRelPath := "./" + DefaultBitriseConfigFileName
	bitriseSecretsFileRelPath := "./" + DefaultSecretsFileName

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

	defaultExpand := true
	projectSettingsEnvs := []envmanModels.EnvironmentItemModel{}
	if val, err := goinp.AskForString("What's the BITRISE_PROJECT_TITLE?"); err != nil {
		log.Fatalln(err)
	} else {
		projectTitleEnv := envmanModels.EnvironmentItemModel{
			"BITRISE_PROJECT_TITLE": val,
			"opts": envmanModels.EnvironmentItemOptionsModel{
				IsExpand: &defaultExpand,
			},
		}
		projectSettingsEnvs = append(projectSettingsEnvs, projectTitleEnv)
	}
	if val, err := goinp.AskForString("What's your primary development branch's name?"); err != nil {
		log.Fatalln(err)
	} else {
		devBranchEnv := envmanModels.EnvironmentItemModel{
			"BITRISE_DEV_BRANCH": val,
			"opts": envmanModels.EnvironmentItemOptionsModel{
				IsExpand: &defaultExpand,
			},
		}
		projectSettingsEnvs = append(projectSettingsEnvs, devBranchEnv)
	}

	// TODO:
	//  generate a couple of base steps
	//  * timestamp gen
	//  * bash script - hello world

	scriptStepTitle := "Hello Bitrise!"
	scriptStepContent := `#!/bin/bash
echo "Welcome to Bitrise!"`
	bitriseConf := models.BitriseDataModel{
		FormatVersion:        c.App.Version,
		DefaultStepLibSource: defaultStepLibSource,
		App: models.AppModel{
			Environments: projectSettingsEnvs,
		},
		Workflows: map[string]models.WorkflowModel{
			"primary": models.WorkflowModel{
				Steps: []models.StepListItemModel{
					models.StepListItemModel{
						"script": stepmanModels.StepModel{
							Title: &scriptStepTitle,
							Inputs: []envmanModels.EnvironmentItemModel{
								envmanModels.EnvironmentItemModel{
									"content": scriptStepContent,
								},
							},
						},
					},
				},
			},
		},
	}

	if err := bitrise.SaveConfigToFile(bitriseConfigFileRelPath, bitriseConf); err != nil {
		log.Fatalln("Failed to init the bitrise config file:", err)
	} else {
		fmt.Println()
		fmt.Println("# NOTES about the " + DefaultBitriseConfigFileName + " config file:")
		fmt.Println()
		fmt.Println("We initialized a " + DefaultBitriseConfigFileName + " config file for you.")
		fmt.Println("If you're in this folder you can use this config file")
		fmt.Println(" with bitrise automatically, you don't have to")
		fmt.Println(" specify it's path.")
		fmt.Println()
	}

	if initialized, err := saveSecretsToFile(bitriseSecretsFileRelPath, defaultSecretsContent); err != nil {
		log.Fatalln("Failed to init the secrets file:", err)
	} else if initialized {
		fmt.Println()
		fmt.Println("# NOTES about the " + DefaultSecretsFileName + " secrets file:")
		fmt.Println()
		fmt.Println("We also created a " + DefaultSecretsFileName + " file")
		fmt.Println(" in this directory, to keep your passwords, absolute path configurations")
		fmt.Println(" and other secrets separate from your")
		fmt.Println(" main configuration file.")
		fmt.Println("This way you can safely commit and share your configuration file")
		fmt.Println(" and ignore this secrets file, so nobody else will")
		fmt.Println(" know about your secrets.")
		fmt.Println(colorstring.Yellow("You should NEVER commit this secrets file into your repository!!"))
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("Hurray, you're good to go!")
	fmt.Println("You can simply run:")
	fmt.Println("-> bitrise run primary")
	fmt.Println("to test the sample configuration (which contains")
	fmt.Println("an example workflow called 'primary').")
	fmt.Println()
	fmt.Println("Once you tested this sample setup you can")
	fmt.Println(" open the " + DefaultBitriseConfigFileName + " config file,")
	fmt.Println(" modify it and then run a workflow with:")
	fmt.Println("-> bitrise run YOUR-WORKFLOW-NAME")
}

func saveSecretsToFile(pth, secretsStr string) (bool, error) {
	if exists, err := pathutil.IsPathExists(pth); err != nil {
		return false, err
	} else if exists {
		ask := fmt.Sprintf("A secrets file already exists at %s - do you want to overwrite it?", pth)
		if val, err := goinp.AskForBool(ask); err != nil {
			return false, err
		} else if !val {
			log.Infoln("Init canceled, existing file (" + pth + ") won't be overwritten.")
			return false, nil
		}
	}

	if err := bitrise.WriteStringToFile(pth, secretsStr); err != nil {
		return false, err
	}
	return true, nil
}

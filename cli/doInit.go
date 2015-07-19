package cli

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/bitrise"
	models "github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/codegangsta/cli"
)

var defaultSecretsContent = `envs:
- MY_HOME: $HOME
- MY_SECRET_PASSWORD: XyZ
  is_expand: no
  # Hint: You can use is_expand: no
  #  if you want to make it sure that
  #  the value is preserved as-it-is, and won't be
  #  expanded before use.
  # For example if your password contains the dollar sign ($)
  #  it would (by default) be expanded as an environment variable.
  # You can prevent this with is_expand: no`

func doInit(c *cli.Context) {
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
	projectSettingsEnvs := []models.EnvironmentItemModel{}
	if val, err := goinp.AskForString("What's the BITRISE_PROJECT_TITLE?"); err != nil {
		log.Fatalln(err)
	} else {
		projectTitleEnv := models.EnvironmentItemModel{
			EnvKey:   "BITRISE_PROJECT_TITLE",
			Value:    val,
			IsExpand: defaultExpand,
		}
		projectSettingsEnvs = append(projectSettingsEnvs, projectTitleEnv)
	}
	if val, err := goinp.AskForString("What's your primary development branch's name?"); err != nil {
		log.Fatalln(err)
	} else {
		devBranchEnv := models.EnvironmentItemModel{
			EnvKey:   "BITRISE_DEV_BRANCH",
			Value:    val,
			IsExpand: defaultExpand,
		}
		projectSettingsEnvs = append(projectSettingsEnvs, devBranchEnv)
	}

	// TODO:
	//  generate a couple of base steps
	//  * timestamp gen
	//  * bash script - hello world

	bitriseConf := models.BitriseDataModel{
		FormatVersion: "1.0.0", // TODO: move this into a project config file!
		App: models.AppModel{
			Environments: projectSettingsEnvs,
		},
		Workflows: map[string]models.WorkflowModel{
			"primary": models.WorkflowModel{},
		},
	}

	if err := saveConfigToFile(bitriseConfigFileRelPath, bitriseConf); err != nil {
		log.Fatalln("Failed to init the bitrise config file:", err)
	} else {
		fmt.Println()
		fmt.Println("# NOTES about the " + DefaultBitriseConfigFileName + " config file:")
		fmt.Println()
		fmt.Println("We initialized a " + DefaultBitriseConfigFileName + " config file for you.")
		fmt.Println("If you're in this folder you can use this config file")
		fmt.Println(" with bitrise-cli automatically, you don't have to")
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
		fmt.Println("You should NEVER commit this secrets file into your repository!!")
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("Hurray, you're good to go!")
	fmt.Println("You can simply run:")
	fmt.Println("-> bitrise-cli run primary")
	fmt.Println("to test the sample configuration (which contains")
	fmt.Println("an example workflow called 'primary').")
	fmt.Println()
	fmt.Println("Once you tested this sample setup you can")
	fmt.Println(" open the " + DefaultBitriseConfigFileName + " config file,")
	fmt.Println(" modify it and then run a workflow with:")
	fmt.Println("-> bitrise-cli run YOUR-WORKFLOW-NAME")
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

func saveConfigToFile(pth string, bitriseConf models.BitriseDataModel) error {
	confModel := bitriseConf.ToBitriseConfigSerializeModel()
	contBytes, err := generateYAML(confModel)
	if err != nil {
		return err
	}
	if err := bitrise.WriteBytesToFile(pth, contBytes); err != nil {
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

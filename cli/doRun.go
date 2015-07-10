package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/bitrise"
	models "github.com/bitrise-io/bitrise-cli/models/models_0_9_0"
	"github.com/bitrise-io/go-pathutil"
	"github.com/codegangsta/cli"
)

func exportWorkflowEnvironments(workflow models.WorkflowModel) error {
	log.Info("[BITRISE_CLI] - Exporting workflow environments")

	for _, env := range workflow.Environments {
		if env.Value != "" {
			if err := bitrise.RunEnvmanAdd(env.MappedTo, env.Value); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return err
			}
		}
	}
	return nil
}

func activateAndRunSteps(workflow models.WorkflowModel) error {
	log.Info("[BITRISE_CLI] - Activating and running steps")

	for _, step := range workflow.Steps {
		stepDir := "./steps/" + step.ID + "/" + step.VersionTag + "/"

		if err := bitrise.RunStepmanSetup(step.StepLibSource); err != nil {
			log.Error("Failed to setup stepman:", err)
		}

		if err := bitrise.RunStepmanActivate(step.StepLibSource, step.ID, step.VersionTag, stepDir); err != nil {
			log.Errorln("[BITRISE_CLI] - Failed to run stepman activate")
			return err
		}

		log.Infof("[BITRISE_CLI] - Step activated: %s (%s)", step.ID, step.VersionTag)

		if err := runStep(step); err != nil {
			log.Errorln("[BITRISE_CLI] - Failed to run step")
			return err
		}
	}
	return nil
}

func runStep(step models.StepModel) error {
	log.Infof("[BITRISE_CLI] - Running step: %s (%s)", step.ID, step.VersionTag)

	// Add step envs
	for _, input := range step.Inputs {
		if input.Value != "" {
			if err := bitrise.RunEnvmanAdd(input.MappedTo, input.Value); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return err
			}
		}
	}

	stepDir := "./steps/" + step.ID + "/" + step.VersionTag + "/"
	//stepCmd := fmt.Sprintf("%sstep.sh", stepDir)
	stepCmd := "step.sh"
	cmd := []string{"bash", stepCmd}

	if err := bitrise.RunEnvmanRunInDir(stepDir, cmd); err != nil {
		log.Errorln("[BITRISE_CLI] - Failed to run envman run")
		return err
	}

	log.Infof("[BITRISE_CLI] - Step executed: %s (%s)", step.ID, step.VersionTag)
	return nil
}

func doRun(c *cli.Context) {
	log.Info("[BITRISE_CLI] - Run")

	// Input validation
	workflowJSONPath := c.String(PathKey)
	if workflowJSONPath == "" {
		log.Infoln("[BITRISE_CLI] - Workflow path not defined, searching for bitrise.json in current folder...")

		if exist, err := pathutil.IsPathExists("./bitrise.json"); err != nil {
			log.Fatalln("[BITRISE_CLI] - Failed to check path:", err)
		} else if !exist {
			log.Fatalln("[BITRISE_CLI] - No workflow json found")
		}
		workflowJSONPath = "./bitrise.json"
	}

	// Envman setup
	if err := os.Setenv(bitrise.EnvstorePathEnvKey, bitrise.EnvstorePath); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to add env:", err)
	}

	if err := os.Setenv(bitrise.FormattedOutputPathEnvKey, bitrise.FormattedOutputPath); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to add env:", err)
	}

	if err := bitrise.RunEnvmanInit(); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to run envman init")
	}

	// Run work flow
	if workflow, err := bitrise.ReadWorkflowJSON(workflowJSONPath); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to read work flow:", err)
	} else {
		if err := exportWorkflowEnvironments(workflow); err != nil {
			log.Fatalln("[BITRISE_CLI] - Failed to export environments:", err)
		}

		if err := activateAndRunSteps(workflow); err != nil {
			log.Fatalln("[BITRISE_CLI] - Failed to activate steps:", err)
		}
	}
}

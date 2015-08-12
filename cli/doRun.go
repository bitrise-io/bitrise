package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	models "github.com/bitrise-io/bitrise/models/models_1_0_0"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/versions"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/codegangsta/cli"
)

const (
	// DefaultBitriseConfigFileName ...
	DefaultBitriseConfigFileName = "bitrise.yml"
	// DefaultSecretsFileName ...
	DefaultSecretsFileName = ".bitrise.secrets.yml"

	depManagerBrew = "brew"
)

func runStep(step stepmanModels.StepModel, stepIDData models.StepIDData, stepDir string) (int, error) {
	log.Debugf("[BITRISE_CLI] - Try running step: %s (%s)", stepIDData.IDorURI, stepIDData.Version)

	// Check dependencies
	for _, dep := range step.Dependencies {
		switch dep.Manager {
		case depManagerBrew:
			err := bitrise.InstallWithBrewIfNeeded(dep.Name)
			if err != nil {
				return 1, err
			}
			break
		default:
			return 1, errors.New("Not supported dependency (" + dep.Manager + ") (" + dep.Name + ")")
		}
	}

	// Add step envs
	for _, input := range step.Inputs {
		key, value, err := input.GetKeyValuePair()
		if err != nil {
			return 1, err
		}

		opts, err := input.GetOptions()
		if err != nil {
			return 1, err
		}

		if value != "" {
			log.Debugf("Input: %#v\n", input)
			if err := bitrise.RunEnvmanAdd(key, value, *opts.IsExpand); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return 1, err
			}
		}
	}

	stepCmd := stepDir + "/" + "step.sh"
	cmd := []string{"bash", stepCmd}
	log.Debug("OUTPUT:")
	if exit, err := bitrise.RunEnvmanRunInDir(bitrise.CurrentDir, cmd, "panic"); err != nil {
		return exit, err
	}

	log.Debugf("[BITRISE_CLI] - Step executed: %s (%s)", stepIDData.IDorURI, stepIDData.Version)
	return 0, nil
}

func activateAndRunSteps(workflow models.WorkflowModel, defaultStepLibSource string, buildRunResults models.BuildRunResultsModel) models.BuildRunResultsModel {
	log.Debugln("[BITRISE_CLI] - Activating and running steps")

	var stepStartTime time.Time

	registerStepRunResults := func(step stepmanModels.StepModel, resultCode int, exitCode int, err error) {
		stepResults := models.StepRunResultsModel{
			StepName: *step.Title,
			Error:    err,
			ExitCode: exitCode,
		}

		switch resultCode {
		case bitrise.StepRunResultCodeSuccess:
			buildRunResults.SuccessSteps = append(buildRunResults.SuccessSteps, stepResults)
			break
		case bitrise.StepRunResultCodeFailed:
			log.Errorf("Step (%s) failed, error: (%v)", *step.Title, err)
			buildRunResults.FailedSteps = append(buildRunResults.FailedSteps, stepResults)
			break
		case bitrise.StepRunResultCodeFailedSkippable:
			log.Warnf("Step (%s) failed, but was marked as skippable, error: (%v)", *step.Title, err)
			buildRunResults.FailedSkippableSteps = append(buildRunResults.FailedSkippableSteps, stepResults)
			break
		case bitrise.StepRunResultCodeSkipped:
			log.Warnf("A previous step failed, and this step (%s) was not marked as IsAlwaysRun, skipped", *step.Title)
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		case bitrise.StepRunResultCodeSkippedWithRunIf:
			log.Warn("The step's (" + *step.Title + ") Run-If expression evaluated to false - skipping")
			log.Info("The Run-If expression was: ", colorstring.Blue(*step.RunIf))
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		default:
			log.Error("Unkown result code")
			return
		}

		bitrise.PrintStepSummary(*step.Title, resultCode, time.Now().Sub(stepStartTime), exitCode)
	}

	registerStepListItemRunResults := func(stepListItem models.StepListItemModel, resultCode int, exitCode int, err error) {
		name := ""
		for key := range stepListItem {
			name = key
			break
		}

		stepResults := models.StepRunResultsModel{
			StepName: name,
			Error:    err,
			ExitCode: exitCode,
		}

		switch resultCode {
		case bitrise.StepRunResultCodeSuccess:
			buildRunResults.SuccessSteps = append(buildRunResults.SuccessSteps, stepResults)
			break
		case bitrise.StepRunResultCodeFailed:
			log.Errorf("Step (%s) failed, error: (%v)", name, err)
			buildRunResults.FailedSteps = append(buildRunResults.FailedSteps, stepResults)
			break
		case bitrise.StepRunResultCodeFailedSkippable:
			log.Warnf("Step (%s) failed, but was marked as skippable, error: (%v)", name, err)
			buildRunResults.FailedSkippableSteps = append(buildRunResults.FailedSkippableSteps, stepResults)
			break
		case bitrise.StepRunResultCodeSkipped:
			log.Warnf("A previous step failed, and this step (%s) was not marked as IsAlwaysRun, skipped", name)
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		case bitrise.StepRunResultCodeSkippedWithRunIf:
			log.Warn("The step's (" + name + ") Run-If expression evaluated to false - skipping")
			buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
			break
		default:
			log.Error("Unkown result code")
			return
		}

		bitrise.PrintStepSummary(name, resultCode, time.Now().Sub(stepStartTime), exitCode)
	}

	for idx, stepListItm := range workflow.Steps {
		stepStartTime = time.Now()

		if err := bitrise.SetBuildFailedEnv(buildRunResults.IsBuildFailed()); err != nil {
			log.Error("Failed to set Build Status envs")
		}
		compositeStepIDStr, workflowStep, err := models.GetStepIDStepDataPair(stepListItm)
		if err != nil {
			registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
			continue
		}
		stepIDData, err := models.CreateStepIDDataFromString(compositeStepIDStr, defaultStepLibSource)
		if err != nil {
			registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
			continue
		}

		log.Debugf("[BITRISE_CLI] - Running Step: %#v", workflowStep)

		if err := bitrise.CleanupStepWorkDir(); err != nil {
			registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
			continue
		}

		stepDir := bitrise.BitriseWorkStepsDirPath
		stepYMLPth := bitrise.BitriseWorkDirPath + "/current_step.yml"

		if stepIDData.SteplibSource == "path" {
			log.Debugf("[BITRISE_CLI] - Local step found: (path:%s)", stepIDData.IDorURI)
			stepAbsLocalPth, err := pathutil.AbsPath(stepIDData.IDorURI)
			if err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}

			log.Debugln("stepAbsLocalPth:", stepAbsLocalPth, "|stepDir:", stepDir)

			if err := cmdex.CopyDir(stepAbsLocalPth, stepDir, true); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}
			if err := cmdex.CopyFile(stepAbsLocalPth+"/step.yml", stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}
		} else if stepIDData.SteplibSource == "git" {
			log.Debugf("[BITRISE_CLI] - Remote step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)
			if err := cmdex.GitCloneTagOrBranch(stepIDData.IDorURI, stepDir, stepIDData.Version); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}
			if err := cmdex.CopyFile(stepDir+"/step.yml", stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}
		} else if stepIDData.SteplibSource != "" {
			log.Debugf("[BITRISE_CLI] - Steplib (%s) step (id:%s) (version:%s) found, activating step", stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version)
			if err := bitrise.RunStepmanSetup(stepIDData.SteplibSource); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}

			if err := bitrise.RunStepmanActivate(stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version, stepDir, stepYMLPth); err != nil {
				registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			} else {
				log.Debugf("[BITRISE_CLI] - Step activated: (ID:%s) (version:%s)", stepIDData.IDorURI, stepIDData.Version)
			}
		} else {
			registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, fmt.Errorf("Invalid stepIDData: No SteplibSource or LocalPath defined (%v)", stepIDData))
			continue
		}

		specStep, err := bitrise.ReadSpecStep(stepYMLPth)
		log.Debugf("Spec read from YML: %#v\n", specStep)
		if err != nil {
			registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
			continue
		}

		mergedStep, err := models.MergeStepWith(specStep, workflowStep)
		if err != nil {
			registerStepListItemRunResults(stepListItm, bitrise.StepRunResultCodeFailed, 1, err)
			continue
		}

		bitrise.PrintRunningStep(*mergedStep.Title, idx)
		if mergedStep.RunIf != nil && *mergedStep.RunIf != "" {
			isRun, err := bitrise.EvaluateStepTemplateToBool(*mergedStep.RunIf, buildRunResults)
			if err != nil {
				registerStepRunResults(mergedStep, bitrise.StepRunResultCodeFailed, 1, err)
				continue
			}
			if !isRun {
				registerStepRunResults(mergedStep, bitrise.StepRunResultCodeSkippedWithRunIf, 0, err)
				continue
			}
		}
		if buildRunResults.IsBuildFailed() && !*mergedStep.IsAlwaysRun {
			registerStepRunResults(mergedStep, bitrise.StepRunResultCodeSkipped, 0, err)
		} else {
			exit, err := runStep(mergedStep, stepIDData, stepDir)
			if err != nil {
				if *mergedStep.IsSkippable {
					registerStepRunResults(mergedStep, bitrise.StepRunResultCodeFailedSkippable, exit, err)
				} else {
					registerStepRunResults(mergedStep, bitrise.StepRunResultCodeFailed, exit, err)
				}
			} else {
				registerStepRunResults(mergedStep, bitrise.StepRunResultCodeSuccess, 0, nil)
			}
		}
	}

	return buildRunResults
}

func runWorkflow(workflow models.WorkflowModel, steplibSource string, buildRunResults models.BuildRunResultsModel) models.BuildRunResultsModel {
	bitrise.PrintRunningWorkflow(workflow.Title)

	if err := bitrise.ExportEnvironmentsList(workflow.Environments); err != nil {
		bitrise.PrintBuildFailedFatal(buildRunResults.StartTime, errors.New("[BITRISE_CLI] - Failed to export Workflow environments: "+err.Error()))
	}

	buildRunResults = activateAndRunSteps(workflow, steplibSource, buildRunResults)
	return buildRunResults
}

func activateAndRunWorkflow(workflow models.WorkflowModel, bitriseConfig models.BitriseDataModel, buildRunResults models.BuildRunResultsModel) models.BuildRunResultsModel {
	// Run these workflows before running the target workflow
	for _, beforeWorkflowName := range workflow.BeforeRun {
		beforeWorkflow, exist := bitriseConfig.Workflows[beforeWorkflowName]
		if !exist {
			bitrise.PrintBuildFailedFatal(buildRunResults.StartTime, errors.New("[BITRISE_CLI] - Specified Workflow ("+beforeWorkflowName+") does not exist!"))
		}
		if beforeWorkflow.Title == "" {
			beforeWorkflow.Title = beforeWorkflowName
		}
		buildRunResults = activateAndRunWorkflow(beforeWorkflow, bitriseConfig, buildRunResults)
	}

	// Run the target workflow
	buildRunResults = runWorkflow(workflow, bitriseConfig.DefaultStepLibSource, buildRunResults)

	// Run these workflows after running the target workflow
	for _, afterWorkflowName := range workflow.AfterRun {
		afterWorkflow, exist := bitriseConfig.Workflows[afterWorkflowName]
		if !exist {
			bitrise.PrintBuildFailedFatal(buildRunResults.StartTime, errors.New("[BITRISE_CLI] - Specified Workflow ("+afterWorkflowName+") does not exist!"))
		}
		if afterWorkflow.Title == "" {
			afterWorkflow.Title = afterWorkflowName
		}
		buildRunResults = activateAndRunWorkflow(afterWorkflow, bitriseConfig, buildRunResults)
	}

	return buildRunResults
}

func doRun(c *cli.Context) {
	PrintBitriseHeaderASCIIArt(c.App.Version)
	log.Debugln("[BITRISE_CLI] - Run")

	if !bitrise.CheckIsSetupWasDoneForVersion(c.App.Version) {
		log.Warnln(colorstring.Yellow("Setup was not performed for this version of bitrise, doing it now..."))
		if err := bitrise.RunSetup(c.App.Version); err != nil {
			log.Fatalln("Setup failed:", err)
		}
	}

	startTime := time.Now()

	// Input validation
	bitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" {
		log.Debugln("[BITRISE_CLI] - Workflow path not defined, searching for " + DefaultBitriseConfigFileName + " in current folder...")
		bitriseConfigPath = bitrise.CurrentDir + "/" + DefaultBitriseConfigFileName

		if exist, err := pathutil.IsPathExists(bitriseConfigPath); err != nil {
			bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to check path:"+err.Error()))
		} else if !exist {
			log.Fatalln("[BITRISE_CLI] - No workflow yml found")
			bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - No workflow yml found"))
		}
	}

	inventoryPath := c.String(InventoryKey)
	if inventoryPath == "" {
		log.Debugln("[BITRISE_CLI] - Inventory path not defined, searching for " + DefaultSecretsFileName + " in current folder...")
		inventoryPath = bitrise.CurrentDir + "/" + DefaultSecretsFileName

		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to check path: "+err.Error()))
		} else if !exist {
			log.Debugln("[BITRISE_CLI] - No inventory yml found")
			inventoryPath = ""
		}
	} else {
		if exist, err := pathutil.IsPathExists(inventoryPath); err != nil {
			bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to check path: "+err.Error()))
		} else if !exist {
			bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - No inventory yml found"))
		}
	}
	if inventoryPath != "" {
		if err := bitrise.RunEnvmanEnvstoreTest(inventoryPath); err != nil {
			bitrise.PrintBuildFailedFatal(startTime, errors.New("Invalid invetory format: "+err.Error()))
		}

		if err := cmdex.CopyFile(inventoryPath, bitrise.EnvstorePath); err != nil {
			bitrise.PrintBuildFailedFatal(startTime, errors.New("Failed to copy inventory: "+err.Error()))
		}
	}

	// Workflow selection
	workflowToRunName := ""
	if len(c.Args()) < 1 {
		log.Infoln("No workfow specified!")
	} else {
		workflowToRunName = c.Args()[0]
	}

	// Envman setup
	if err := os.Setenv(bitrise.EnvstorePathEnvKey, bitrise.EnvstorePath); err != nil {
		bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to add env: "+err.Error()))
	}

	if err := os.Setenv(bitrise.FormattedOutputPathEnvKey, bitrise.FormattedOutputPath); err != nil {
		bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to add env: "+err.Error()))
	}

	if inventoryPath == "" {
		if err := bitrise.RunEnvmanInit(); err != nil {
			bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to run envman init"))
		}
	}

	// Run work flow
	bitriseConfig, err := bitrise.ReadBitriseConfig(bitriseConfigPath)
	if err != nil {
		bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to read Workflow: "+err.Error()))
	}

	isConfigVersionOK, err := versions.IsVersionGreaterOrEqual(c.App.Version, bitriseConfig.FormatVersion)
	if err != nil {
		log.Warn("bitrise CLI version: ", c.App.Version)
		log.Warn("bitrise.yml Format Version: ", bitriseConfig.FormatVersion)
		log.Fatalln("Failed to compare bitrise CLI's version with the bitrise.yml FormatVersion: ", err)
	}
	if !isConfigVersionOK {
		log.Warnf("The bitrise.yml has a higher Format Version (%s) than the bitrise CLI's version (%s).", bitriseConfig.FormatVersion, c.App.Version)
		log.Fatalln("This bitrise.yml was created with and for a newer version of bitrise CLI, please upgrade your bitrise CLI to use this bitrise.yml!")
	}

	// Check workflow
	if workflowToRunName == "" {
		// no workflow specified
		//  list all the available ones and then exit
		log.Infoln("The following workflows are available:")
		for wfName := range bitriseConfig.Workflows {
			log.Infoln(" * " + wfName)
		}
		fmt.Println()
		log.Infoln("You can run a selected workflow with:")
		log.Infoln("-> bitrise run the-workflow-name")
		os.Exit(1)
	}

	// App level environment
	if err := bitrise.ExportEnvironmentsList(bitriseConfig.App.Environments); err != nil {
		bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Failed to export App environments: "+err.Error()))
	}

	workflowToRun, exist := bitriseConfig.Workflows[workflowToRunName]
	if !exist {
		bitrise.PrintBuildFailedFatal(startTime, errors.New("[BITRISE_CLI] - Specified Workflow ("+workflowToRunName+") does not exist!"))
	}
	if workflowToRun.Title == "" {
		workflowToRun.Title = workflowToRunName
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: startTime,
	}
	buildRunResults = activateAndRunWorkflow(workflowToRun, bitriseConfig, buildRunResults)

	// // Build finished
	bitrise.PrintSummary(buildRunResults)
	if len(buildRunResults.FailedSteps) > 0 {
		log.Fatal("[BITRISE_CLI] - Workflow FINISHED but a couple of steps failed - Ouch")
	} else {
		if len(buildRunResults.FailedSkippableSteps) > 0 {
			log.Warn("[BITRISE_CLI] - Workflow FINISHED but a couple of non imporatant steps failed")
		}
	}
}

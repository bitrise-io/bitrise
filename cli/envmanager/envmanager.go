package envmanager

import (
	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/models"
	envmanModels "github.com/bitrise-io/envman/v2/models"
)

type WorkflowEnvManager struct {
	secrets        []envmanModels.EnvironmentItemModel
	workflowEnvs   []envmanModels.EnvironmentItemModel
	stepBundleEnvs []envmanModels.EnvironmentItemModel

	currentStepBundleUUID string
	buildFailed           bool
}

func NewWorkflowEnvManager(secrets []envmanModels.EnvironmentItemModel, appEnvs []envmanModels.EnvironmentItemModel, workflowID, workflowTitle string, workflowEnv []envmanModels.EnvironmentItemModel) *WorkflowEnvManager {
	var initialEnv []envmanModels.EnvironmentItemModel
	// envman setup envs
	initialEnv = append(initialEnv,
		envmanModels.EnvironmentItemModel{configs.EnvstorePathEnvKey: configs.OutputEnvstorePath},
		envmanModels.EnvironmentItemModel{configs.FormattedOutputPathEnvKey: configs.FormattedOutputPathEnvKey},
	)

	// App level environment
	initialEnv = append(secrets, appEnvs...)
	initialEnv = append(initialEnv, bitrise.BuildStatusEnvs(false)...)

	// Target workflow environment
	initialEnv = append(initialEnv,
		envmanModels.EnvironmentItemModel{"BITRISE_TRIGGERED_WORKFLOW_ID": workflowID},
		envmanModels.EnvironmentItemModel{"BITRISE_TRIGGERED_WORKFLOW_TITLE": workflowTitle},
	)

	// TODO: reconsider adding workflow envs here:
	//  a, we add target workflow envs twice (here and in WorkflowStart)
	//  b, we expose target workflow envs to before_run workflows.
	initialEnv = append(initialEnv, workflowEnv...)

	return &WorkflowEnvManager{
		secrets:      copyenvs(secrets),
		workflowEnvs: initialEnv,
	}
}

func (em *WorkflowEnvManager) WorkflowStart(workflowEnv []envmanModels.EnvironmentItemModel) {
	em.workflowEnvs = append(em.workflowEnvs, workflowEnv...)
}

func (em *WorkflowEnvManager) EnvsForStepStart(stepPlan models.StepExecutionPlan) []envmanModels.EnvironmentItemModel {
	if stepPlan.StepBundleUUID != em.currentStepBundleUUID {
		// new bundle started / old bundle ended
		em.currentStepBundleUUID = stepPlan.StepBundleUUID

		if stepPlan.StepBundleUUID != "" {
			// new bundle started -> reset step bundle envs
			em.stepBundleEnvs = append(copyenvs(em.workflowEnvs), stepPlan.StepBundleEnvs...)
		} else {
			// old bundle ended -> clear step bundle envs
			em.stepBundleEnvs = nil
		}
	}

	if em.currentStepBundleUUID != "" {
		return copyenvs(em.stepBundleEnvs)
	}

	return copyenvs(em.workflowEnvs)
}

func (em *WorkflowEnvManager) UpdateWithStepFinished(outputEnvs []envmanModels.EnvironmentItemModel, buildRunResult models.BuildRunResultsModel) {
	envsToAdd := copyenvs(outputEnvs)

	// Check if workflow started failing
	if em.buildFailed == false && buildRunResult.IsBuildFailed() {
		em.buildFailed = true

		// Add Failed step related envs
		if len(buildRunResult.FailedSteps) == 1 {
			failedStepRunResult := buildRunResult.FailedSteps[0]
			failedStepEnvs := bitrise.FailedStepEnvs(failedStepRunResult)
			envsToAdd = append(envsToAdd, failedStepEnvs...)
		}

		// Add Failed build related envs
		buildStatusEnvs := bitrise.BuildStatusEnvs(true)
		envsToAdd = append(envsToAdd, buildStatusEnvs...)
	}

	em.workflowEnvs = append(em.workflowEnvs, envsToAdd...)

	if em.currentStepBundleUUID != "" {
		em.stepBundleEnvs = append(em.stepBundleEnvs, envsToAdd...)
	}
}

func copyenvs(envItems []envmanModels.EnvironmentItemModel) []envmanModels.EnvironmentItemModel {
	copied := make([]envmanModels.EnvironmentItemModel, len(envItems))
	copy(copied, envItems)
	return copied
}

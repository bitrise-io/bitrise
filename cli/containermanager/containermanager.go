package containermanager

import (
	"github.com/bitrise-io/bitrise/v2/cli/docker"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/tools"
	envmanModels "github.com/bitrise-io/envman/v2/models"
)

type Manager struct {
	executionContainerDefinitions map[string]models.Container
	serviceContainerDefinitions   map[string]models.Container

	dockerManager *docker.ContainerManager

	// keeps track of the current with group ID
	currentWithGroupID string
}

func NewManager(executionContainers map[string]models.Container, serviceContainers map[string]models.Container, dockerManager *docker.ContainerManager) Manager {
	return Manager{
		executionContainerDefinitions: executionContainers,
		serviceContainerDefinitions:   serviceContainers,
		dockerManager:                 dockerManager,
	}
}

func (m *Manager) UpdateWithStepStarted(stepPlan models.StepExecutionPlan, environments []envmanModels.EnvironmentItemModel, workflowTitle string) {
	if stepPlan.WithGroupUUID != m.currentWithGroupID {
		if stepPlan.WithGroupUUID != "" {
			if len(stepPlan.ContainerID) > 0 || len(stepPlan.ServiceIDs) > 0 {
				m.startContainersForStepGroup(stepPlan.ContainerID, stepPlan.ServiceIDs, environments, stepPlan.WithGroupUUID, workflowTitle)
			}
		}

		m.currentWithGroupID = stepPlan.WithGroupUUID
	}
}

func (m *Manager) UpdateWithStepFinished(stepIDX int, plan models.WorkflowExecutionPlan) {
	// Shut down containers if the step is in a 'With group', and it's the last step in the group
	if m.currentWithGroupID != "" {
		isLastStepInWorkflow := stepIDX == len(plan.Steps)-1
		doesStepGroupChange := stepIDX < len(plan.Steps)-1 && m.currentWithGroupID != plan.Steps[stepIDX+1].WithGroupUUID
		if isLastStepInWorkflow || doesStepGroupChange {
			m.stopContainersForStepGroup(m.currentWithGroupID, plan.WorkflowTitle)
			m.currentWithGroupID = ""
		}
	}
}

func (m *Manager) GetExecutionContainerFroStepGroup(groupID string) *docker.RunningContainer {
	return m.dockerManager.GetExecutionContainerForStepGroup(groupID)
}

func (m *Manager) DestroyAllContainers() error {
	return m.dockerManager.DestroyAllContainers()
}

func (m *Manager) startContainersForStepGroup(containerID string, serviceIDs []string, environments []envmanModels.EnvironmentItemModel, groupID, workflowTitle string) {
	if containerID == "" && len(serviceIDs) == 0 {
		return
	}

	envList := m.initEnvs(environments)

	if containerID != "" {
		containerDef := m.getExecutionContainerDefinition(containerID)
		if containerDef != nil {
			log.Infof("ℹ️ Running workflow in docker container: %s", containerDef.Image)

			_, err := m.dockerManager.StartExecutionContainerForStepGroup(*containerDef, groupID, envList)
			if err != nil {
				log.Errorf("Could not start the specified docker image for workflow: %s", workflowTitle)
			}
		}
	}

	if len(serviceIDs) > 0 {
		servicesDefs := m.getServiceContainerDefinitions(serviceIDs...)
		_, err := m.dockerManager.StartServiceContainersForStepGroup(servicesDefs, groupID, envList)
		if err != nil {
			log.Errorf("❌ Some services failed to start properly!")
		}
	}
}

func (m *Manager) stopContainersForStepGroup(groupID, workflowTitle string) {
	if container := m.dockerManager.GetExecutionContainerForStepGroup(groupID); container != nil {
		// TODO: Feature idea, make this configurable, so that we can keep the container for debugging purposes.
		if err := container.Destroy(); err != nil {
			log.Errorf("Attempted to stop the docker container for workflow: %s: %s", workflowTitle, err)
		}
	}

	if services := m.dockerManager.GetServiceContainersForStepGroup(groupID); services != nil {
		for _, container := range services {
			if err := container.Destroy(); err != nil {
				log.Errorf("Attempted to stop the docker container for service: %s: %s", container.Name, err)
			}
		}
	}
}

func (m *Manager) initEnvs(environments []envmanModels.EnvironmentItemModel) envmanModels.EnvsJSONListModel {
	if err := tools.EnvmanInit(configs.InputEnvstorePath, true); err != nil {
		log.Debugf("Couldn't initialize envman.")
	}
	if err := tools.EnvmanAddEnvs(configs.InputEnvstorePath, environments); err != nil {
		log.Debugf("Couldn't add envs.")
	}

	envList, err := tools.EnvmanReadEnvList(configs.InputEnvstorePath)
	if err != nil {
		log.Debugf("Couldn't read envs from envman.")
	}

	return envList
}

func (m *Manager) getExecutionContainerDefinition(id string) *models.Container {
	container, ok := m.executionContainerDefinitions[id]
	if ok {
		return &container
	}
	return nil
}

func (m *Manager) getServiceContainerDefinitions(ids ...string) map[string]models.Container {
	services := map[string]models.Container{}
	for _, id := range ids {
		service, ok := m.serviceContainerDefinitions[id]
		if ok {
			services[id] = service
		}
	}
	return services
}

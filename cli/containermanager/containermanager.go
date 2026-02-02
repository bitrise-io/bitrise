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
	services           map[string]models.Container
	containers         map[string]models.Container
	dockerManager      *docker.ContainerManager
	currentStepGroupID string
}

func NewManager(services map[string]models.Container, containers map[string]models.Container, dockerManager *docker.ContainerManager) Manager {
	return Manager{
		services:      services,
		containers:    containers,
		dockerManager: dockerManager,
	}
}

func (m *Manager) UpdateWithStepStarted(stepPlan models.StepExecutionPlan, environments []envmanModels.EnvironmentItemModel, workflowTitle string) {
	if stepPlan.WithGroupUUID != m.currentStepGroupID {
		if stepPlan.WithGroupUUID != "" {
			if len(stepPlan.ContainerID) > 0 || len(stepPlan.ServiceIDs) > 0 {
				m.startContainersForStepGroup(stepPlan.ContainerID, stepPlan.ServiceIDs, environments, stepPlan.WithGroupUUID, workflowTitle)
			}
		}

		m.currentStepGroupID = stepPlan.WithGroupUUID
	}
}

func (m *Manager) UpdateWithStepFinished(stepIDX int, plan models.WorkflowExecutionPlan) {
	isLastStepInWorkflow := stepIDX == len(plan.Steps)-1

	if m.currentStepGroupID != "" {
		doesStepGroupChange := stepIDX < len(plan.Steps)-1 && m.currentStepGroupID != plan.Steps[stepIDX+1].WithGroupUUID
		if isLastStepInWorkflow || doesStepGroupChange {
			m.stopContainersForStepGroup(m.currentStepGroupID, plan.WorkflowTitle)
		}
	}
}

func (m *Manager) GetExecutionContainer(groupID string) *docker.RunningContainer {
	return m.dockerManager.GetExecutionContainer(groupID)
}

func (m *Manager) DestroyAllContainers() error {
	return m.dockerManager.DestroyAllContainers()
}

func (m *Manager) startContainersForStepGroup(containerID string, serviceIDs []string, environments []envmanModels.EnvironmentItemModel, groupID, workflowTitle string) {
	if containerID == "" && len(serviceIDs) == 0 {
		return
	}

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

	if containerID != "" {
		containerDef := m.ContainerDefinition(containerID)
		if containerDef != nil {
			log.Infof("ℹ️ Running workflow in docker container: %s", containerDef.Image)

			_, err := m.dockerManager.StartExecutionContainer(*containerDef, groupID, envList)
			if err != nil {
				log.Errorf("Could not start the specified docker image for workflow: %s", workflowTitle)
			}
		}
	}

	if len(serviceIDs) > 0 {
		servicesDefs := m.ServiceDefinitions(serviceIDs...)
		_, err := m.dockerManager.StartServiceContainers(servicesDefs, groupID, envList)
		if err != nil {
			log.Errorf("❌ Some services failed to start properly!")
		}
	}
}

func (m *Manager) stopContainersForStepGroup(groupID, workflowTitle string) {
	if container := m.dockerManager.GetExecutionContainer(groupID); container != nil {
		// TODO: Feature idea, make this configurable, so that we can keep the container for debugging purposes.
		if err := container.Destroy(); err != nil {
			log.Errorf("Attempted to stop the docker container for workflow: %s: %s", workflowTitle, err)
		}
	}

	if services := m.dockerManager.GetServiceContainers(groupID); services != nil {
		for _, container := range services {
			if err := container.Destroy(); err != nil {
				log.Errorf("Attempted to stop the docker container for service: %s: %s", container.Name, err)
			}
		}
	}
}

func (m *Manager) ContainerDefinition(id string) *models.Container {
	container, ok := m.containers[id]
	if ok {
		return &container
	}
	return nil
}

func (m *Manager) ServiceDefinitions(ids ...string) map[string]models.Container {
	services := map[string]models.Container{}
	for _, id := range ids {
		service, ok := m.services[id]
		if ok {
			services[id] = service
		}
	}
	return services
}

package containermanager

import (
	"github.com/bitrise-io/bitrise/v2/cli/docker"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/tools"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/google/uuid"
)

type Manager struct {
	executionContainers map[string]models.Container
	serviceContainers   map[string]models.Container
	dockerManager       *docker.ContainerManager

	plan                   models.WorkflowRunPlan
	legacyContainerisation bool

	// keeps track of the current with group ID
	currentWithGroupID string

	// Keeps track of the currently running containers
	currentExecutionContainer string
	currentServiceContainers  []string
}

func NewManager(executionContainers map[string]models.Container, serviceContainers map[string]models.Container, dockerManager *docker.ContainerManager) Manager {
	return Manager{
		executionContainers: executionContainers,
		serviceContainers:   serviceContainers,
		dockerManager:       dockerManager,
	}
}

func (m *Manager) SetLegacyContainerisation(useLegacy bool) {
	m.legacyContainerisation = useLegacy
}

func (m *Manager) UpdateWithStepStarted(stepPlan models.StepExecutionPlan, environments []envmanModels.EnvironmentItemModel, workflowTitle string) {
	if m.legacyContainerisation {
		if stepPlan.WithGroupUUID != m.currentWithGroupID {
			if stepPlan.WithGroupUUID != "" {
				if len(stepPlan.ContainerID) > 0 || len(stepPlan.ServiceIDs) > 0 {
					m.startContainersForStepGroup(stepPlan.ContainerID, stepPlan.ServiceIDs, environments, stepPlan.WithGroupUUID, workflowTitle)
				}
			}

			m.currentWithGroupID = stepPlan.WithGroupUUID
		}
	} else {
		var startExecutionContainer string
		if m.shouldStartExecutionContainer(stepPlan) {
			startExecutionContainer = stepPlan.ExecutionContainer.ContainerID
		}

		var startServiceContainers []string
		// TODO: implement service container transitions

		if startExecutionContainer != "" {
			newGroupUUID := uuid.New()
			m.startContainersForStepGroup(startExecutionContainer, startServiceContainers, environments, newGroupUUID.String(), workflowTitle)
		}

		/*
			possible transitions for execution containers:
			A, no running container:
				- step requires a container: start it
				- step requires no container: do nothing
			B, running container:
				- step requires no container:
					- first next step with container requirement requires the same container: do nothing
					- first next step with container requirement requires the same container with recreate: stop running container, start new container later
					- first next step with container requirement requires different container: stop running container, start new container later
					- no more steps with container requirement: stop running container
				- step requires a container:
					- step requires same container: do nothing
					- step requires different container: stop running container, start new container

			possible transitions for execution containers at step will start event:
			A, no running container:
				- step requires a container: start it
				- step requires no container: do nothing
			B, running container:
				- step requires no container:
					- first next step with container requirement requires the same container: do nothing
					- first next step with container requirement requires the same container with recreate: stop running container, start new container later
					- first next step with container requirement requires different container: stop running container, start new container later
					- no more steps with container requirement: stop running container
				- step requires a container:
					- step requires same container: do nothing
					- step requires different container: stop running container, start new container
		*/
	}
}

func (m *Manager) shouldStartExecutionContainer(stepPlan models.StepExecutionPlan) bool {
	return m.currentExecutionContainer == "" && stepPlan.ExecutionContainer != nil
}

func (m *Manager) shouldStopExecutionContainer(stepPlan models.StepExecutionPlan, stepIDX int) bool {
	return false
}

func (m *Manager) UpdateWithStepFinished(stepIDX int, plan models.WorkflowExecutionPlan) {
	if !m.legacyContainerisation {
		if m.currentWithGroupID != "" {
			isLastStepInWorkflow := stepIDX == len(plan.Steps)-1
			doesStepGroupChange := stepIDX < len(plan.Steps)-1 && m.currentWithGroupID != plan.Steps[stepIDX+1].WithGroupUUID
			if isLastStepInWorkflow || doesStepGroupChange {
				m.stopContainersForStepGroup(m.currentWithGroupID, plan.WorkflowTitle)
				m.currentWithGroupID = ""
			}
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
		containerDef := m.executionContainerDefinition(containerID)
		if containerDef != nil {
			log.Infof("ℹ️ Running workflow in docker container: %s", containerDef.Image)

			_, err := m.dockerManager.StartExecutionContainer(*containerDef, groupID, envList)
			if err != nil {
				log.Errorf("Could not start the specified docker image for workflow: %s", workflowTitle)
			}
		}
	}

	if len(serviceIDs) > 0 {
		servicesDefs := m.serviceContainerDefinitions(serviceIDs...)
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

func (m *Manager) executionContainerDefinition(id string) *models.Container {
	container, ok := m.executionContainers[id]
	if ok {
		return &container
	}
	return nil
}

func (m *Manager) serviceContainerDefinitions(ids ...string) map[string]models.Container {
	services := map[string]models.Container{}
	for _, id := range ids {
		service, ok := m.serviceContainers[id]
		if ok {
			services[id] = service
		}
	}
	return services
}

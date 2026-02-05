package containermanager

import (
	"fmt"
	"sync"

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
	logger        *docker.Logger

	// keeps track of the current with group ID and the currently running containers for the group
	currentWithGroupID  string
	executionContainers map[string]*docker.RunningContainer
	serviceContainers   map[string][]*docker.RunningContainer

	mu       sync.Mutex
	released bool
}

func NewManager(executionContainers map[string]models.Container, serviceContainers map[string]models.Container, dockerManager *docker.ContainerManager, logger *docker.Logger) *Manager {
	return &Manager{
		executionContainerDefinitions: executionContainers,
		serviceContainerDefinitions:   serviceContainers,
		dockerManager:                 dockerManager,
		logger:                        logger,
		executionContainers:           make(map[string]*docker.RunningContainer),
		serviceContainers:             make(map[string][]*docker.RunningContainer),
	}
}

func (m *Manager) UpdateWithStepStarted(stepPlan models.StepExecutionPlan, environments []envmanModels.EnvironmentItemModel) {
	if stepPlan.WithGroupUUID != m.currentWithGroupID {
		if stepPlan.WithGroupUUID != "" {
			if len(stepPlan.ContainerID) > 0 || len(stepPlan.ServiceIDs) > 0 {
				m.startContainersForStepGroup(stepPlan.ContainerID, stepPlan.ServiceIDs, environments, stepPlan.WithGroupUUID)
			}
		}

		m.currentWithGroupID = stepPlan.WithGroupUUID
	} else {
		/*
			possible transitions for execution containers:
			A, no running container:
				- step requires a container: start it
				- step requires no container: do nothing
			B, running container:
				- first next step with container requirement requires the same container: do nothing
				- first next step with container requirement requires the same container with recreate: stop running container, start new container later
				- first next step with container requirement requires different container: stop running container, start new container later
				- no more steps with container requirement: stop running container
		*/
	}
}

func (m *Manager) UpdateWithStepFinished(stepIDX int, plan models.WorkflowExecutionPlan) {
	// Shut down containers if the step is in a 'With group', and it's the last step in the group
	if m.currentWithGroupID != "" {
		isLastStepInWorkflow := stepIDX == len(plan.Steps)-1
		doesStepGroupChange := stepIDX < len(plan.Steps)-1 && m.currentWithGroupID != plan.Steps[stepIDX+1].WithGroupUUID
		if isLastStepInWorkflow || doesStepGroupChange {
			m.stopContainersForStepGroup(m.currentWithGroupID)
			m.currentWithGroupID = ""
		}
	}
}

func (m *Manager) GetExecutionContainerForStepGroup(groupID string) *docker.RunningContainer {
	return m.executionContainers[groupID]
}

func (m *Manager) DestroyAllContainers() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.released = true

	for _, executionContainer := range m.executionContainers {
		m.logger.Infof("ℹ️ Removing execution container: %s", executionContainer.Name)
		if err := executionContainer.Destroy(); err != nil {
			return fmt.Errorf("destroy execution container: %w", err)
		}
	}

	for _, serviceContainers := range m.serviceContainers {
		for _, serviceContainer := range serviceContainers {
			if serviceContainer == nil {
				continue
			}
			m.logger.Infof("Removing service container: %s", serviceContainer.Name)
			if err := serviceContainer.Destroy(); err != nil {
				return fmt.Errorf("destroy service container: %w", err)
			}
		}
	}

	return nil
}

func (m *Manager) startContainersForStepGroup(containerID string, serviceIDs []string, environments []envmanModels.EnvironmentItemModel, groupID string) {
	if containerID == "" && len(serviceIDs) == 0 {
		return
	}

	envList := m.initEnvs(environments)

	if containerID != "" {
		containerDef := m.getExecutionContainerDefinition(containerID)
		if containerDef != nil {
			log.Infof("ℹ️ Running step group in docker container: %s", containerDef.Image)

			_, err := m.startExecutionContainerForStepGroup(*containerDef, groupID, envList)
			if err != nil {
				log.Errorf("Could not start the specified docker image: %s", containerDef.Image)
			}
		}
	}

	if len(serviceIDs) > 0 {
		servicesDefs := m.getServiceContainerDefinitions(serviceIDs...)
		_, err := m.startServiceContainersForStepGroup(servicesDefs, groupID, envList)
		if err != nil {
			log.Errorf("❌ Some services failed to start properly!")
		}
	}
}

func (m *Manager) startExecutionContainerForStepGroup(container models.Container, groupID string, envs map[string]string) (*docker.RunningContainer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	runningContainer, err := m.loginAndRunContainer(docker.ExecutionContainerType, container, fmt.Sprintf("bitrise-workflow-%s", groupID), envs)
	// Even on failure we save the reference to make sure containers will be cleaned up
	if runningContainer != nil {
		m.executionContainers[groupID] = runningContainer
	}
	return runningContainer, err
}

func (m *Manager) startServiceContainersForStepGroup(containers map[string]models.Container, groupID string, envs map[string]string) ([]*docker.RunningContainer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var runningContainers []*docker.RunningContainer
	failedServices := make(map[string]error)

	for containerID := range containers {
		serviceContainer := containers[containerID]

		runningContainer, err := m.loginAndRunContainer(docker.ServiceContainerType, serviceContainer, containerID, envs)
		if runningContainer != nil {
			runningContainers = append(runningContainers, runningContainer)
		}
		if err != nil {
			failedServices[containerID] = err
		}
	}

	// Even on failure we save the references to make sure containers will be cleaned up
	m.serviceContainers[groupID] = append(m.serviceContainers[groupID], runningContainers...)

	if len(failedServices) != 0 {
		errServices := fmt.Errorf("failed to start services")
		for containerID, err := range failedServices {
			errServices = fmt.Errorf("%v: %w", errServices, err)
			m.logger.Errorf("Failed to start service container (%s): %s", containerID, err)
		}
		return runningContainers, errServices
	}
	return runningContainers, nil
}

func (m *Manager) loginAndRunContainer(t docker.ContainerType, containerDef models.Container, containerName string, envs map[string]string) (*docker.RunningContainer, error) {
	if m.released {
		return nil, fmt.Errorf("container manager was released already")
	}

	return m.dockerManager.LoginAndRunContainer(t, containerDef, containerName, envs)
}

func (m *Manager) stopContainersForStepGroup(groupID string) {
	if container := m.GetExecutionContainerForStepGroup(groupID); container != nil {
		// TODO: Feature idea, make this configurable, so that we can keep the container for debugging purposes.
		if err := container.Destroy(); err != nil {
			log.Errorf("Attempted to stop the docker container for step group: %s", err)
		}
	}

	if services := m.getServiceContainersForStepGroup(groupID); services != nil {
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

func (m *Manager) getServiceContainersForStepGroup(groupID string) []*docker.RunningContainer {
	return m.serviceContainers[groupID]
}

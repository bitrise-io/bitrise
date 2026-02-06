package containermanager

import (
	"fmt"
	"sync"

	"github.com/bitrise-io/bitrise/v2/cli/docker"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/tools"
	envmanModels "github.com/bitrise-io/envman/v2/models"
)

type Manager struct {
	executionContainerDefinitions map[string]models.Container
	serviceContainerDefinitions   map[string]models.Container

	dockerManager *docker.ContainerManager
	logger        *docker.Logger

	workflowRunPlan        models.WorkflowRunPlan
	legacyContainerisation bool

	// Keeps track of the current with group ID and the currently running containers for the group
	currentWithGroupID  string
	executionContainers map[string]*docker.RunningContainer
	serviceContainers   map[string][]*docker.RunningContainer

	// Keeps track of all running containers with the new containerisation mode.
	// Running containers are mapped by the container ID defined in the config.
	runningExecutionContainer map[string]*docker.RunningContainer
	runningServiceContainers  map[string]*docker.RunningContainer

	mu       sync.Mutex
	released bool
}

func NewManager(executionContainers map[string]models.Container, serviceContainers map[string]models.Container, dockerManager *docker.ContainerManager, logger *docker.Logger) *Manager {
	return &Manager{
		executionContainerDefinitions: executionContainers,
		serviceContainerDefinitions:   serviceContainers,
		dockerManager:                 dockerManager,
		logger:                        logger,
		// Legacy containerisation mode
		currentWithGroupID:  "",
		executionContainers: make(map[string]*docker.RunningContainer),
		serviceContainers:   make(map[string][]*docker.RunningContainer),
		// New containerisation mode
		runningExecutionContainer: make(map[string]*docker.RunningContainer),
		runningServiceContainers:  make(map[string]*docker.RunningContainer),
	}
}

func (m *Manager) SetWorkflowRunPlan(plan models.WorkflowRunPlan) {
	m.workflowRunPlan = plan
}

// TODO: merge it with SetWorkflowRunPlan.
func (m *Manager) SetLegacyContainerisation(legacyContainerisation bool) {
	m.legacyContainerisation = legacyContainerisation

	if !legacyContainerisation {
		// In new containerisation mode, both service and execution containers are defined under the "containers" key,
		// so we need to separate them based on their type.
		executionContainers := map[string]models.Container{}
		serviceContainers := map[string]models.Container{}

		for containerID, container := range m.executionContainerDefinitions {
			switch container.Type {
			case models.ContainerTypeExecution:
				executionContainers[containerID] = container
			case models.ContainerTypeService:
				serviceContainers[containerID] = container
			default:
				executionContainers[containerID] = container
				serviceContainers[containerID] = container
			}
		}

		m.executionContainerDefinitions = executionContainers
		m.serviceContainerDefinitions = serviceContainers
	}
}

func (m *Manager) UpdateWithStepStarted(stepPlan models.StepExecutionPlan, environments []envmanModels.EnvironmentItemModel) {
	if m.legacyContainerisation {
		if stepPlan.WithGroupUUID != m.currentWithGroupID {
			if stepPlan.WithGroupUUID != "" {
				if len(stepPlan.ContainerID) > 0 || len(stepPlan.ServiceIDs) > 0 {
					m.startContainersForStepGroup(stepPlan.ContainerID, stepPlan.ServiceIDs, environments, stepPlan.WithGroupUUID)
				}
			}

			m.currentWithGroupID = stepPlan.WithGroupUUID
		}
		return
	}

	var executionContainerToStart string
	var serviceContainersToStart []string

	if len(m.runningExecutionContainer) == 0 {
		// No running execution container, we need to check the current step's container requirement and start the container if needed.
		// Previous execution container is stopped at the end of the step execution, so we don't need to check if the required container is different from the currently running one.
		executionContainerToStart = stepPlan.ContainerID
	}

	for _, serviceContainerConfig := range stepPlan.ServiceContainers {
		// If there is no running container for the required service container, we need to start it.
		// Unused service containers are stopped at the end of the step execution, so we don't need to check if the required container is different from the currently running one.
		if _, isRunning := m.runningServiceContainers[serviceContainerConfig.ContainerID]; !isRunning {
			serviceContainersToStart = append(serviceContainersToStart, serviceContainerConfig.ContainerID)
		}
	}

	m.startContainers(executionContainerToStart, serviceContainersToStart, environments)

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

func (m *Manager) UpdateWithStepFinished(stepIDX int, plan models.WorkflowExecutionPlan, stepPlan models.StepExecutionPlan) {
	if m.legacyContainerisation {
		// Shut down containers if the step is in a 'With group', and it's the last step in the group
		if m.currentWithGroupID != "" {
			isLastStepInWorkflow := stepIDX == len(plan.Steps)-1
			doesStepGroupChange := stepIDX < len(plan.Steps)-1 && m.currentWithGroupID != plan.Steps[stepIDX+1].WithGroupUUID
			if isLastStepInWorkflow || doesStepGroupChange {
				m.stopContainersForStepGroup(m.currentWithGroupID)
				m.currentWithGroupID = ""
			}
		}
		return
	}

	if len(m.runningExecutionContainer) > 0 {
		// Find next step with execution container requirement
		currentStepFound := false
		var nextStepWithExecutionContainerRequirement *models.StepExecutionPlan
		for _, workflow := range m.workflowRunPlan.ExecutionPlan {
			for _, step := range workflow.Steps {
				if step.UUID == stepPlan.UUID {
					currentStepFound = true
					continue
				}

				if currentStepFound {
					if step.ExecutionContainer != nil {
						nextStepWithExecutionContainerRequirement = &step
						break
					}
				}
			}
			if nextStepWithExecutionContainerRequirement != nil {
				break
			}
		}

		// Current execution container ID
		currentExecutionContainerID := ""
		for containerID := range m.runningExecutionContainer {
			currentExecutionContainerID = containerID
			break
		}

		stopCurrentExecutionContainer := false
		if nextStepWithExecutionContainerRequirement == nil || nextStepWithExecutionContainerRequirement.ExecutionContainer == nil {
			// TODO: Stop current container
			stopCurrentExecutionContainer = true
		} else {
			// Next execution container ID and recreate option
			nextExecutionContainerID := ""
			shouldRestartExecutionContainer := false
			if nextStepWithExecutionContainerRequirement.ExecutionContainer != nil {
				nextExecutionContainerID = nextStepWithExecutionContainerRequirement.ExecutionContainer.ContainerID
				shouldRestartExecutionContainer = nextStepWithExecutionContainerRequirement.ExecutionContainer.Recreate
			}

			if nextExecutionContainerID != currentExecutionContainerID {
				// Next step requires a different execution container, stop the currently running execution container.
				// TODO: Stop current container
				stopCurrentExecutionContainer = true
			} else if nextExecutionContainerID == currentExecutionContainerID && shouldRestartExecutionContainer {
				// Next step requires the same execution container but with recreate option, stop the currently running execution container.
				// TODO: Stop current container
				stopCurrentExecutionContainer = true
			}
		}

		if stopCurrentExecutionContainer && currentExecutionContainerID != "" {
			if container := m.runningExecutionContainer[currentExecutionContainerID]; container != nil {
				if err := container.Destroy(); err != nil {
					m.logger.Errorf("Attempted to stop the docker container for step group: %s", err)
				}
			}
		}

	}

}

func (m *Manager) GetExecutionContainerForStepGroup(groupID string) *docker.RunningContainer {
	return m.executionContainers[groupID]
}

func (m *Manager) GetExecutionContainer(containerID string) *docker.RunningContainer {
	return m.runningExecutionContainer[containerID]
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
			m.logger.Infof("ℹ️ Running step group in docker container: %s", containerDef.Image)

			_, err := m.startExecutionContainerForStepGroup(*containerDef, groupID, envList)
			if err != nil {
				m.logger.Errorf("Could not start the specified docker image: %s", containerDef.Image)
			}
		}
	}

	if len(serviceIDs) > 0 {
		servicesDefs := m.getServiceContainerDefinitions(serviceIDs...)
		_, err := m.startServiceContainersForStepGroup(servicesDefs, groupID, envList)
		if err != nil {
			m.logger.Errorf("❌ Some services failed to start properly!")
		}
	}
}

func (m *Manager) startContainers(containerID string, serviceIDs []string, environments []envmanModels.EnvironmentItemModel) {
	if containerID == "" && len(serviceIDs) == 0 {
		return
	}

	envList := m.initEnvs(environments)

	if containerID != "" {
		containerDef := m.getExecutionContainerDefinition(containerID)
		if containerDef != nil {
			m.logger.Infof("ℹ️ Running step group in docker container: %s", containerDef.Image)

			_, err := m.startExecutionContainer(*containerDef, containerID, envList)
			if err != nil {
				m.logger.Errorf("Could not start the specified docker image: %s", containerDef.Image)
			}
		}
	}

	if len(serviceIDs) > 0 {
		servicesDefs := m.getServiceContainerDefinitions(serviceIDs...)
		_, err := m.startServiceContainers(servicesDefs, envList)
		if err != nil {
			m.logger.Errorf("❌ Some services failed to start properly!")
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

func (m *Manager) startExecutionContainer(container models.Container, containerID string, envs map[string]string) (*docker.RunningContainer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	runningContainer, err := m.loginAndRunContainer(docker.ExecutionContainerType, container, containerID, envs)
	// Even on failure we save the reference to make sure containers will be cleaned up
	if runningContainer != nil {
		m.runningExecutionContainer[containerID] = runningContainer
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

func (m *Manager) startServiceContainers(containers map[string]models.Container, envs map[string]string) ([]*docker.RunningContainer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var runningContainers []*docker.RunningContainer
	failedServices := make(map[string]error)

	for containerID := range containers {
		serviceContainer := containers[containerID]

		runningContainer, err := m.loginAndRunContainer(docker.ServiceContainerType, serviceContainer, containerID, envs)
		if runningContainer != nil {
			runningContainers = append(runningContainers, runningContainer)
			// Even on failure we save the references to make sure containers will be cleaned up
			m.runningServiceContainers[containerID] = runningContainer
		}
		if err != nil {
			failedServices[containerID] = err
		}
	}

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
			m.logger.Errorf("Attempted to stop the docker container for step group: %s", err)
		}
	}

	if services := m.getServiceContainersForStepGroup(groupID); services != nil {
		for _, container := range services {
			if err := container.Destroy(); err != nil {
				m.logger.Errorf("Attempted to stop the docker container for service: %s: %s", container.Name, err)
			}
		}
	}
}

func (m *Manager) initEnvs(environments []envmanModels.EnvironmentItemModel) envmanModels.EnvsJSONListModel {
	if err := tools.EnvmanInit(configs.InputEnvstorePath, true); err != nil {
		m.logger.Debugf("Couldn't initialize envman.")
	}
	if err := tools.EnvmanAddEnvs(configs.InputEnvstorePath, environments); err != nil {
		m.logger.Debugf("Couldn't add envs.")
	}

	envList, err := tools.EnvmanReadEnvList(configs.InputEnvstorePath)
	if err != nil {
		m.logger.Debugf("Couldn't read envs from envman.")
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

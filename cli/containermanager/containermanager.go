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

	if len(plan.WithGroupPlans) > 0 {
		m.legacyContainerisation = true
		m.logger.Debugf("Using legacy containerisation mode")
	} else {
		m.logger.Debugf("Using new containerisation mode")
		// In new containerisation mode, both service and execution containers are defined under the "containers" key,
		// so we need to separate them based on their type.
		m.executionContainerDefinitions, m.serviceContainerDefinitions = models.ProcessContainerList(m.executionContainerDefinitions)
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

	// In the new containerisation mode, containers are not tied to a step group,
	// so we check the container requirements of the current step to decide if we need to start any new containers.
	// Here we only check what to start, stopping is handled in UpdateWithStepFinished.
	var executionContainerToStart string
	var serviceContainersToStart []string

	if stepPlan.ExecutionContainer != nil && stepPlan.ExecutionContainer.ContainerID != "" {
		if _, isRunning := m.runningExecutionContainer[stepPlan.ExecutionContainer.ContainerID]; !isRunning {
			executionContainerToStart = stepPlan.ExecutionContainer.ContainerID
		}
	}

	for _, serviceContainerConfig := range stepPlan.ServiceContainers {
		if serviceContainerConfig.ContainerID == "" {
			continue
		}

		// If there is no running container for the required service container, we need to start it.
		if _, isRunning := m.runningServiceContainers[serviceContainerConfig.ContainerID]; !isRunning {
			serviceContainersToStart = append(serviceContainersToStart, serviceContainerConfig.ContainerID)
		}
	}

	m.debugLogRunningContainers(stepPlan)
	m.startContainers(executionContainerToStart, serviceContainersToStart, environments)
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

	// In the new containerisation mode, containers are not tied to a step group,
	// so we check the container requirements of the upcoming steps to decide if we need to stop the currently running containers.
	// Here we only check what to stop, starting is handled in UpdateWithStepStarted.
	if len(m.runningExecutionContainer) > 0 {
		for containerID := range m.runningExecutionContainer {
			if m.shouldStopExecutionContainer(containerID, stepPlan) {
				if container := m.runningExecutionContainer[containerID]; container != nil {
					m.logger.Infof("ℹ️ Removing execution container: %s", containerID)
					if err := container.Destroy(); err != nil {
						m.logger.Errorf("Attempted to stop execution container: %s", err)
					}
					delete(m.runningExecutionContainer, containerID)
				}
			}
		}
	}

	if len(m.runningServiceContainers) > 0 {
		for containerID := range m.runningServiceContainers {
			if m.shouldStopServiceContainer(containerID, stepPlan) {
				if container := m.runningServiceContainers[containerID]; container != nil {
					m.logger.Infof("ℹ️ Removing service container: %s", container.Name)
					if err := container.Destroy(); err != nil {
						m.logger.Errorf("Attempted to stop service container: %s", err)
					}
					delete(m.runningServiceContainers, containerID)
				}
			}
		}
	}
}

func (m *Manager) GetExecutionContainerForStep(UUID string) (*models.Container, *docker.RunningContainer) {
	stepPlan := m.findStepPlan(UUID)
	if stepPlan == nil {
		// This should not happen, but in case it does, we return nil to avoid breaking the execution.
		return nil, nil
	}

	if m.legacyContainerisation {
		if stepPlan.WithGroupUUID == "" || stepPlan.ContainerID == "" {
			return nil, nil
		}

		containerDefinition, ok := m.executionContainerDefinitions[stepPlan.ContainerID]
		if !ok {
			// This should not happen, but in case it does, we return nil to avoid breaking the execution.
			return nil, nil
		}
		runningContainer := m.executionContainers[stepPlan.WithGroupUUID]

		return &containerDefinition, runningContainer
	}

	if stepPlan.ExecutionContainer == nil || stepPlan.ExecutionContainer.ContainerID == "" {
		return nil, nil
	}

	containerDefinition, ok := m.executionContainerDefinitions[stepPlan.ExecutionContainer.ContainerID]
	if !ok {
		// This should not happen, but in case it does, we return nil to avoid breaking the execution.
		return nil, nil
	}

	runningContainer := m.runningExecutionContainer[stepPlan.ExecutionContainer.ContainerID]

	return &containerDefinition, runningContainer
}

func (m *Manager) DestroyAllContainers() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.released = true

	for _, executionContainer := range m.executionContainers {
		if executionContainer == nil {
			continue
		}
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

	for _, executionContainer := range m.runningExecutionContainer {
		if executionContainer == nil {
			continue
		}
		m.logger.Infof("ℹ️ Removing execution container: %s", executionContainer.Name)
		if err := executionContainer.Destroy(); err != nil {
			return fmt.Errorf("destroy execution container: %w", err)
		}
	}

	for _, serviceContainer := range m.runningServiceContainers {
		if serviceContainer == nil {
			continue
		}
		m.logger.Infof("Removing service container: %s", serviceContainer.Name)
		if err := serviceContainer.Destroy(); err != nil {
			return fmt.Errorf("destroy service container: %w", err)
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
			m.logger.Infof("ℹ️ Running step in docker container: %s", containerDef.Image)

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
	if container := m.getExecutionContainerForStepGroup(groupID); container != nil {
		// TODO: Feature idea, make this configurable, so that we can keep the container for debugging purposes.
		m.logger.Infof("ℹ️ Removing execution container: %s", container.Name)
		if err := container.Destroy(); err != nil {
			m.logger.Errorf("Attempted to stop the docker container for step group: %s", err)
		}
	}

	if services := m.getServiceContainersForStepGroup(groupID); services != nil {
		for _, container := range services {
			m.logger.Infof("ℹ️ Removing service container: %s", container.Name)
			if err := container.Destroy(); err != nil {
				m.logger.Errorf("Attempted to stop the docker container for service: %s: %s", container.Name, err)
			}
		}
	}
}

func (m *Manager) shouldStopExecutionContainer(containerID string, currentStepPlan models.StepExecutionPlan) bool {
	nextStepPlan, containerCfg := m.findNextStepPlanWithExecutionContainerRequirement(containerID, currentStepPlan)
	if nextStepPlan == nil || containerCfg == nil {
		// No more steps with requirement for the given execution container
		return true
	}

	if nextStepPlan.ExecutionContainer.Recreate {
		// Next step requires the same execution container but with recreate option
		return true
	}

	// Next step requires the same execution container without recreate option
	return false
}

func (m *Manager) debugLogRunningContainers(stepPlan models.StepExecutionPlan) {
	for containerID := range m.runningExecutionContainer {
		reuseContainer := stepPlan.ExecutionContainer != nil && stepPlan.ExecutionContainer.ContainerID == containerID
		if reuseContainer {
			m.logger.Debugf("Reusing execution container: %s", containerID)
		} else {
			m.logger.Debugf("Keep running execution container: %s", containerID)
		}
	}

	for containerID := range m.runningServiceContainers {
		reuseContainer := false
		for _, serviceContainerConfig := range stepPlan.ServiceContainers {
			if serviceContainerConfig.ContainerID == containerID {
				reuseContainer = true
				break
			}
		}
		if reuseContainer {
			m.logger.Debugf("Reusing service container: %s", containerID)
		} else {
			m.logger.Debugf("Keep running service container: %s", containerID)
		}
	}
}

func (m *Manager) findStepPlan(UUID string) *models.StepExecutionPlan {
	for _, workflow := range m.workflowRunPlan.ExecutionPlan {
		for _, step := range workflow.Steps {
			if step.UUID == UUID {
				return &step
			}
		}
	}
	return nil
}

func (m *Manager) findNextStepPlanWithExecutionContainerRequirement(containerID string, currentStepPlan models.StepExecutionPlan) (*models.StepExecutionPlan, *models.ContainerConfig) {
	currentStepFound := false
	for _, workflow := range m.workflowRunPlan.ExecutionPlan {
		for _, step := range workflow.Steps {
			if step.UUID == currentStepPlan.UUID {
				currentStepFound = true
				continue
			}

			if currentStepFound && step.ExecutionContainer != nil && step.ExecutionContainer.ContainerID == containerID {
				return &step, step.ExecutionContainer
			}
		}
	}

	return nil, nil
}

func (m *Manager) shouldStopServiceContainer(containerID string, currentStepPlan models.StepExecutionPlan) bool {
	nextStepPlan, containerCfg := m.findNextStepPlanWithServiceContainerRequirement(containerID, currentStepPlan)
	if nextStepPlan == nil || containerCfg == nil {
		// No more steps with requirement for the given service container
		return true
	}

	if containerCfg.Recreate {
		// Next step requires the same service container but with recreate option
		return true
	}

	// Next step requires the same service container without recreate option
	return false
}

func (m *Manager) findNextStepPlanWithServiceContainerRequirement(containerID string, currentStepPlan models.StepExecutionPlan) (*models.StepExecutionPlan, *models.ContainerConfig) {
	currentStepFound := false
	for _, workflow := range m.workflowRunPlan.ExecutionPlan {
		for _, step := range workflow.Steps {
			if step.UUID == currentStepPlan.UUID {
				currentStepFound = true
				continue
			}

			if currentStepFound {
				for _, containerCfg := range step.ServiceContainers {
					if containerCfg.ContainerID == "" {
						continue
					}

					if containerCfg.ContainerID == containerID {
						return &step, &containerCfg
					}
				}
			}
		}
	}

	return nil, nil
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

func (m *Manager) getExecutionContainerForStepGroup(groupID string) *docker.RunningContainer {
	return m.executionContainers[groupID]
}

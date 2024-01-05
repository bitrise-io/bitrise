package docker

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/command"
)

type RunningContainer struct {
	Name string // TODO refactor to use docker sdk, and return container ID instead of name
}

func (rc *RunningContainer) Destroy() error {
	_, err := command.New("docker", "rm", "--force", rc.Name).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		// rc.logger.Errorf(out)
		return fmt.Errorf("remove docker container: %w", err)
	}
	return nil
}

func (rc *RunningContainer) ExecuteCommandArgs(envs []string) []string {
	args := []string{"exec"}

	for _, env := range envs {
		args = append(args, "-e", env)
	}

	args = append(args, rc.Name)

	return args
}

type ContainerManager struct {
	logger             log.Logger
	workflowContainers map[string]*RunningContainer
	serviceContainers  map[string][]*RunningContainer
}

func NewContainerManager(logger log.Logger) *ContainerManager {
	return &ContainerManager{
		logger:             logger,
		workflowContainers: make(map[string]*RunningContainer),
		serviceContainers:  make(map[string][]*RunningContainer),
	}
}

func (cm *ContainerManager) Login(container models.Container, envs map[string]string) error {
	cm.logger.Infof("Running workflow in docker container: %s", container.Image)
	cm.logger.Debugf("Docker cred: %s", container.Credentials)

	if container.Credentials.Username != "" && container.Credentials.Password != "" {
		cm.logger.Debugf("Logging into docker registry: %s", container.Image)

		password := container.Credentials.Password
		if strings.HasPrefix(password, "$") {
			if value, ok := envs[strings.TrimPrefix(container.Credentials.Password, "$")]; ok {
				password = value
			}
		}

		args := []string{"login", "--username", container.Credentials.Username, "--password", password}

		if container.Credentials.Server != "" {
			args = append(args, container.Credentials.Server)
		} else if len(strings.Split(container.Image, "/")) > 2 {
			args = append(args, container.Image)
		}

		cm.logger.Debugf("Running command: docker %s", strings.Join(args, " "))

		out, err := command.New("docker", args...).RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			cm.logger.Errorf(out)
			return fmt.Errorf("run docker login: %w", err)
		}
	}
	return nil
}

func (cm *ContainerManager) StartWorkflowContainer(container models.Container, workflowID string) (*RunningContainer, error) {
	containerName := fmt.Sprintf("workflow-%s", workflowID)
	dockerMountOverrides := strings.Split(os.Getenv("BITRISE_DOCKER_MOUNT_OVERRIDES"), ",")
	// TODO: make sure the sleep command works across OS flavours
	runningContainer, err := cm.startContainer(container, containerName, dockerMountOverrides, "sleep infinity", "/bitrise/src")
	if err != nil {
		return nil, fmt.Errorf("start workflow container: %w", err)
	}
	cm.workflowContainers[workflowID] = runningContainer
	return runningContainer, nil
}

func (cm *ContainerManager) StartServiceContainer(service models.Container, workflowID string, serviceName string) (*RunningContainer, error) {
	// Naming the container other than the service name, can cause issues with network calls
	runningContainer, err := cm.startContainer(service, serviceName, []string{}, "", "")
	if err != nil {
		return nil, fmt.Errorf("start service container: %w", err)
	}
	cm.serviceContainers[workflowID] = append(cm.serviceContainers[workflowID], runningContainer)
	return runningContainer, nil
}

func (cm *ContainerManager) GetWorkflowContainer(workflowID string) *RunningContainer {
	return cm.workflowContainers[workflowID]
}

func (cm *ContainerManager) GetServiceContainers(workflowID string) []*RunningContainer {
	return cm.serviceContainers[workflowID]
}

func (cm *ContainerManager) DestroyAllContainers() error {
	for _, container := range cm.workflowContainers {
		if err := container.Destroy(); err != nil {
			return fmt.Errorf("destroy workflow container: %w", err)
		}
	}

	for _, containers := range cm.serviceContainers {
		for _, container := range containers {
			if err := container.Destroy(); err != nil {
				return fmt.Errorf("destroy service container: %w", err)
			}
		}
	}

	return nil
}

func (cm *ContainerManager) startContainer(container models.Container,
	name string,
	volumes []string,
	commandArgs, workingDir string,
) (*RunningContainer, error) {
	dockerRunArgs := []string{"run",
		"--platform", "linux/amd64",
		"--network=bitrise",
		"-d",
	}

	for _, o := range volumes {
		dockerRunArgs = append(dockerRunArgs, "-v", o)
	}

	for _, env := range container.Envs {
		for name, value := range env {
			dockerRunArgs = append(dockerRunArgs, "-e", fmt.Sprintf("%s=%s", name, value))
		}
	}

	for _, port := range container.Ports {
		dockerRunArgs = append(dockerRunArgs, "-p", port)
	}

	if workingDir != "" {
		dockerRunArgs = append(dockerRunArgs, "-w", workingDir)
	}

	dockerRunArgs = append(dockerRunArgs,
		fmt.Sprintf("--name=%s", name),
		container.Image,
	)

	// TODO: think about enabling setting this on the public UI as well
	if commandArgs != "" {
		commandArgsList := strings.Split(commandArgs, " ")
		dockerRunArgs = append(dockerRunArgs, commandArgsList...)
	}

	log.Infof("Running command: docker %s", strings.Join(dockerRunArgs, " "))
	out, err := command.New("docker", dockerRunArgs...).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		log.Errorf(out)
		return nil, fmt.Errorf("run docker container: %w", err)
	}

	runningContainer := &RunningContainer{
		Name: name,
	}
	return runningContainer, nil
}

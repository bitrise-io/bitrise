package docker

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/gofrs/uuid"
)

type RunningContainer struct {
	name        string // TODO refactor to use docker sdk, and return container ID instead of name
	initialized bool
}

type ContainerManager struct {
	logger    log.Logger
	container *RunningContainer
}

func NewContainerManager(logger log.Logger) *ContainerManager {
	return &ContainerManager{
		logger: logger,
		container: &RunningContainer{
			initialized: false,
		},
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

func (cm *ContainerManager) StartContainer(container models.Container) error {
	dockerMountOverrides := strings.Split(os.Getenv("BITRISE_DOCKER_MOUNT_OVERRIDES"), ",")
	dockerRunArgs := []string{"run",
		"--platform", "linux/amd64",
		"--network=bitrise",
		"-d",
	}

	for _, o := range dockerMountOverrides {
		dockerRunArgs = append(dockerRunArgs, "-v", o)
	}

	randomContainerSuffix := fmt.Sprintf("%s", uuid.Must(uuid.NewV4()))[0:8]
	containerName := fmt.Sprintf("workflow-container-%s", randomContainerSuffix)
	dockerRunArgs = append(dockerRunArgs,
		"-w", "/bitrise/src", // BitriseSourceDir
		fmt.Sprintf("--name=%s", containerName),
		container.Image,
		"sleep", "infinity",
	)

	out, err := command.New("docker", dockerRunArgs...).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		log.Errorf(out)
		return fmt.Errorf("run docker container: %w", err)
	}

	cm.container.initialized = true
	cm.container.name = containerName

	return nil
}

func (cm *ContainerManager) Destroy() error {
	if cm.container == nil || !cm.container.initialized {
		// TODO handle err
	}

	out, err := command.New("docker", "rm", "--force", cm.container.name).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		cm.logger.Errorf(out)
		return fmt.Errorf("remove docker container: %w", err)
	}

	cm.container.initialized = false
	cm.container.name = ""

	return nil
}

func (cm *ContainerManager) ExecuteCommandArgs(envs []string) []string {
	if cm.container == nil || !cm.container.initialized {
		// TODO handle err
	}

	args := []string{"exec"}

	for _, env := range envs {
		args = append(args, "-e", env)
	}

	args = append(args, cm.container.name)

	return args
}

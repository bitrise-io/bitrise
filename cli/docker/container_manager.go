package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bitrise-io/bitrise/v2/models/yml"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/v2/redactwriter"
)

type RunningContainer struct {
	ID   string
	Name string
}

type containerCreateOptions struct {
	name       string
	volumes    []string
	command    string
	workingDir string
	user       string
}

func (rc *RunningContainer) Destroy() error {
	_, err := command.New("docker", "rm", "--force", "--volumes", rc.Name).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
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

const bitriseNetwork = "bitrise"

type ContainerManager struct {
	logger             Logger
	workflowContainers map[string]*RunningContainer
	serviceContainers  map[string][]*RunningContainer
	client             *client.Client

	mu       sync.Mutex
	released bool
}

type Logger struct {
	logger  log.Logger
	secrets []string
}

func (dl *Logger) Infof(format string, args ...interface{}) {
	log := fmt.Sprintf(format, args...)
	redacted, _ := dl.Redact(log)
	dl.logger.Info(redacted)
}

func (dl *Logger) Errorf(format string, args ...interface{}) {
	log := fmt.Sprintf(format, args...)
	redacted, _ := dl.Redact(log)
	dl.logger.Error(redacted)
}

func (dl *Logger) Warnf(format string, args ...interface{}) {
	log := fmt.Sprintf(format, args...)
	redacted, _ := dl.Redact(log)
	dl.logger.Warn(redacted)
}

func (dl *Logger) Redact(s string) (string, error) {
	src := bytes.NewReader([]byte(s))
	dstBuf := new(bytes.Buffer)
	logger := log.NewUtilsLogAdapter()
	redactWriterDst := redactwriter.New(dl.secrets, dstBuf, &logger)

	if _, err := io.Copy(redactWriterDst, src); err != nil {
		return "", fmt.Errorf("failed to redact secrets, stream copy failed: %s", err)
	}
	if err := redactWriterDst.Close(); err != nil {
		return "", fmt.Errorf("failed to redact secrets, closing the stream failed: %s", err)
	}

	redactedValue := dstBuf.String()
	return redactedValue, nil
}

func NewContainerManager(logger log.Logger, secrets []string) *ContainerManager {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Warnf("Docker client failed to initialize (possibly running on unsupported environment): %s", err)
	}

	return &ContainerManager{
		logger: Logger{
			logger:  logger,
			secrets: secrets,
		},
		workflowContainers: make(map[string]*RunningContainer),
		serviceContainers:  make(map[string][]*RunningContainer),
		client:             dockerClient,
	}
}

func (cm *ContainerManager) StartContainerForStepGroup(
	container yml.Container,
	groupID string,
	envs map[string]string,
) (*RunningContainer, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := cm.login(container, envs); err != nil {
		log.Errorf("docker credentials provided, but the authentication failed.")
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	containerName := fmt.Sprintf("bitrise-workflow-%s", groupID)

	// TODO: handle default mounts if BITRISE_DOCKER_MOUNT_OVERRIDES is not provided
	dockerMountOverrides := strings.Split(os.Getenv("BITRISE_DOCKER_MOUNT_OVERRIDES"), ",")

	runningContainer, err := cm.runContainer(container, containerCreateOptions{
		name:       containerName,
		volumes:    dockerMountOverrides,
		command:    "sleep infinity",
		workingDir: "/bitrise/src",
		user:       "root",
	}, envs)

	// Even on failure we save the reference to make sure containers will be cleaned up
	if runningContainer != nil {
		cm.workflowContainers[groupID] = runningContainer
	}

	if err != nil {
		return runningContainer, fmt.Errorf("start workflow container: %w", err)
	}

	if err := cm.healthCheckContainer(context.Background(), runningContainer); err != nil {
		return runningContainer, fmt.Errorf("container health check: %w", err)
	}

	return runningContainer, nil
}

func (cm *ContainerManager) StartServiceContainersForStepGroup(
	services map[string]yml.Container,
	groupID string,
	envs map[string]string,
) ([]*RunningContainer, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var containers []*RunningContainer
	failedServices := make(map[string]error)

	for serviceName := range services {
		serviceContainer := services[serviceName]

		if err := cm.login(serviceContainer, envs); err != nil {
			failedServices[serviceName] = err
			continue
		}

		// Naming the container other than the service name, can cause issues with network calls
		runningContainer, err := cm.runContainer(serviceContainer, containerCreateOptions{
			name: serviceName,
		}, envs)
		if runningContainer != nil {
			containers = append(containers, runningContainer)
		}
		if err != nil {
			failedServices[serviceName] = err
		}
	}
	// Even on failure we save the references to make sure containers will be cleaned up
	cm.serviceContainers[groupID] = append(cm.serviceContainers[groupID], containers...)

	if len(failedServices) != 0 {
		errServices := fmt.Errorf("failed to start services")
		for serviceName, err := range failedServices {
			errServices = fmt.Errorf("%v: %w", errServices, err)
			cm.logger.Errorf("Failed to start service container (%s): %s", serviceName, err)
		}
		return containers, errServices
	}

	for _, container := range containers {
		if err := cm.healthCheckContainer(context.Background(), container); err != nil {
			return containers, fmt.Errorf("container health check: %w", err)
		}
	}

	return containers, nil
}

func (cm *ContainerManager) GetContainerForStepGroup(groupID string) *RunningContainer {
	return cm.workflowContainers[groupID]
}

func (cm *ContainerManager) GetServiceContainersForStepGroup(groupID string) []*RunningContainer {
	return cm.serviceContainers[groupID]
}

func (cm *ContainerManager) DestroyAllContainers() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.released = true
	for _, container := range cm.workflowContainers {
		cm.logger.Infof("ℹ️ Removing workflow container: %s", container.Name)
		if err := container.Destroy(); err != nil {
			return fmt.Errorf("destroy workflow container: %w", err)
		}
	}

	for _, containers := range cm.serviceContainers {
		for _, container := range containers {
			if container == nil {
				continue
			}
			cm.logger.Infof("Removing service container: %s", container.Name)
			if err := container.Destroy(); err != nil {
				return fmt.Errorf("destroy service container: %w", err)
			}
		}
	}

	return nil
}

func (cm *ContainerManager) login(container yml.Container, envs map[string]string) error {
	if container.Credentials.Username != "" && container.Credentials.Password != "" {
		cm.logger.Infof("ℹ️ Logging into docker registry: %s", container.Image)

		resolvedPassword := resolveEnvVariable(container.Credentials.Password, envs)
		resolvedUsername := resolveEnvVariable(container.Credentials.Username, envs)
		args := []string{"login", "--username", resolvedUsername, "--password", resolvedPassword}

		if container.Credentials.Server != "" {
			args = append(args, container.Credentials.Server)
		} else {
			args = append(args, container.Image)
		}

		cm.logger.Infof("ℹ️ Running command: docker %s", strings.Join(args, " "))

		out, err := command.New("docker", args...).RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			cm.logger.Errorf("%s", out)
			return fmt.Errorf("run docker login: %w", err)
		}
	}
	return nil
}

// We are not using the docker sdk for pull, start and create commands because:
//   - We want to make sure the end user can easily debug using the same docker command we issue
//     (hard to convert between sdk and cli api)
//   - We'd like to support options generically,
//     with the SDK we would need to parse the string ourselves to convert them properly to their own type
//   - SDK differs from the CLI in some cases, for example pulling from private registry requires the exact token
//     it can't automatically use the docker config
func (cm *ContainerManager) runContainer(
	container yml.Container,
	options containerCreateOptions,
	envs map[string]string,
) (*RunningContainer, error) {
	if cm.released {
		return nil, fmt.Errorf("container manager was released already")
	}

	if err := cm.ensureNetwork(); err != nil {
		return nil, fmt.Errorf("ensure bitrise docker network: %w", err)
	}

	cm.logger.Infof("ℹ️ Pulling docker image: %s", container.Image)
	err := cm.pullImageWithRetry(container)
	if err != nil {
		return nil, fmt.Errorf("pull docker image: %w", err)
	}
	cm.logger.Infof("✅ Docker image pulled: %s", container.Image)

	cm.logger.Infof("ℹ️ Creating docker container: %s", container.Image)
	err = cm.createContainer(container, options, envs)
	if err != nil {
		return nil, fmt.Errorf("create docker container: %w", err)
	}
	cm.logger.Infof("✅ Docker container created: %s", container.Image)

	cm.logger.Infof("ℹ️ Starting docker container: %s", container.Image)
	runningContainer, err := cm.startContainer(options)
	if err != nil {
		return runningContainer, fmt.Errorf("start docker container: %w", err)
	}
	cm.logger.Infof("✅ Container (%s) is running (%s)", runningContainer.Name, runningContainer.ID)

	return runningContainer, nil
}

func (cm *ContainerManager) startContainer(options containerCreateOptions) (*RunningContainer, error) {
	// At this point the container has been created, but it's not running yet
	// Even if we can't start it we need to return the container reference to make sure it will be cleaned up
	runningContainer := &RunningContainer{
		Name: options.name,
	}

	cm.logger.Infof("ℹ️ Running command: docker start %s", options.name)
	out, err := command.New("docker", "start", options.name).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		cm.logger.Errorf("%s", out)
		return runningContainer, fmt.Errorf("start docker container (%s): %w", options.name, err)
	}

	// We need to get the container ID to be able to check the health
	// This also serves as a validation that the container is running
	result, err := cm.getRunningContainer(context.Background(), options.name)
	if result != nil {
		runningContainer.ID = result.ID
	}
	if err != nil {
		return runningContainer, fmt.Errorf("container (%s) unable to start properly: %w", options.name, err)
	}
	return runningContainer, nil
}

func (cm *ContainerManager) createContainer(
	container yml.Container,
	options containerCreateOptions,
	envs map[string]string,
) error {
	dockerRunArgs := []string{
		"create",
		"--platform", "linux/amd64",
		fmt.Sprintf("--network=%s", bitriseNetwork),
	}

	for _, o := range options.volumes {
		dockerRunArgs = append(dockerRunArgs, "-v", o)
	}

	for _, env := range container.Envs {
		for name, value := range env {
			resolvedValue := resolveEnvVariable(fmt.Sprintf("%s", value), envs)
			dockerRunArgs = append(dockerRunArgs, "-e", fmt.Sprintf("%s=%s", name, resolvedValue))
		}
	}

	for _, port := range container.Ports {
		dockerRunArgs = append(dockerRunArgs, "-p", port)
	}

	if options.workingDir != "" {
		dockerRunArgs = append(dockerRunArgs, "-w", options.workingDir)
	}

	if options.user != "" {
		dockerRunArgs = append(dockerRunArgs, "-u", options.user)
	}

	if container.Options != "" {
		// This regex splits the string by spaces, but keeps quoted strings together
		// For example --health-cmd "redis-cli ping" will be split into: "--health-cmd", "redis-cli ping"
		r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
		result := r.FindAllString(container.Options, -1)

		// Remove quotes from the strings
		var options []string
		for _, result := range result {
			options = append(options, strings.ReplaceAll(result, "\"", ""))
		}

		dockerRunArgs = append(dockerRunArgs, options...)
	}

	dockerRunArgs = append(dockerRunArgs,
		fmt.Sprintf("--name=%s", options.name),
		container.Image,
	)

	if options.command != "" {
		commandArgsList := strings.Split(options.command, " ")
		dockerRunArgs = append(dockerRunArgs, commandArgsList...)
	}

	cm.logger.Infof("ℹ️ Running command: docker %s", strings.Join(dockerRunArgs, " "))

	out, err := command.New("docker", dockerRunArgs...).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		cm.logger.Errorf("%s", out)
		return fmt.Errorf("create container (%s): %w", options.name, err)
	}

	return nil
}

func (cm *ContainerManager) pullImageWithRetry(container yml.Container) error {
	pulling := true
	defer func() {
		pulling = false
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)
			if !pulling {
				return
			}
			cm.logger.Infof("⏳ Still pulling image: %s ...", container.Image)
		}
	}()

	// In case of pull error we retry 3 times
	var err error
	retries := 0
	for retries < 3 {
		err = cm.pullImage(container)
		if err != nil {
			cm.logger.Warnf("❌ Error during image pull: %s", err.Error())
			cm.logger.Warnf("⏳ Failed to pull image, retrying (retry %d/3) ... ", retries+1)
		} else {
			break
		}
		retries++
	}
	return err
}

func (cm *ContainerManager) pullImage(container yml.Container) error {
	images, err := cm.client.ImageList(context.Background(), image.ListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", container.Image)),
	})
	if err != nil {
		cm.logger.Warnf("Failed to check whether local image exist already, pulling...: %s", err.Error())
	} else if len(images) > 0 {
		cm.logger.Infof("ℹ️ Image (%s) already exists locally", container.Image)
		return nil
	}

	dockerRunArgs := []string{"pull", "--platform", "linux/amd64", container.Image}
	cm.logger.Infof("ℹ️ Running command: docker %s", strings.Join(dockerRunArgs, " "))
	out, err := command.New("docker", dockerRunArgs...).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		cm.logger.Errorf("%s", out)
		return fmt.Errorf("pull container (%s): %w", container.Image, err)
	}
	return nil
}

func (cm *ContainerManager) getRunningContainer(ctx context.Context, name string) (*container.Summary, error) {
	containers, err := cm.client.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", name)),
	})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	if len(containers) != 1 {
		return nil, fmt.Errorf("multiple containers with the same name found: %s", name)
	}

	if containers[0].State != "running" {
		logs, err := cm.client.ContainerLogs(ctx, containers[0].ID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		})
		if err != nil {
			return &containers[0], fmt.Errorf("container is not running: failed to get container logs: %w", err)
		}

		content, err := io.ReadAll(logs)
		if err != nil {
			return &containers[0], fmt.Errorf("container is not running: failed to read container logs: %w", err)
		}
		cm.logger.Errorf("Failed container (%s) logs:\n %s\n", name, string(content))
		return &containers[0], fmt.Errorf("container (%s) is not running", name)
	}
	return &containers[0], nil
}

func (cm *ContainerManager) healthCheckContainer(ctx context.Context, container *RunningContainer) error {
	inspect, err := cm.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return fmt.Errorf("inspect container (%s): %w", container.Name, err)
	}

	if inspect.State.Health == nil {
		cm.logger.Infof("✅ No healthcheck is defined for container (%s), assuming healthy...", container.Name)
		return nil
	}

	retries := 0
	for inspect.State.Health.Status != "healthy" {
		if inspect.State.Health.Status == "unhealthy" {
			cm.logger.Errorf("❌ Container (%s) is unhealthy...", container.Name)
			return fmt.Errorf("container (%s) is unhealthy", container.Name)
		}
		// this solution prefers quick retries at the beginning and constant for the rest
		sleep := 5
		if retries < 5 {
			sleep = retries + 1
		}
		time.Sleep(time.Duration(sleep) * time.Second)

		cm.logger.Infof("⏳ Waiting for container (%s) to be healthy... (retry: %ds)", container.Name, sleep)
		inspect, err = cm.client.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			return fmt.Errorf("inspect container (%s): %w", container.Name, err)
		}
		retries++

	}

	cm.logger.Infof("✅ Container (%s) is healthy...", container.Name)
	return nil
}

func (cm *ContainerManager) ensureNetwork() error {
	networks, err := cm.client.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", bitriseNetwork)),
	})
	if err != nil {
		return fmt.Errorf("list networks: %w", err)
	}

	if len(networks) > 0 {
		return nil
	}

	if _, err := cm.client.NetworkCreate(context.Background(), bitriseNetwork, network.CreateOptions{}); err != nil {
		return fmt.Errorf("create network: %w", err)
	}

	return nil
}

func resolveEnvVariable(value string, envs map[string]string) string {
	if strings.HasPrefix(value, "$") {
		if value, ok := envs[strings.TrimPrefix(value, "$")]; ok {
			return value
		}
	}
	return value
}

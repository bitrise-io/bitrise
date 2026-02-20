package models

import (
	"fmt"

	envmanModels "github.com/bitrise-io/envman/v2/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

type ContainerType string

const (
	ContainerTypeExecution ContainerType = "execution"
	ContainerTypeService   ContainerType = "service"
)

// Container model defines a Docker container configuration.
type Container struct {
	Type        ContainerType                       `json:"type,omitempty" yaml:"type,omitempty"`
	Image       string                              `json:"image,omitempty" yaml:"image,omitempty"`
	Credentials DockerCredentials                   `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	Ports       []string                            `json:"ports,omitempty" yaml:"ports,omitempty"`
	Envs        []envmanModels.EnvironmentItemModel `json:"envs,omitempty" yaml:"envs,omitempty"`
	Options     string                              `json:"options,omitempty" yaml:"options,omitempty"`
}

type DockerCredentials struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Server   string `json:"server,omitempty" yaml:"server,omitempty"`
}

type Containerisable struct {
	step       *stepmanModels.StepModel
	stepBundle *StepBundleListItemModel
}

func newContainerisableFromStep(step stepmanModels.StepModel) Containerisable {
	return Containerisable{step: &step}
}

func newContainerisableFromStepBundle(stepBundle StepBundleListItemModel) Containerisable {
	return Containerisable{stepBundle: &stepBundle}
}

func (c Containerisable) GetExecutionContainerConfig() (*ContainerConfig, error) {
	var executionContainer stepmanModels.ContainerReference
	if c.stepBundle != nil {
		executionContainer = c.stepBundle.ExecutionContainer
	} else if c.step != nil {
		executionContainer = c.step.ExecutionContainer
	}
	if executionContainer == nil {
		return nil, nil
	}
	ctrConfig, err := stepmanModels.GetContainerConfig(executionContainer)
	if err != nil {
		return nil, err
	}
	if ctrConfig == nil {
		return nil, nil
	}
	return &ContainerConfig{
		ContainerID: ctrConfig.ContainerID,
		Recreate:    ctrConfig.Recreate,
	}, nil
}

func (c Containerisable) GetServiceContainerConfigs() ([]ContainerConfig, error) {
	var serviceContainers []stepmanModels.ContainerReference
	if c.stepBundle != nil {
		serviceContainers = c.stepBundle.ServiceContainers
	} else if c.step != nil {
		serviceContainers = c.step.ServiceContainers
	}
	if serviceContainers == nil {
		return nil, nil
	}

	var containerConfigs []ContainerConfig
	for _, containerDef := range serviceContainers {
		ctrConfig, err := stepmanModels.GetContainerConfig(containerDef)
		if err != nil {
			return nil, err
		}
		if ctrConfig != nil {
			containerConfigs = append(containerConfigs, ContainerConfig{
				ContainerID: ctrConfig.ContainerID,
				Recreate:    ctrConfig.Recreate,
			})
		}
	}
	return containerConfigs, nil
}

type containerValidationContext struct {
	ExecutionContainers map[string]Container
	ServiceContainers   map[string]Container
}

func validateContainerReferences(containerisable Containerisable, validationContext containerValidationContext) error {
	executionContainerCfg, err := containerisable.GetExecutionContainerConfig()
	if err != nil {
		return fmt.Errorf("invalid execution container definition: %w", err)
	}
	if executionContainerCfg != nil {
		if _, ok := validationContext.ExecutionContainers[executionContainerCfg.ContainerID]; !ok {
			return fmt.Errorf("undefined execution container (%s) referenced", executionContainerCfg.ContainerID)
		}
	}

	serviceContainerIDs := map[string]bool{}
	serviceContainerCfgs, err := containerisable.GetServiceContainerConfigs()
	if err != nil {
		return fmt.Errorf("invalid service container definition: %w", err)
	}
	for _, serviceContainerCfg := range serviceContainerCfgs {
		if _, ok := serviceContainerIDs[serviceContainerCfg.ContainerID]; ok {
			return fmt.Errorf("duplicate service container reference: %s", serviceContainerCfg.ContainerID)
		}
		serviceContainerIDs[serviceContainerCfg.ContainerID] = true

		if _, ok := validationContext.ServiceContainers[serviceContainerCfg.ContainerID]; !ok {
			return fmt.Errorf("undefined service container (%s) referenced", serviceContainerCfg.ContainerID)
		}
	}
	return nil
}

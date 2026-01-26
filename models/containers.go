package models

import (
	"fmt"

	envmanModels "github.com/bitrise-io/envman/v2/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// Container model defines a Docker container configuration.
type Container struct {
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
	Step       *stepmanModels.StepModel
	StepBundle *StepBundleListItemModel
}

func newContainerisableFromStep(step stepmanModels.StepModel) Containerisable {
	return Containerisable{
		Step: &step,
	}
}

func newContainerisableFromStepBundle(stepBundle StepBundleListItemModel) Containerisable {
	return Containerisable{
		StepBundle: &stepBundle,
	}
}

func (step Containerisable) GetExecutionContainerConfig() (*ContainerConfig, error) {
	var executionContainer stepmanModels.ContainerReference
	if step.StepBundle != nil {
		executionContainer = step.StepBundle.ExecutionContainer
	} else if step.Step != nil {
		executionContainer = step.Step.ExecutionContainer
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

func (step Containerisable) GetServiceContainerConfigs() ([]ContainerConfig, error) {
	var serviceContainers []stepmanModels.ContainerReference
	if step.StepBundle != nil {
		serviceContainers = step.StepBundle.ServiceContainers
	} else if step.Step != nil {
		serviceContainers = step.Step.ServiceContainers
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

	serviceContainerCfgs, err := containerisable.GetServiceContainerConfigs()
	if err != nil {
		return fmt.Errorf("invalid service container definition: %w", err)
	}
	for _, serviceContainerCfg := range serviceContainerCfgs {
		if _, ok := validationContext.ServiceContainers[serviceContainerCfg.ContainerID]; !ok {
			return fmt.Errorf("undefined service container (%s) referenced", serviceContainerCfg.ContainerID)
		}
	}
	return nil
}

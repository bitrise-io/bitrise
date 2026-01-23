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
	if step.StepBundle != nil {
		return getContainerConfig(step.StepBundle.ExecutionContainer)
	} else if step.Step != nil {
		return getContainerConfig(step.Step.ExecutionContainer)
	}
	return nil, nil
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
		ctrConfig, err := getContainerConfig(containerDef)
		if err != nil {
			return nil, err
		}
		if ctrConfig != nil {
			containerConfigs = append(containerConfigs, *ctrConfig)
		}
	}
	return containerConfigs, nil
}

/*
Get ContainerConfig from container definition which can be either a string or a map.
Examples:
  - redis
  - postgres: { recreate: true }
*/
func getContainerConfig(container any) (*ContainerConfig, error) {
	if container == nil {
		return nil, nil
	}

	if ctrStr, ok := container.(string); ok {
		return &ContainerConfig{
			ContainerID: ctrStr,
			Recreate:    false,
		}, nil
	}

	var id string
	var recreate bool
	if ctrMap, ok := container.(map[any]any); ok {
		for k, v := range ctrMap {
			id, ok = k.(string)
			if !ok {
				return nil, fmt.Errorf("invalid container config ID type: %T", k)
			}

			if ctrCfg, ok := v.(map[any]any); ok {
				recreateVal, ok := ctrCfg["recreate"]
				if ok {
					recreate, ok = recreateVal.(bool)
					if !ok {
						return nil, fmt.Errorf("invalid recreate value type: %T", recreateVal)
					}
				}
			}

			break
		}

		return &ContainerConfig{
			ContainerID: id,
			Recreate:    recreate,
		}, nil
	}

	return nil, nil
}

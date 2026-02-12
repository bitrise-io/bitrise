package models

import "fmt"

// ContainerReference is a reference to a container. The value is either the container id, or a map with the container id and additional configuration.
//
//	service_containers:
//	- redis
//	- postgres:
//	    recreate: true
type ContainerReference any

type ContainerConfig struct {
	ContainerID string
	Recreate    bool
}

func GetContainerConfig(ref ContainerReference) (*ContainerConfig, error) {
	if ref == nil {
		return nil, nil
	}

	switch ref := ref.(type) {
	case string:
		return getContainerConfigFromString(ref), nil
	case map[any]any:
		return getContainerConfigFromMap(ref)
	case map[string]any:
		return getContainerConfigFromMap(ref)
	default:
		return nil, fmt.Errorf("invalid container config type: %T", ref)
	}
}

func getContainerConfigFromString(ctrStr string) *ContainerConfig {
	if ctrStr == "" {
		return nil
	}

	return &ContainerConfig{
		ContainerID: ctrStr,
		Recreate:    false,
	}
}

func getContainerConfigFromMap(ctr any) (*ContainerConfig, error) {
	ctrMap, ok := toStrMap(ctr)
	if !ok {
		return nil, fmt.Errorf("invalid container config map type: %T", ctr)
	}

	if len(ctrMap) != 1 {
		return nil, fmt.Errorf("invalid container config map length: %d", len(ctrMap))
	}

	var id string
	var recreate bool

	for k, v := range ctrMap {
		id = k

		ctrCfg, ok := toStrMap(v)
		if !ok {
			return nil, fmt.Errorf("invalid container config value type: %T", v)
		}
		if len(ctrCfg) > 1 {
			return nil, fmt.Errorf("invalid container config value map length: %d", len(ctrCfg))
		}

		if len(ctrCfg) == 1 {
			recreateVal, ok := ctrCfg["recreate"]
			if !ok {
				return nil, fmt.Errorf("missing recreate key in container config")
			}

			recreate, ok = recreateVal.(bool)
			if !ok {
				return nil, fmt.Errorf("invalid recreate value type: %T", recreateVal)
			}
		}
		break
	}

	return &ContainerConfig{
		ContainerID: id,
		Recreate:    recreate,
	}, nil
}

func toStrMap(v any) (map[string]any, bool) {
	if m, ok := v.(map[string]any); ok {
		return m, ok
	}
	m, ok := v.(map[any]any)
	if !ok {
		return nil, false
	}

	strMap := make(map[string]any, len(m))
	for k, v := range m {
		strKey, ok := k.(string)
		if !ok {
			return nil, false
		}
		strMap[strKey] = v
	}
	return strMap, true
}

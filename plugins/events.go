package plugins

import (
	"encoding/json"
	"fmt"

	"github.com/bitrise-io/bitrise/configs"
)

// TriggerEventName ...
type TriggerEventName string

const (
	// DidFinishRun ...
	DidFinishRun TriggerEventName = "DidFinishRun"
)

// TriggerEvent ...
func TriggerEvent(name TriggerEventName, payload interface{}) error {
	// Create plugin input
	bitriseVersion, err := configs.BitriseVersion()
	if err != nil {
		return err
	}

	pluginInputModel := map[string]interface{}{
		"version": bitriseVersion.String(),
		"payload": payload,
	}

	pluginInputBytes, err := json.Marshal(pluginInputModel)
	if err != nil {
		return err
	}
	pluginInputStr := string(pluginInputBytes)

	// Load plugins
	plugins, err := LoadPlugins(string(name))
	if err != nil {
		return err
	}

	// Run plugins
	for _, plugin := range plugins {
		if err := RunPlugin(plugin, []string{}, pluginInputStr); err != nil {
			return err
		}
	}

	return nil
}

// LoadPlugins ...
func LoadPlugins(eventName string) ([]Plugin, error) {
	routing, err := readPluginRouting()
	if err != nil {
		return []Plugin{}, err
	}

	pluginNames := []string{}
	for name, route := range routing.RouteMap {
		if route.TriggerEvent == eventName {
			pluginNames = append(pluginNames, name)
		}
	}

	plugins := []Plugin{}
	for _, name := range pluginNames {
		plugin, found, err := LoadPlugin(name)
		if err != nil {
			return []Plugin{}, err
		}
		if !found {
			return []Plugin{}, fmt.Errorf("Plugin (%s) exist in routing, but not found", name)
		}
		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

package toolprovider

import (
	"fmt"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

func getToolRequests(config models.BitriseDataModel, workflowID string) ([]provider.ToolRequest, error) {
	globalTools := config.Tools
	workflow, ok := config.Workflows[workflowID]
	if !ok {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}
	workflowTools := workflow.Tools

	mergedTools := globalTools
	if mergedTools == nil {
		mergedTools = workflowTools
	}
	for toolID, toolVersion := range workflowTools {
		if toolVersion == "unset" {
			delete(mergedTools, toolID)
		} else {
			mergedTools[toolID] = toolVersion
		}
	}

	var toolRequests []provider.ToolRequest
	for toolID, toolVersion := range mergedTools {
		v, strategy, err := ParseVersionString(toolVersion)
		if err != nil {
			return nil, fmt.Errorf("parse %s version: %w", toolID, err)
		}

		var pluginIdentifier *string
		if config.ToolConfig != nil && config.ToolConfig.ExtraPlugins != nil {
			if pluginID, ok := config.ToolConfig.ExtraPlugins[toolID]; ok {
				pluginIdentifier = &pluginID
			}
		}

		toolRequests = append(toolRequests, provider.ToolRequest{
			ToolName:           provider.ToolID(toolID),
			UnparsedVersion:    v,
			ResolutionStrategy: strategy,
			PluginURL:          pluginIdentifier,
		})
	}

	return toolRequests, nil
}

func defaultToolConfig() models.ToolConfigModel {
	return models.ToolConfigModel{
		Provider: "mise",
	}
}

func stackStatusDependentToolConfig() models.ToolConfigModel {
	isEdge := configs.IsEdgeStack()
	if isEdge {
		return models.ToolConfigModel{
			Provider: "mise",
		}
	}
	return models.ToolConfigModel{
		Provider:                       "mise",
		ExperimentalDisableFastInstall: true,
	}
}

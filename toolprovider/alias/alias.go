package alias

import "github.com/bitrise-io/bitrise/v2/toolprovider/provider"

var toolAliasMap = map[provider.ToolID]provider.ToolID{
	"go":   "golang",
	"node": "nodejs",
}

func GetCanonicalToolID(id provider.ToolID) provider.ToolID {
	if canonicalID, exists := toolAliasMap[id]; exists {
		return canonicalID
	}
	return id
}

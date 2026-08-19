package alias

import (
	"slices"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

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

// AliasesFor returns the aliases that resolve to the given canonical tool ID,
// sorted for stable output. Returns nil if the tool has no aliases.
func AliasesFor(canonical provider.ToolID) []provider.ToolID {
	var aliases []provider.ToolID
	for aliasID, canonicalID := range toolAliasMap {
		if canonicalID == canonical {
			aliases = append(aliases, aliasID)
		}
	}
	slices.Sort(aliases)
	return aliases
}

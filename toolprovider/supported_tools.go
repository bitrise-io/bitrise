package toolprovider

import (
	"slices"

	"github.com/bitrise-io/bitrise/v2/toolprovider/alias"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// ToolInfo describes a supported tool: its canonical name and any accepted
// aliases (e.g. "go" for "golang", "node" for "nodejs"), which the CLI treats
// as equivalent to the canonical name.
type ToolInfo struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases,omitempty"`
}

// canonicalToolNames is the source-of-truth list of supported tools, by canonical
// name (e.g. "golang" not "go", "nodejs" not "node").
// It is predominantly composed of mise core tools, but can include others, such
// as flutter.
func canonicalToolNames() []string {
	return []string{
		"bun",
		"deno",
		"elixir",
		"erlang",
		"flutter",
		"golang",
		"java",
		"nodejs",
		"python",
		"ruby",
		"rust",
		"swift",
		"zig",
	}
}

// SupportedTools returns the catalog of tools the CLI advertises and accepts
// for the "versions" and "list-tools" commands, each with its canonical name
// and accepted aliases.
func SupportedTools() []ToolInfo {
	names := canonicalToolNames()
	infos := make([]ToolInfo, 0, len(names))
	for _, name := range names {
		var aliases []string
		for _, aliasID := range alias.AliasesFor(provider.ToolID(name)) {
			aliases = append(aliases, string(aliasID))
		}
		infos = append(infos, ToolInfo{Name: name, Aliases: aliases})
	}
	return infos
}

// IsSupported reports whether the given tool name is supported, resolving
// aliases (e.g. "go", "node") to their canonical name first.
func IsSupported(toolName string) bool {
	canonical := string(alias.GetCanonicalToolID(provider.ToolID(toolName)))
	return slices.Contains(canonicalToolNames(), canonical)
}

// nonMiseCoreExceptions lists tools in SupportedTools that are intentionally
// not mise core tools (e.g. registry-only or third-party tools).
// Update this list when adding a supported tool that isn't a mise core tool.
var nonMiseCoreExceptions = []string{
	"flutter",
}

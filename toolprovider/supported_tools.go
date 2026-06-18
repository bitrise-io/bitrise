package toolprovider

// SupportedTools returns the list of tools the CLI advertises and accepts
// for the "versions" and "list-tools" commands.
// Uses canonical names (e.g. "golang" not "go", "nodejs" not "node").
// It is predominantly composed of mise core tools, but can include
// others, such as flutter.
func SupportedTools() []string {
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

// nonMiseCoreExceptions lists tools in SupportedTools that are intentionally
// not mise core tools (e.g. registry-only or third-party tools).
// Update this list when adding a supported tool that isn't a mise core tool.
var nonMiseCoreExceptions = []string{
	"flutter",
}

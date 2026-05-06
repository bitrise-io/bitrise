package toolprovider

// SupportedTools returns the list of tools the CLI advertises and accepts
// for the "versions" and "list-tools" commands.
// Uses canonical names (e.g. "golang" not "go", "nodejs" not "node").
// It is predominently composed of mise core tools, but can include
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

package cmdutil

// OSDisplayName maps a backend OS token to a human-friendly display name
// (e.g. "macos" → "macOS", "linux" → "Linux"); unknown values pass through
// unchanged. Shared by the RDE stack list and the `rde claude` stack picker so
// the capitalization stays consistent. This is display-only — the raw token is
// still what `--output json` and the wire carry.
func OSDisplayName(os string) string {
	switch os {
	case "macos":
		return "macOS"
	case "linux":
		return "Linux"
	default:
		return os
	}
}

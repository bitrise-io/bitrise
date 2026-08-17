package cmdutil

import "testing"

func TestOSDisplayName(t *testing.T) {
	for in, want := range map[string]string{
		"macos":   "macOS",
		"linux":   "Linux",
		"windows": "windows", // unknown → unchanged
		"":        "",
	} {
		if got := OSDisplayName(in); got != want {
			t.Errorf("OSDisplayName(%q) = %q, want %q", in, got, want)
		}
	}
}

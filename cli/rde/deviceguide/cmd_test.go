package deviceguide

import (
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/cli/cmdtest"
	"github.com/bitrise-io/bitrise/v2/output"
)

func TestDeviceGuide(t *testing.T) {
	for _, tc := range []struct{ arg, want string }{
		{"", "PREVIEW_DEVICE_STATE_READY"},
		{"ios", "serve-sim"},
		{"android", "uiautomator"},
	} {
		var args []string
		if tc.arg != "" {
			args = []string{tc.arg}
		}
		stdout, _, err := cmdtest.Run(t, NewCmd(), cmdtest.Opts{Args: args})
		if err != nil {
			t.Fatalf("%q: %v", tc.arg, err)
		}
		if !strings.Contains(stdout, tc.want) {
			t.Errorf("%q: output missing %q", tc.arg, tc.want)
		}
	}
	if _, _, err := cmdtest.Run(t, NewCmd(), cmdtest.Opts{Args: []string{"tvos"}}); err == nil {
		t.Errorf("unknown platform must error")
	}
}

// TestDeviceGuide_NoMirrorHeader: the mirror files carry a maintainer-only
// HTML comment on line 1; what the command prints must start at the guide's
// title, not at a note about syncing.
func TestDeviceGuide_NoMirrorHeader(t *testing.T) {
	for _, args := range [][]string{nil, {"ios"}, {"android"}} {
		stdout, _, err := cmdtest.Run(t, NewCmd(), cmdtest.Opts{Args: args})
		if err != nil {
			t.Fatalf("args %v: %v", args, err)
		}
		if strings.HasPrefix(stdout, "<!--") || strings.Contains(stdout, "do not edit here") {
			t.Errorf("args %v: output still carries the mirror header:\n%.120s", args, stdout)
		}
		if !strings.HasPrefix(stdout, "# ") {
			t.Errorf("args %v: output must start at the guide title, got %.80q", args, stdout)
		}
	}
	for _, tc := range []struct{ in, want string }{
		{"<!-- note -->\n\n# Title\n", "# Title\n"},
		{"# Title\n", "# Title\n"},
		{"<!-- unterminated\n# Title\n", "<!-- unterminated\n# Title\n"},
	} {
		if got := stripMirrorHeader(tc.in); got != tc.want {
			t.Errorf("stripMirrorHeader(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestDeviceGuide_RejectsNonRawFormat: the inherited --format flag has no
// structured shape for a Markdown guide, so the command must refuse any
// non-raw format instead of printing raw Markdown where a caller expects a
// JSON/YAML object — matching `session logs`'s rejection.
func TestDeviceGuide_RejectsNonRawFormat(t *testing.T) {
	for _, format := range []string{output.FormatJSON, output.FormatYML} {
		for _, args := range [][]string{nil, {"ios"}} {
			stdout, _, err := cmdtest.Run(t, NewCmd(), cmdtest.Opts{Args: args, Format: format})
			if err == nil || !strings.Contains(err.Error(), format) {
				t.Errorf("format %s, args %v: error = %v, want a --format %s rejection", format, args, err, format)
			}
			// cobra echoes the usage on error; the guide body itself must not
			// have been printed.
			if strings.Contains(stdout, "PREVIEW_DEVICE_STATE_READY") || strings.Contains(stdout, "serve-sim") {
				t.Errorf("format %s, args %v: guide must not be printed when rejected:\n%s", format, args, stdout)
			}
		}
	}
}

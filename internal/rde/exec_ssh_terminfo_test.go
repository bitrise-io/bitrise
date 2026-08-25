package rde

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

// TestResolveRemoteTerm covers the branches that don't require a live SSH
// connection: an unset $TERM and ubiquitous terminals return without touching
// the network, and a $TERM whose terminfo can't be dumped locally falls back to
// the default before any session is opened (so a nil client is never reached).
func TestResolveRemoteTerm(t *testing.T) {
	tests := []struct {
		name      string
		localTerm string
		want      string
	}{
		{name: "unset falls back to default", localTerm: "", want: defaultTermType},
		{name: "ubiquitous xterm-256color passes through", localTerm: "xterm-256color", want: "xterm-256color"},
		{name: "ubiquitous xterm passes through", localTerm: "xterm", want: "xterm"},
		{name: "ubiquitous vt100 passes through", localTerm: "vt100", want: "vt100"},
		// A term that no terminfo database knows: infocmp -x fails, so
		// forwardTerminfo returns before c.client is dereferenced and the
		// caller falls back to the default.
		{name: "unknown term falls back to default", localTerm: "no-such-term-zzz-12345", want: defaultTermType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Nil client is safe: every case here returns before any session
			// is opened. A regression that forwards a ubiquitous/unknown term
			// would instead nil-panic, failing the test loudly.
			c := &sshClient{}
			if got := c.resolveRemoteTerm(context.Background(), tt.localTerm); got != tt.want {
				t.Errorf("resolveRemoteTerm(%q) = %q, want %q", tt.localTerm, got, tt.want)
			}
		})
	}
}

func TestLocalTerminfoSource(t *testing.T) {
	if _, err := exec.LookPath("infocmp"); err != nil {
		t.Skip("infocmp not available")
	}

	t.Run("known term yields non-empty source", func(t *testing.T) {
		src, err := localTerminfoSource(context.Background(), "xterm-256color")
		if err != nil {
			t.Fatalf("localTerminfoSource: unexpected error: %v", err)
		}
		if len(src) == 0 {
			t.Fatal("localTerminfoSource: empty source for xterm-256color")
		}
		// infocmp prints the entry name(s) in the header line; a usable
		// terminfo source must reference the term it describes.
		if !strings.Contains(string(src), "xterm-256color") {
			t.Errorf("localTerminfoSource: source does not mention xterm-256color:\n%s", src)
		}
	})

	t.Run("unknown term errors", func(t *testing.T) {
		if _, err := localTerminfoSource(context.Background(), "no-such-term-zzz-12345"); err == nil {
			t.Error("localTerminfoSource: expected error for unknown term, got nil")
		}
	})
}

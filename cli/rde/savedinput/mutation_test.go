package savedinput

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/cli/cmdtest"
	"github.com/bitrise-io/bitrise/v2/output"
)

func TestUpdateCmd_RequiresAField(t *testing.T) {
	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{RDEAPIBaseURL: "http://unused", Args: []string{"sv-1"}})
	if err == nil || !strings.Contains(err.Error(), "--value") {
		t.Errorf("error = %v, want at-least-one-field error", err)
	}
}

func TestUpdateCmd_ValueOnlyOmitsSecret(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/saved-inputs/sv-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"savedInput":{"id":"sv-1","key":"repo","value":"new"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, Args: []string{"sv-1", "--value", "new"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotBody["value"] != "new" {
		t.Errorf("value = %v, want new", gotBody["value"])
	}
	// --secret wasn't passed, so isSecret must be omitted (not reset to false).
	if _, ok := gotBody["isSecret"]; ok {
		t.Errorf("isSecret should be omitted, body=%v", gotBody)
	}
}

func TestUpdateCmd_ValueStdin(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"savedInput":{"id":"sv-1","key":"repo","isSecret":true,"value":"***"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL: srv.URL,
		Args:          []string{"sv-1", "--value-stdin"},
		Stdin:         "new-secret\n",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotBody["value"] != "new-secret" {
		t.Errorf("value = %v, want new-secret (read from stdin)", gotBody["value"])
	}
}

// `--value-stdin < /dev/null` used to blank the stored value silently. Empty
// piped input is a mistake (a missing file, a failed upstream command); --value
// "" stays the deliberate way to clear.
func TestUpdateCmd_ValueStdinRejectsEmpty(t *testing.T) {
	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL: "http://unused",
		Args:          []string{"sv-1", "--value-stdin"},
		Stdin:         "",
	})
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("error = %v, want an empty-value error", err)
	}
}

// A piped value is trimmed the same way --value is (not at all beyond the line
// terminator), so leading/trailing spaces that are part of the secret survive.
func TestUpdateCmd_ValueStdinPreservesSurroundingSpace(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"savedInput":{"id":"sv-1","key":"repo"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL: srv.URL,
		Args:          []string{"sv-1", "--value-stdin"},
		Stdin:         "  pad ded  \n",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotBody["value"] != "  pad ded  " {
		t.Errorf("value = %q, want %q (only the newline trimmed)", gotBody["value"], "  pad ded  ")
	}
}

func TestUpdateCmd_SecretFlagSendsBool(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"savedInput":{"id":"sv-1","key":"repo","isSecret":true,"value":"***"}}`)
	}))
	defer srv.Close()

	// --secret without --value: only the secret flag is patched.
	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, Args: []string{"sv-1", "--secret"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotBody["isSecret"] != true {
		t.Errorf("isSecret = %v, want true", gotBody["isSecret"])
	}
	if _, ok := gotBody["value"]; ok {
		t.Errorf("value should be omitted, body=%v", gotBody)
	}
}

func TestUpdateCmd_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"savedInput":{"id":"sv-1","key":"repo","value":"new"}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL: srv.URL,
		Args:          []string{"sv-1", "--value", "new"},
		Format:        output.FormatJSON,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if got["id"] != "sv-1" || got["value"] != "new" {
		t.Errorf("unexpected JSON: %v", got)
	}
}

func TestUpdateCmd_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{RDEAPIBaseURL: "http://unused", Args: []string{"--value", "x"}})
	if err == nil {
		t.Fatal("expected error when SAVED_INPUT_ID is missing")
	}
}

func TestDeleteCmd_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/saved-inputs/sv-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	stdout, stderr, err := cmdtest.Run(t, newDeleteCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, Args: []string{"sv-1"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stderr, "Deleted saved input sv-1") {
		t.Errorf("stderr missing confirmation: %q", stderr)
	}
	if stdout != "" {
		t.Errorf("stdout should be empty for delete, got: %q", stdout)
	}
}

func TestDeleteCmd_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, newDeleteCmd(), cmdtest.Opts{RDEAPIBaseURL: "http://unused"})
	if err == nil {
		t.Fatal("expected error when SAVED_INPUT_ID is missing")
	}
}

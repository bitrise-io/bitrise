package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPath_HonorsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
	got, err := Path()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/custom/xdg", "bitrise", "cli", "config.yml"), got)
}

func TestPath_FallsBackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	got, err := Path()
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(got, filepath.Join(".config", "bitrise", "cli", "config.yml")))
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	want := Config{
		SetupVersion:           "2.1.0",
		LastCLIUpdateCheck:     time.Now().UTC().Truncate(time.Second),
		LastPluginUpdateChecks: map[string]time.Time{"init": time.Now().UTC().Truncate(time.Second)},
	}
	require.NoError(t, Save(want))

	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, want.SetupVersion, got.SetupVersion)
	assert.True(t, got.LastCLIUpdateCheck.Equal(want.LastCLIUpdateCheck))
	assert.Len(t, got.LastPluginUpdateChecks, 1)
	assert.True(t, got.LastPluginUpdateChecks["init"].Equal(want.LastPluginUpdateChecks["init"]))

	if runtime.GOOS != "windows" {
		p := filepath.Join(dir, "bitrise", "cli", "config.yml")
		info, err := os.Stat(p)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}
}

func TestLoad_MissingFileIsZeroValue(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, Config{}, got)
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise", "cli"), 0o700))
	bad := filepath.Join(dir, "bitrise", "cli", "config.yml")
	require.NoError(t, os.WriteFile(bad, []byte("this: is :: not yaml"), 0o600))
	_, err := Load()
	assert.Error(t, err)
}

func TestLoad_FallsBackToPredecessorConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	predecessorDir := filepath.Join(dir, "bitrise")
	require.NoError(t, os.MkdirAll(predecessorDir, 0o700))
	predecessorFile := filepath.Join(predecessorDir, "config.yaml")
	require.NoError(t, os.WriteFile(predecessorFile, []byte(""+
		"app_slug: my-app-slug\n"+
		"default_workspace_slug: my-workspace\n"+
		"theme: dark\n"+
		"output: json\n"+
		"api_base_url: https://api.example.test\n"), 0o600))

	fallbackAnnounceOnce = sync.Once{}
	var announced strings.Builder
	origWriter := fallbackWriter
	fallbackWriter = &announced
	t.Cleanup(func() { fallbackWriter = origWriter })

	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "my-app-slug", got.AppID, "app_slug must alias to app_id")
	assert.Equal(t, "my-workspace", got.DefaultWorkspaceID, "default_workspace_slug must alias to default_workspace_id")
	assert.Equal(t, "dark", got.Theme)
	assert.Equal(t, "json", got.Output)
	assert.Equal(t, "https://api.example.test", got.APIBaseURL)

	_, err = os.ReadFile(predecessorFile)
	require.NoError(t, err, "the predecessor file must survive Load — it's read-only")
	_, statErr := os.Stat(filepath.Join(dir, "bitrise", "cli", "config.yml"))
	assert.True(t, os.IsNotExist(statErr), "Load must not write the new config file as a side effect")

	assert.Contains(t, announced.String(), predecessorFile, "the fallback should be announced once")
}

func TestLoad_PrefersNewConfigWhenPresent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bitrise", "config.yaml"), []byte("theme: dark\n"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise", "cli"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bitrise", "cli", "config.yml"), []byte("theme: light\n"), 0o600))

	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "light", got.Theme, "the new file must win once it exists, regardless of the predecessor file")
}

func TestActivePath_PredecessorWhileFallbackLive(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise"), 0o700))
	predecessorFile := filepath.Join(dir, "bitrise", "config.yaml")
	require.NoError(t, os.WriteFile(predecessorFile, []byte("theme: dark\n"), 0o600))

	got, err := ActivePath()
	require.NoError(t, err)
	assert.Equal(t, predecessorFile, got)
}

func TestActivePath_NewPathOnceWritten(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bitrise", "config.yaml"), []byte("theme: dark\n"), 0o600))
	require.NoError(t, Save(Config{Theme: "light"}))

	got, err := ActivePath()
	require.NoError(t, err)
	newPath, err := Path()
	require.NoError(t, err)
	assert.Equal(t, newPath, got, "once the new file has been written, ActivePath must name it even though the predecessor file still exists")
}

func TestActivePath_NewPathWhenNeitherExists(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	got, err := ActivePath()
	require.NoError(t, err)
	newPath, err := Path()
	require.NoError(t, err)
	assert.Equal(t, newPath, got)
}

func TestLoadDir_FindsAncestorFile(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b", "c")
	require.NoError(t, os.MkdirAll(deep, 0o755)) //nolint:gosec // test-only tempdir, perms don't matter

	cfgPath := filepath.Join(root, "a", DirFileName)
	require.NoError(t, os.WriteFile(cfgPath, []byte("setup_version: 1.2.3\n"), 0o644)) //nolint:gosec // test-only tempfile

	got, found, err := loadDirFrom(deep)
	require.NoError(t, err)
	assert.Equal(t, cfgPath, found)
	assert.Equal(t, "1.2.3", got.SetupVersion)
}

func TestLoadDir_NoFileReturnsZero(t *testing.T) {
	got, found, err := loadDirFrom(t.TempDir())
	require.NoError(t, err)
	assert.Empty(t, found)
	assert.Equal(t, Config{}, got)
}

// TestSaveYAML_ConcurrentWritesDontCorrupt writes to the same path from many
// goroutines at once. Each writer used to share a fixed ".tmp" name, so one
// writer's rename could race another's still-in-progress write; the fix
// gives each writer its own temp file. The file left behind must always be
// one full, valid write — never a half-written blend of two.
func TestSaveYAML_ConcurrentWritesDontCorrupt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "concurrent.yaml")

	type payload struct {
		Value string `yaml:"value"`
	}

	const writers = 20
	var wg sync.WaitGroup
	errs := make([]error, writers)
	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = SaveYAML(path, payload{Value: strings.Repeat("x", 100)})
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "writer %d", i)
	}

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var got payload
	require.NoError(t, yaml.Unmarshal(data, &got))
	assert.Equal(t, strings.Repeat("x", 100), got.Value)
}

// TestWriteAtomic_SweepsStaleTmpSiblings simulates a crash between a
// previous writeAtomic's CreateTemp and Rename: a *.tmp file matching its
// naming pattern, old enough to no longer be a plausible in-flight write, is
// left on disk. The next write must clean it up rather than leaking it
// forever.
func TestWriteAtomic_SweepsStaleTmpSiblings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	stale := path + ".deadbeef.tmp"
	require.NoError(t, os.WriteFile(stale, []byte("leftover"), 0o600))
	oldTime := time.Now().Add(-2 * staleTmpAge)
	require.NoError(t, os.Chtimes(stale, oldTime, oldTime))

	require.NoError(t, writeAtomic(path, []byte("value: 1\n")))

	_, err := os.Stat(stale)
	assert.True(t, os.IsNotExist(err), "stale tmp file should have been swept")
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "value: 1\n", string(got))
}

// TestWriteAtomic_KeepsFreshTmpSiblings guards against the sweep in
// writeAtomic mistaking a concurrent writer's still-in-flight temp file
// (same *.tmp glob pattern, just created) for a crash leftover and deleting
// it out from under that writer.
func TestWriteAtomic_KeepsFreshTmpSiblings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	fresh := path + ".inflight.tmp"
	require.NoError(t, os.WriteFile(fresh, []byte("still being written"), 0o600))

	require.NoError(t, writeAtomic(path, []byte("value: 1\n")))

	got, err := os.ReadFile(fresh)
	require.NoError(t, err, "fresh tmp file should not have been swept")
	assert.Equal(t, "still being written", string(got))
}

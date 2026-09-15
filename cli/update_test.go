package cli

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/bitrise-io/bitrise/v2/version"
	ver "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)

func TestUpdaterPublishedVersions(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want []string
	}{
		{
			name: "sorted by version, not by the order the API returns them",
			tags: []string{"2.44.0", "2.45.0", "2.43.4"},
			want: []string{"2.43.4", "2.44.0", "2.45.0"},
		},
		{
			name: "a higher major is not special",
			tags: []string{"3.0.0", "2.45.0"},
			want: []string{"2.45.0", "3.0.0"},
		},
		{
			name: "pre-releases are left out",
			tags: []string{"2.45.0", "3.0.0-rc.1", "3.0.0-beta"},
			want: []string{"2.45.0"},
		},
		{
			name: "tags that are not versions are left out",
			tags: []string{"2.45.0", "not-a-version"},
			want: []string{"2.45.0"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			u := testUpdater(t, test.tags...)

			versions, err := u.publishedVersions()

			require.NoError(t, err)
			require.Equal(t, test.want, versionStrings(versions))
		})
	}
}

func TestUpdaterPublishedVersionsAsksForAFullPage(t *testing.T) {
	var gotPerPage string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPerPage = r.URL.Query().Get("per_page")
		if _, err := w.Write([]byte("[]")); err != nil {
			t.Errorf("failed to write tags response: %s", err)
		}
	}))
	t.Cleanup(server.Close)

	u := newUpdater()
	u.tagsURL = server.URL
	u.client = server.Client()

	_, err := u.publishedVersions()

	require.NoError(t, err)
	require.Equal(t, "100", gotPerPage)
}

func TestNewAvailableUpdates(t *testing.T) {
	tests := []struct {
		name          string
		current       string
		published     []string
		wantSameMajor string
		wantNewMajor  string
	}{
		{
			name:          "newer patch within the running major",
			current:       "2.44.0",
			published:     []string{"2.44.0", "2.45.0"},
			wantSameMajor: "2.45.0",
		},
		{
			name:      "up to date",
			current:   "2.45.0",
			published: []string{"2.44.0", "2.45.0"},
		},
		{
			name:      "running a version newer than the published one",
			current:   "2.46.0",
			published: []string{"2.45.0"},
		},
		{
			name:          "a new major is reported on its own, not as the update to install",
			current:       "2.45.0",
			published:     []string{"2.45.0", "3.0.0"},
			wantNewMajor:  "3.0.0",
			wantSameMajor: "",
		},
		{
			name:          "both a patch and a new major",
			current:       "2.44.0",
			published:     []string{"2.45.0", "3.0.0"},
			wantSameMajor: "2.45.0",
			wantNewMajor:  "3.0.0",
		},
		{
			name:          "the highest of each major wins",
			current:       "2.43.0",
			published:     []string{"2.44.0", "2.45.0", "3.0.0", "3.1.0"},
			wantSameMajor: "2.45.0",
			wantNewMajor:  "3.1.0",
		},
		{
			name:      "already on the newest major",
			current:   "3.0.0",
			published: []string{"2.45.0", "3.0.0"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updates := newAvailableUpdates(parseVersion(t, test.current), parseVersions(t, test.published))

			require.Equal(t, test.wantSameMajor, versionString(updates.sameMajor))
			require.Equal(t, test.wantNewMajor, versionString(updates.newMajor))
		})
	}
}

func TestUpdaterAvailableUpdates(t *testing.T) {
	t.Run("resolves the highest version of each major", func(t *testing.T) {
		setVersion(t, "2.44.0")
		u := testUpdater(t, "2.45.0", "2.44.0", "3.0.0", "3.0.1-rc.1")

		updates, err := u.availableUpdates()

		require.NoError(t, err)
		require.Equal(t, "2.45.0", versionString(updates.sameMajor))
		require.Equal(t, "3.0.0", versionString(updates.newMajor))
	})

	t.Run("dev build has no comparable version", func(t *testing.T) {
		setVersion(t, "dev")
		u := testUpdater(t, "2.45.0", "3.0.0")

		updates, err := u.availableUpdates()

		require.NoError(t, err)
		require.Nil(t, updates.sameMajor)
		require.Nil(t, updates.newMajor)
	})
}

func TestMajorUpdateCommand(t *testing.T) {
	newMajor := parseVersion(t, "3.0.0")

	require.Equal(t, "brew upgrade bitrise", majorUpdateCommand(newMajor, true))
	require.Equal(t, "bitrise update --version 3.0.0", majorUpdateCommand(newMajor, false))
}

func TestAssetName(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{goos: "darwin", goarch: "amd64", want: "bitrise-Darwin-x86_64"},
		{goos: "darwin", goarch: "arm64", want: "bitrise-Darwin-arm64"},
		{goos: "linux", goarch: "amd64", want: "bitrise-Linux-x86_64"},
		{goos: "linux", goarch: "arm64", want: "bitrise-Linux-arm64"},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			asset, err := assetName(test.goos, test.goarch)

			require.NoError(t, err)
			require.Equal(t, test.want, asset)
		})
	}
}

func TestAssetNameOfAnArchitectureWeDoNotPublish(t *testing.T) {
	_, err := assetName("linux", "386")

	require.EqualError(t, err, "no Bitrise CLI release is published for linux/386")
}

func TestUpdaterBinaryURL(t *testing.T) {
	url, err := newUpdater().binaryURL("2.45.0", "darwin", "arm64")

	require.NoError(t, err)
	require.Equal(t, "https://github.com/bitrise-io/bitrise/releases/download/v2.45.0/bitrise-Darwin-arm64", url)
}

func TestReplaceBinary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bitrise")
	require.NoError(t, os.WriteFile(path, []byte("old binary"), 0755))

	require.NoError(t, replaceBinary(path, strings.NewReader("new binary")))

	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "new binary", string(contents))

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0755), info.Mode().Perm())

	requireOnlyFile(t, dir, "bitrise")
}

func TestReplaceBinaryKeepsTheOriginalWhenTheDownloadFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bitrise")
	require.NoError(t, os.WriteFile(path, []byte("old binary"), 0755))

	err := replaceBinary(path, iotest.ErrReader(errors.New("connection reset")))

	require.ErrorContains(t, err, "connection reset")

	contents, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	require.Equal(t, "old binary", string(contents))

	requireOnlyFile(t, dir, "bitrise")
}

// Tags are served in the given order, and prefixed with "v", the way the GitHub
// tags API returns them.
func testUpdater(t *testing.T, tags ...string) updater {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		payload := make([]map[string]string, 0, len(tags))
		for _, tag := range tags {
			payload = append(payload, map[string]string{"name": "v" + tag})
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Errorf("failed to write tags response: %s", err)
		}
	}))
	t.Cleanup(server.Close)

	u := newUpdater()
	u.tagsURL = server.URL
	u.client = server.Client()
	u.isBrewInstall = func() (bool, error) { return false, nil }

	return u
}

func requireOnlyFile(t *testing.T, dir, name string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	require.Equal(t, []string{name}, names)
}

func setVersion(t *testing.T, v string) {
	t.Helper()

	original := version.VERSION
	version.VERSION = v
	t.Cleanup(func() { version.VERSION = original })
}

func parseVersion(t *testing.T, v string) *ver.Version {
	t.Helper()

	parsed, err := ver.NewVersion(v)
	require.NoError(t, err)

	return parsed
}

func parseVersions(t *testing.T, versions []string) []*ver.Version {
	t.Helper()

	parsed := make([]*ver.Version, 0, len(versions))
	for _, v := range versions {
		parsed = append(parsed, parseVersion(t, v))
	}

	return parsed
}

func versionStrings(versions []*ver.Version) []string {
	strs := make([]string, 0, len(versions))
	for _, v := range versions {
		strs = append(strs, v.String())
	}

	return strs
}

func versionString(v *ver.Version) string {
	if v == nil {
		return ""
	}

	return v.String()
}

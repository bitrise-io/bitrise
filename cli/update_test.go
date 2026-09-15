package cli

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

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

func TestUpdaterPublishedVersionsReadsEveryPage(t *testing.T) {
	// The newest version sits on the last page, where a single-page read would
	// miss it.
	tags := make([]string, 0, tagsPerPage*2+1)
	for i := range tagsPerPage * 2 {
		tags = append(tags, fmt.Sprintf("1.0.%d", i))
	}
	tags = append(tags, "2.45.0")

	u := testUpdater(t, tags...)

	versions, err := u.publishedVersions()

	require.NoError(t, err)
	require.Len(t, versions, len(tags))
	require.Equal(t, "2.45.0", versions[len(versions)-1].String())
}

func TestUpdaterPublishedVersionsStopsAtThePageCap(t *testing.T) {
	var requestedPages int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestedPages++

		page := make([]map[string]string, 0, tagsPerPage)
		for i := range tagsPerPage {
			page = append(page, map[string]string{"name": fmt.Sprintf("v1.0.%d", i)})
		}
		if err := json.NewEncoder(w).Encode(page); err != nil {
			t.Errorf("failed to write tags response: %s", err)
		}
	}))
	t.Cleanup(server.Close)

	u := newUpdater()
	u.tagsURL = server.URL
	u.client = server.Client()

	_, err := u.publishedVersions()

	require.NoError(t, err)
	require.Equal(t, maxTagPages, requestedPages)
}

func TestUpdaterTagPageParameters(t *testing.T) {
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		if _, err := w.Write([]byte("[]")); err != nil {
			t.Errorf("failed to write tags response: %s", err)
		}
	}))
	t.Cleanup(server.Close)

	u := newUpdater()
	u.tagsURL = server.URL
	u.client = server.Client()

	_, err := u.tagPage(3)

	require.NoError(t, err)
	require.Equal(t, "100", gotQuery.Get("per_page"))
	require.Equal(t, "3", gotQuery.Get("page"))
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
	require.Equal(t, "bitrise-Darwin-x86_64", assetName("darwin"))
	require.Equal(t, "bitrise-Linux-x86_64", assetName("linux"))
}

func TestUpdaterBinaryURL(t *testing.T) {
	require.Equal(t,
		"https://github.com/bitrise-io/bitrise/releases/download/v2.45.0/bitrise-Darwin-x86_64",
		newUpdater().binaryURL("2.45.0", "darwin"))
}

// Tags are served in the given order, and prefixed with "v", the way the GitHub
// tags API returns them.
func testUpdater(t *testing.T, tags ...string) updater {
	t.Helper()

	server := httptest.NewServer(tagsHandler(t, tags))
	t.Cleanup(server.Close)

	u := newUpdater()
	u.tagsURL = server.URL
	u.client = server.Client()
	u.isBrewInstall = func() (bool, error) { return false, nil }

	return u
}

// Serves the tags the way the GitHub tags API does: prefixed with "v", split
// into pages, in the order given.
func tagsHandler(t *testing.T, tags []string) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		page, err := strconv.Atoi(cmp.Or(r.URL.Query().Get("page"), "1"))
		if err != nil {
			t.Errorf("invalid page parameter: %s", err)
			return
		}

		start := min((page-1)*tagsPerPage, len(tags))
		end := min(start+tagsPerPage, len(tags))

		payload := make([]map[string]string, 0, end-start)
		for _, tag := range tags[start:end] {
			payload = append(payload, map[string]string{"name": "v" + tag})
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Errorf("failed to write tags response: %s", err)
		}
	}
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

package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/stretchr/testify/require"
)

func TestUpdaterLatestTag(t *testing.T) {
	u := testUpdater(t, "2.45.0", "2.44.0", "2.43.4")

	latest, err := u.latestTag()

	require.NoError(t, err)
	require.Equal(t, "2.45.0", latest.String())
}

func TestUpdaterLatestTagTakesTheFirstEntryOfTheResponse(t *testing.T) {
	u := testUpdater(t, "2.44.0", "2.45.0")

	latest, err := u.latestTag()

	require.NoError(t, err)
	require.Equal(t, "2.44.0", latest.String())
}

func TestUpdaterNewCLIVersion(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		tags           []string
		want           string
	}{
		{
			name:           "newer version published",
			currentVersion: "2.44.0",
			tags:           []string{"2.45.0"},
			want:           "2.45.0",
		},
		{
			name:           "up to date",
			currentVersion: "2.45.0",
			tags:           []string{"2.45.0"},
			want:           "",
		},
		{
			name:           "running a version newer than the published one",
			currentVersion: "2.46.0",
			tags:           []string{"2.45.0"},
			want:           "",
		},
		{
			name:           "dev build has no comparable version",
			currentVersion: "dev",
			tags:           []string{"2.45.0"},
			want:           "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setVersion(t, test.currentVersion)
			u := testUpdater(t, test.tags...)

			newVersion, err := u.newCLIVersion()

			require.NoError(t, err)
			require.Equal(t, test.want, newVersion)
		})
	}
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

func setVersion(t *testing.T, v string) {
	t.Helper()

	original := version.VERSION
	version.VERSION = v
	t.Cleanup(func() { version.VERSION = original })
}

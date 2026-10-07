package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name          string
		pages         [][]Release
		current       string
		wantUpdate    *Available
		wantNewMajor  *Available
		wantMajorSeen bool
		wantRequests  int64
	}{
		{
			name:          "the newest release of the major updates, a higher major is reported apart",
			pages:         [][]Release{{release("v3.0.0"), release("v2.46.0"), release("v2.45.0")}},
			current:       "2.45.0",
			wantUpdate:    available("2.46.0"),
			wantNewMajor:  available("3.0.0"),
			wantMajorSeen: true,
			wantRequests:  1,
		},
		{
			name:          "the max semver wins over the list order",
			pages:         [][]Release{{release("v2.45.0"), release("v2.46.1"), release("v2.46.0")}},
			current:       "2.44.0",
			wantUpdate:    available("2.46.1"),
			wantMajorSeen: true,
			wantRequests:  1,
		},
		{
			name: "drafts and pre-releases are skipped",
			pages: [][]Release{{
				prerelease("v2.47.0"),
				draft("v2.46.0"),
				release("v2.45.1"),
			}},
			current:       "2.45.0",
			wantUpdate:    available("2.45.1"),
			wantMajorSeen: true,
			wantRequests:  1,
		},
		{
			name:         "a tag that is not MAJOR.MINOR.PATCH is skipped",
			pages:        [][]Release{{release("v2.46.0-rc1"), release("v2.46"), release("latest")}},
			current:      "2.45.0",
			wantRequests: 1,
		},
		{
			name:         "an empty release list offers nothing",
			pages:        [][]Release{{}},
			current:      "2.45.0",
			wantRequests: 1,
		},
		{
			name:          "the newest release of the major being the running one offers nothing",
			pages:         [][]Release{{release("v2.46.0"), release("v2.45.0")}},
			current:       "2.46.0",
			wantMajorSeen: true,
			wantRequests:  1,
		},
		{
			name:          "a release older than the running one offers nothing",
			pages:         [][]Release{{release("v2.45.0")}},
			current:       "2.46.0",
			wantMajorSeen: true,
			wantRequests:  1,
		},
		{
			name:          "a full page without the running major is followed by the next page",
			pages:         [][]Release{majorReleases(3, releasesPerPage), {release("v2.45.0")}},
			current:       "2.44.0",
			wantUpdate:    available("2.45.0"),
			wantNewMajor:  available(fmt.Sprintf("3.%d.0", releasesPerPage-1)),
			wantMajorSeen: true,
			wantRequests:  2,
		},
		{
			name:          "a full page carrying the running major is not followed by the next page",
			pages:         [][]Release{majorReleases(2, releasesPerPage), {release("v2.100.0")}},
			current:       "2.44.0",
			wantUpdate:    available(fmt.Sprintf("2.%d.0", releasesPerPage-1)),
			wantMajorSeen: true,
			wantRequests:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, requests := newReleasesClient(t, tt.pages)

			versions, err := Resolve(t.Context(), client, tt.current)

			require.NoError(t, err)
			require.Equal(t, tt.wantUpdate, versions.Update)
			require.Equal(t, tt.wantNewMajor, versions.NewMajor)
			require.Equal(t, tt.wantMajorSeen, versions.CurrentMajorFound)
			require.Equal(t, tt.wantRequests, requests.Load())
		})
	}
}

func TestResolve_NotComparableCurrentVersionSendsNoRequest(t *testing.T) {
	client, requests := newReleasesClient(t, [][]Release{{release("v2.46.0")}})

	_, err := Resolve(t.Context(), client, "dev")

	require.Error(t, err)
	require.Zero(t, requests.Load())
}

func TestResolve_PagingIsBounded(t *testing.T) {
	pages := make([][]Release, maxReleasePages+1)
	for i := range pages {
		pages[i] = majorReleases(3, releasesPerPage)
	}
	client, requests := newReleasesClient(t, pages)

	versions, err := Resolve(t.Context(), client, "2.44.0")

	require.NoError(t, err)
	require.Nil(t, versions.Update)
	require.False(t, versions.CurrentMajorFound, "a nil Update must not read as up to date when the major was never seen")
	require.EqualValues(t, maxReleasePages, requests.Load())
}

func TestResolve_RequestError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)
	client := testClient(srv.URL)

	_, err := Resolve(t.Context(), client, "2.45.0")

	require.Error(t, err)
}

func TestResolveTarget(t *testing.T) {
	tests := []struct {
		name         string
		pages        [][]Release
		current      string
		requested    string
		want         Target
		wantRequests int64
	}{
		{
			name:         "the newest release of the running major is installed, a higher major only reported",
			pages:        [][]Release{{release("v3.0.0"), release("v2.46.0"), release("v2.45.0")}},
			current:      "2.45.0",
			want:         Target{Version: "2.46.0", NewMajor: available("3.0.0")},
			wantRequests: 1,
		},
		{
			name:         "the newest release of the running major being the running one is up to date",
			pages:        [][]Release{{release("v2.46.0"), release("v2.45.0")}},
			current:      "2.46.0",
			want:         Target{Version: "2.46.0", UpToDate: true},
			wantRequests: 1,
		},
		{
			name:         "a higher major is reported even when the running major has nothing newer",
			pages:        [][]Release{{release("v3.0.0"), release("v2.46.0")}},
			current:      "2.46.0",
			want:         Target{Version: "2.46.0", UpToDate: true, NewMajor: available("3.0.0")},
			wantRequests: 1,
		},
		{
			name:      "a requested version crossing a major is installed",
			pages:     [][]Release{{release("v2.46.0")}},
			current:   "2.45.0",
			requested: "3.0.0",
			want:      Target{Version: "3.0.0"},
		},
		{
			name:      "a requested version keeps its pre-release",
			current:   "2.45.0",
			requested: "3.0.0-rc.1",
			want:      Target{Version: "3.0.0-rc.1"},
		},
		{
			name:      "a requested version is accepted with the tag's v prefix",
			current:   "2.45.0",
			requested: "v2.31.0",
			want:      Target{Version: "2.31.0"},
		},
		{
			name:      "a requested version below the running one is installed",
			current:   "2.46.0",
			requested: "2.31.0",
			want:      Target{Version: "2.31.0"},
		},
		{
			name:      "a requested version naming the running one is up to date",
			current:   "2.46.0",
			requested: "v2.46.0",
			want:      Target{Version: "2.46.0", UpToDate: true},
		},
		{
			name:      "a requested version is installed on a build that has no comparable version",
			current:   "dev",
			requested: "2.31.0",
			want:      Target{Version: "2.31.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, requests := newReleasesClient(t, tt.pages)

			target, err := ResolveTarget(t.Context(), client, tt.current, tt.requested)

			require.NoError(t, err)
			require.Equal(t, tt.want, target)
			require.Equal(t, tt.wantRequests, requests.Load())
		})
	}
}

func TestResolveTarget_NotComparableCurrentVersionIsRejectedWithoutAskingGitHub(t *testing.T) {
	for _, current := range []string{"dev", "2.46.1-next"} {
		t.Run(current, func(t *testing.T) {
			client, requests := newReleasesClient(t, [][]Release{{release("v2.46.0")}})

			_, err := ResolveTarget(t.Context(), client, current, "")

			require.ErrorContains(t, err, "bitrise update --version X.Y.Z")
			require.Zero(t, requests.Load())
		})
	}
}

func TestResolveTarget_InvalidRequestedVersionIsRejectedWithoutAskingGitHub(t *testing.T) {
	for _, requested := range []string{"invalid", "latest", "2.46", "2.46.0.1", "vv2.46.0", "2.46.0+build.7"} {
		t.Run(requested, func(t *testing.T) {
			client, requests := newReleasesClient(t, [][]Release{{release("v2.46.0")}})

			_, err := ResolveTarget(t.Context(), client, "2.45.0", requested)

			require.ErrorContains(t, err, "expected MAJOR.MINOR.PATCH")
			require.Zero(t, requests.Load())
		})
	}
}

// newReleasesClient serves pages[page-1] for every requested page, and counts
// the requests it answered.
func newReleasesClient(t *testing.T, pages [][]Release) (*Client, *atomic.Int64) {
	t.Helper()

	var requests atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		releases := []Release{}
		if page >= 1 && page <= len(pages) {
			releases = pages[page-1]
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(releases)
	}))
	t.Cleanup(srv.Close)

	return testClient(srv.URL), &requests
}

func majorReleases(major, count int) []Release {
	releases := make([]Release, 0, count)
	for minor := count - 1; minor >= 0; minor-- {
		releases = append(releases, release(fmt.Sprintf("v%d.%d.0", major, minor)))
	}
	return releases
}

func release(tag string) Release {
	return Release{TagName: tag, HTMLURL: releaseNotesURL(tag)}
}

func draft(tag string) Release {
	r := release(tag)
	r.Draft = true
	return r
}

func prerelease(tag string) Release {
	r := release(tag)
	r.Prerelease = true
	return r
}

func available(version string) *Available {
	return &Available{Version: version, URL: releaseNotesURL("v" + version)}
}

func releaseNotesURL(tag string) string {
	return "https://github.com/bitrise-io/bitrise/releases/tag/" + tag
}

func TestResolve_MajorMissingFromTheReadPagesIsNotUpToDate(t *testing.T) {
	client, _ := newReleasesClient(t, [][]Release{{release("v3.0.0"), release("v2.45.0")}})

	versions, err := Resolve(t.Context(), client, "1.2.3")

	require.NoError(t, err)
	require.Nil(t, versions.Update)
	require.False(t, versions.CurrentMajorFound)
	require.Equal(t, available("3.0.0"), versions.NewMajor)
}

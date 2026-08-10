package bitriseapi

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRawRequest_MethodAndBodyPassthrough(t *testing.T) {
	var gotMethod, gotBody string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		buf := make([]byte, 64)
		n, _ := r.Body.Read(buf)
		gotBody = string(buf[:n])
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	resp, err := mustClient(t, srv.URL, "t").RawRequest(context.Background(), http.MethodPost, "/things", nil, nil, strings.NewReader(`{"a":1}`))
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, `{"a":1}`, gotBody)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, `{"ok":true}`, string(resp.Body))
}

func TestRawRequest_DefaultHeadersAndUserOverride(t *testing.T) {
	var gotAuth, gotAccept string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		_, _ = w.Write([]byte(`{}`))
	})

	header := http.Header{"Accept": []string{"text/plain"}}
	_, err := mustClient(t, srv.URL, "t").RawRequest(context.Background(), http.MethodGet, "/x", nil, header, nil)
	require.NoError(t, err)
	assert.Equal(t, "token t", gotAuth)
	assert.Equal(t, "text/plain", gotAccept, "user header should override the default Accept")
}

func TestRawRequest_PathJoiningAndQueryMerging(t *testing.T) {
	cases := map[string]struct {
		path      string
		query     url.Values
		wantPath  string
		wantQuery string
	}{
		"leading slash":        {path: "/apps", wantPath: "/apps"},
		"no leading slash":     {path: "apps", wantPath: "/apps"},
		"query in path merged": {path: "/apps?limit=10", query: url.Values{"sort": {"name"}}, wantPath: "/apps", wantQuery: "limit=10&sort=name"},
		"explicit query only":  {path: "/apps", query: url.Values{"sort": {"name"}}, wantPath: "/apps", wantQuery: "sort=name"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var gotPath, gotQuery string
			srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotQuery = r.URL.RawQuery
				_, _ = w.Write([]byte(`{}`))
			})

			_, err := mustClient(t, srv.URL, "t").RawRequest(context.Background(), http.MethodGet, tc.path, tc.query, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, tc.wantPath, gotPath)
			assert.Equal(t, tc.wantQuery, gotQuery)
		})
	}
}

func TestRawRequest_AbsoluteURLIgnoresConfiguredBaseButKeepsAuth(t *testing.T) {
	var gotAuth, gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{}`))
	})

	client := mustClient(t, "https://unused.invalid", "t")
	_, err := client.RawRequest(context.Background(), http.MethodGet, srv.URL+"/absolute", nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "/absolute", gotPath)
	assert.Equal(t, "token t", gotAuth)
}

func TestRawRequest_NonSuccessStatusIsNotAnError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	resp, err := mustClient(t, srv.URL, "t").RawRequest(context.Background(), http.MethodGet, "/missing", nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, `{"message":"not found"}`, string(resp.Body))
}

package httpfetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/bartventer/httpcache"
	"github.com/stretchr/testify/require"
)

// inventoryServer serves body under the given Cache-Control with a fixed ETag,
// counting how many requests actually reach it and answering 304 to a matching
// If-None-Match.
type inventoryServer struct {
	hits         atomic.Int32
	conditionals atomic.Int32
	cacheControl string
	etag         string
	body         string
}

func (s *inventoryServer) start(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.hits.Add(1)
		w.Header().Set("Cache-Control", s.cacheControl)
		w.Header().Set("ETag", s.etag)
		if r.Header.Get("If-None-Match") == s.etag {
			s.conditionals.Add(1)
			w.WriteHeader(http.StatusNotModified)
			return
		}
		if _, err := fmt.Fprint(w, s.body); err != nil {
			t.Errorf("write test response: %s", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// cachingClientFor builds the caching client but pointed at a test server, so
// the cache sits above the server's transport instead of the retrying one.
func cachingClientFor(t *testing.T, srv *httptest.Server) Client {
	t.Helper()
	c, err := NewCachingWithClient(testLogger{t}, srv.Client())
	require.NoError(t, err)
	return c
}

// getStatus reads the body and the cache's own verdict on the request, so a
// test can assert HIT/MISS directly rather than inferring it from how many
// requests reached the origin.
func getStatus(t *testing.T, c Client, url string) (string, string) {
	t.Helper()
	rt := c.(*client).httpClient.Transport
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(b), resp.Header.Get(httpcache.CacheStatusHeader)
}

func getBody(t *testing.T, c Client, url string) string {
	t.Helper()
	body, err := c.Get(context.Background(), url)
	require.NoError(t, err)
	defer func() { require.NoError(t, body.Close()) }()
	b, err := io.ReadAll(body)
	require.NoError(t, err)
	return string(b)
}

type testLogger struct{ t *testing.T }

func (l testLogger) Debugf(format string, v ...any) { l.t.Logf(format, v...) }

// A fresh entry is served without touching the origin at all: this is the case
// that removes most of a run's inventory requests.
func TestCacheServesFreshWithoutOrigin(t *testing.T) {
	origin := &inventoryServer{cacheControl: "public, max-age=60, must-revalidate", etag: `"v1"`, body: `{"step_ids":["script"]}`}
	srv := origin.start(t)
	c := cachingClientFor(t, srv)

	body, status := getStatus(t, c, srv.URL)
	require.Equal(t, `{"step_ids":["script"]}`, body)
	require.Equal(t, "MISS", status, "the first read has nothing to serve")

	for range 4 {
		body, status = getStatus(t, c, srv.URL)
		require.Equal(t, `{"step_ids":["script"]}`, body)
		require.Equal(t, "HIT", status)
	}
	require.EqualValues(t, 1, origin.hits.Load(), "five reads of a fresh object must hit the origin once")
	require.EqualValues(t, 0, origin.conditionals.Load(), "a fresh entry must not be revalidated")
}

// An expired entry revalidates with If-None-Match and takes the 304, which the
// transport turns back into a usable 200 - httpfetch.Get rejects any non-2xx,
// so a 304 reaching it would surface as a StatusError.
func TestCacheRevalidatesWithETagAnd304(t *testing.T) {
	origin := &inventoryServer{cacheControl: "public, max-age=0, must-revalidate", etag: `"v1"`, body: `{"step_ids":["script"]}`}
	srv := origin.start(t)
	c := cachingClientFor(t, srv)

	require.Equal(t, `{"step_ids":["script"]}`, getBody(t, c, srv.URL))
	// Immediately stale (max-age=0), so the next read must revalidate and
	// still return the body from cache rather than erroring on the 304.
	body, status := getStatus(t, c, srv.URL)
	require.Equal(t, `{"step_ids":["script"]}`, body)
	require.Equal(t, "REVALIDATED", status)

	require.EqualValues(t, 1, origin.conditionals.Load(), "the second read must send If-None-Match")
	require.EqualValues(t, 2, origin.hits.Load())
}

// An immutable response that is still fresh is served without revalidating.
func TestCacheNeverRevalidatesFreshImmutable(t *testing.T) {
	origin := &inventoryServer{cacheControl: "public, max-age=86400, immutable", etag: `"v1"`, body: `{"title":"Script"}`}
	srv := origin.start(t)
	c := cachingClientFor(t, srv)

	for range 3 {
		require.Equal(t, `{"title":"Script"}`, getBody(t, c, srv.URL))
	}
	require.EqualValues(t, 1, origin.hits.Load())
	require.EqualValues(t, 0, origin.conditionals.Load())
}

// NewCachingClient must wire the cache to the right half of the pair: inventory
// reads are cached, step archives and executables never are.
func TestNewCachingClientCachesInventoryOnly(t *testing.T) {
	clients, err := NewCachingClient(testLogger{t})
	require.NoError(t, err)

	for _, tt := range []struct {
		name     string
		client   Client
		wantHits int
	}{
		{"inventory caches", clients.Caching, 1},
		{"downloads do not cache", clients.Passthrough, 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			origin := &inventoryServer{cacheControl: "public, max-age=60, must-revalidate", etag: `"v1"`, body: `{"step_ids":["script"]}`}
			srv := origin.start(t)

			for range 3 {
				require.Equal(t, `{"step_ids":["script"]}`, getBody(t, tt.client, srv.URL))
			}
			require.EqualValues(t, tt.wantHits, origin.hits.Load(), "three reads through %s", tt.name)
		})
	}
}

// A failing origin must not be cached. retryablehttp uses PassthroughErrorHandler,
// so a 5xx that outlives the retries arrives as a normal response, and a 404 is a
// legitimate inventory answer (unknown step or version). Either one carries the
// inventory's freshness directive, so an unguarded cache would store it and replay
// it for the rest of the run.
func TestCacheDoesNotStoreFailures(t *testing.T) {
	for _, status := range []int{http.StatusBadGateway, http.StatusNotFound} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var hits atomic.Int32
			var failing atomic.Bool
			failing.Store(true)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				hits.Add(1)
				w.Header().Set("Cache-Control", "public, max-age=60, must-revalidate")
				if failing.Load() {
					w.WriteHeader(status)
					return
				}
				if _, err := fmt.Fprint(w, `{"step_ids":["script"]}`); err != nil {
					t.Errorf("write test response: %s", err)
				}
			}))
			t.Cleanup(srv.Close)
			c := cachingClientFor(t, srv)

			// The second read must reach the origin rather than be served the stored failure.
			for range 2 {
				_, err := c.Get(context.Background(), srv.URL)
				require.Error(t, err)
			}
			require.EqualValues(t, 2, hits.Load(), "a failure must not be cached and replayed")

			// Once the origin recovers the very next read must succeed - a pinned
			// failure would keep failing for the rest of its freshness window.
			failing.Store(false)
			require.Equal(t, `{"step_ids":["script"]}`, getBody(t, c, srv.URL))
			require.EqualValues(t, 3, hits.Load())
		})
	}
}

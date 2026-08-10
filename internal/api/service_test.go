package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

func TestDo_GetFieldsBecomeQuery(t *testing.T) {
	var gotQuery string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{}`))
	})

	_, err := NewService(mustClient(t, srv.URL)).Do(context.Background(), Request{
		Method: http.MethodGet,
		Path:   "/apps",
		Fields: []KeyValue{{Key: "sort_by", Value: "last_build_at"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "sort_by=last_build_at", gotQuery)
}

func TestDo_NonGetFieldsBecomeJSONBodyWithDefaultContentType(t *testing.T) {
	var gotBody, gotContentType string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		_, _ = w.Write([]byte(`{}`))
	})

	_, err := NewService(mustClient(t, srv.URL)).Do(context.Background(), Request{
		Method: http.MethodPost,
		Path:   "/apps",
		Fields: []KeyValue{{Key: "title", Value: "My App"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "application/json", gotContentType)
	assert.JSONEq(t, `{"title":"My App"}`, gotBody)
}

func TestDo_CallerContentTypeWins(t *testing.T) {
	var gotContentType string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		_, _ = w.Write([]byte(`{}`))
	})

	_, err := NewService(mustClient(t, srv.URL)).Do(context.Background(), Request{
		Method:  http.MethodPost,
		Path:    "/apps",
		Fields:  []KeyValue{{Key: "title", Value: "My App"}},
		Headers: http.Header{"Content-Type": []string{"application/vnd.custom+json"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "application/vnd.custom+json", gotContentType)
}

func TestDo_BodyPassthrough(t *testing.T) {
	var gotBody string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		_, _ = w.Write([]byte(`{}`))
	})

	_, err := NewService(mustClient(t, srv.URL)).Do(context.Background(), Request{
		Method: http.MethodPost,
		Path:   "/apps",
		Body:   strings.NewReader(`{"raw":"body"}`),
	})
	require.NoError(t, err)
	assert.Equal(t, `{"raw":"body"}`, gotBody)
}

func TestDo_InputBodyGetsDefaultContentType(t *testing.T) {
	var gotContentType string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		_, _ = w.Write([]byte(`{}`))
	})

	_, err := NewService(mustClient(t, srv.URL)).Do(context.Background(), Request{
		Method: http.MethodPost,
		Path:   "/apps",
		Body:   strings.NewReader(`{"nested":{"a":1}}`),
	})
	require.NoError(t, err)
	assert.Equal(t, "application/json", gotContentType)
}

func TestDo_PaginateRejectedOnNonGet(t *testing.T) {
	_, err := NewService(mustClient(t, "https://unused.invalid")).Do(context.Background(), Request{
		Method:   http.MethodPost,
		Path:     "/apps",
		Paginate: true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--all is only supported for GET requests")
}

func TestDo_PaginateRejectedWithInputBody(t *testing.T) {
	_, err := NewService(mustClient(t, "https://unused.invalid")).Do(context.Background(), Request{
		Method:   http.MethodGet,
		Path:     "/apps",
		Body:     strings.NewReader(`{"raw":"body"}`),
		Paginate: true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--all cannot be combined with --input")
}

func TestDo_PaginateMergesPages(t *testing.T) {
	var requests []string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.RawQuery)
		if r.URL.Query().Get("next") == "cursor2" {
			_, _ = w.Write([]byte(`{"data":[3],"paging":{"next":""}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[1,2],"paging":{"next":"cursor2"}}`))
	})

	resp, err := NewService(mustClient(t, srv.URL)).Do(context.Background(), Request{
		Method:   http.MethodGet,
		Path:     "/apps",
		Paginate: true,
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"data":[1,2,3]}`, string(resp.Body))
	require.Len(t, requests, 2)
	assert.Equal(t, "next=cursor2", requests[1])
}

func TestDo_PaginateNoOpOnNonListResponse(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"foo":"bar"}`))
	})

	resp, err := NewService(mustClient(t, srv.URL)).Do(context.Background(), Request{
		Method:   http.MethodGet,
		Path:     "/me",
		Paginate: true,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, `{"foo":"bar"}`, string(resp.Body))
}

func TestDo_PaginateStopsOnErrorPage(t *testing.T) {
	var calls int
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			_, _ = w.Write([]byte(`{"data":[1],"paging":{"next":"cursor2"}}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp, err := NewService(mustClient(t, srv.URL)).Do(context.Background(), Request{
		Method:   http.MethodGet,
		Path:     "/apps",
		Paginate: true,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, `{"message":"boom"}`, string(resp.Body))
	assert.Equal(t, 2, calls)
}

func TestDo_PaginateEmptyListStaysAnArray(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[],"paging":{"next":""}}`))
	})

	resp, err := NewService(mustClient(t, srv.URL)).Do(context.Background(), Request{
		Method:   http.MethodGet,
		Path:     "/apps",
		Paginate: true,
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"data":[]}`, string(resp.Body), "an empty list must not collapse to a null data field")
}

func TestDo_PaginateKeepsFieldsAcrossPages(t *testing.T) {
	var requests []url.Values
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Query())
		if r.URL.Query().Get("next") == "cursor2" {
			_, _ = w.Write([]byte(`{"data":[2],"paging":{"next":""}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[1],"paging":{"next":"cursor2"}}`))
	})

	_, err := NewService(mustClient(t, srv.URL)).Do(context.Background(), Request{
		Method:   http.MethodGet,
		Path:     "/apps",
		Fields:   []KeyValue{{Key: "sort_by", Value: "last_build_at"}},
		Paginate: true,
	})
	require.NoError(t, err)
	require.Len(t, requests, 2)
	assert.Equal(t, "last_build_at", requests[1].Get("sort_by"))
	assert.Equal(t, "cursor2", requests[1].Get("next"))
}

func TestDo_PaginateReportsLastPageHeaders(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("next") == "cursor2" {
			w.Header().Set("X-Ratelimit-Remaining", "41")
			w.Header().Set("Etag", `W/"page-2"`)
			_, _ = w.Write([]byte(`{"data":[2],"paging":{"next":""}}`))
			return
		}
		w.Header().Set("X-Ratelimit-Remaining", "42")
		_, _ = w.Write([]byte(`{"data":[1],"paging":{"next":"cursor2"}}`))
	})

	resp, err := NewService(mustClient(t, srv.URL)).Do(context.Background(), Request{
		Method:   http.MethodGet,
		Path:     "/apps",
		Paginate: true,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "41", resp.Header.Get("X-Ratelimit-Remaining"))
	assert.Empty(t, resp.Header.Get("Content-Length"), "the last page's length must not describe the merged body")
	assert.Empty(t, resp.Header.Get("Etag"), "the last page's validator must not describe the merged body")
}

func TestDo_PaginateStopsOnRepeatedCursor(t *testing.T) {
	var calls int
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"data":[1],"paging":{"next":"stuck"}}`))
	})

	_, err := NewService(mustClient(t, srv.URL)).Do(context.Background(), Request{
		Method:   http.MethodGet,
		Path:     "/apps",
		Paginate: true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pagination stalled")
	assert.Equal(t, 2, calls)
}

func newFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func mustClient(t *testing.T, baseURL string) *bitriseapi.Client {
	t.Helper()
	c, err := bitriseapi.New(baseURL, "t")
	require.NoError(t, err)
	return c
}

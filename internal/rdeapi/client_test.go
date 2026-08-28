package rdeapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_TrimsTrailingSlashFromBaseURL(t *testing.T) {
	c, err := New("https://api.bitrise.io/rde/", "tok")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.baseURL != "https://api.bitrise.io/rde" {
		t.Errorf("baseURL = %q, want trailing slash trimmed", c.baseURL)
	}
}

func TestNew_RejectsPlainHTTP(t *testing.T) {
	if _, err := New("http://api.bitrise.io/rde", "tok"); err == nil {
		t.Error("expected error for non-loopback plaintext http")
	}
}

func TestNew_DefaultHTTPClientTimeout(t *testing.T) {
	c, err := New("https://api.bitrise.io/rde", "tok")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.httpClient.Timeout != defaultTimeout {
		t.Errorf("httpClient.Timeout = %v, want %v", c.httpClient.Timeout, defaultTimeout)
	}
}

func TestDo_SetsRequiredHeaders(t *testing.T) {
	rs := newRecordingServer(t, `{"savedInputs":[]}`)

	if _, err := rs.client().ListSavedInputs(context.Background()); err != nil {
		t.Fatalf("ListSavedInputs: %v", err)
	}
	checks := map[string]string{
		"Authorization":    "Bearer tok",
		"Accept":           "application/json",
		"User-Agent":       UserAgent,
		"X-Request-Source": requestSource,
	}
	for h, want := range checks {
		if got := rs.lastHeader.Get(h); got != want {
			t.Errorf("header %s = %q, want %q", h, got, want)
		}
	}
}

func TestGetRequest_HasNoContentType(t *testing.T) {
	rs := newRecordingServer(t, `{"savedInputs":[]}`)

	if _, err := rs.client().ListSavedInputs(context.Background()); err != nil {
		t.Fatalf("ListSavedInputs: %v", err)
	}
	// getJSON sends no body, so Content-Type must be absent.
	if ct := rs.lastHeader.Get("Content-Type"); ct != "" {
		t.Errorf("GET Content-Type = %q, want empty", ct)
	}
}

func TestPostRequest_SetsJSONContentType(t *testing.T) {
	rs := newRecordingServer(t, `{"savedInput":{"id":"x","key":"k"}}`)

	if _, err := rs.client().CreateSavedInput(context.Background(), CreateSavedInputRequest{Key: "k", Value: "v"}); err != nil {
		t.Fatalf("CreateSavedInput: %v", err)
	}
	if ct := rs.lastHeader.Get("Content-Type"); ct != "application/json" {
		t.Errorf("POST Content-Type = %q, want application/json", ct)
	}
}

func TestAPIError_ExtractsMessageFromEnvelope(t *testing.T) {
	rs := newRecordingServer(t, `{"code":5,"message":"session not found"}`)
	rs.status = http.StatusNotFound

	_, err := rs.client().ListSavedInputs(context.Background())
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if apiErr.Message != "session not found" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "session not found")
	}
	if apiErr.Body != "" {
		t.Errorf("Body = %q, want empty when message was parsed", apiErr.Body)
	}
	if want := "GET /v1/saved-inputs"; apiErr.RequestInfo != want {
		t.Errorf("RequestInfo = %q, want %q", apiErr.RequestInfo, want)
	}
	if want := "GET /v1/saved-inputs: RDE API 404: session not found"; apiErr.Error() != want {
		t.Errorf("Error() = %q, want %q", apiErr.Error(), want)
	}
}

func TestAPIError_IncludesFieldViolations(t *testing.T) {
	rs := newRecordingServer(t, `{"code":3,"message":"Bad request.","details":[{"@type":"type.googleapis.com/google.rpc.BadRequest","fieldViolations":[{"field":"session_inputs","description":"missing required input: BUILD_TOKEN"}]}]}`)
	rs.status = http.StatusBadRequest

	_, err := rs.client().ListSavedInputs(context.Background())
	if err == nil {
		t.Fatal("expected error on 400")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %T", err)
	}
	if got, want := len(apiErr.Violations), 1; got != want {
		t.Fatalf("len(Violations) = %d, want %d", got, want)
	}
	if want := "missing required input: BUILD_TOKEN"; apiErr.Violations[0] != want {
		t.Errorf("Violations[0] = %q, want %q", apiErr.Violations[0], want)
	}
	if apiErr.Body != "" {
		t.Errorf("Body = %q, want empty when violations were parsed", apiErr.Body)
	}
	if want := "GET /v1/saved-inputs: RDE API 400: Bad request.: missing required input: BUILD_TOKEN"; apiErr.Error() != want {
		t.Errorf("Error() = %q, want %q", apiErr.Error(), want)
	}
}

func TestAPIError_FieldViolationFallsBackToFieldName(t *testing.T) {
	rs := newRecordingServer(t, `{"details":[{"@type":"type.googleapis.com/google.rpc.BadRequest","fieldViolations":[{"field":"name"}]}]}`)
	rs.status = http.StatusBadRequest

	_, err := rs.client().ListSavedInputs(context.Background())
	if err == nil {
		t.Fatal("expected error on 400")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %T", err)
	}
	// No message, no description — the field name carries the only signal,
	// so the raw body must not be used as a fallback.
	if want := "GET /v1/saved-inputs: RDE API 400: name"; apiErr.Error() != want {
		t.Errorf("Error() = %q, want %q", apiErr.Error(), want)
	}
}

func TestAPIError_FallsBackToRawBody(t *testing.T) {
	rs := newRecordingServer(t, "upstream exploded")
	rs.status = http.StatusInternalServerError

	_, err := rs.client().ListSavedInputs(context.Background())
	if err == nil {
		t.Fatal("expected error on 500")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %T", err)
	}
	// No JSON message field, so the raw body is preserved.
	if apiErr.Message != "" {
		t.Errorf("Message = %q, want empty", apiErr.Message)
	}
	if apiErr.Body != "upstream exploded" {
		t.Errorf("Body = %q, want %q", apiErr.Body, "upstream exploded")
	}
	if want := "GET /v1/saved-inputs: RDE API 500: upstream exploded"; apiErr.Error() != want {
		t.Errorf("Error() = %q, want %q", apiErr.Error(), want)
	}
}

func TestAPIError_BareStatusWhenNoMessageOrBody(t *testing.T) {
	e := &APIError{StatusCode: http.StatusBadGateway}
	if want := "RDE API 502"; e.Error() != want {
		t.Errorf("Error() = %q, want %q", e.Error(), want)
	}
}

// doStream has no production caller in this package yet — it lands ahead of
// its first consumer (session log streaming, a later PR) because the client
// core needs to be complete on its own. Test it directly.

func TestDoStream_ReturnsLiveResponseOn2xx(t *testing.T) {
	rs := newRecordingServer(t, "line one\nline two\n")
	c := rs.client()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rs.srv.URL+"/v1/stream", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := c.doStream(req)
	if err != nil {
		t.Fatalf("doStream: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if want := "line one\nline two\n"; string(body) != want {
		t.Errorf("body = %q, want %q", body, want)
	}
	for h, want := range map[string]string{
		"Authorization":    "Bearer tok",
		"Accept":           "application/json",
		"User-Agent":       UserAgent,
		"X-Request-Source": requestSource,
	} {
		if got := rs.lastHeader.Get(h); got != want {
			t.Errorf("header %s = %q, want %q", h, got, want)
		}
	}
}

func TestDoStream_ReturnsAPIErrorOnNon2xx(t *testing.T) {
	rs := newRecordingServer(t, `{"message":"stream not available"}`)
	rs.status = http.StatusNotFound
	c := rs.client()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rs.srv.URL+"/v1/stream", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	_, err = c.doStream(req)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %T", err)
	}
	if apiErr.Message != "stream not available" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "stream not available")
	}
}

func TestContextCancellation(t *testing.T) {
	rs := newRecordingServer(t, `{"savedInputs":[]}`)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := rs.client().ListSavedInputs(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestWsPath(t *testing.T) {
	if got := wsPath("ws-1", "/sessions"); got != "/v1/workspaces/ws-1/sessions" {
		t.Errorf("wsPath = %q", got)
	}
	// Missing leading slash is tolerated.
	if got := wsPath("ws-1", "sessions"); got != "/v1/workspaces/ws-1/sessions" {
		t.Errorf("wsPath (no leading slash) = %q", got)
	}
	// Workspace IDs are path-escaped.
	if got := wsPath("a b", "/x"); got != "/v1/workspaces/a%20b/x" {
		t.Errorf("wsPath (escaping) = %q", got)
	}
}

func TestUserPath(t *testing.T) {
	if got := userPath("/saved-inputs"); got != "/v1/saved-inputs" {
		t.Errorf("userPath = %q", got)
	}
	if got := userPath("saved-inputs"); got != "/v1/saved-inputs" {
		t.Errorf("userPath (no leading slash) = %q", got)
	}
}

// recordingServer spins up an httptest server, captures the last request
// (method, path, headers, body), and replies with a canned status + body.
// Shared by client_test.go and saved_inputs_test.go.
type recordingServer struct {
	t          *testing.T
	srv        *httptest.Server
	status     int
	response   string
	lastMethod string
	lastPath   string
	lastBody   []byte
	lastHeader http.Header
	hits       int
}

func newRecordingServer(t *testing.T, response string) *recordingServer {
	t.Helper()
	rs := &recordingServer{t: t, status: http.StatusOK, response: response}
	rs.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rs.hits++
		rs.lastMethod = r.Method
		rs.lastPath = r.URL.Path
		rs.lastHeader = r.Header.Clone()
		rs.lastBody, _ = io.ReadAll(r.Body)
		if rs.status != http.StatusOK {
			w.WriteHeader(rs.status)
		}
		_, _ = io.WriteString(w, rs.response)
	}))
	t.Cleanup(rs.srv.Close)
	return rs
}

func (rs *recordingServer) client() *Client {
	c, err := New(rs.srv.URL, "tok")
	if err != nil {
		rs.t.Fatalf("New: %v", err)
	}
	return c
}

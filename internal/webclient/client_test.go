package webclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestNew_RejectsInvalidBaseURLs(t *testing.T) {
	for _, raw := range []string{
		"app.bitrise.io",        // no scheme
		"http://app.bitrise.io", // plaintext against a non-loopback host
		"ftp://app.bitrise.io",  // wrong scheme entirely
		"not a url \x7f",        // unparsable
	} {
		if _, err := New(raw); err == nil {
			t.Fatalf("New(%q) = nil error, want an error", raw)
		}
	}
}

func TestNew_AllowsHTTPOnLoopback(t *testing.T) {
	for _, raw := range []string{"http://localhost:1234", "http://127.0.0.1:1234"} {
		if _, err := New(raw); err != nil {
			t.Fatalf("New(%q): %v", raw, err)
		}
	}
}

func TestNew_TrimsTrailingSlash(t *testing.T) {
	c, err := New("https://app.bitrise.io/")
	if err != nil {
		t.Fatal(err)
	}
	if c.baseURL != "https://app.bitrise.io" {
		t.Fatalf("baseURL = %q, want no trailing slash", c.baseURL)
	}
}

func TestPrime_SendsUserAgent(t *testing.T) {
	var capturedUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUA = r.Header.Get("User-Agent")
		_, _ = io.WriteString(w, `<meta name="csrf-token" content="t" />`)
	}))
	defer srv.Close()

	c, err := New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Prime(context.Background(), "/users/sign_up"); err != nil {
		t.Fatalf("Prime: %v", err)
	}
	if capturedUA != userAgent {
		t.Fatalf("User-Agent = %q, want %q", capturedUA, userAgent)
	}
}

func TestPrime_ExtractsMetaCSRFAndStoresCookies(t *testing.T) {
	const wantToken = "csrf-meta-value-abc"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/sign_up" {
			http.NotFound(w, r)
			return
		}
		setTestCookie(w, "CSRF-TOKEN", "cookie-value")
		setTestCookie(w, "_concrete_website_session", "sess1")
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, `<html><meta name="csrf-token" content="`+wantToken+`" /></html>`)
	}))
	defer srv.Close()

	c, err := New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if c.httpClient.Timeout != defaultTimeout {
		t.Fatalf("httpClient.Timeout = %v, want %v", c.httpClient.Timeout, defaultTimeout)
	}
	if c.httpClient.Jar == nil {
		t.Fatal("httpClient.Jar is nil, want a cookie jar for the Prime/PostJSON flow")
	}
	if err := c.Prime(context.Background(), "/users/sign_up"); err != nil {
		t.Fatalf("Prime: %v", err)
	}
	if c.csrfToken != wantToken {
		t.Fatalf("csrfToken = %q, want %q", c.csrfToken, wantToken)
	}
}

func TestPrime_ReturnsErrorOnNon2xxStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `<html>maintenance</html>`)
	}))
	defer srv.Close()

	c, err := New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Prime(context.Background(), "/users/sign_in"); err == nil {
		t.Fatal("expected Prime to error on a 503 response")
	}
	if c.csrfToken != "" {
		t.Fatalf("csrfToken = %q, want unset after a failed prime", c.csrfToken)
	}
}

func TestPrime_ReturnsErrorWhenNoCSRFTokenFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `<html><body>no csrf meta here</body></html>`)
	}))
	defer srv.Close()

	c, err := New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Prime(context.Background(), "/users/sign_in"); err == nil {
		t.Fatal("expected Prime to error when no CSRF meta tag is present")
	}
}

func TestPostJSON_SendsCSRFAndCookies(t *testing.T) {
	var capturedCSRF string
	var capturedCookieNames []string
	var capturedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/sign_up":
			setTestCookie(w, "CSRF-TOKEN", "cookie-value")
			w.Header().Set("Content-Type", "text/html")
			_, _ = io.WriteString(w, `<html><meta name="csrf-token" content="meta-token" /></html>`)
		case "/users":
			capturedCSRF = r.Header.Get("X-CSRF-Token")
			for _, ck := range r.Cookies() {
				capturedCookieNames = append(capturedCookieNames, ck.Name)
			}
			_ = json.NewDecoder(r.Body).Decode(&capturedBody)
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"id":1}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c, err := New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Prime(context.Background(), "/users/sign_up"); err != nil {
		t.Fatalf("Prime: %v", err)
	}
	resp, err := c.PostJSON(context.Background(), "/users", map[string]any{"x": 1})
	if err != nil {
		t.Fatalf("PostJSON: %v", err)
	}
	if resp.Status != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.Status)
	}
	if capturedCSRF != "meta-token" {
		t.Fatalf("X-CSRF-Token = %q, want %q", capturedCSRF, "meta-token")
	}
	if !slices.Contains(capturedCookieNames, "CSRF-TOKEN") {
		t.Fatalf("CSRF-TOKEN cookie not forwarded; got cookies %v", capturedCookieNames)
	}
	if capturedBody["x"].(float64) != 1 {
		t.Fatalf("body x = %v, want 1", capturedBody["x"])
	}
}

func TestPostJSON_OmitsCSRFWhenNotPrimed(t *testing.T) {
	var capturedCSRF string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCSRF = r.Header.Get("X-CSRF-Token")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c, err := New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.PostJSON(context.Background(), "/users/sign_in", map[string]any{}); err != nil {
		t.Fatalf("PostJSON: %v", err)
	}
	if capturedCSRF != "" {
		t.Fatalf("X-CSRF-Token = %q, want empty", capturedCSRF)
	}
}

func TestPostJSON_ReturnsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	c, err := New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.PostJSON(context.Background(), "/anything", map[string]any{})
	if err != nil {
		t.Fatalf("PostJSON: %v", err)
	}
	if resp.Status != http.StatusCreated {
		t.Fatalf("status = %d", resp.Status)
	}
	if !strings.Contains(string(resp.Body), `"ok":true`) {
		t.Fatalf("body = %s", resp.Body)
	}
}

func TestPostJSON_RedirectSurfacesAsNonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/sign_in" {
			http.Redirect(w, r, "/users/sign_in", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c, err := New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.PostJSON(context.Background(), "/users/sign_in", map[string]any{})
	if err != nil {
		t.Fatalf("PostJSON: %v", err)
	}
	if resp.Status != http.StatusFound {
		t.Fatalf("status = %d, want %d (redirect not followed)", resp.Status, http.StatusFound)
	}
}

// setTestCookie writes a test cookie. Not Secure: the test server runs over
// plain HTTP, and pre-1.26 Go's cookiejar won't forward a Secure cookie
// there even for localhost.
func setTestCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
}

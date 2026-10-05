package update

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReleasesPage(t *testing.T) {
	var gotQuery, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		gotAccept = r.Header.Get("Accept")
		_, _ = w.Write([]byte(`[{"tag_name":"v2.46.0","html_url":"https://example.com/v2.46.0","draft":false,"prerelease":true}]`))
	}))
	t.Cleanup(srv.Close)
	client := NewClient(WithReleasesURL(srv.URL), WithHTTPClient(srv.Client()))

	releases, err := client.ReleasesPage(t.Context(), 2)

	require.NoError(t, err)
	require.Equal(t, "page=2&per_page=100", gotQuery)
	require.Equal(t, "application/vnd.github+json", gotAccept)
	require.Equal(t, []Release{{
		TagName:    "v2.46.0",
		HTMLURL:    "https://example.com/v2.46.0",
		Prerelease: true,
	}}, releases)
}

func TestReleasesPage_Errors(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr string
	}{
		{
			name: "non-2xx response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
			},
			wantErr: "GitHub releases API 403: {\"message\":\"API rate limit exceeded\"}",
		},
		{
			name: "malformed JSON",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"not":"a list"`))
			},
			wantErr: "parse releases response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			t.Cleanup(srv.Close)
			client := NewClient(WithReleasesURL(srv.URL), WithHTTPClient(srv.Client()))

			_, err := client.ReleasesPage(t.Context(), 1)

			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestReleasesPage_UnreachableHost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	addr := srv.URL
	srv.Close()
	client := NewClient(WithReleasesURL(addr))

	_, err := client.ReleasesPage(t.Context(), 1)

	require.ErrorContains(t, err, "request releases")
}

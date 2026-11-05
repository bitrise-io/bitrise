package quarantinedtests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalapp "github.com/bitrise-io/bitrise/v3/internal/app"
)

func TestView_Text(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(quarantinedTestJSON))
	})

	out, _, err := runCmd(t, NewViewCommand(), srv.URL, "qt-1")
	require.NoError(t, err)
	assert.Equal(t, "/apps/my-app/quarantined-tests/qt-1", gotPath)
	for _, want := range []string{
		"ID:           qt-1\n",
		"Test case:    test_login\n",
		"Suites:       AppTests › LoginTests\n",
		"Class:        LoginSpec\n",
		"Branches:     main, release/*\n",
		"Targets:      all\n",
		"Comment:      Flaky on CI\n",
		" by jdoe\n",
		"From:         build #42 (build-1)\n",
	} {
		assert.Contains(t, out, want)
	}
	assert.NotContains(t, out, "Identifier:", "an empty identifier is left out")
}

func TestView_NotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error_msg":"Not Found"}`))
	})

	_, _, err := runCmd(t, NewViewCommand(), srv.URL, "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting quarantined test failed")
	assert.Contains(t, err.Error(), "404")
}

func TestView_RequiresID(t *testing.T) {
	_, _, err := runCmd(t, NewViewCommand(), "http://127.0.0.1:1")
	assert.Error(t, err)
}

func TestSourceText(t *testing.T) {
	tests := map[string]struct {
		src  internalapp.QuarantineSource
		want string
	}{
		"api":              {src: internalapp.QuarantineSource{Source: "api"}, want: "API"},
		"insights":         {src: internalapp.QuarantineSource{Source: "insights"}, want: "Insights"},
		"build":            {src: internalapp.QuarantineSource{Source: "build", ID: "b-1", Number: 7}, want: "build #7 (b-1)"},
		"build, no num":    {src: internalapp.QuarantineSource{Source: "build", ID: "b-1"}, want: "build b-1"},
		"pipeline":         {src: internalapp.QuarantineSource{Source: "pipeline", ID: "p-1", Number: 3}, want: "pipeline #3 (p-1)"},
		"pipeline, no num": {src: internalapp.QuarantineSource{Source: "pipeline", ID: "p-1"}, want: "pipeline p-1"},
		"unknown source":   {src: internalapp.QuarantineSource{Source: "webhook"}, want: "webhook"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, sourceText(tt.src))
		})
	}
}

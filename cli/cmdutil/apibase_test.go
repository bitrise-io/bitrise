package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestResolveAPIBaseURL_EnvWinsOverContext(t *testing.T) {
	t.Setenv(EnvAPIBaseURL, "https://env.example")

	cmd := &cobra.Command{}
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolved{Config: config.Config{APIBaseURL: "https://ctx.example"}}))

	assert.Equal(t, "https://env.example", ResolveAPIBaseURL(cmd))
}

func TestResolveAPIBaseURL_ContextWinsOverDefault(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolved{Config: config.Config{APIBaseURL: "https://ctx.example"}}))

	assert.Equal(t, "https://ctx.example", ResolveAPIBaseURL(cmd))
}

func TestResolveAPIBaseURL_NilContextFallsBackToDefault(t *testing.T) {
	cmd := &cobra.Command{}

	assert.Equal(t, config.DefaultAPIBaseURL, ResolveAPIBaseURL(cmd))
}

func TestResolveAPIBaseURL_EmptyResolvedContextFallsBackToDefault(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())

	assert.Equal(t, config.DefaultAPIBaseURL, ResolveAPIBaseURL(cmd))
}

func TestResolveAPIBaseURL_TrimsTrailingSlash(t *testing.T) {
	t.Setenv(EnvAPIBaseURL, "https://env.example/")
	cmd := &cobra.Command{}
	assert.Equal(t, "https://env.example", ResolveAPIBaseURL(cmd), "callers concatenate a path onto the result and must not get a double slash")

	t.Setenv(EnvAPIBaseURL, "")
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolved{Config: config.Config{APIBaseURL: "https://ctx.example/"}}))
	assert.Equal(t, "https://ctx.example", ResolveAPIBaseURL(cmd))
}

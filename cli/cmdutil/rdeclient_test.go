package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestResolveRDEAPIBaseURL_EnvWinsOverContext(t *testing.T) {
	t.Setenv(EnvRDEAPIBaseURL, "https://env.example")

	cmd := &cobra.Command{}
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolved{Config: config.Config{RDEAPIBaseURL: "https://ctx.example"}}))

	assert.Equal(t, "https://env.example", ResolveRDEAPIBaseURL(cmd))
}

func TestResolveRDEAPIBaseURL_ContextWinsOverDefault(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolved{Config: config.Config{RDEAPIBaseURL: "https://ctx.example"}}))

	assert.Equal(t, "https://ctx.example", ResolveRDEAPIBaseURL(cmd))
}

func TestResolveRDEAPIBaseURL_NilContextFallsBackToDefault(t *testing.T) {
	cmd := &cobra.Command{}

	assert.Equal(t, config.DefaultRDEAPIBaseURL, ResolveRDEAPIBaseURL(cmd))
}

func TestResolveRDEAPIBaseURL_EmptyResolvedContextFallsBackToDefault(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())

	assert.Equal(t, config.DefaultRDEAPIBaseURL, ResolveRDEAPIBaseURL(cmd))
}

func TestResolveRDEAPIBaseURL_TrimsTrailingSlash(t *testing.T) {
	t.Setenv(EnvRDEAPIBaseURL, "https://env.example/")
	cmd := &cobra.Command{}
	assert.Equal(t, "https://env.example", ResolveRDEAPIBaseURL(cmd), "callers concatenate a path onto the result and must not get a double slash")

	t.Setenv(EnvRDEAPIBaseURL, "")
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolved{Config: config.Config{RDEAPIBaseURL: "https://ctx.example/"}}))
	assert.Equal(t, "https://ctx.example", ResolveRDEAPIBaseURL(cmd))
}

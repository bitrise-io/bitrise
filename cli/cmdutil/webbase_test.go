package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestResolveWebBaseURL_EnvWinsOverContext(t *testing.T) {
	t.Setenv(EnvWebBaseURL, "https://env.example")

	cmd := &cobra.Command{}
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolved{Config: config.Config{WebBaseURL: "https://ctx.example"}}))

	assert.Equal(t, "https://env.example", ResolveWebBaseURL(cmd))
}

func TestResolveWebBaseURL_ContextWinsOverDefault(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolved{Config: config.Config{WebBaseURL: "https://ctx.example"}}))

	assert.Equal(t, "https://ctx.example", ResolveWebBaseURL(cmd))
}

func TestResolveWebBaseURL_NilContextFallsBackToDefault(t *testing.T) {
	cmd := &cobra.Command{}

	assert.Equal(t, config.DefaultWebBaseURL, ResolveWebBaseURL(cmd))
}

func TestResolveWebBaseURL_EmptyResolvedContextFallsBackToDefault(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())

	assert.Equal(t, config.DefaultWebBaseURL, ResolveWebBaseURL(cmd))
}

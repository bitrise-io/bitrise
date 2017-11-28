package git

import (
	"os"

	"github.com/bitrise-io/go-utils/command"
)

// Git represents a Git project.
type Git struct {
	path string
}

// New creates a new git project.
func New(path string) *Git {
	return &Git{path: path}
}

func (g *Git) command(args ...string) *command.Model {
	cmd := command.New("git", args...)
	cmd.SetDir(g.path)
	cmd.SetEnvs(append(os.Environ(), "GIT_ASKPASS=echo")...)
	return cmd
}

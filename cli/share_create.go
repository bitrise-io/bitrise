package cli

import (
	"errors"

	"github.com/bitrise-io/bitrise/tools"
	"github.com/codegangsta/cli"
)

func create(c *cli.Context) error {
	// Input validation
	tag := c.String(TagKey)
	if tag == "" {
		return errors.New("No step tag specified")
	}

	gitURI := c.String(GitKey)
	if gitURI == "" {
		return errors.New("No step url specified")
	}

	stepID := c.String(StepIDKey)

	return tools.StepmanShareCreate(tag, gitURI, stepID)
}

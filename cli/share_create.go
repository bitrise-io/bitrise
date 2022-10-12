package cli

import (
	"github.com/bitrise-io/bitrise/tools"
	"github.com/urfave/cli"
)

func create(c *cli.Context) error {
	// Input validation
	tag := c.String(TagKey)
	if tag == "" {
		failf("No step tag specified")
	}

	gitURI := c.String(GitKey)
	if gitURI == "" {
		failf("No step url specified")
	}

	stepID := c.String(StepIDKey)

	if err := tools.StepmanShareCreate(tag, gitURI, stepID); err != nil {
		failf("Bitrise share create failed, error: %s", err)
	}

	return nil
}

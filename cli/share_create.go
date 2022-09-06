package cli

import (
	"github.com/bitrise-io/bitrise/tools"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/urfave/cli"
)

func create(c *cli.Context) error {
	// Input validation
	tag := c.String(TagKey)
	if tag == "" {
		log.Fatal("No step tag specified")
	}

	gitURI := c.String(GitKey)
	if gitURI == "" {
		log.Fatal("No step url specified")
	}

	stepID := c.String(StepIDKey)

	if err := tools.StepmanShareCreate(tag, gitURI, stepID); err != nil {
		log.Fatalf("Bitrise share create failed, error: %s", err)
	}

	return nil
}

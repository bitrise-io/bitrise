package cli

import (
	"log"

	"github.com/codegangsta/cli"
)

func create(c *cli.Context) {
	// Input validation
	tag := c.String(TagKey)
	if tag == "" {
		log.Fatalln("[STEPMAN] - No step tag specified")
	}

	URL := c.String(URLKey)
	if URL == "" {
		log.Fatalln("[STEPMAN] - No step url specified")
	}

}

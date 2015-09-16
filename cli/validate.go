package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func validate(c *cli.Context) {
	if c.String(ConfigBase64Key) != "" || c.String(ConfigKey) != "" || c.String(PathKey) != "" {
		// Config validation
		_, err := CreateBitriseConfigFromCLIParams(c)
		if err != nil {
			log.Fatalf("Failed to validat bitrise cofing, err: %s", err)
		}

		log.Info("Valid bitrise config")
	}

	if c.String(InventoryBase64Key) != "" || c.String(InventoryKey) != "" {
		// Inventory validation
		_, err := CreateInventoryFromCLIParams(c)
		if err != nil {
			log.Fatalf("Failed to validat inventory, err: %s", err)
		}

		log.Info("Valid inventory")
	}
}

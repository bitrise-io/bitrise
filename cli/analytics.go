package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/codegangsta/cli"
)

//=======================================
// Utility
//=======================================

func setOptOutAnalytics(enable bool) error {
	config, err := bitrise.ReadConfig()
	if err != nil {
		return err
	}

	config.OptOutAnalytics = enable

	if err := bitrise.SaveConfig(config); err != nil {
		return err
	}

	return nil
}

//=======================================
// Main
//=======================================

func enableAnalytics(c *cli.Context) {
	if err := setOptOutAnalytics(false); err != nil {
		log.Fatalf("Failed to enable analytics, error: %#v", err)
	}
	log.Infoln("Analytics enabled")
}

func disableAnalytics(c *cli.Context) {
	if err := setOptOutAnalytics(true); err != nil {
		log.Fatalf("Failed to disable analytics, error: %#v", err)
	}
	log.Infoln("Analytics disabled")
}

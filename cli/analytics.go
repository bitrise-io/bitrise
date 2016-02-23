package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/codegangsta/cli"
)

//=======================================
// Utility
//=======================================

func saveIsAnalyticsDisabled(disabled bool) error {
	config, err := configs.ReadConfig()
	if err != nil {
		return err
	}

	config.IsAnalyticsDisabled = disabled

	if err := configs.SaveConfig(config); err != nil {
		return err
	}

	return nil
}

//=======================================
// Main
//=======================================

func enableAnalytics(c *cli.Context) {
	if err := saveIsAnalyticsDisabled(false); err != nil {
		log.Fatalf("Failed to enable analytics, error: %#v", err)
	}
	log.Infoln("Analytics enabled")
}

func disableAnalytics(c *cli.Context) {
	if err := saveIsAnalyticsDisabled(true); err != nil {
		log.Fatalf("Failed to disable analytics, error: %#v", err)
	}
	log.Infoln("Analytics disabled")
}

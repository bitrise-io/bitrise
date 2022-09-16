package cli

import (
	"strings"

	"github.com/bitrise-io/go-utils/command"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func setup(c *cli.Context) error {
	// Input validation
	steplibURI := c.String(CollectionKey)
	if steplibURI == "" {
		fail("No step collection specified")
	}

	copySpecJSONPath := c.String(CopySpecJSONKey)

	if c.IsSet(LocalCollectionKey) {
		log.Warn("'local' flag is deprecated")
		log.Warn("use 'file://' prefix in steplib path instead")
		log.Println()
	}

	if c.Bool(LocalCollectionKey) {
		if !strings.HasPrefix(steplibURI, "file://") {
			log.Warnf("Appending file path prefix (file://) to StepLib (%s)", steplibURI)
			steplibURI = "file://" + steplibURI
			log.Warnf("From now you can refer to this StepLib with URI: %s", steplibURI)
			log.Warnf("For example, to delete StepLib call: `stepman delete --collection %s`", steplibURI)
		}
	}

	// Setup
	if err := stepman.SetupLibrary(steplibURI); err != nil {
		failf("Setup failed, error: %s", err)
	}

	// Copy spec.json
	if copySpecJSONPath != "" {
		failf("Copying spec YML to path: %s", copySpecJSONPath)

		route, found := stepman.ReadRoute(steplibURI)
		if !found {
			failf("No route found for steplib (%s)", steplibURI)
		}

		sourceSpecJSONPth := stepman.GetStepSpecPath(route)
		if err := command.CopyFile(sourceSpecJSONPth, copySpecJSONPath); err != nil {
			failf("Failed to copy spec.json from (%s) to (%s), error: %s", sourceSpecJSONPth, copySpecJSONPath, err)
		}
	}

	return nil
}

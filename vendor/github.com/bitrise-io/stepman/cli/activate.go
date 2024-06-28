package cli

import (
	"fmt"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

var activateCommand = cli.Command{
	Name:  "activate",
	Usage: "Copy the step with specified --id, and --version, into provided path. If --version flag is not set, the latest version of the step will be used. If --copyyml flag is set, step.yml will be copied to the given path.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   CollectionKey + ", " + collectionKeyShort,
			Usage:  "Collection of step.",
			EnvVar: CollectionPathEnvKey,
		},
		cli.StringFlag{
			Name:  IDKey + ", " + idKeyShort,
			Usage: "Step id.",
		},
		cli.StringFlag{
			Name:  VersionKey + ", " + versionKeyShort,
			Usage: "Step version.",
		},
		cli.StringFlag{
			Name:  PathKey + ", " + pathKeyShort,
			Usage: "Path where the step will copied.",
		},
		cli.StringFlag{
			Name:  CopyYMLKey + ", " + copyYMLKeyShort,
			Usage: "Path where the activated step's step.yml will be copied.",
		},
		cli.BoolFlag{
			Name:  UpdateKey + ", " + updateKeyShort,
			Usage: "If flag is set, and collection doesn't contains the specified step, the collection will updated.",
		},
	},
	Action: func(c *cli.Context) error {
		if err := activate(c); err != nil {
			failf("Command failed: %s", err)
		}
		return nil
	},
}

func activate(c *cli.Context) error {
	stepLibURI := c.String(CollectionKey)
	if stepLibURI == "" {
		return fmt.Errorf("no steplib specified")
	}

	id := c.String(IDKey)
	if id == "" {
		return fmt.Errorf("no step ID specified")
	}

	path := c.String(PathKey)
	if path == "" {
		return fmt.Errorf("no destination path specified")
	}

	version := c.String(VersionKey)
	copyYML := c.String(CopyYMLKey)
	update := c.Bool(UpdateKey)
	logger := log.NewDefaultLogger(false)
	isOfflineMode := false

	return Activate(stepLibURI, id, version, path, copyYML, update, logger, isOfflineMode)
}

func Activate(stepLibURI, id, version, destination, destinationStepYML string, updateLibrary bool, log stepman.Logger, isOfflineMode bool) error {
	return stepman.Activate(stepLibURI, id, version, destination, destinationStepYML, updateLibrary, log, isOfflineMode)
}

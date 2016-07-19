package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/stepman/output"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func gitClone(uri, pth string) (string, error) {
	if uri == "" {
		return "", errors.New("Git Clone 'uri' missing")
	}
	if pth == "" {
		return "", errors.New("Git Clone 'pth' missing")
	}

	command := exec.Command("git", "clone", "--recursive", uri, pth)
	bytes, err := command.CombinedOutput()
	return string(bytes), err
}

func setupSteplib(steplibURI string, silent bool) error {
	logger := output.NewLogger(silent)

	if exist, err := stepman.RootExistForCollection(steplibURI); err != nil {
		return fmt.Errorf("Failed to check if routing exist for steplib (%s), error: %s", steplibURI, err)
	} else if exist {

		logger.Debugf("Steplib (%s) already initialized, ready to use", steplibURI)
		return nil
	}

	alias := stepman.GenerateFolderAlias()
	route := stepman.SteplibRoute{
		SteplibURI:  steplibURI,
		FolderAlias: alias,
	}

	// Cleanup
	isSuccess := false
	defer func() {
		if !isSuccess {
			if err := stepman.CleanupRoute(route); err != nil {
				logger.Errorf("Failed to cleanup routing for steplib (%s), error: %s", steplibURI, err)
			}
		}
	}()

	// Setup
	isLocalSteplib := strings.HasPrefix(steplibURI, "file://")

	pth := stepman.GetCollectionBaseDirPath(route)
	if !isLocalSteplib {
		if out, err := gitClone(steplibURI, pth); err != nil {
			return fmt.Errorf("Failed to setup steplib (%s), output: %s, error: %s", steplibURI, out, err)
		}
	} else {
		// Local spec path
		logger.Warn("Using local steplib")
		logger.Infof("Creating steplib dir: %s", pth)

		if err := os.MkdirAll(pth, 0777); err != nil {
			return fmt.Errorf("Failed to create steplib dir (%s), error: %s", pth, err)
		}

		logger.Info("Collection dir created - OK")
		stepLibPth := steplibURI
		if strings.HasPrefix(steplibURI, "file://") {
			stepLibPth = strings.TrimPrefix(steplibURI, "file://")
		}
		if err := cmdex.CopyDir(stepLibPth, pth, true); err != nil {
			return fmt.Errorf("Failed to copy dir (%s) to (%s), error: %s", stepLibPth, pth, err)
		}
	}

	if err := stepman.ReGenerateStepSpec(route); err != nil {
		return fmt.Errorf("Failed to re-generate steplib (%s), error: %s", steplibURI, err)
	}

	if err := stepman.AddRoute(route); err != nil {
		return fmt.Errorf("Failed to setup routing: %s", err)
	}

	isSuccess = true

	return nil
}

func setup(c *cli.Context) error {
	log.Debug("Setup")

	// Input validation
	steplibURI := c.String(CollectionKey)
	if steplibURI == "" {
		log.Fatal("No step collection specified")
	}

	copySpecJSONPath := c.String(CopySpecJSONKey)

	if c.IsSet(LocalCollectionKey) {
		log.Warn("'local' flag is deprecated")
		log.Warn("use 'file://' prefix in steplib path instead")
		fmt.Println()
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
	if err := setupSteplib(steplibURI, false); err != nil {
		log.Fatalf("Setup failed, error: %s", err)
	}

	// Copy spec.json
	if copySpecJSONPath != "" {
		log.Infof("Copying spec YML to path: %s", copySpecJSONPath)

		route, found := stepman.ReadRoute(steplibURI)
		if !found {
			log.Fatalf("No route found for steplib (%s)", steplibURI)
		}

		sourceSpecJSONPth := stepman.GetStepSpecPath(route)
		if err := cmdex.CopyFile(sourceSpecJSONPth, copySpecJSONPath); err != nil {
			log.Fatalf("Failed to copy spec.json from (%s) to (%s), error: %s", sourceSpecJSONPth, copySpecJSONPath, err)
		}
	}

	return nil
}

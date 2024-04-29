package stepman

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/stepman/models"
)

const filePathPrefix = "file://"

// Logger ...
type Logger interface {
	Debugf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
}

// SetupLibrary ...
func SetupLibrary(libraryURI string, log Logger) error {
	if exist, err := RootExistForLibrary(libraryURI); err != nil {
		return fmt.Errorf("failed to check if routing exist for library (%s), error: %s", libraryURI, err)
	} else if exist {
		return nil
	}

	alias := GenerateFolderAlias()
	route := SteplibRoute{
		SteplibURI:  libraryURI,
		FolderAlias: alias,
	}

	// Cleanup
	isSuccess := false
	defer func() {
		if !isSuccess {
			if err := CleanupRoute(route); err != nil {
				log.Warnf("Failed to cleanup routing for library (%s), error: %s", libraryURI, err)
			}
		}
	}()

	// Setup
	isLocalLibrary := strings.HasPrefix(libraryURI, filePathPrefix)

	pth := GetLibraryBaseDirPath(route)
	if !isLocalLibrary {
		if err := retry.Times(2).Wait(3 * time.Second).Try(func(attempt uint) error {
			repo, err := git.New(pth)
			if err != nil {
				return err
			}
			return repo.Clone(libraryURI).Run()
		}); err != nil {
			return fmt.Errorf("failed to clone library (%s), error: %s", libraryURI, err)
		}
	} else {
		// Local spec path
		if err := os.MkdirAll(pth, 0777); err != nil {
			return fmt.Errorf("failed to create library dir (%s), error: %s", pth, err)
		}

		libraryFilePath := libraryURI
		if strings.HasPrefix(libraryURI, filePathPrefix) {
			libraryFilePath = strings.TrimPrefix(libraryURI, filePathPrefix)
		}

		if err := command.CopyDir(libraryFilePath, pth, true); err != nil {
			return fmt.Errorf("failed to copy dir (%s) to (%s), error: %s", libraryFilePath, pth, err)
		}
	}

	if err := ReGenerateLibrarySpec(route); err != nil {
		return fmt.Errorf("failed to re-generate library (%s), error: %s", libraryURI, err)
	}

	if err := AddRoute(route); err != nil {
		return fmt.Errorf("failed to add routing, error: %s", err)
	}

	isSuccess = true

	return nil
}

// UpdateLibrary ...
func UpdateLibrary(libraryURI string, log Logger) (models.StepCollectionModel, error) {
	route, found := ReadRoute(libraryURI)
	if !found {
		if err := CleanupDanglingLibrary(libraryURI); err != nil {
			log.Warnf("Failed to cleaning up library (%s), error: %s", libraryURI, err)
		}
		return models.StepCollectionModel{}, fmt.Errorf("no route found for library: %s", libraryURI)
	}

	isLocalLibrary := strings.HasPrefix(libraryURI, filePathPrefix)

	if isLocalLibrary {
		if err := CleanupRoute(route); err != nil {
			return models.StepCollectionModel{}, fmt.Errorf("failed to cleanup route for library (%s), error: %s", libraryURI, err)
		}

		if err := SetupLibrary(libraryURI, log); err != nil {
			return models.StepCollectionModel{}, fmt.Errorf("failed to setup library (%s), error: %s", libraryURI, err)
		}
	} else {
		pth := GetLibraryBaseDirPath(route)
		if exists, err := pathutil.IsPathExists(pth); err != nil {
			return models.StepCollectionModel{}, fmt.Errorf("failed to check if library (%s) directory (%s) exist, error: %s", libraryURI, pth, err)
		} else if !exists {
			return models.StepCollectionModel{}, fmt.Errorf("library (%s) not initialized", libraryURI)
		}

		if err := retry.Times(2).Wait(3 * time.Second).Try(func(attempt uint) error {
			repo, err := git.New(pth)
			if err != nil {
				return err
			}
			return repo.Pull().Run()
		}); err != nil {
			return models.StepCollectionModel{}, fmt.Errorf("failed to pull library (%s), error: %s", libraryURI, err)
		}

		if err := ReGenerateLibrarySpec(route); err != nil {
			return models.StepCollectionModel{}, fmt.Errorf("failed to generate spec for library (%s), error: %s", libraryURI, err)
		}
	}

	return ReadStepSpec(libraryURI)
}

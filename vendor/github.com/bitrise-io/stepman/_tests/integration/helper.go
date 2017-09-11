package integration

import (
	"os"

	"github.com/bitrise-io/go-utils/command"
)

const defaultLibraryURI = "https://github.com/bitrise-io/bitrise-steplib.git"

func binPath() string {
	return os.Getenv("INTEGRATION_TEST_BINARY_PATH")
}

func cleanupLibrary(libraryURI string) error {
	cmd := command.New(binPath(), "delete", "--collection", libraryURI)
	return cmd.Run()
}

func setupLibrary(libraryURI string) error {
	cmd := command.New(binPath(), "setup", "--collection", libraryURI)
	return cmd.Run()
}

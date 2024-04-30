package cli

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/stepman"
)

func compressStep(log stepman.Logger, patchFilePath, targetExecutablePathLatest, targetExecutablePath string) error {
	patchFileExist, err := pathutil.IsPathExists(patchFilePath)
	if err != nil {
		return fmt.Errorf("failed to check if %s path exist: %w", patchFilePath, err)
	}
	if patchFileExist {
		if err := os.Remove(patchFilePath); err != nil {
			return fmt.Errorf("failed to remove existing patch file: %w", err)
		}
	}

	patchFromExist, err := pathutil.IsPathExists(targetExecutablePathLatest)
	if err != nil {
		return fmt.Errorf("failed to check if %s path exist: %s", targetExecutablePathLatest, err)
	}
	if !patchFromExist {
		return fmt.Errorf("Latest Step version used for patch (%s) not found", targetExecutablePathLatest)
	}

	if targetExecutablePath == "" {
		return nil
	}

	compressCmd := command.New("zstd", "--patch-from="+targetExecutablePathLatest, targetExecutablePath, "-o", patchFilePath)
	log.Debugf("[Stepman] compressing step $ %s", compressCmd.PrintableCommandArgs())
	out, err := compressCmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to compress with command (%s), output: %s", compressCmd.PrintableCommandArgs(), out)
	}

	if err := os.Remove(targetExecutablePath); err != nil {
		return fmt.Errorf("failed to remove uncompressed step executable: %w", err)
	}

	return nil
}

func uncompressStepFromCache(patchFromPath, targetVersionPatchPath, targetExecutablePath, checkSumPath string) error {
	for _, path := range []string{patchFromPath, targetVersionPatchPath, checkSumPath} {
		exist, err := pathutil.IsPathExists(path)
		if err != nil {
			return fmt.Errorf("failed to check if %s path exist: %s", path, err)
		}

		if !exist {
			return fmt.Errorf("%s not found in cache", path)
		}
	}

	decompressCmd := command.New("zstd", "-d", "--patch-from", patchFromPath, targetVersionPatchPath, "-o", targetExecutablePath)
	decompressCmd.SetStdout(nil).SetStderr(nil)

	exit, err := decompressCmd.RunAndReturnExitCode()
	if err != nil {
		return fmt.Errorf("failed to apply patch with command (%s), exit code: %d: %s", decompressCmd.PrintableCommandArgs(), exit, err)
	}

	return checkChecksum(targetExecutablePath, checkSumPath)
}

func writeChecksum(patchFilePath, checksumPath string) error {
	checksumExist, err := pathutil.IsPathExists(checksumPath)
	if err != nil {
		return fmt.Errorf("failed to check if path (%s) exist: %w", checksumPath, err)
	}
	if checksumExist {
		if err := os.Remove(checksumPath); err != nil {
			return fmt.Errorf("failed to remove checksum file: %w", err)
		}
	}

	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(patchFilePath)))
	if err := os.WriteFile(checksumPath, []byte(checksum), 0400); err != nil {
		return fmt.Errorf("Failed to write checksum (%s) to file %s: %w", checksum, checksumPath, err)
	}

	return nil
}

func checkChecksum(executablePath, checksumPath string) error {
	checksum, err := os.ReadFile(checksumPath)
	if err != nil {
		return fmt.Errorf("Failed to read checksum from file %s: %w", checksumPath, err)
	}

	calculatedChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(executablePath)))
	if string(checksum) != calculatedChecksum {
		return fmt.Errorf("Checksum mismatch %s expected %s, got %s", executablePath, checksum, calculatedChecksum)
	}

	return nil
}

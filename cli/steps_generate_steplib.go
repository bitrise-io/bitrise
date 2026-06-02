package cli

import (
	"fmt"
	"time"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/stepman/steplibrary/specgen"
	"github.com/urfave/cli"
)

func generateSteplib(c *cli.Context) error {
	steplibURL := c.String("steplib-url")
	outputDir := c.String("output")
	commitSHA := c.String("commit-sha")
	timestamp := c.String("timestamp")

	opts := specgen.Options{
		SteplibCommitSHA: commitSHA,
	}
	if timestamp != "" {
		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return fmt.Errorf("invalid --timestamp (expected RFC3339): %s", err)
		}
		opts.GeneratedAt = t
	}

	log.Infof("Generating V2 step library inventory")
	log.Infof("Steplib: %s", steplibURL)
	log.Infof("Output:  %s", outputDir)
	if commitSHA != "" {
		log.Infof("Commit override: %s", commitSHA)
	}
	if !opts.GeneratedAt.IsZero() {
		log.Infof("Time:    %s", opts.GeneratedAt.Format(time.RFC3339))
	}

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	stats, err := specgen.Generate(steplibURL, outputDir, opts, logger)
	if err != nil {
		return err
	}

	log.Donef("Generated %d steps (%d versions) — %d files, %d bytes in %s",
		stats.StepCount, stats.VersionCount,
		stats.FilesWritten, stats.BytesWritten, stats.Duration)
	return nil
}

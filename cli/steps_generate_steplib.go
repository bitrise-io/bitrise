package cli

import (
	"fmt"
	"time"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/stepman/steplibrary/indexgen"
	"github.com/urfave/cli"
)

func generateSteplib(c *cli.Context, logger log.Logger) error {
	steplibURL := c.String("steplib-url")
	outputDir := c.String("output")

	opts, err := buildIndexgenOpts(c.String("commit-sha"), c.String("timestamp"))
	if err != nil {
		return err
	}

	logger.Infof("Generating V2 step library inventory")
	logger.Infof("Steplib: %s", steplibURL)
	logger.Infof("Output:  %s", outputDir)
	if opts.SteplibCommitSHA != "" {
		logger.Infof("Commit override: %s", opts.SteplibCommitSHA)
	}
	if !opts.GeneratedAt.IsZero() {
		logger.Infof("Time:    %s", opts.GeneratedAt.Format(time.RFC3339))
	}

	stats, err := indexgen.Generate(steplibURL, outputDir, opts, logger)
	if err != nil {
		return err
	}

	logger.Donef("Generated %d steps (%d versions) — %d files, %d bytes in %s",
		stats.StepCount, stats.VersionCount,
		stats.FilesWritten, stats.BytesWritten, stats.Duration)
	return nil
}

func buildIndexgenOpts(commitSHA, timestamp string) (indexgen.Options, error) {
	opts := indexgen.Options{SteplibCommitSHA: commitSHA}
	if timestamp != "" {
		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return indexgen.Options{}, fmt.Errorf("invalid --timestamp (expected RFC3339): %s", err)
		}
		opts.GeneratedAt = t
	}
	return opts, nil
}

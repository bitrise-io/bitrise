// Package indexgen generates the V2 step library inventory tree from a
// bitrise-steplib source. The wire-format types it emits live in
// steplibrary/steplibindex.
//
// Generation stages the whole tree in a sibling temp directory, runs Validate
// against the staged tree, and only then atomically publishes it with a single
// rename. Validation is unconditional: an invalid
// staged tree is never published, so any existing inventory at the output dir
// is left untouched on a validation failure, and a successful Generate
// guarantees the published inventory passes Validate.
package indexgen

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
	"github.com/bitrise-io/stepman/stepman"
)

// Options control generator behavior. Zero values are filled with sensible
// defaults; callers (Generate, tests) override what they need.
type Options struct {
	// GeneratedAt is written to meta.json (RFC3339). Optional: when zero it
	// defaults to now. Normalized to UTC whole seconds either way. Tests set it
	// for deterministic output.
	GeneratedAt time.Time
	// SteplibCommitSHA is written to meta.json. Optional: the URI entry point
	// fills it from the clone's HEAD commit when empty.
	SteplibCommitSHA string
}

// Stats summarizes a successful generation.
type Stats struct {
	StepCount    int
	VersionCount int
	FilesWritten int
	BytesWritten int64
	Duration     time.Duration
}

// Generate sets up the steplib identified by steplibURI (cloning it into
// stepman's local cache via stepman.SetupLibrary if not already present) and
// writes the V2 inventory tree to outputDir. It is the URI-based entry point
// (the bitrise steps generate-steplib subcommand calls it); generateFromSteplibClone
// is the lower-level core that reads from an already-available filesystem.
func Generate(steplibURI, outputDir string, opts Options, log stepman.Logger) (Stats, error) {
	if err := stepman.SetupLibrary(steplibURI, log); err != nil {
		return Stats{}, fmt.Errorf("setup steplib %s: %w", steplibURI, err)
	}
	route, found := stepman.ReadRoute(steplibURI)
	if !found {
		return Stats{}, fmt.Errorf("no route for steplib %s after setup", steplibURI)
	}
	libDir := stepman.GetLibraryBaseDirPath(route)

	opts, err := withDefaults(opts, libDir)
	if err != nil {
		return Stats{}, err
	}
	return generateFromSteplibClone(os.DirFS(libDir), outputDir, opts, log)
}

// withDefaults fills zero-valued options. steplibDir is the steplib's git
// working copy, used to default SteplibCommitSHA to its HEAD commit when the
// caller didn't pin one; pass "" to leave SteplibCommitSHA untouched (e.g. when
// generating from a non-git fs.FS).
func withDefaults(o Options, steplibDir string) (Options, error) {
	if o.GeneratedAt.IsZero() {
		o.GeneratedAt = time.Now()
	}
	// meta.json's updated_at is RFC3339; normalize to UTC whole seconds so it
	// serializes as clean RFC3339, never with sub-second digits.
	o.GeneratedAt = o.GeneratedAt.UTC().Truncate(time.Second)
	if o.SteplibCommitSHA == "" && steplibDir != "" {
		sha, err := headCommitSHA(steplibDir)
		if err != nil {
			return o, fmt.Errorf("resolve steplib HEAD commit: %w", err)
		}
		o.SteplibCommitSHA = sha
	}
	return o, nil
}

// headCommitSHA returns the HEAD commit hash of the git working copy at dir.
func headCommitSHA(dir string) (string, error) {
	repo, err := git.New(dir)
	if err != nil {
		return "", err
	}
	return repo.RevParse("HEAD").RunAndReturnTrimmedCombinedOutput()
}

// generateFromSteplibClone reads a bitrise-steplib clone from inputFS and writes
// the V2 inventory tree to outputDir. The tree is staged in a sibling temp
// directory and published with a single rename on success, so a failure
// mid-generation never leaves a half-written inventory at outputDir; any
// existing tree at outputDir is replaced wholesale.
func generateFromSteplibClone(inputFS fs.FS, outputDir string, opts Options, log stepman.Logger) (_ Stats, err error) {
	start := time.Now()
	// No git dir here (fs.FS source); SteplibCommitSHA is defaulted by Generate.
	opts, err = withDefaults(opts, "")
	if err != nil {
		return Stats{}, err
	}

	steplibYML, err := readSteplibYML(inputFS)
	if err != nil {
		return Stats{}, fmt.Errorf("read steplib.yml: %w", err)
	}

	steps, err := collectSteps(inputFS, log)
	if err != nil {
		return Stats{}, err
	}

	staging, err := createStagingDir(outputDir)
	if err != nil {
		return Stats{}, err
	}
	defer func() {
		// On success staging has been renamed away, so RemoveAll is a no-op;
		// on failure it removes the partial tree.
		if rmErr := os.RemoveAll(staging); rmErr != nil {
			err = errors.Join(err, fmt.Errorf("clean staging dir %s: %w", staging, rmErr))
		}
	}()

	// The writer is rooted at the inventory root (staging); every path it gets
	// comes from steplibindex, which roots files under the format-version dir
	// (e.g. v2/), so the published outputDir contains <version>/{meta,index,steps}.
	w := newWriter(staging, fileutil.NewFileManager())
	if err := writeInventory(w, inputFS, steps, steplibYML, opts); err != nil {
		return Stats{}, err
	}

	// Validate the fully-staged tree before publishing: an invalid tree is never
	// published, so any existing inventory at outputDir is left untouched. staging
	// is the dir CONTAINING the version dir (v2/), the root Validate expects.
	if violations := Validate(os.DirFS(staging)); len(violations) > 0 {
		errs := make([]error, len(violations))
		for i, v := range violations {
			errs[i] = v
		}
		return Stats{}, fmt.Errorf("staged inventory failed validation (%d violations):\n%w", len(violations), errors.Join(errs...))
	}

	if err := publish(staging, outputDir); err != nil {
		return Stats{}, err
	}

	return buildStats(steps, w, start), nil
}

// createStagingDir makes a fresh staging directory as a sibling of outputDir
// (same filesystem, so publish's rename is atomic and never cross-device).
func createStagingDir(outputDir string) (string, error) {
	parent := filepath.Dir(outputDir)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return "", fmt.Errorf("create output parent %s: %w", parent, err)
	}
	staging, err := os.MkdirTemp(parent, ".indexgen-staging-*")
	if err != nil {
		return "", fmt.Errorf("create staging dir: %w", err)
	}
	return staging, nil
}

// writeInventory emits the full V2 tree through w: per-step source files, the
// derived index files, and meta.json.
func writeInventory(w *writer, inputFS fs.FS, steps []parsedStep, steplibYML models.StepCollectionModel, opts Options) error {
	for _, s := range steps {
		if err := writeStepFiles(w, inputFS, s); err != nil {
			return fmt.Errorf("write step %s: %w", s.id, err)
		}
	}
	if err := writeIndexFiles(w, steps); err != nil {
		return fmt.Errorf("write index files: %w", err)
	}
	meta := steplibindex.Meta{
		FormatVersion:     steplibindex.FormatVersion,
		UpdatedAt:         opts.GeneratedAt,
		SteplibCommitSHA:  opts.SteplibCommitSHA,
		SteplibSource:     steplibYML.SteplibSource,
		DownloadLocations: steplibYML.DownloadLocations,
	}
	if err := w.writeJSON(steplibindex.MetaPath().FS(), meta); err != nil {
		return fmt.Errorf("write meta.json: %w", err)
	}
	return nil
}

// publish atomically swaps the staged tree in for any existing one at outputDir.
func publish(stagingDir, outputDir string) error {
	if _, err := os.Stat(stagingDir); err != nil {
		return fmt.Errorf("staging dir (%s): %w", stagingDir, err)
	}
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("clear output dir %s: %w", outputDir, err)
	}
	if err := os.Rename(stagingDir, outputDir); err != nil {
		return fmt.Errorf("publish inventory to %s: %w", outputDir, err)
	}
	return nil
}

// buildStats summarizes a completed generation.
func buildStats(steps []parsedStep, w *writer, start time.Time) Stats {
	versionCount := 0
	for _, s := range steps {
		versionCount += len(s.versions)
	}
	return Stats{
		StepCount:    len(steps),
		VersionCount: versionCount,
		FilesWritten: w.fileCount,
		BytesWritten: w.byteCount,
		Duration:     time.Since(start),
	}
}

// Package indexgen generates the V2 step library inventory tree from a
// bitrise-steplib source. The wire-format types it emits live in
// steplibrary/steplibindex.
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
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
	"github.com/bitrise-io/stepman/stepman"
)

// Options control generator behavior. Zero values are filled with sensible
// defaults; callers (CLI / tests) override what they need.
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

// withDefaults fills zero-valued options. steplibDir is the steplib's git
// working copy, used to default SteplibCommitSHA to its HEAD commit when the
// caller didn't pin one; pass "" to leave SteplibCommitSHA untouched (e.g. when
// generating from a non-git fs.FS).
func withDefaults(o Options, steplibDir string) (Options, error) {
	if o.GeneratedAt.IsZero() {
		o.GeneratedAt = time.Now()
	}
	// meta.json's updated_at is RFC3339; normalize to UTC whole seconds so a
	// defaulted now() matches the precision of a -timestamp value (parsed as
	// RFC3339) and never serializes sub-second digits.
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

	// Stage in a sibling of outputDir (same filesystem, so the publish rename
	// is atomic and never cross-device).
	parent := filepath.Dir(outputDir)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return Stats{}, fmt.Errorf("create output parent %s: %w", parent, err)
	}
	staging, err := os.MkdirTemp(parent, ".steplib-gen-staging-*")
	if err != nil {
		return Stats{}, fmt.Errorf("create staging dir: %w", err)
	}
	defer func() {
		// On success staging has been renamed away, so RemoveAll is a no-op;
		// on failure it removes the partial tree.
		if rmErr := os.RemoveAll(staging); rmErr != nil {
			err = errors.Join(err, fmt.Errorf("clean staging dir %s: %w", staging, rmErr))
		}
	}()

	// Root the tree under the format-version dir (e.g. v2/) inside staging, so
	// the published outputDir contains <version>/{meta.json,spec,steps}.
	w := &writer{outputDir: filepath.Join(staging, steplibindex.VersionDir()), fw: realFileWriter{}, fm: fileutil.NewFileManager(), fileCount: 0, byteCount: 0}

	for _, s := range steps {
		if err := writeStepFiles(w, inputFS, s); err != nil {
			return Stats{}, fmt.Errorf("write step %s: %w", s.id, err)
		}
	}

	if err := writeIndexFiles(w, steps); err != nil {
		return Stats{}, fmt.Errorf("write index files: %w", err)
	}

	meta := steplibindex.Meta{
		FormatVersion:     steplibindex.FormatVersion,
		UpdatedAt:         opts.GeneratedAt,
		SteplibCommitSHA:  opts.SteplibCommitSHA,
		SteplibSource:     steplibYML.SteplibSource,
		DownloadLocations: steplibYML.DownloadLocations,
	}
	if err := w.writeJSON("meta.json", meta); err != nil {
		return Stats{}, fmt.Errorf("write meta.json: %w", err)
	}

	// Publish: swap the freshly staged tree in for any existing one.
	if err := os.RemoveAll(outputDir); err != nil {
		return Stats{}, fmt.Errorf("clear output dir %s: %w", outputDir, err)
	}
	if err := os.Rename(staging, outputDir); err != nil {
		return Stats{}, fmt.Errorf("publish inventory to %s: %w", outputDir, err)
	}

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
	}, nil
}

// Generate sets up the steplib identified by steplibURI (cloning it into
// stepman's local cache via stepman.SetupLibrary if not already present) and
// writes the V2 inventory tree to outputDir. It is the URI-based entry point
// used by the CLI; generateFromSteplibClone is the lower-level core that reads
// from an already-available filesystem.
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

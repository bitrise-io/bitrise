package steplibrary

import (
	"context"
	"fmt"
	"os"

	"github.com/bitrise-io/stepman/stepman"
)

// sourceProvider resolves the local directory holding a step's extracted
// source for a resolved version. Steplib depends on this interface so the
// source layer can be faked in tests without a function field.
type sourceProvider interface {
	stepSourceDir(ctx context.Context, step ResolvedStepVersion) (string, error)
}

// v1Source resolves step source through stepman's V1 on-disk cache.
type v1Source struct {
	steplibURI    string
	isOfflineMode bool
	log           stepman.Logger
}

// stepSourceDir returns the local directory holding the extracted step source
// for the given version. The cache is immutable per version, so a present dir
// is a hit and returned as-is; on a miss the source is fetched via
// stepman.DownloadStep — the same V1 code path that resolves download_locations
// (zip + git fallback), retries, verifies the git commit hash, and extracts
// into the V1 cache dir.
func (p v1Source) stepSourceDir(_ context.Context, step ResolvedStepVersion) (string, error) {
	route, found := stepman.ReadRoute(p.steplibURI)
	if !found {
		return "", fmt.Errorf("no route found for %s steplib", p.steplibURI)
	}
	cacheDir := stepman.GetStepCacheDirPath(route, step.ID, step.Version)

	// Cache hit: source already extracted (immutable per version).
	if info, err := os.Stat(cacheDir); err == nil && info.IsDir() {
		return cacheDir, nil
	}

	if p.isOfflineMode {
		return "", fmt.Errorf("step %s@%s is not in the local cache and offline mode is set", step.ID, step.Version)
	}

	collection, err := stepman.ReadStepSpec(p.steplibURI)
	if err != nil {
		return "", fmt.Errorf("read steplib spec for %s: %w", p.steplibURI, err)
	}

	stepModel, stepFound, versionFound := collection.GetStep(step.ID, step.Version)
	if !stepFound || !versionFound {
		return "", fmt.Errorf("%s steplib does not contain %s@%s", p.steplibURI, step.ID, step.Version)
	}
	commit := ""
	if stepModel.Source != nil {
		commit = stepModel.Source.Commit
	}

	if err := stepman.DownloadStep(p.steplibURI, collection, step.ID, step.Version, commit, p.log); err != nil {
		return "", fmt.Errorf("download step %s@%s: %w", step.ID, step.Version, err)
	}
	return cacheDir, nil
}

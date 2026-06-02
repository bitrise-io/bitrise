package steplibrary

import (
	"context"
	"fmt"
	"os"

	"github.com/bitrise-io/stepman/stepman"
)

// getStepSourceDir returns the local directory holding the extracted step
// source for the given version, using stepman's V1 on-disk cache. The cache is
// immutable per version, so a present dir is a hit and returned as-is; on a
// miss the source is fetched via stepman.DownloadStep — the same V1 code path
// that resolves download_locations (zip + git fallback), retries, verifies the
// git commit hash, and extracts into the V1 cache dir.
func (s *Steplib) getStepSourceDir(_ context.Context, step ResolvedStepVersion) (string, error) {
	route, found := stepman.ReadRoute(s.steplibURI)
	if !found {
		return "", fmt.Errorf("no route found for %s steplib", s.steplibURI)
	}
	cacheDir := stepman.GetStepCacheDirPath(route, step.ID, step.Version)

	// Cache hit: source already extracted (immutable per version).
	if info, err := os.Stat(cacheDir); err == nil && info.IsDir() {
		return cacheDir, nil
	}

	if s.isOfflineMode {
		return "", fmt.Errorf("step %s@%s is not in the local cache and offline mode is set", step.ID, step.Version)
	}

	collection, err := stepman.ReadStepSpec(s.steplibURI)
	if err != nil {
		return "", fmt.Errorf("read steplib spec for %s: %w", s.steplibURI, err)
	}

	stepModel, stepFound, versionFound := collection.GetStep(step.ID, step.Version)
	if !stepFound || !versionFound {
		return "", fmt.Errorf("%s steplib does not contain %s@%s", s.steplibURI, step.ID, step.Version)
	}
	commit := ""
	if stepModel.Source != nil {
		commit = stepModel.Source.Commit
	}

	if err := stepman.DownloadStep(s.steplibURI, collection, step.ID, step.Version, commit, s.log); err != nil {
		return "", fmt.Errorf("download step %s@%s: %w", step.ID, step.Version, err)
	}
	return cacheDir, nil
}

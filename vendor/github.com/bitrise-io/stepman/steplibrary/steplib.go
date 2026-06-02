package steplibrary

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/stepman"
)

type Steplib struct {
	log stepman.Logger
	// steplibURI is the steplib *identity* — the URI the user references in
	// bitrise.yml (e.g. the official git URL). It keys the V1 on-disk cache
	// and route used by source-fallback (see source.go) and is reported as
	// StepInfoModel.Library. It is NOT the URL the V2 inventory is fetched
	// from; that is the inventory URL held by the HTTP API.
	steplibURI       string
	isOfflineMode    bool
	api              API
	fileManager      fileutil.FileManager
	fetcher          httpfetch.Client
	fetchSourceDirFn func(ctx context.Context, step ResolvedStepVersion) (string, error)
}

type ActivateOutputPaths struct {
	YMLPath, CodePath string
}

// New builds a Steplib. steplibURI is the steplib identity (the user's
// bitrise.yml URI, used for the V1 cache and source fallback); inventoryURL is
// the base URL the V2 inventory JSON is fetched from. They differ for the
// official steplib, whose git identity is rewritten to a compiled-in V2 host.
func New(log stepman.Logger, steplibURI, inventoryURL string, isOfflineMode bool, fileManager fileutil.FileManager) *Steplib {
	api := NewHTTPAPI(inventoryURL, v2CacheDir(inventoryURL), nil, log)
	s := &Steplib{
		log:              log,
		steplibURI:       steplibURI,
		isOfflineMode:    isOfflineMode,
		api:              api,
		fileManager:      fileManager,
		fetcher:          httpfetch.NewClient(nil, log),
		fetchSourceDirFn: nil,
	}
	s.fetchSourceDirFn = s.getStepSourceDir
	return s
}

// v2CacheDir returns a stable on-disk cache directory for a given steplib URL.
// Keyed by a sha256 prefix so different URLs don't collide and the directory
// name is filesystem-safe.
func v2CacheDir(steplibURI string) string {
	sum := sha256.Sum256([]byte(steplibURI))
	return filepath.Join(stepman.GetStepmanDirPath(), "v2-cache", hex.EncodeToString(sum[:8]))
}

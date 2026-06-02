package steplibrary

import (
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/stepman"
)

type Steplib struct {
	log stepman.Logger
	// steplibURI is the steplib *identity* — the URI the user references in
	// bitrise.yml (e.g. the official git URL). It is reported as
	// StepInfoModel.Library. It is NOT the URL the V2 inventory is fetched
	// from; that is the inventory URL held by the HTTP API.
	steplibURI  string
	api         API
	fileManager fileutil.FileManager
	fetcher     httpfetch.Client
	source      sourceProvider
}

type ActivateOutputPaths struct {
	YMLPath, CodePath string
}

// New builds a Steplib. steplibURI is the steplib identity (the user's
// bitrise.yml URI, used for the V1 cache and source fallback); inventoryURL is
// the base URL the V2 inventory JSON is fetched from. They differ for the
// official steplib, whose git identity is rewritten to a compiled-in V2 host.
func New(log stepman.Logger, steplibURI, inventoryURL string, isOfflineMode bool, fileManager fileutil.FileManager) *Steplib {
	return &Steplib{
		log:         log,
		steplibURI:  steplibURI,
		api:         NewHTTPAPI(inventoryURL, httpfetch.NewClient(log)),
		fileManager: fileManager,
		fetcher:     httpfetch.NewClient(log),
		source:      v1Source{steplibURI: steplibURI, isOfflineMode: isOfflineMode, log: log},
	}
}

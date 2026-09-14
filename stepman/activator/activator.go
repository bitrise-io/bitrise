package activator

import (
	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/steplibrary"
	"github.com/bitrise-io/stepman/stepman"
)

const (
	bitriseSteplibURL    = "https://github.com/bitrise-io/bitrise-steplib.git"
	bitriseSteplibAPIURL = "https://steplib.bitrise.io/api"
)

// Options is the configuration of an Activator.
type Options struct {
	// DisableSteplibAPI routes canonical Bitrise StepLib steps through a local
	// git clone instead of the StepLib API. Custom/self-hosted StepLibs use
	// the git-clone path either way.
	DisableSteplibAPI bool

	// SteplibAPIURL is the base URL of the StepLib API. Empty means the
	// canonical Bitrise inventory.
	SteplibAPIURL string

	// DisablePrecompiled builds every step from source, even where the library
	// offers a prebuilt executable.
	DisablePrecompiled bool

	// PrecompiledStorageURLs are the base URLs tried in order for precompiled
	// executables. Empty means steplib.DefaultPrecompiledStorageURLs.
	PrecompiledStorageURLs []string

	// IsOfflineMode forbids network access, restricting activation to what is
	// already in the local StepLib cache. It is not supported together with
	// the StepLib API, which has no local inventory to read from.
	IsOfflineMode bool
}

// withDefaults fills in the values that Options leaves optional
func withDefaults(opts Options) Options {
	if opts.SteplibAPIURL == "" {
		opts.SteplibAPIURL = bitriseSteplibAPIURL
	}
	return opts
}

// Activator activates steps.
type Activator struct {
	log     stepman.Logger
	opts    Options
	fetcher httpfetch.Client

	// library serves the StepLib inventory
	library steplibrary.Client
}

// New builds an Activator. Reuse the returned value for every step of a run:
// building one per step is what this constructor exists to avoid.
func New(log stepman.Logger, opts Options) *Activator {
	opts = withDefaults(opts)
	clients, err := httpfetch.NewCachingClient(log)
	if err != nil {
		log.Warnf("Continuing without the in-memory inventory cache: %s", err)
		plain := httpfetch.NewClient(log)
		clients = httpfetch.Clients{Caching: plain, Passthrough: plain}
	}

	return &Activator{
		log:     log,
		opts:    opts,
		fetcher: clients.Passthrough,
		library: steplibrary.New(log, opts.SteplibAPIURL, clients.Caching),
	}
}

// useSteplibAPIFor reports whether steplibURI's steps are served by the StepLib API.
func (a *Activator) useSteplibAPIFor(steplibURI string) bool {
	return !a.opts.DisableSteplibAPI && steplibURI == bitriseSteplibURL
}

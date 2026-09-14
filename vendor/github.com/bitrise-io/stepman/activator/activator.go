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

// Options is the configuration of an Activator. It is resolved once, when the
// Activator is built, rather than re-read on every step activation.
//
// The flags are opt-outs, so the zero value is what production wants: the
// StepLib V2 API and prebuilt executables both on.
type Options struct {
	// DisableSteplibAPI routes canonical Bitrise StepLib steps through a local
	// git clone instead of the StepLib V2 API. Custom/self-hosted StepLibs use
	// the git-clone path either way.
	DisableSteplibAPI bool

	// SteplibAPIURL is the base URL of the StepLib V2 API. Empty means the
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
	// the StepLib V2 API, which has no local inventory to read from.
	IsOfflineMode bool
}

// withDefaults fills in the values that Options leaves optional, so the rest of
// the package can rely on them being set. PrecompiledStorageURLs is not among
// them: steplib owns that list and defaults it itself.
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

	// library serves the StepLib V2 inventory. Always present, including when
	// UseSteplibAPI is off: whether a given step may use it depends on that
	// step's StepLib source too, so the choice is made per activation rather
	// than here. See useSteplibAPIFor.
	library steplibrary.Client
}

// New builds an Activator. Reuse the returned value for every step of a run:
// building one per step is what this constructor exists to avoid.
func New(log stepman.Logger, opts Options) *Activator {
	opts = withDefaults(opts)
	fetcher := httpfetch.NewClient(log)

	return &Activator{
		log:     log,
		opts:    opts,
		fetcher: fetcher,
		library: steplibrary.New(log, opts.SteplibAPIURL, fetcher),
	}
}

// useSteplibAPIFor reports whether steplibURI's steps are served by the StepLib
// V2 API. The API only hosts the canonical Bitrise StepLib, so custom StepLibs
// stay on the git-clone path regardless of configuration.
func (a *Activator) useSteplibAPIFor(steplibURI string) bool {
	return !a.opts.DisableSteplibAPI && steplibURI == bitriseSteplibURL
}

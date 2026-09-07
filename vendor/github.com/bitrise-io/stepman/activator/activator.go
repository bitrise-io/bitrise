package activator

import (
	"os"
	"strings"

	"github.com/bitrise-io/stepman/activator/steplib"
	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/steplibrary"
	"github.com/bitrise-io/stepman/stepman"
)

const (
	bitriseSteplibURL    = "https://github.com/bitrise-io/bitrise-steplib.git"
	bitriseSteplibAPIURL = "https://steplib.bitrise.io/api"

	useSteplibAPIEnv               = "BITRISE_STEPLIB_USE_API"
	precompiledStepsEnv            = "BITRISE_STEPLIB_USE_BINARY"
	precompiledStepsStorageURLsEnv = "BITRISE_STEPLIB_STORAGE_URLS"
)

// Options is the configuration of an Activator. It is resolved once, when the
// Activator is built, rather than re-read on every step activation.
type Options struct {
	// UseSteplibAPI activates steps of the canonical Bitrise StepLib over the
	// StepLib V2 API instead of a local git clone. It has no effect on
	// custom/self-hosted StepLibs, which always use the git-clone path.
	UseSteplibAPI bool

	// SteplibAPIURL is the base URL of the StepLib V2 API. Empty means the
	// canonical Bitrise inventory.
	SteplibAPIURL string

	// UsePrecompiled allows activating a step from a prebuilt executable
	// instead of downloading and building its source.
	UsePrecompiled bool

	// PrecompiledStorageURLs are the base URLs tried in order for precompiled
	// executables. Empty means steplib.DefaultPrecompiledStorageURLs.
	PrecompiledStorageURLs []string
}

// OptionsFromEnv resolves the Options that stepman reads from the environment.
// Both feature flags default to on and are disabled by setting them to "false"
// or "0". Callers that configure the activator themselves do not need this.
func OptionsFromEnv() Options {
	return Options{
		UseSteplibAPI:          !isEnvDisabled(useSteplibAPIEnv),
		SteplibAPIURL:          "",
		UsePrecompiled:         !isEnvDisabled(precompiledStepsEnv),
		PrecompiledStorageURLs: splitStorageURLs(os.Getenv(precompiledStepsStorageURLsEnv)),
	}
}

// isEnvDisabled reports whether an opt-out flag is switched off. Any other
// value, including an unset variable, leaves the feature enabled.
func isEnvDisabled(key string) bool {
	return os.Getenv(key) == "false" || os.Getenv(key) == "0"
}

// splitStorageURLs parses the comma-separated storage URL override. An empty
// override yields nil, which withDefaults resolves to the built-in list.
func splitStorageURLs(override string) []string {
	if override == "" {
		return nil
	}
	return strings.Split(override, ",")
}

// withDefaults fills in the values that Options leaves optional, so the rest of
// the package can rely on them being set.
func withDefaults(opts Options) Options {
	if opts.SteplibAPIURL == "" {
		opts.SteplibAPIURL = bitriseSteplibAPIURL
	}
	if len(opts.PrecompiledStorageURLs) == 0 {
		opts.PrecompiledStorageURLs = steplib.DefaultPrecompiledStorageURLs
	}
	return opts
}

// Activator activates steps. It owns the collaborators that are worth keeping
// alive across activations — most importantly a single HTTP client, so that
// every step in a run shares one connection pool — and is therefore meant to be
// built once per workflow run and reused, not per step.
type Activator struct {
	log     stepman.Logger
	opts    Options
	fetcher httpfetch.Client

	// library serves the StepLib V2 inventory. It is built even when
	// UseSteplibAPI is off, because whether a given step may use it also
	// depends on that step's StepLib source; see useSteplibAPIFor.
	library *steplibrary.Client
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
	return a.opts.UseSteplibAPI && steplibURI == bitriseSteplibURL
}

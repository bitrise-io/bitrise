package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/stepman/activator"
	"github.com/bitrise-io/stepman/stepid"
)

// stepActivator wraps the stepman activator. Both it and the activator it
// holds are built once per run, so every step shares one HTTP client.
type stepActivator struct {
	activator *activator.Activator
	logger    log.Logger
}

func newStepActivator(logger log.Logger) stepActivator {
	return stepActivator{
		activator: activator.New(logger, activatorOptionsFromEnv()),
		logger:    logger,
	}
}

// activatorOptionsFromEnv resolves how steps are activated for this run.
// Stepman takes these as explicit options and reads no environment of its own,
// so the mapping lives here, at the edge, and happens once per run.
//
// Both feature flags default to on; only "false" or "0" opts out.
func activatorOptionsFromEnv() activator.Options {
	return activator.Options{
		UseSteplibAPI:          !isEnvDisabled(configs.SteplibUseAPIEnvKey),
		SteplibAPIURL:          "",
		UsePrecompiled:         !isEnvDisabled(configs.SteplibUseBinaryEnvKey),
		PrecompiledStorageURLs: splitStorageURLs(os.Getenv(configs.SteplibStorageURLsEnvKey)),
	}
}

// isEnvDisabled reports whether an opt-out flag is switched off. Any other
// value, including an unset variable, leaves the feature enabled.
func isEnvDisabled(key string) bool {
	return os.Getenv(key) == "false" || os.Getenv(key) == "0"
}

// splitStorageURLs parses the comma-separated storage URL override. An empty
// override yields nil, which leaves stepman's built-in list in place.
func splitStorageURLs(override string) []string {
	if override == "" {
		return nil
	}
	return strings.Split(override, ",")
}

// Note: even when err != nil, the ActivatedStep struct will be returned with a valid DidStepLibUpdate value
func (a stepActivator) activateStep(
	stepIDData stepid.CanonicalID,
	isStepLibUpdated bool,
	stepDir string, // $TMPDIR/bitrise/step_src
	workDir string, // $TMPDIR/bitrise
) (activator.ActivatedStep, error) {
	if stepIDData.SteplibSource == "path" {
		log.Debugf("[BITRISE_CLI] - Local step found: (path:%s)", stepIDData.IDorURI)

		activatedStep, err := activator.ActivatePathRefStep(
			a.logger,
			stepIDData,
			stepDir,
			workDir,
		)
		if err != nil {
			return activator.ActivatedStep{ActivationType: activator.ActivationTypePathRef}, fmt.Errorf("activate local step: %w", err)
		}
		return activatedStep, nil
	} else if stepIDData.SteplibSource == "git" {
		log.Debugf("[BITRISE_CLI] - Remote step, with direct git uri: (uri:%s) (tag-or-branch:%s)", stepIDData.IDorURI, stepIDData.Version)

		activatedStep, err := activator.ActivateGitRefStep(
			a.logger,
			stepIDData,
			stepDir,
			workDir,
		)
		if err != nil {
			return activator.ActivatedStep{ActivationType: activator.ActivationTypeGitRef}, fmt.Errorf("activate git step reference: %w", err)
		}
		return activatedStep, nil
	} else if stepIDData.SteplibSource != "" {
		activatedStep, err := a.activator.ActivateSteplibRefStep(
			stepIDData,
			stepDir,
			workDir,
			isStepLibUpdated,
		)
		if err != nil {
			// Note: we return the partial result on purpose because DidStepLibUpdate is important
			// even in case of an error
			return activatedStep, fmt.Errorf("activate steplib step: %w", err)
		}
		return activatedStep, nil
	} else {
		return activator.ActivatedStep{}, fmt.Errorf("invalid stepIDData: no SteplibSource or LocalPath defined (%v)", stepIDData)
	}
}

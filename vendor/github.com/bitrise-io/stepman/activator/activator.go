package activator

import "github.com/bitrise-io/stepman/activator/result"

// Backwards-compatible aliases for the activation result types. The canonical
// definitions live in activator/result; this file lets downstream consumers
// (notably the bitrise CLI) keep importing `activator.ActivationType` and
// `activator.ActivatedStep` without churn.
type (
	ActivatedStep  = result.ActivatedStep
	ActivationType = result.ActivationType
)

const (
	ActivationTypeSteplibExecutable = result.ActivationTypeSteplibExecutable
	ActivationTypeSteplibSource     = result.ActivationTypeSteplibSource
	ActivationTypePathRef           = result.ActivationTypePathRef
	ActivationTypeGitRef            = result.ActivationTypeGitRef
)

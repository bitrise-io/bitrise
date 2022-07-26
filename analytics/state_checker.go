package analytics

import (
	"github.com/bitrise-io/go-utils/v2/env"
)

// DisabledEnvKey controls both the old (analytics plugin) and new (v2) implementations
const DisabledEnvKey = "BITRISE_ANALYTICS_DISABLED"

// V2DisabledEnvKey controls only the new (v2) implementation
const V2DisabledEnvKey = "BITRISE_ANALYTICS_V2_DISABLED"

// StateChecker ...
type StateChecker interface {
	Enabled() bool
}

type stateChecker struct {
	envRepository env.Repository
}

// NewStateChecker ...
func NewStateChecker(repository env.Repository) StateChecker {
	return stateChecker{envRepository: repository}
}

// Enabled ...
func (s stateChecker) Enabled() bool {
	if s.envRepository.Get(V2DisabledEnvKey) == "true" {
		return false
	}

	return s.envRepository.Get(DisabledEnvKey) != "true"
}

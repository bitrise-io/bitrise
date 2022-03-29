package analytics

import (
	"github.com/bitrise-io/go-utils/v2/env"
)

// DisabledEnvKey ...
const DisabledEnvKey = "BITRISE_ANALYTICS_DISABLED"

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
	return s.envRepository.Get(DisabledEnvKey) != "true"
}

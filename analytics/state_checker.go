package analytics

import (
	"github.com/bitrise-io/go-utils/v2/env"
)

const (
	DisabledEnvKey = "BITRISE_ANALYTICS_DISABLED"

	trueEnv = "true"
)

type StateChecker interface {
	Enabled() bool
}

type stateChecker struct {
	envRepository env.Repository
}

func NewStateChecker(repository env.Repository) StateChecker {
	return stateChecker{envRepository: repository}
}

func (s stateChecker) Enabled() bool {
	return s.envRepository.Get(DisabledEnvKey) != trueEnv
}

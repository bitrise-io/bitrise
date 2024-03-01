package analytics

import (
	"github.com/bitrise-io/go-utils/v2/env"
)

const (
	// DisabledEnvKey controls both the old (analytics plugin) and new (v2) implementations
	DisabledEnvKey = "BITRISE_ANALYTICS_DISABLED"
	// V2DisabledEnvKey controls only the new (v2) implementation
	V2DisabledEnvKey = "BITRISE_ANALYTICS_V2_DISABLED"
	// V2AsyncDisabledEnvKey can be used to disable the default async queries
	V2AsyncDisabledEnvKey = "BITRISE_ANALYTICS_V2_ASYNC_DISABLED"

	trueEnv = "true"
)

type StateChecker interface {
	Enabled() bool
	UseAsync() bool
}

type stateChecker struct {
	envRepository env.Repository
}

func NewStateChecker(repository env.Repository) StateChecker {
	return stateChecker{envRepository: repository}
}

func (s stateChecker) Enabled() bool {
	if s.envRepository.Get(V2DisabledEnvKey) == trueEnv {
		return false
	}

	return s.envRepository.Get(DisabledEnvKey) != trueEnv
}

func (s stateChecker) UseAsync() bool {
	return s.envRepository.Get(V2AsyncDisabledEnvKey) != trueEnv
}

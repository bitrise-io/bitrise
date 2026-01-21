package configs

import (
	"os"
	"strings"
)

// IsEdgeStack checks if the current stack is an edge stack based on environment variables. If we run outside of a Bitrise CI env, it also returns false.
func IsEdgeStack() bool {
	if stackStatus, ok := os.LookupEnv("BITRISEIO_STACK_STATUS"); ok && strings.Contains(stackStatus, "edge") {
		return true
	}

	return false
}

package update

import (
	"strings"

	"github.com/bitrise-io/bitrise/v3/log/corelog"
	"github.com/bitrise-io/bitrise/v3/output"
)

const (
	// Either one opts out, set to any value, empty included.
	EnvNoUpdateNotifier    = "BITRISE_NO_UPDATE_NOTIFIER"
	EnvNoCLIUpdateNotifier = "BITRISE_CLI_NO_UPDATE_NOTIFIER"

	updateCommandName = "update"
)

// Eligibility is passed as a struct so the policy stays a pure function: no
// cobra, no environment of its own, no network.
type Eligibility struct {
	CurrentVersion string
	// CommandPath is cobra's cmd.CommandPath(), e.g. "bitrise yml update".
	CommandPath string
	// OutputFormat is the value of the root --output flag.
	OutputFormat string
	// LogFormat is `run --output-format`, empty for commands without it.
	LogFormat   string
	Quiet       bool
	StderrIsTTY bool
	CIMode      bool
	// A nil LookupEnv reads as none of the opt-out variables being set.
	LookupEnv func(string) (string, bool)
}

// The version check comes first, so a build that cannot be compared against the
// releases costs no request and no cache write.
func IsEligible(e Eligibility) bool {
	if !isComparable(e.CurrentVersion) {
		return false
	}
	if e.CIMode || e.Quiet || !e.StderrIsTTY {
		return false
	}
	// corelog is the leaf that owns the logger type names; the log package above it
	// would pull the workflow models into this one.
	if isMachineReadableOutput(e.OutputFormat) || e.LogFormat == string(corelog.JSONLogger) {
		return false
	}
	// `update` reports the versions it found itself, and the process still carries
	// the version it started with after a successful one.
	if isSelfUpdateCommand(e.CommandPath) {
		return false
	}
	if e.LookupEnv != nil {
		for _, key := range []string{EnvNoUpdateNotifier, EnvNoCLIUpdateNotifier} {
			if _, ok := e.LookupEnv(key); ok {
				return false
			}
		}
	}
	return true
}

func isMachineReadableOutput(format string) bool {
	return format == output.FormatJSON || format == output.FormatYML
}

// Matched by depth, not by the whole path: the first segment is the binary name,
// which the user can rename.
func isSelfUpdateCommand(commandPath string) bool {
	segments := strings.Fields(commandPath)
	return len(segments) == 2 && segments[1] == updateCommandName
}

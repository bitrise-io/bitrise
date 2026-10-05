package update

import (
	"strings"

	"github.com/bitrise-io/bitrise/v3/log/corelog"
	"github.com/bitrise-io/bitrise/v3/output"
)

const (
	// EnvNoUpdateNotifier and EnvNoCLIUpdateNotifier opt out of the notice.
	// Either one counts, set to any value, empty included.
	EnvNoUpdateNotifier    = "BITRISE_NO_UPDATE_NOTIFIER"
	EnvNoCLIUpdateNotifier = "BITRISE_CLI_NO_UPDATE_NOTIFIER"

	updateCommandName = "update"
)

// Eligibility is everything the notice policy decides on. It is passed as a
// struct so the policy stays a pure function: no cobra, no environment of its
// own, no network.
type Eligibility struct {
	CurrentVersion string
	// CommandPath is the path of the command that ran, cobra's
	// cmd.CommandPath(), for example "bitrise yml update".
	CommandPath string
	// OutputFormat is the value of the root --output flag.
	OutputFormat string
	// LogFormat is the value of `run --output-format`, empty for the commands
	// that do not have it.
	LogFormat   string
	Quiet       bool
	StderrIsTTY bool
	CIMode      bool
	// LookupEnv reads the opt-out variables, os.LookupEnv in production. A nil
	// value reads as none of them being set.
	LookupEnv func(string) (string, bool)
}

// IsEligible reports whether the update notice may be checked for and printed.
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

// isSelfUpdateCommand matches the root's own `update` by its depth rather than
// by the whole path: the first segment is the binary name, which the user can
// rename, and every other `update` belongs to a command group that updates
// something else.
func isSelfUpdateCommand(commandPath string) bool {
	segments := strings.Fields(commandPath)
	return len(segments) == 2 && segments[1] == updateCommandName
}

package cli

import (
	"errors"
	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/spf13/cobra"
)

// envmanCommand exists only to register "envman" so help and completion list it.
// Real invocations are intercepted before cobra by envmanPassthrough (see Run),
// which forwards them to runEnvman, so this RunE is unreachable. DisableFlagParsing
// means that if cobra ever did dispatch here, it would hit the guard below rather
// than failing on envman's own flags first.
var envmanCommand = &cobra.Command{
	Use:                "envman",
	Short:              "Runs an envman command.",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("envman must be dispatched before cobra by envmanPassthrough; reaching this point is a command dispatch bug")
	},
}

// envmanPassthrough reports whether the invocation targets the envman command
// (the first non-global-flag token is "envman") and, if so, returns the args
// that follow it, to be forwarded verbatim.
func envmanPassthrough(rawArgs []string) ([]string, bool) {
	i := commandTokenIndex(rawArgs)
	if i < len(rawArgs) && rawArgs[i] == envmanCommand.Name() {
		return rawArgs[i+1:], true
	}
	return nil, false
}

func runEnvman(root *cobra.Command, rawArgs []string, envmanArgs []string) {
	applyGlobalFlagsFromArgs(root, rawArgs)
	if err := before(root, nil); err != nil {
		failf(err.Error())
	}

	logCommandParameters(envmanCommand)

	if err := runCommandWith("envman", envmanArgs); err != nil {
		failf("Command failed, error: %s", err)
	}
}

func runCommandWith(toolName string, args []string) error {
	cmd := command.NewWithStandardOuts(toolName, args...).SetStdin(os.Stdin)
	return cmd.Run()
}

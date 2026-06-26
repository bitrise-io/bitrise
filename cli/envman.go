package cli

import (
	"os"

	"github.com/bitrise-io/bitrise/v2/cli/legacy"
	"github.com/bitrise-io/go-utils/command"
	"github.com/spf13/cobra"
)

// envmanCommand registers "envman" so help and completion list it. Real
// invocations are normally intercepted before cobra by envmanPassthrough (see
// Run), which strips the leading bitrise global flags and forwards the args after
// "envman" to runEnvman. cobra reaches this RunE only when a non-global flag
// precedes "envman" (so the pre-cobra check doesn't fire); DisableFlagParsing lets
// us forward the leftover args (the stray flag included) verbatim, so envman itself
// reports the unexpected flag rather than bitrise emitting an internal error.
var envmanCommand = &cobra.Command{
	Use:                "envman",
	Short:              "Runs an envman command.",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logCommandParameters(cmd)
		if err := runCommandWith("envman", args); err != nil {
			failf("Command failed, error: %s", err)
		}
		return nil
	},
}

// envmanPassthrough reports whether the invocation targets the envman command
// (the first non-global-flag token is "envman") and, if so, returns the args
// that follow it, to be forwarded verbatim.
func envmanPassthrough(rawArgs []string) ([]string, bool) {
	i := legacy.CommandTokenIndex(rawArgs, globalFlagNames)
	if i < len(rawArgs) && rawArgs[i] == envmanCommand.Name() {
		return rawArgs[i+1:], true
	}
	return nil, false
}

func runEnvman(root *cobra.Command, rawArgs []string, envmanArgs []string) {
	legacy.ApplyGlobalFlagsFromArgs(root, rawArgs, globalFlagNames)
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

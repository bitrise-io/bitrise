package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/cli/legacy"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func Test_loggerParameters(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		wantIsRunCommand bool
		wantOutputFormat log.LoggerType
	}{
		{
			name:             "Empty test",
			args:             []string{},
			wantIsRunCommand: false,
			wantOutputFormat: "",
		},
		{
			name:             "Run command",
			args:             []string{"run"},
			wantIsRunCommand: true,
		},
		{
			name:             "Output format json with one dash syntax",
			args:             []string{"-output-format", "json"},
			wantOutputFormat: "json",
		},
		{
			name:             "Output format console with two dash syntax",
			args:             []string{"--output-format", "console"},
			wantOutputFormat: "console",
		},
		{
			name:             "Output format json value with one dash syntax",
			args:             []string{"-output-format=json"},
			wantOutputFormat: "json",
		},
		{
			name:             "Output format console value with two dash syntax",
			args:             []string{"--output-format=console"},
			wantOutputFormat: "console",
		},
		{
			name:             "Output format invalid syntax",
			args:             []string{"-output-format", "--log-level"},
			wantOutputFormat: "",
		},
		{
			name:             "Output format invalid value",
			args:             []string{"-output-format", "invalid"},
			wantOutputFormat: "",
		},
		{
			name:             "Invalid flag",
			args:             []string{"-output-format-invalid=json"},
			wantOutputFormat: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isRunCommand, outputFormat := loggerParameters(tt.args)
			assert.Equalf(t, tt.wantIsRunCommand, isRunCommand, "loggerParameters(%v)", tt.args)
			assert.Equalf(t, tt.wantOutputFormat, outputFormat, "loggerParameters(%v)", tt.args)
		})
	}
}

func Test_detectPlugin(t *testing.T) {
	root := newRootCommand()

	tests := []struct {
		name       string
		args       []string
		wantName   string
		wantArgs   []string
		wantPlugin bool
	}{
		{
			name:       "plugin with no args",
			args:       []string{":analytics"},
			wantName:   "analytics",
			wantArgs:   []string{},
			wantPlugin: true,
		},
		{
			// Global-named flags after the plugin name belong to the plugin and
			// must be forwarded, not consumed as bitrise globals.
			name:       "flag sharing a global name after the plugin name is forwarded to the plugin",
			args:       []string{":analytics", "--debug", "on"},
			wantName:   "analytics",
			wantArgs:   []string{"--debug", "on"},
			wantPlugin: true,
		},
		{
			name:       "leading global flag is skipped, trailing flags forwarded",
			args:       []string{"--debug", ":analytics", "--ci", "on"},
			wantName:   "analytics",
			wantArgs:   []string{"--ci", "on"},
			wantPlugin: true,
		},
		{
			name:       "known command with a colon arg stays a command",
			args:       []string{"run", "a:b"},
			wantPlugin: false,
		},
		{
			name:       "known command after a leading global flag",
			args:       []string{"--ci", "run"},
			wantPlugin: false,
		},
		{
			name:       "command alias is a known command",
			args:       []string{"r"},
			wantPlugin: false,
		},
		{
			name:       "only global flags, no command token",
			args:       []string{"--debug"},
			wantPlugin: false,
		},
		{
			name:       "no args",
			args:       []string{},
			wantPlugin: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, args, isPlugin := detectPlugin(root, tt.args)
			assert.Equal(t, tt.wantPlugin, isPlugin)
			if tt.wantPlugin {
				assert.Equal(t, tt.wantName, name)
				assert.Equal(t, tt.wantArgs, args)
			}
		})
	}
}

func Test_envmanPassthrough(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantArgs  []string
		wantMatch bool
	}{
		{
			name:      "envman with passthrough args",
			args:      []string{"envman", "add", "--key", "FOO"},
			wantArgs:  []string{"add", "--key", "FOO"},
			wantMatch: true,
		},
		{
			// A global flag after envman belongs to the passthrough and must be
			// forwarded verbatim, while a leading one is consumed by bitrise.
			name:      "leading global flag is consumed, rest forwarded verbatim",
			args:      []string{"--debug", "envman", "add", "--ci"},
			wantArgs:  []string{"add", "--ci"},
			wantMatch: true,
		},
		{
			name:      "envman with no args",
			args:      []string{"envman"},
			wantArgs:  []string{},
			wantMatch: true,
		},
		{
			name:      "another command is not envman",
			args:      []string{"run"},
			wantMatch: false,
		},
		{
			name:      "only global flags is not envman",
			args:      []string{"--ci"},
			wantMatch: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, isEnvman := envmanPassthrough(tt.args)
			assert.Equal(t, tt.wantMatch, isEnvman)
			if tt.wantMatch {
				assert.Equal(t, tt.wantArgs, args)
			}
		})
	}
}

func Test_applyGlobalFlagsFromArgs_onlyLeadingApplied(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantDebug bool
		wantCI    bool
		wantPR    bool
	}{
		{
			// A plugin's own --debug after the command token must not set bitrise's
			// persistent --debug flag (nor seed CI/PR). The early debug logger
			// (legacy.IsDebugMode) scans unbounded and still reacts — a kept compat
			// quirk that matches the pre-cobra CLI; it is not asserted here.
			name:      "global-named flag after the command token is not applied to bitrise",
			args:      []string{":plugin", "--debug"},
			wantDebug: false,
		},
		{
			name:      "leading global flags are applied",
			args:      []string{"--debug", "--ci", ":plugin"},
			wantDebug: true,
			wantCI:    true,
		},
		{
			name:   "leading global flag with explicit value",
			args:   []string{"--pr=true", "envman", "add"},
			wantPR: true,
		},
		{
			name:   "global flag after envman is forwarded, not applied",
			args:   []string{"envman", "--pr"},
			wantPR: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newRootCommand()
			legacy.ApplyGlobalFlagsFromArgs(root, tt.args, globalFlagNames)

			debug, _ := root.PersistentFlags().GetBool(DebugModeKey)
			ci, _ := root.PersistentFlags().GetBool(CIKey)
			pr, _ := root.PersistentFlags().GetBool(PRKey)
			assert.Equal(t, tt.wantDebug, debug, "debug")
			assert.Equal(t, tt.wantCI, ci, "ci")
			assert.Equal(t, tt.wantPR, pr, "pr")
		})
	}
}

// urfave/cli ignored an unrecognised flag that followed a positional argument;
// the migration reproduces that via FParseErrWhitelist on every command (cobra
// does not inherit it). Guard against a command — including a nested subcommand —
// being added without the leniency.
func Test_unknownFlagPassthroughEnabledOnWholeTree(t *testing.T) {
	var check func(c *cobra.Command)
	check = func(c *cobra.Command) {
		assert.Truef(t, c.FParseErrWhitelist.UnknownFlags, "command %q must tolerate unknown flags", c.CommandPath())
		for _, sub := range c.Commands() {
			check(sub)
		}
	}
	check(newRootCommand())
}

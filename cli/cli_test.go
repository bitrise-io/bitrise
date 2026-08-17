package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/output"
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
			name:             "Output format json with space syntax",
			args:             []string{"--output-format", "json"},
			wantOutputFormat: "json",
		},
		{
			name:             "Output format console with space syntax",
			args:             []string{"--output-format", "console"},
			wantOutputFormat: "console",
		},
		{
			name:             "Output format json value with equals syntax",
			args:             []string{"--output-format=json"},
			wantOutputFormat: "json",
		},
		{
			name:             "Output format console value with equals syntax",
			args:             []string{"--output-format=console"},
			wantOutputFormat: "console",
		},
		{
			name:             "Single-dash long flag is not recognised",
			args:             []string{"-output-format", "json"},
			wantOutputFormat: "",
		},
		{
			name:             "Output format invalid syntax",
			args:             []string{"--output-format", "--log-level"},
			wantOutputFormat: "",
		},
		{
			name:             "Output format invalid value",
			args:             []string{"--output-format", "invalid"},
			wantOutputFormat: "",
		},
		{
			name:             "Invalid flag",
			args:             []string{"--output-format-invalid=json"},
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
		{
			// A value-taking global flag written with a space must not have its
			// value ("json") mistaken for the plugin/command token.
			name:       "leading value flag with space syntax is skipped",
			args:       []string{"--output", "json", ":analytics"},
			wantName:   "analytics",
			wantArgs:   []string{},
			wantPlugin: true,
		},
		{
			name:       "leading value flag with equals syntax is skipped",
			args:       []string{"--output=json", ":analytics", "--flag"},
			wantName:   "analytics",
			wantArgs:   []string{"--flag"},
			wantPlugin: true,
		},
		{
			name:       "leading theme flag with space syntax is skipped",
			args:       []string{"--theme", "dark", ":analytics"},
			wantName:   "analytics",
			wantArgs:   []string{},
			wantPlugin: true,
		},
		{
			name:       "leading value shorthand with space syntax is skipped",
			args:       []string{"-o", "json", ":analytics"},
			wantName:   "analytics",
			wantArgs:   []string{},
			wantPlugin: true,
		},
		{
			name:       "leading value shorthand with attached value is skipped",
			args:       []string{"-ojson", ":analytics", "--flag"},
			wantName:   "analytics",
			wantArgs:   []string{"--flag"},
			wantPlugin: true,
		},
		{
			name:       "leading bool shorthand cluster is skipped",
			args:       []string{"-q", ":analytics"},
			wantName:   "analytics",
			wantArgs:   []string{},
			wantPlugin: true,
		},
		{
			// A cluster is all-or-nothing: -x is not a bitrise global, so -qx
			// is not consumed as one. plugins.ParseArgs then scans past it to
			// the ":" token, so the plugin still runs — without the flag.
			name:       "shorthand cluster with an unknown letter is not a global",
			args:       []string{"-qx", ":analytics"},
			wantName:   "analytics",
			wantArgs:   []string{},
			wantPlugin: true,
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
		{
			name:      "leading value flag with space syntax is skipped",
			args:      []string{"--output", "json", "envman", "add"},
			wantArgs:  []string{"add"},
			wantMatch: true,
		},
		{
			// Shorthands are the spelling users reach for, and forwarding one
			// into envman makes it reject a flag bitrise owns.
			name:      "leading value shorthand with space syntax is skipped",
			args:      []string{"-o", "json", "envman", "add"},
			wantArgs:  []string{"add"},
			wantMatch: true,
		},
		{
			name:      "leading bool shorthand is skipped",
			args:      []string{"-q", "envman", "add"},
			wantArgs:  []string{"add"},
			wantMatch: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, isEnvman := envmanPassthrough(newRootCommand(), tt.args)
			assert.Equal(t, tt.wantMatch, isEnvman)
			if tt.wantMatch {
				assert.Equal(t, tt.wantArgs, args)
			}
		})
	}
}

func Test_applyGlobalFlagsFromArgs_onlyLeadingApplied(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantDebug  bool
		wantCI     bool
		wantPR     bool
		wantQuiet  bool
		wantOutput string
		wantTheme  string
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
		{
			name:       "leading value flag with space syntax sets the value, not \"true\"",
			args:       []string{"--output", "json", ":plugin"},
			wantOutput: "json",
		},
		{
			name:       "leading value flag with equals syntax",
			args:       []string{"--output=yml", ":plugin"},
			wantOutput: "yml",
		},
		{
			name:      "leading theme flag with space syntax",
			args:      []string{"--theme", "dark", ":plugin"},
			wantTheme: "dark",
		},
		{
			// A plugin's own --output after the command token must not set
			// bitrise's persistent --output flag.
			name:       "value flag after the command token is not applied to bitrise",
			args:       []string{":plugin", "--output", "json"},
			wantOutput: "",
		},
		{
			name:       "value shorthand with space syntax",
			args:       []string{"-o", "json", ":plugin"},
			wantOutput: "json",
		},
		{
			name:       "value shorthand with attached value",
			args:       []string{"-oyml", ":plugin"},
			wantOutput: "yml",
		},
		{
			name:       "value shorthand with equals syntax",
			args:       []string{"-o=json", ":plugin"},
			wantOutput: "json",
		},
		{
			name:      "bool shorthand",
			args:      []string{"--debug", ":plugin"},
			wantDebug: true,
		},
		{
			// A cluster ending in a value flag sets every flag in it.
			name:       "shorthand cluster of a bool and a value flag",
			args:       []string{"-qo", "json", ":plugin"},
			wantQuiet:  true,
			wantOutput: "json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newRootCommand()
			cmdutil.ApplyGlobalFlagsFromArgs(root, tt.args, cmdutil.GlobalFlagNames)

			debug, _ := root.PersistentFlags().GetBool(cmdutil.DebugModeKey)
			ci, _ := root.PersistentFlags().GetBool(cmdutil.CIKey)
			pr, _ := root.PersistentFlags().GetBool(cmdutil.PRKey)
			quiet, _ := root.PersistentFlags().GetBool(cmdutil.FlagQuiet)
			outputVal, _ := root.PersistentFlags().GetString(cmdutil.FlagOutput)
			themeVal, _ := root.PersistentFlags().GetString(cmdutil.FlagTheme)
			assert.Equal(t, tt.wantDebug, debug, "debug")
			assert.Equal(t, tt.wantCI, ci, "ci")
			assert.Equal(t, tt.wantPR, pr, "pr")
			assert.Equal(t, tt.wantQuiet, quiet, "quiet")
			assert.Equal(t, tt.wantOutput, outputVal, "output")
			assert.Equal(t, tt.wantTheme, themeVal, "theme")
		})
	}
}

// runEnvman and runPlugin call before() directly, without ever going through
// cobra's Execute()/ExecuteC() — the only place that seeds cmd.Context() with
// context.Background() when nil. Regression test for a panic ("cannot create
// context from nil parent") that this caused in config.WithResolved.
func Test_before_calledWithoutExecute_doesNotPanic(t *testing.T) {
	root := newRootCommand()

	assert.NotPanics(t, func() {
		err := before(root, nil)
		assert.NoError(t, err)
	})
}

// Test_before_outputPrecedence pins the precedence documented in
// cli/config/cmd.go: root flag > $BITRISE_OUTPUT > the "output" config key >
// raw — env above config, matching every other resolver in cli/cmdutil.
func Test_before_outputPrecedence(t *testing.T) {
	t.Cleanup(func() {
		output.SetDefault(output.FormatRaw)
		require.NoError(t, output.ConfigureOutputFormat(output.FormatRaw))
	})

	tests := []struct {
		name       string
		rootFlag   string
		configured string
		env        string
		want       string
	}{
		{name: "root flag wins over env and config", rootFlag: "raw", configured: "json", env: "yml", want: "raw"},
		{name: "env wins over config key when no flag", configured: "json", env: "yml", want: "yml"},
		{name: "config key used when neither flag nor env set", configured: "json", want: "json"},
		{name: "raw default when nothing set", want: "raw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			if tt.env != "" {
				t.Setenv(cmdutil.EnvOutput, tt.env)
			}
			if tt.configured != "" {
				require.NoError(t, internalconfig.Save(internalconfig.Config{Output: tt.configured}))
			}

			root := newRootCommand()
			if tt.rootFlag != "" {
				require.NoError(t, root.PersistentFlags().Set(cmdutil.FlagOutput, tt.rootFlag))
			}
			require.NoError(t, before(root, nil))
			require.NoError(t, output.ConfigureOutputFormat(""))
			assert.Equal(t, tt.want, output.Format)
		})
	}
}

// Test_ymlMergeOutputFlag_ShadowsGlobalOutputFlag guards the shorthand
// collision called out in the master rde-migration plan: yml merge's own
// --output/-o (an output directory) must keep working unchanged once the
// root gains a persistent --output/-o (an output format) — cobra's flag
// merge order (local flags are registered before parent persistent flags are
// merged in) makes the local one win, with no panic on the shared shorthand.
func Test_ymlMergeOutputFlag_ShadowsGlobalOutputFlag(t *testing.T) {
	root := newRootCommand()

	var mergeCmd *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == "yml" {
			for _, sub := range c.Commands() {
				if sub.Name() == "merge" {
					mergeCmd = sub
				}
			}
		}
	}
	require.NotNil(t, mergeCmd, "yml merge command must be registered")

	require.NotPanics(t, func() {
		require.NoError(t, mergeCmd.ParseFlags([]string{"-o", "/tmp/merged"}))
	})

	got, err := mergeCmd.Flags().GetString("output")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/merged", got, "-o must resolve to yml merge's own output-directory flag")

	rootOutput, _ := root.PersistentFlags().GetString(cmdutil.FlagOutput)
	assert.Equal(t, "", rootOutput, "the global --output flag must be untouched by yml merge's local -o")
}

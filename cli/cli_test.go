package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/log"
	"github.com/stretchr/testify/assert"
)

func TestCommandInfo(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantCommand    string
		wantSubcommand string
		wantFlags      []string
	}{
		{
			name:           "Empty command",
			args:           []string{},
			wantCommand:    "",
			wantSubcommand: "",
			wantFlags:      nil,
		},
		{
			name:           "CLI command",
			args:           []string{"run", "e2e"},
			wantCommand:    "run",
			wantSubcommand: "",
			wantFlags:      nil,
		},
		{
			name:           "Plugin command",
			args:           []string{":plugin", "do", "something"},
			wantCommand:    ":plugin",
			wantSubcommand: "do",
			wantFlags:      nil,
		},
		{
			name:           "Flags",
			args:           []string{"run", "--A", "-a", "--B=true", "-b false", "--C /path/to/something"},
			wantCommand:    "run",
			wantSubcommand: "",
			wantFlags:      []string{"A", "a", "B", "b", "C"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, subcommand, flags := commandExecutionInfo(tt.args)
			assert.Equalf(t, tt.wantCommand, command, "commandExecutionInfo(%v)", tt.args)
			assert.Equalf(t, tt.wantSubcommand, subcommand, "commandExecutionInfo(%v)", tt.args)
			assert.Equalf(t, tt.wantFlags, flags, "commandExecutionInfo(%v)", tt.args)
		})
	}
}

func Test_loggerParameters(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		wantIsRunCommand bool
		wantOutputFormat log.LoggerType
		wantDebugMode    bool
	}{
		{
			name:             "Empty test",
			args:             []string{},
			wantIsRunCommand: false,
			wantOutputFormat: "",
			wantDebugMode:    false,
		},
		{
			name:          "Debug mode on with one dash syntax",
			args:          []string{"-debug"},
			wantDebugMode: true,
		},
		{
			name:          "Debug mode on with two dash syntax",
			args:          []string{"--debug"},
			wantDebugMode: true,
		},
		{
			name:          "Debug mode on with value syntax",
			args:          []string{"-debug=true"},
			wantDebugMode: true,
		},
		{
			name:          "Debug mode off with value syntax",
			args:          []string{"--debug=true"},
			wantDebugMode: true,
		},
		{
			name:          "Debug mode invalid syntax",
			args:          []string{"--debug true"},
			wantDebugMode: false,
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
			isRunCommand, outputFormat, debugMode := loggerParameters(tt.args)
			assert.Equalf(t, tt.wantIsRunCommand, isRunCommand, "loggerParameters(%v)", tt.args)
			assert.Equalf(t, tt.wantOutputFormat, outputFormat, "loggerParameters(%v)", tt.args)
			assert.Equalf(t, tt.wantDebugMode, debugMode, "loggerParameters(%v)", tt.args)
		})
	}
}

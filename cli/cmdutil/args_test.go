package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDetectSingleDashLongFlag(t *testing.T) {
	newCmd := func() *cobra.Command {
		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().BoolP(FlagQuiet, "q", false, "quiet")
		root.PersistentFlags().StringP(FlagOutput, "o", "", "output")
		root.PersistentFlags().Bool(DebugModeKey, false, "debug")

		child := &cobra.Command{Use: "run"}
		child.Flags().StringP(ConfigKey, "c", "", "config")
		child.Flags().StringP(InventoryKey, "i", "", "inventory")
		child.Flags().String(WorkflowKey, "", "workflow")
		root.AddCommand(child)

		return child
	}

	for _, tc := range []struct {
		name         string
		args         []string
		wantArg      string
		wantFlagName string
		wantFound    bool
	}{
		{
			name:         "single-dash long flag name misparsed by pflag as a shorthand cluster",
			args:         []string{"run", "-config", "bitrise.yml"},
			wantArg:      "-config",
			wantFlagName: "config",
			wantFound:    true,
		},
		{
			name:         "single-dash long flag name with attached value",
			args:         []string{"run", "-inventory=secrets.yml"},
			wantArg:      "-inventory=secrets.yml",
			wantFlagName: "inventory",
			wantFound:    true,
		},
		{
			name:         "single-dash long flag with no colliding shorthand",
			args:         []string{"run", "-workflow", "primary"},
			wantArg:      "-workflow",
			wantFlagName: "workflow",
			wantFound:    true,
		},
		{
			name:         "single-dash long flag inherited from a persistent parent flag",
			args:         []string{"run", "-debug"},
			wantArg:      "-debug",
			wantFlagName: "debug",
			wantFound:    true,
		},
		{
			name:      "double-dash long flag is untouched",
			args:      []string{"run", "--config", "bitrise.yml"},
			wantFound: false,
		},
		{
			name:      "legitimate value shorthand with space syntax is untouched",
			args:      []string{"run", "-c", "bitrise.yml"},
			wantFound: false,
		},
		{
			name:      "legitimate value shorthand with attached value is untouched",
			args:      []string{"run", "-i", "secrets.yml"},
			wantFound: false,
		},
		{
			name:      "legitimate bool shorthand cluster is untouched",
			args:      []string{"run", "-qo", "json"},
			wantFound: false,
		},
		{
			name:      "bare dash is untouched",
			args:      []string{"run", "-"},
			wantFound: false,
		},
		{
			name:      "positional args after -- are not scanned",
			args:      []string{"run", "--", "-config"},
			wantFound: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			arg, flagName, found := DetectSingleDashLongFlag(newCmd(), tc.args)
			assert.Equal(t, tc.wantFound, found)
			if tc.wantFound {
				assert.Equal(t, tc.wantArg, arg)
				assert.Equal(t, tc.wantFlagName, flagName)
			}
		})
	}
}

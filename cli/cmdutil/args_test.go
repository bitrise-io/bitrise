package cmdutil

import (
	"io"
	"slices"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// Test_walkShorthands_matchesPflag runs the shared grammar model and a real
// pflag.FlagSet over the same token and asserts they agree on every count.
// Three separate bugs in this file reached review because each hand-rolled
// model was only checked by hand; this checks it against the parser it models.
func Test_walkShorthands_matchesPflag(t *testing.T) {
	const sentinel = "SENTINEL"

	tokens := []string{
		"-q", "-qq", "-qh", "-h",
		"-o", "-ojson", "-o=json", "-o=", "-oq", "-oqjson",
		"-q=", "-q=true", "-qo", "-qojson", "-qo=json",
		"-c", "-cfile", "-c=file",
		"-config", "-qconfig", "-hconfig", "-qqconfig", "-qconfig=x",
		"-x", "-qx", "-=",
	}

	for _, token := range tokens {
		t.Run(token, func(t *testing.T) {
			fs := newPflagFixture()
			predicted, wantsNextArg, ok := walkShorthands(token[1:], func(c string) *pflag.Flag {
				return fs.ShorthandLookup(c)
			})

			actual := newPflagFixture()
			parseErr := actual.Parse([]string{token, sentinel})

			if !ok {
				assert.Error(t, parseErr, "model rejected the token, pflag accepted it")
				return
			}
			require.NoError(t, parseErr, "model accepted the token, pflag rejected it")

			consumed := !slices.Contains(actual.Args(), sentinel)
			assert.Equal(t, consumed, wantsNextArg, "disagreement on whether the next argument is the value")

			for i, a := range predicted {
				value := a.value
				// Only the flag that ends the cluster can take the next
				// argument; the bools before it are already satisfied.
				if wantsNextArg && i == len(predicted)-1 {
					value = sentinel
				}
				assert.Equal(t, actual.Lookup(a.name).Value.String(), value, "disagreement on %s's value", a.name)
			}
		})
	}
}

// newPflagFixture mirrors the shorthands that actually collide in the real
// tree: a bool, help, and two value flags whose shorthands lead long names
// (-c/--config, -o/--output).
func newPflagFixture() *pflag.FlagSet {
	fs := pflag.NewFlagSet("fixture", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolP(FlagQuiet, "q", false, "quiet")
	fs.BoolP("help", "h", false, "help")
	fs.StringP(FlagOutput, "o", "", "output")
	fs.StringP("config", "c", "", "config")
	return fs
}

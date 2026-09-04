package savedinput

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

func newCreateCmd() *cobra.Command {
	var (
		key        string
		value      string
		valueStdin bool
		isSecret   bool
		format     string
	)
	c := &cobra.Command{
		Use:   "create",
		Short: "Create a new saved input",
		Long: `Create a new saved input.

The value can be supplied three ways:
  --value VALUE   use VALUE literally (pass --value - to store a literal dash)
  --value-stdin   read the value from stdin without prompting; keeps secrets
                  out of shell history
  neither         prompt for the value interactively; input is masked when
                  stdin is a terminal`,
		Example: `  bitrise rde saved-input create --key repo-name --value my-app
  echo -n "ghp_xxx" | bitrise rde saved-input create --key gh-token --value-stdin --secret
  bitrise rde saved-input create --key gh-token --secret   # prompts for the value`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			if key == "" {
				return fmt.Errorf("--key is required")
			}
			valueChanged := cmd.Flags().Changed("value")
			value, _, err := resolveValue(cmd, value, valueChanged, valueStdin, true)
			if err != nil {
				return err
			}
			// Empty is only allowed when the caller explicitly passed --value "".
			// An empty prompt or empty piped stdin is treated as a mistake.
			if value == "" && !valueChanged {
				return fmt.Errorf("value is empty")
			}
			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			in, err := internalrde.NewService(client).CreateSavedInput(cmd.Context(), internalrde.CreateSavedInputRequest{
				Key:      key,
				Value:    value,
				IsSecret: isSecret,
			})
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, in, renderDetail)
		},
	}
	c.Flags().StringVar(&key, "key", "", "saved-input key (required)")
	c.Flags().StringVar(&value, "value", "", "value to store (literal)")
	c.Flags().BoolVar(&valueStdin, "value-stdin", false, "read the value from stdin without prompting")
	c.Flags().BoolVar(&isSecret, "secret", false, "encrypt value at rest; the value will be masked in subsequent reads")
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	c.MarkFlagsMutuallyExclusive("value", "value-stdin")
	return c
}

// resolveValue determines a saved-input value from the --value / --value-stdin
// flags, which the caller marks mutually exclusive. When --value-stdin is set it
// reads a single line from stdin without prompting. When --value was passed it
// returns the literal flag value. Otherwise, if promptIfMissing is true it prompts
// interactively (masked when stdin is a terminal); if false it reports
// provided=false so the caller can leave the value untouched.
//
// This mirrors the secret-input convention used by `bitrise auth login`
// (--with-token / --password-stdin plus a masked interactive default), down to
// requiring --value-stdin to actually be piped: falling through to an unmasked
// read on a terminal would echo the secret with no prompt to explain the wait.
//
// The stdin read uses ReadPasswordInput, which trims only the trailing line
// terminator, so a piped value is treated the same as one passed to --value
// (which is not trimmed at all). The interactive prompt keeps ReadSecretInput's
// full trim, where stripping copy/paste whitespace is what the user wants.
func resolveValue(cmd *cobra.Command, value string, valueChanged, valueStdin, promptIfMissing bool) (resolved string, provided bool, err error) {
	switch {
	case valueStdin:
		if cerr := cmdutil.CheckValueStdinPiped(valueStdin, cmdutil.IsTerminal(cmd.InOrStdin()), "bitrise rde saved-input "+cmd.Name()); cerr != nil {
			return "", false, cerr
		}
		v, rerr := cmdutil.ReadPasswordInput(cmd.InOrStdin(), cmd.ErrOrStderr(), "", true)
		return v, true, rerr
	case valueChanged:
		return value, true, nil
	case promptIfMissing:
		v, rerr := cmdutil.ReadSecretInput(cmd.InOrStdin(), cmd.ErrOrStderr(), "Value: ", false)
		return v, true, rerr
	default:
		return "", false, nil
	}
}

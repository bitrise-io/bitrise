package build

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v3/internal/build"
	"github.com/bitrise-io/bitrise/v3/output"
)

// NewTriggerCommand returns the `build trigger` subcommand.
func NewTriggerCommand() *cobra.Command {
	var (
		workflow      string
		pipeline      string
		branch        string
		branchDest    string
		tag           string
		commitHash    string
		commitMessage string
		envJSON       string
		priority      int
		pullRequestID int
		stack         string
		machineType   string
		licensePool   string
		wait          bool
		watch         bool
		interval      time.Duration
	)

	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Start a new build",
		Long: `Start a new build on the given app.

The app is resolved via --app ID, BITRISE_APP_ID, or "bitrise config set app_id ID".

If neither --workflow nor --pipeline is given, Bitrise selects the
appropriate workflow from the trigger map.

--wait blocks until the build finishes without streaming logs; with --format
json/yml the final build record is written to stdout.

--watch waits for the build to finish, showing progress: with --format
json/yml, build logs stream to stderr as plain text and the final build
record is written to stdout; otherwise, on a terminal, an interactive status
display is shown instead of raw log lines, and when stdout isn't a terminal,
logs stream as plain text there instead.`,
		Example: `  bitrise build trigger --app my-app-id --workflow primary
  bitrise build trigger --app my-app-id --workflow deploy --branch release/1.2 --format json
  bitrise build trigger --app my-app-id --pipeline my-pipeline --branch main
  bitrise build trigger --app my-app-id --workflow primary --tag v1.2.3
  bitrise build trigger --app my-app-id --workflow primary --branch-dest main --pull-request-id 42
  bitrise build trigger --app my-app-id --workflow primary --env '{"MY_VAR":"hello","OTHER":"world"}'
  bitrise build trigger --app my-app-id --workflow primary --env '{"API_URL":"https://example.com","PRICE":{"value":"$5","is_expand":false}}'
  bitrise build trigger --app my-app-id --workflow primary --stack osx-xcode-16.0.x --machine-type g2-m1.4core
  bitrise build trigger --app my-app-id --pipeline my-pipeline --priority 0
  bitrise build trigger --app my-app-id --workflow primary --wait
  bitrise build trigger --app my-app-id --workflow primary --watch`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}
			appSlug, err := cmdutil.ResolveAndLookupAppSlug(cmd, client)
			if err != nil {
				return err
			}

			if branch == "" && tag == "" && commitHash == "" {
				branch = "main"
			}

			envs, err := parseEnvFlag(envJSON)
			if err != nil {
				return err
			}

			var priorityOverride *int
			if cmd.Flags().Changed("priority") {
				priorityOverride = &priority
			}

			svc := internalbuild.NewService(client)

			b, err := svc.Trigger(cmd.Context(), internalbuild.TriggerRequest{
				AppSlug:       appSlug,
				Workflow:      workflow,
				Pipeline:      pipeline,
				Branch:        branch,
				BranchDest:    branchDest,
				Tag:           tag,
				CommitHash:    commitHash,
				CommitMessage: commitMessage,
				PullRequestID: pullRequestID,
				Priority:      priorityOverride,
				Stack:         stack,
				MachineTypeID: machineType,
				LicensePoolID: licensePool,
				Environments:  envs,
			})
			if err != nil {
				return err
			}

			return runAfterTrigger(cmd, svc, b, wait, watch, interval)
		},
	}

	cmd.Flags().StringVar(&workflow, "workflow", "", "workflow ID to trigger (mutually exclusive with --pipeline)")
	cmd.Flags().StringVar(&pipeline, "pipeline", "", "pipeline ID to trigger (mutually exclusive with --workflow)")
	cmd.Flags().StringVar(&branch, "branch", "", `branch to build (default "main" unless --tag or --commit-hash is given)`)
	cmd.Flags().StringVar(&branchDest, "branch-dest", "", "target branch for pull-request builds")
	cmd.Flags().StringVar(&tag, "tag", "", "tag to build")
	cmd.Flags().StringVar(&commitHash, "commit-hash", "", "commit hash to build")
	cmd.Flags().StringVar(&commitMessage, "commit-message", "", "commit message to record")
	cmd.Flags().StringVar(&envJSON, "env", "", `environment variables as a JSON object, e.g. '{"KEY":"value"}'; $VAR references in values are expanded, use '{"KEY":{"value":"$5","is_expand":false}}' to pass a value verbatim`)
	cmd.Flags().IntVar(&priority, "priority", 0, "build priority from -100 to 100, overrides the bitrise.yml and trigger map priority even when 0; omit to keep those (available on certain plans only)")
	cmd.Flags().IntVar(&pullRequestID, "pull-request-id", 0, "pull request ID for PR builds")
	cmd.Flags().StringVar(&stack, "stack", "", "stack ID to run the build on, overrides the workflow's stack (see 'bitrise stack list')")
	cmd.Flags().StringVar(&machineType, "machine-type", "", "machine type ID to run the build on, overrides the workflow's machine type")
	cmd.Flags().StringVar(&licensePool, "license-pool", "", "license pool ID to run the build with")
	cmd.Flags().BoolVar(&wait, "wait", false, "block until the build finishes without streaming logs (exit code reflects build outcome)")
	cmd.Flags().BoolVar(&watch, "watch", false, "wait for the build to finish, showing progress (exit code reflects build outcome)")
	cmd.Flags().DurationVar(&interval, "interval", 3*time.Second, "polling interval when --wait or --watch is active")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	cmd.MarkFlagsMutuallyExclusive("workflow", "pipeline")
	cmd.MarkFlagsMutuallyExclusive("wait", "watch")

	return cmd
}

func parseEnvFlag(envJSON string) ([]internalbuild.TriggerEnv, error) {
	if envJSON == "" {
		return nil, nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(envJSON), &raw); err != nil {
		return nil, fmt.Errorf("--env: invalid JSON object: %w", err)
	}

	envs := make([]internalbuild.TriggerEnv, 0, len(raw))
	for key, rawValue := range raw {
		env := internalbuild.TriggerEnv{Key: key, IsExpand: true}
		if err := json.Unmarshal(rawValue, &env.Value); err != nil {
			var item struct {
				Value    *string `json:"value"`
				IsExpand *bool   `json:"is_expand"`
			}
			dec := json.NewDecoder(bytes.NewReader(rawValue))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&item); err != nil || item.Value == nil {
				return nil, fmt.Errorf(`--env: %q must be a string or {"value":"...","is_expand":false}`, key)
			}
			env.Value = *item.Value
			if item.IsExpand != nil {
				env.IsExpand = *item.IsExpand
			}
		}
		envs = append(envs, env)
	}
	return envs, nil
}

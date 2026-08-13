package build

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v2/internal/build"
	"github.com/bitrise-io/bitrise/v2/output"
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
		wait          bool
		watch         bool
		interval      time.Duration
	)

	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Start a new build",
		Long: `Start a new build on the given app.

Optional flags:
  --workflow ID          workflow ID (mutually exclusive with --pipeline); Bitrise
                         selects the appropriate workflow from the trigger map if omitted
  --pipeline ID          pipeline ID (mutually exclusive with --workflow)
  --branch BRANCH        branch to build (default "main" for branch builds)
  --branch-dest BRANCH   target branch for pull-request builds
  --tag TAG              tag to build
  --commit-hash HASH     commit hash to build from
  --commit-message MSG   commit message to record
  --pull-request-id ID   pull request ID for PR builds
  --priority N           build priority (-1 = low, 0 = normal, 1 = high)
  --env JSON             environment variables as a JSON object, e.g. '{"KEY":"value"}'
  --wait                 wait for the build to finish without streaming logs; exits 0 on
                         success, non-zero on failure. With --format json/yml the final
                         build record is written to stdout.
  --watch                stream build logs until the build finishes; exits 0 on success,
                         non-zero on failure. With --format json/yml logs go to stderr and
                         the final build record is written to stdout.
  --interval DURATION    polling interval when --wait or --watch is active (default 3s)`,
		Example: `  bitrise build trigger --app my-app-id --workflow primary
  bitrise build trigger --app my-app-id --workflow deploy --branch release/1.2 --format json
  bitrise build trigger --app my-app-id --pipeline my-pipeline --branch main
  bitrise build trigger --app my-app-id --workflow primary --tag v1.2.3
  bitrise build trigger --app my-app-id --workflow primary --env '{"MY_VAR":"hello","OTHER":"world"}'
  bitrise build trigger --app my-app-id --workflow primary --wait
  bitrise build trigger --app my-app-id --workflow primary --watch`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			appSlug, err := cmdutil.ResolveAppSlug(cmd)
			if err != nil {
				return err
			}

			// Default to branch "main" for branch builds when no tag is given.
			if branch == "" && tag == "" {
				branch = "main"
			}

			var envs []internalbuild.TriggerEnv
			if envJSON != "" {
				var raw map[string]string
				if err := json.Unmarshal([]byte(envJSON), &raw); err != nil {
					return fmt.Errorf("--env: invalid JSON object: %w", err)
				}
				for k, v := range raw {
					envs = append(envs, internalbuild.TriggerEnv{Key: k, Value: v})
				}
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
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
				Priority:      priority,
				Environments:  envs,
			})
			if err != nil {
				return err
			}

			if !wait && !watch {
				return output.Render(cmd.OutOrStdout(), output.Format, b, printTriggerText)
			}

			if watch {
				logWriter := io.Writer(cmd.OutOrStdout())
				if output.Format == output.FormatJSON || output.Format == output.FormatYML {
					logWriter = cmd.ErrOrStderr()
				}
				return runWatch(cmd, svc, b, interval, logWriter, output.Format)
			}

			// --wait: silent block until the build finishes; no log output.
			header := fmt.Sprintf("Waiting for build #%d to finish\n", b.BuildNumber)
			if url := buildDetailURL(cmd, b); url != "" {
				header += fmt.Sprintf("%s\n", url)
			}
			if _, err := fmt.Fprint(cmd.ErrOrStderr(), header); err != nil {
				return err
			}

			finalBuild, err := svc.WaitForCompletion(cmd.Context(), b.AppSlug, b.Slug, interval)
			if errors.Is(err, context.Canceled) {
				return writeDetachNotice(cmd.ErrOrStderr(), "build watch "+b.Slug)
			}
			if err != nil {
				return err
			}

			if output.Format == output.FormatJSON || output.Format == output.FormatYML {
				if err := output.Render(cmd.OutOrStdout(), output.Format, finalBuild, printBuildText); err != nil {
					return err
				}
			} else {
				footer := fmt.Sprintf("Build #%d finished: %s%s\n", finalBuild.BuildNumber, finalBuild.Status, buildElapsed(finalBuild))
				if _, err := fmt.Fprint(cmd.ErrOrStderr(), footer); err != nil {
					return err
				}
			}

			// The exit code reflects the build outcome in every mode, including
			// --format json/yml: stdout already carries the build record above.
			if finalBuild.Status != "success" && finalBuild.Status != "aborted-with-success" {
				return fmt.Errorf("build %s", finalBuild.Status)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&workflow, "workflow", "", "workflow ID to trigger (mutually exclusive with --pipeline)")
	cmd.Flags().StringVar(&pipeline, "pipeline", "", "pipeline ID to trigger (mutually exclusive with --workflow)")
	cmd.Flags().StringVar(&branch, "branch", "", `branch to build (default "main" for branch builds)`)
	cmd.Flags().StringVar(&branchDest, "branch-dest", "", "target branch for pull-request builds")
	cmd.Flags().StringVar(&tag, "tag", "", "tag to build")
	cmd.Flags().StringVar(&commitHash, "commit-hash", "", "commit hash to build")
	cmd.Flags().StringVar(&commitMessage, "commit-message", "", "commit message to record")
	cmd.Flags().StringVar(&envJSON, "env", "", `environment variables as a JSON object, e.g. '{"KEY":"value"}'`)
	cmd.Flags().IntVar(&priority, "priority", 0, "build priority (-1 = low, 0 = normal, 1 = high)")
	cmd.Flags().IntVar(&pullRequestID, "pull-request-id", 0, "pull request ID for PR builds")
	cmd.Flags().BoolVar(&wait, "wait", false, "block until the build finishes without streaming logs (exit code reflects build outcome)")
	cmd.Flags().BoolVar(&watch, "watch", false, "stream build logs until the build finishes (exit code reflects build outcome)")
	cmd.Flags().DurationVar(&interval, "interval", 3*time.Second, "polling interval when --wait or --watch is active")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	cmd.MarkFlagsMutuallyExclusive("workflow", "pipeline")
	cmd.MarkFlagsMutuallyExclusive("wait", "watch")

	_ = cmd.RegisterFlagCompletionFunc("priority", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"-1\tlow priority", "0\tnormal priority", "1\thigh priority"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

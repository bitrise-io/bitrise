package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/urfave/cli"
)

const (
	// PerformanceJSONPath is the path where performance data is saved
	PerformanceJSONPath = ".bitrise/performance_data.json"
)

// SavePerformanceMetrics saves the performance metrics to a JSON file
func SavePerformanceMetrics(buildResults models.BuildRunResultsModel) error {
	if buildResults.PerformanceMetrics == nil {
		return nil
	}

	// Create the directory if it doesn't exist
	dir := ".bitrise"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Marshal the performance metrics to JSON
	data, err := json.MarshalIndent(buildResults.PerformanceMetrics, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(PerformanceJSONPath, data, 0644)
}

// LoadPerformanceMetrics loads performance metrics from the JSON file
func LoadPerformanceMetrics() (*models.BuildPerformanceMetrics, error) {
	// Check if file exists
	if _, err := os.Stat(PerformanceJSONPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no performance data found - run a workflow first")
	}

	// Read the file
	data, err := os.ReadFile(PerformanceJSONPath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the data
	var metrics models.BuildPerformanceMetrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// GenerateTraceProfile generates a Chrome trace profile file from performance metrics
func GenerateTraceProfile(metrics *models.BuildPerformanceMetrics, outputPath string) error {
	// Create a new trace profile
	profile := models.NewTraceProfile(metrics)

	// Marshal to JSON
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Write to file
	return os.WriteFile(outputPath, data, 0644)
}

// PerformanceCommand returns the performance command
func PerformanceCommand() cli.Command {
	return cli.Command{
		Name:    "performance",
		Aliases: []string{"perf"},
		Usage:   "Show performance metrics for the last build",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "top",
				Usage: "Show only the top N slowest steps",
				Value: 10,
			},
			cli.BoolFlag{
				Name:  "summary",
				Usage: "Show only the summary",
			},
			cli.BoolFlag{
				Name:  "detailed",
				Usage: "Show detailed phase timing for each step",
			},
			cli.StringFlag{
				Name:  "trace",
				Usage: "Generate Chrome trace profile and save to the specified file",
			},
		},
		Action: func(c *cli.Context) error {
			// Set up the logger
			logger := log.NewLogger(log.GetGlobalLoggerOpts())

			// Load the performance metrics
			metrics, err := LoadPerformanceMetrics()
			if err != nil {
				return err
			}

			// Initialize tabwriter for formatted output
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			// Print build summary
			fmt.Fprintf(w, "Build Performance Summary:\n")
			fmt.Fprintf(w, "------------------------\n")
			fmt.Fprintf(w, "Total Build Time:\t%s\n", models.FormatDuration(metrics.TotalTime))
			fmt.Fprintf(w, "Number of Workflows:\t%d\n", len(metrics.Workflows))

			var totalSteps int
			for _, wf := range metrics.Workflows {
				totalSteps += len(wf.Steps)
			}
			fmt.Fprintf(w, "Total Steps:\t%d\n\n", totalSteps)
			w.Flush()

			if !c.Bool("summary") {
				// Print workflow details
				fmt.Fprintf(w, "Workflow Performance:\n")
				fmt.Fprintf(w, "--------------------\n")

				// Sort workflows by execution time
				sort.Slice(metrics.Workflows, func(i, j int) bool {
					return metrics.Workflows[i].TotalTime > metrics.Workflows[j].TotalTime
				})

				for _, wf := range metrics.Workflows {
					fmt.Fprintf(w, "Workflow: %s\n", wf.WorkflowTitle)
					fmt.Fprintf(w, "Duration: %s\n", models.FormatDuration(wf.TotalTime))
					fmt.Fprintf(w, "Steps: %d\n\n", len(wf.Steps))
				}
				w.Flush()
			}

			// Get top slowest steps
			topN := c.Int("top")
			slowestSteps := metrics.GetTopNSlowestSteps(topN)

			fmt.Fprintf(w, "Top %d Slowest Steps:\n", topN)
			fmt.Fprintf(w, "-------------------\n")
			fmt.Fprintf(w, "Step\tDuration\t%% of Build\n")
			fmt.Fprintf(w, "----\t--------\t----------\n")

			for _, step := range slowestSteps {
				percentOfBuild := float64(step.TotalTime.Milliseconds()) / float64(metrics.TotalTime.Milliseconds()) * 100
				fmt.Fprintf(w, "%s\t%s\t%.1f%%\n", step.StepTitle, models.FormatDuration(step.TotalTime), percentOfBuild)
			}
			w.Flush()

			if c.Bool("detailed") {
				fmt.Fprintf(w, "\nDetailed Phase Timing:\n")
				fmt.Fprintf(w, "--------------------\n")

				for _, step := range slowestSteps {
					fmt.Fprintf(w, "Step: %s (Total: %s)\n", step.StepTitle, models.FormatDuration(step.TotalTime))

					// Sort phases by duration
					sort.Slice(step.Phases, func(i, j int) bool {
						return step.Phases[i].Duration > step.Phases[j].Duration
					})

					for _, phase := range step.Phases {
						percentOfStep := float64(phase.Duration.Milliseconds()) / float64(step.TotalTime.Milliseconds()) * 100
						fmt.Fprintf(w, "  %s:\t%s\t(%.1f%%)\n", phase.Phase, models.FormatDuration(phase.Duration), percentOfStep)
					}
					fmt.Fprintln(w, "")
				}
				w.Flush()
			}

			// Generate trace profile if requested
			if tracePath := c.String("trace"); tracePath != "" {
				if err := GenerateTraceProfile(metrics, tracePath); err != nil {
					logger.Errorf("Failed to generate trace profile: %s", err)
				} else {
					logger.Infof("\nTrace profile saved to: %s", tracePath)
					logger.Infof("You can open this file in Chrome by navigating to chrome://tracing")
				}
			}

			logger.Infof("\nTo improve build performance, focus on optimizing the steps that take the longest time.")
			return nil
		},
	}
}

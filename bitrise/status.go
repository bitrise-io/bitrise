package bitrise

import (
	"fmt"
	"github.com/bitrise-io/bitrise/models"
)

func HumanReadableStatus(status int) string {
	switch status {
	case models.StepRunStatusCodeSuccess:
		return "success"
	case models.StepRunStatusCodeFailed:
		return "failed"
	case models.StepRunStatusCodeFailedSkippable:
		return "failed_skippable"
	case models.StepRunStatusCodeSkipped:
		return "skipped"
	case models.StepRunStatusCodeSkippedWithRunIf:
		return "skipped_with_run_if"
	case models.StepRunStatusCodePreparationFailed:
		return "preparation_failed"
	case models.StepRunStatusAbortedWithCustomTimeout:
		return "aborted_with_custom_timeout"
	case models.StepRunStatusAbortedWithNoOutputTimeout:
		return "aborted_with_no_output"
	default:
		return "unknown"
	}
}

func InternalStatus(status string) int {
	switch status {
	case "success":
		return models.StepRunStatusCodeSuccess
	case "failed":
		return models.StepRunStatusCodeFailed
	case "failed_skippable":
		return models.StepRunStatusCodeFailedSkippable
	case "skipped":
		return models.StepRunStatusCodeSkipped
	case "skipped_with_run_if":
		return models.StepRunStatusCodeSkippedWithRunIf
	case "preparation_failed":
		return models.StepRunStatusCodePreparationFailed
	case "aborted_with_custom_timeout":
		return models.StepRunStatusAbortedWithCustomTimeout
	case "aborted_with_no_output":
		return models.StepRunStatusAbortedWithNoOutputTimeout
	default:
		return -1
	}
}

func StatusReason(status, code int) string {
	switch status {
	case models.StepRunStatusCodeSuccess, models.StepRunStatusCodeSkipped, models.StepRunStatusCodeSkippedWithRunIf:
		return ""
	case models.StepRunStatusCodeFailed, models.StepRunStatusCodePreparationFailed, models.StepRunStatusCodeFailedSkippable:
		return fmt.Sprintf("exit code: %d", code)
	case models.StepRunStatusAbortedWithCustomTimeout:
		return "timed out"
	case models.StepRunStatusAbortedWithNoOutputTimeout:
		return "timed out due to no output"
	default:
		return "unknown result code"
	}

}

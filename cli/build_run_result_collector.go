package cli

import (
	"errors"
	"time"

	"github.com/bitrise-io/bitrise/analytics"
	"github.com/bitrise-io/bitrise/exitcode"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/tools/timeoutcmd"
	"github.com/bitrise-io/bitrise/utils"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/pointers"
	coreanalytics "github.com/bitrise-io/go-utils/v2/analytics"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

type buildRunResultCollector struct {
	tracker analytics.Tracker
}

func newBuildRunResultCollector(tracker analytics.Tracker) buildRunResultCollector {
	return buildRunResultCollector{tracker: tracker}
}

func (r buildRunResultCollector) registerStepRunResults(
	buildRunResults *models.BuildRunResultsModel,
	stepExecutionId string,
	stepStartTime time.Time,
	step stepmanModels.StepModel,
	stepInfoPtr stepmanModels.StepInfoModel,
	stepIdxPtr int,
	runIf string,
	status models.StepRunStatus,
	exitCode int,
	err error,
	isLastStep bool,
	printStepHeader bool,
	redactedStepInputs map[string]string,
	properties coreanalytics.Properties) {

	timeout, noOutputTimeout := time.Duration(-1), time.Duration(-1)
	if status == models.StepRunStatusCodeFailed {
		// Forward the status of a Step or a wrapped bitrise process.
		switch exitCode {
		case exitcode.CLIAbortedWithCustomTimeout:
			status = models.StepRunStatusAbortedWithCustomTimeout
		case exitcode.CLIAbortedWithNoOutputTimeout:
			status = models.StepRunStatusAbortedWithNoOutputTimeout
		}

		var timeoutErr timeoutcmd.TimeoutError
		if ok := errors.As(err, &timeoutErr); ok {
			status = models.StepRunStatusAbortedWithCustomTimeout
			timeout = timeoutErr.Timeout
		}

		var noOutputTimeoutErr timeoutcmd.NoOutputTimeoutError
		if ok := errors.As(err, &noOutputTimeoutErr); ok {
			status = models.StepRunStatusAbortedWithNoOutputTimeout
			noOutputTimeout = noOutputTimeoutErr.Timeout
		}
	}

	stepInfoCopy := stepmanModels.StepInfoModel{
		Library:         stepInfoPtr.Library,
		ID:              stepInfoPtr.ID,
		Version:         stepInfoPtr.Version,
		OriginalVersion: stepInfoPtr.OriginalVersion,
		LatestVersion:   stepInfoPtr.LatestVersion,
		GroupInfo:       stepInfoPtr.GroupInfo,
		Step:            stepInfoPtr.Step,
		DefinitionPth:   stepInfoPtr.DefinitionPth,
	}

	isExitStatusError := true
	if err != nil {
		isExitStatusError = errorutil.IsExitStatusError(err)
	}

	// Print step preparation errors before the step header box,
	// other errors are printed within the step box.
	if status == models.StepRunStatusCodePreparationFailed && err != nil {
		if !isExitStatusError {
			log.Errorf("Preparing Step (%s) failed: %s", pointers.StringWithDefault(stepInfoCopy.Step.Title, "missing title"), err)
		}
	}
	if printStepHeader {
		logStepStarted(stepInfoPtr, step, stepIdxPtr, stepExecutionId, stepStartTime)
	}

	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	stepResults := models.StepRunResultsModel{
		StepInfo:   stepInfoCopy,
		StepInputs: redactedStepInputs,
		Status:     status,
		Idx:        buildRunResults.ResultsCount(),
		RunTime:    time.Since(stepStartTime),
		ErrorStr:   errStr,
		ExitCode:   exitCode,
		StartTime:  stepStartTime,
	}

	r.tracker.SendStepFinishedEvent(properties, analytics.StepResult{
		Info:            prepareAnalyticsStepInfo(step, stepInfoPtr),
		Status:          status,
		ErrorMessage:    errStr,
		Timeout:         timeout,
		NoOutputTimeout: noOutputTimeout,
	})

	switch status {
	case models.StepRunStatusCodeSuccess:
		buildRunResults.SuccessSteps = append(buildRunResults.SuccessSteps, stepResults)
	case models.StepRunStatusCodePreparationFailed:
		buildRunResults.FailedSteps = append(buildRunResults.FailedSteps, stepResults)
	case models.StepRunStatusCodeFailed:
		if !isExitStatusError {
			log.Errorf("Step (%s) failed: %s", pointers.StringWithDefault(stepInfoCopy.Step.Title, "missing title"), err)
		}
		buildRunResults.FailedSteps = append(buildRunResults.FailedSteps, stepResults)
	case models.StepRunStatusCodeFailedSkippable:
		if !isExitStatusError {
			log.Warnf("Step (%s) failed, but was marked as skippable: %s", pointers.StringWithDefault(stepInfoCopy.Step.Title, "missing title"), err)
		} else {
			log.Warnf("Step (%s) failed, but was marked as skippable", pointers.StringWithDefault(stepInfoCopy.Step.Title, "missing title"))
		}
		buildRunResults.FailedSkippableSteps = append(buildRunResults.FailedSkippableSteps, stepResults)
	case models.StepRunStatusAbortedWithCustomTimeout, models.StepRunStatusAbortedWithNoOutputTimeout:
		log.Errorf("Step (%s) aborted: %s", pointers.StringWithDefault(stepInfoCopy.Step.Title, "missing title"), err)
		buildRunResults.FailedSteps = append(buildRunResults.FailedSteps, stepResults)
	case models.StepRunStatusCodeSkipped:
		log.Warnf("A previous step failed, and this step (%s) was not marked as IsAlwaysRun, skipped", pointers.StringWithDefault(stepInfoCopy.Step.Title, "missing title"))
		buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
	case models.StepRunStatusCodeSkippedWithRunIf:
		log.Warn("The step's (" + pointers.StringWithDefault(stepInfoCopy.Step.Title, "missing title") + ") Run-If expression evaluated to false - skipping")
		if runIf != "" {
			log.Info("The Run-If expression was: ", colorstring.Blue(runIf))
		}
		buildRunResults.SkippedSteps = append(buildRunResults.SkippedSteps, stepResults)
	default:
		log.Error("Unknown result code")
		return
	}

	logStepFinished(stepResults, stepExecutionId, isLastStep)
}

func logStepFinished(stepResults models.StepRunResultsModel, stepExecutionId string, isLastStep bool) {
	params := stepFinishedParamsFromResults(stepResults, stepExecutionId, isLastStep)
	log.PrintStepFinishedEvent(params)
}

func stepFinishedParamsFromResults(results models.StepRunResultsModel, stepExecutionId string, isLastStep bool) log.StepFinishedParams {
	title := ""
	if results.StepInfo.Step.Title != nil {
		title = *results.StepInfo.Step.Title
	}

	supportURL := ""
	if results.StepInfo.Step.SupportURL != nil {
		supportURL = *results.StepInfo.Step.SupportURL
	}

	sourceURL := ""
	if results.StepInfo.Step.SourceCodeURL != nil {
		sourceURL = *results.StepInfo.Step.SourceCodeURL
	}

	var errors []log.StepError
	if results.ErrorStr != "" && results.Status.IsFailure() {
		errors = append(errors, log.StepError{
			Code:    results.ExitCode,
			Message: results.ErrorStr,
		})
	}

	var stepUpdate *log.StepUpdate
	updateAvailable, _ := utils.IsUpdateAvailable(results.StepInfo.Version, results.StepInfo.LatestVersion)
	if updateAvailable {
		stepUpdate = &log.StepUpdate{
			OriginalVersion: results.StepInfo.OriginalVersion,
			ResolvedVersion: results.StepInfo.Version,
			LatestVersion:   results.StepInfo.LatestVersion,
			ReleasesURL:     utils.RepoReleasesURL(sourceURL),
		}
	}

	var stepDeprecation *log.StepDeprecation
	if results.StepInfo.GroupInfo.RemovalDate != "" || results.StepInfo.GroupInfo.DeprecateNotes != "" {
		stepDeprecation = &log.StepDeprecation{
			RemovalDate: results.StepInfo.GroupInfo.RemovalDate,
			Note:        results.StepInfo.GroupInfo.DeprecateNotes,
		}
	}

	return log.StepFinishedParams{
		ExecutionId:    stepExecutionId,
		InternalStatus: int(results.Status),
		Status:         results.Status.HumanReadableStatus(),
		StatusReason:   results.Status.Reason(results),
		Title:          title,
		RunTime:        results.RunTime.Milliseconds(),
		SupportURL:     supportURL,
		SourceCodeURL:  sourceURL,
		Errors:         errors,
		Update:         stepUpdate,
		Deprecation:    stepDeprecation,
		LastStep:       isLastStep,
	}
}

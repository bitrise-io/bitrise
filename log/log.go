package log

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v2/log/corelog"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/colorstring"
)

const rfc3339MicroTimeLayout = "2006-01-02T15:04:05.999999Z07:00"

const consoleTimeLayout = "15:04:05"

type LoggerType corelog.LoggerType

const (
	JSONLogger    = LoggerType(corelog.JSONLogger)
	ConsoleLogger = LoggerType(corelog.ConsoleLogger)
)

type Producer corelog.Producer

const (
	BitriseCLI = Producer(corelog.BitriseCLI)
	Step       = Producer(corelog.Step)
)

// defaultLogger ...
type defaultLogger struct {
	opts   LoggerOpts
	logger corelog.Logger
}

type ConsoleLoggerOpts struct {
	Timestamp bool
}

type LoggerOpts struct {
	LoggerType        LoggerType
	Producer          Producer
	ProducerID        string
	ConsoleLoggerOpts ConsoleLoggerOpts
	DebugLogEnabled   bool
	Writer            io.Writer
	TimeProvider      func() time.Time
}

// NewLogger ...
func NewLogger(opts LoggerOpts) Logger {
	return newLogger(opts)
}

func newLogger(opts LoggerOpts) *defaultLogger {
	logger := corelog.NewLogger(corelog.LoggerType(opts.LoggerType), opts.Writer)
	return &defaultLogger{
		opts:   opts,
		logger: logger,
	}
}

// Error ...
func (m *defaultLogger) Error(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.ErrorLevel)
}

// Errorf ...
func (m *defaultLogger) Errorf(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.ErrorLevel)
}

// Warn ...
func (m *defaultLogger) Warn(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.WarnLevel)
}

// Warnf ...
func (m *defaultLogger) Warnf(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.WarnLevel)
}

// Info ...
func (m *defaultLogger) Info(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.InfoLevel)
}

// Infof ...
func (m *defaultLogger) Infof(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.InfoLevel)
}

// Done ...
func (m *defaultLogger) Done(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.DoneLevel)
}

// Donef ...
func (m *defaultLogger) Donef(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.DoneLevel)
}

// Print ...
func (m *defaultLogger) Print(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.NormalLevel)
}

// Printf ...
func (m *defaultLogger) Printf(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.NormalLevel)
}

// Debug ...
func (m *defaultLogger) Debug(args ...interface{}) {
	if !m.opts.DebugLogEnabled {
		return
	}
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.DebugLevel)
}

// Debugf ...
func (m *defaultLogger) Debugf(format string, args ...interface{}) {
	if !m.opts.DebugLogEnabled {
		return
	}
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.DebugLevel)
}

// LogMessage ...
func (m *defaultLogger) LogMessage(message string, level corelog.Level) {
	if level == corelog.DebugLevel && !m.opts.DebugLogEnabled {
		return
	}

	m.logMessage(message, level)
}

func (m *defaultLogger) PrintBitriseStartedEvent(plan models.WorkflowRunPlan) {
	if m.opts.LoggerType == JSONLogger {
		m.logger.LogEvent(plan, corelog.EventLogFields{
			Timestamp: m.opts.TimeProvider().Format(rfc3339MicroTimeLayout),
			EventType: "bitrise_started",
		})
	} else {
		m.Print()
		m.Printf("Invocation started at %s", colorstring.Cyan("%s", m.opts.TimeProvider().Format(consoleTimeLayout)))
		m.Printf("Bitrise CLI version: %s", colorstring.Cyan("%s", plan.Version))
		m.Print()
		m.Infof("Run modes:")
		m.Printf("CI mode: %v", colorstring.Cyan("%v", plan.CIMode))
		m.Printf("PR mode: %v", colorstring.Cyan("%v", plan.PRMode))
		m.Printf("Debug mode: %v", colorstring.Cyan("%v", plan.DebugMode))
		m.Printf("Secret filtering mode: %v", colorstring.Cyan("%v", plan.SecretFilteringMode))
		m.Printf("Secret Envs filtering mode: %v", colorstring.Cyan("%v", plan.SecretEnvsFilteringMode))
		m.Printf("Using Step library in offline mode: %v", colorstring.Cyan("%v", plan.IsSteplibOfflineMode))
		m.Printf("No output timeout mode: %v", colorstring.Cyan("%v", plan.NoOutputTimeoutMode))
		m.Print()
		var workflowIDs []string
		for _, workflowPlan := range plan.ExecutionPlan {
			workflowID := workflowPlan.WorkflowID
			workflowIDs = append(workflowIDs, workflowID)
		}
		var prefix string
		if len(workflowIDs) == 1 {
			prefix = "Running workflow"
		} else {
			prefix = "Running workflows"
		}

		m.Printf("%s: %s", prefix, colorstring.Cyan("%s", strings.Join(workflowIDs, " → ")))
	}
}

func (m *defaultLogger) PrintStepStartedEvent(params StepStartedParams) {
	if m.opts.LoggerType == JSONLogger {
		m.logger.LogEvent(params, corelog.EventLogFields{
			Timestamp: m.opts.TimeProvider().Format(rfc3339MicroTimeLayout),
			EventType: "step_started",
		})
	} else {
		lines := generateStepStartedHeaderLines(params)
		for _, line := range lines {
			m.Print(line)
		}
	}
}

func (m *defaultLogger) PrintStepFinishedEvent(params StepFinishedParams) {
	if m.opts.LoggerType == JSONLogger {
		m.logger.LogEvent(params, corelog.EventLogFields{
			Timestamp: m.opts.TimeProvider().Format(rfc3339MicroTimeLayout),
			EventType: "step_finished",
		})
	} else {
		lines := generateStepFinishedFooterLines(params)
		for _, line := range lines {
			m.Print(line)
		}
	}
}

func (m *defaultLogger) logMessage(message string, level corelog.Level) {
	fields := m.createMessageFields(level)
	m.logger.LogMessage(message, corelog.MessageLogFields(fields))
}

func (m *defaultLogger) createMessageFields(level corelog.Level) MessageFields {
	if m.opts.LoggerType == JSONLogger {
		return createJSONLogMessageFields(m.opts.Producer, m.opts.ProducerID, level, m.opts.TimeProvider)
	}

	var tProvider func() time.Time
	if m.opts.ConsoleLoggerOpts.Timestamp {
		tProvider = m.opts.TimeProvider
	}
	return createConsoleLogMessageFields(level, tProvider)
}

func createJSONLogMessageFields(producer Producer, producerID string, level corelog.Level, timeProvider func() time.Time) MessageFields {
	return MessageFields{
		Timestamp:  timeProvider().Format(rfc3339MicroTimeLayout),
		Producer:   corelog.Producer(producer),
		ProducerID: producerID,
		Level:      level,
	}
}

func createConsoleLogMessageFields(level corelog.Level, timeProvider func() time.Time) MessageFields {
	fields := MessageFields{
		Level: level,
	}
	if timeProvider != nil {
		fields.Timestamp = timeProvider().Format(consoleTimeLayout)
	}
	return fields
}

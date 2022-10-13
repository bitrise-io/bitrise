package log

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/log/corelog"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/colorstring"
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
		m.PrintBitriseASCIIArt()
		m.Warnf("CI mode: %v", plan.CIMode)
		m.Warnf("PR mode: %v", plan.PRMode)
		m.Warnf("Debug mode: %v", plan.DebugMode)
		m.Warnf("Secret filtering mode: %v", plan.SecretFilteringMode)
		m.Warnf("Secret Envs filtering mode: %v", plan.SecretEnvsFilteringMode)
		m.Warnf("No output timeout mode: %v", plan.NoOutputTimeoutMode)
		m.Print()
		var workflowIDs []string
		for _, workflowPlan := range plan.ExecutionPlan {
			workflowID := workflowPlan.WorkflowID
			if workflowPlan.WorkflowID == plan.TargetWorkflowID {
				workflowID = colorstring.Green(workflowPlan.WorkflowID)
			}
			workflowIDs = append(workflowIDs, workflowID)
		}
		var prefix string
		if len(workflowIDs) == 1 {
			prefix = colorstring.Blue("Running workflow")
		} else {
			prefix = colorstring.Blue("Running workflows")
		}

		m.Printf("%s: %s", prefix, strings.Join(workflowIDs, " -->  "))
	}
}

// PrintBitriseASCIIArt ...
func (m *defaultLogger) PrintBitriseASCIIArt() {
	m.Print(`
██████╗ ██╗████████╗██████╗ ██╗███████╗███████╗
██╔══██╗██║╚══██╔══╝██╔══██╗██║██╔════╝██╔════╝
██████╔╝██║   ██║   ██████╔╝██║███████╗█████╗
██╔══██╗██║   ██║   ██╔══██╗██║╚════██║██╔══╝
██████╔╝██║   ██║   ██║  ██║██║███████║███████╗
╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚══════╝╚══════╝`)
	m.Infof("version: %s", colorstring.Green(version.VERSION))
	m.Print()
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

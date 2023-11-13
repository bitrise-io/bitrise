package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/stretchr/testify/require"
)

const (
	debugLogMessage = "This is a debug message"
)

func TestStepDebugLogMessagesAppear(t *testing.T) {
	consoleLog, consoleErr := createConsoleLog(t, "debug_log")
	require.NoError(t, consoleErr)
	require.True(t, strings.Contains(consoleLog, debugLogMessage))

	jsonLog, jsonErr := createJSONLog(t, "debug_log")
	require.NoError(t, jsonErr)
	require.True(t, strings.Contains(string(jsonLog), debugLogMessage))
}

func TestConsoleLogCanBeRestoredFromJSONLog(t *testing.T) {
	consoleLog, consoleErr := createConsoleLog(t, "fail_test")
	require.EqualError(t, consoleErr, "exit status 1")

	jsonLog, jsonErr := createJSONLog(t, "fail_test")
	require.EqualError(t, jsonErr, "exit status 1")

	convertedConsoleLog := restoreConsoleLog(t, jsonLog)
	require.Equal(t, replaceVariableParts(consoleLog), replaceVariableParts(convertedConsoleLog))
}

func createConsoleLog(t *testing.T, workflow string) (string, error) {
	{
		cmd := exec.Command(binPath(), "setup")
		outBytes, err := cmd.CombinedOutput()
		require.NoError(t, err, string(outBytes))
	}

	{
		cmd := exec.Command(binPath(), ":analytics", "off")
		outBytes, err := cmd.CombinedOutput()
		require.NoError(t, err, string(outBytes))
	}

	cmd := exec.Command(binPath(), "run", workflow, "--config", "log_format_test_bitrise.yml")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func createJSONLog(t *testing.T, workflow string) ([]byte, error) {
	{
		cmd := exec.Command(binPath(), "setup")
		outBytes, err := cmd.CombinedOutput()
		require.NoError(t, err, string(outBytes))
	}

	{
		cmd := exec.Command(binPath(), ":analytics", "off")
		outBytes, err := cmd.CombinedOutput()
		require.NoError(t, err, string(outBytes))
	}

	cmd := exec.Command(binPath(), "run", workflow, "--config", "log_format_test_bitrise.yml", "--output-format", "json")
	return cmd.CombinedOutput()
}

func restoreConsoleLog(t *testing.T, log []byte) string {
	var consoleLog string
	lines := bytes.Split(log, []byte("\n"))
	for _, line := range lines {
		if string(line) == "" {
			continue
		}

		msg, err := convertMessageLog(line)
		if err != nil {
			msg, err = convertEventLog(line)
			if err != nil {
				t.Fatalf("log can't be parsed as message log nor as event log: %s", string(line))
			}
		}

		consoleLog += msg
	}
	return consoleLog
}

func convertEventLog(line []byte) (string, error) {
	logLine, err := convertBitriseStartedEventLog(line)
	if err == nil {
		return logLine, nil
	}

	logLine, err = convertStepStartedEventLog(line)
	if err == nil {
		return logLine, nil
	}

	logLine, err = convertStepFinishedEventLog(line)
	if err == nil {
		return logLine, nil
	}

	return "", fmt.Errorf("unknown event log")
}

func convertBitriseStartedEventLog(line []byte) (string, error) {
	type EventLog struct {
		Timestamp   string                 `json:"timestamp"`
		MessageType string                 `json:"type"`
		EventType   string                 `json:"event_type"`
		Content     models.WorkflowRunPlan `json:"content"`
	}

	var eventLog EventLog
	err := json.Unmarshal(line, &eventLog)
	if err != nil {
		return "", err
	}

	if eventLog.Content.LogFormatVersion == "" {
		return "", fmt.Errorf("invalid message log")
	}

	var buf bytes.Buffer
	logger := log.NewLogger(log.LoggerOpts{LoggerType: log.ConsoleLogger, Writer: &buf, TimeProvider: time.Now})
	logger.PrintBitriseStartedEvent(eventLog.Content)

	return buf.String(), nil
}

func convertStepStartedEventLog(line []byte) (string, error) {
	if !strings.Contains(string(line), "\"event_type\":\"step_started\"") {
		return "", fmt.Errorf("not a step started event")
	}

	type EventLog struct {
		Timestamp   string                `json:"timestamp"`
		MessageType string                `json:"type"`
		EventType   string                `json:"event_type"`
		Content     log.StepStartedParams `json:"content"`
	}

	var eventLog EventLog
	err := json.Unmarshal(line, &eventLog)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	logger := log.NewLogger(log.LoggerOpts{LoggerType: log.ConsoleLogger, Writer: &buf})
	logger.PrintStepStartedEvent(eventLog.Content)

	return buf.String(), nil
}

func convertStepFinishedEventLog(line []byte) (string, error) {
	if !strings.Contains(string(line), "\"event_type\":\"step_finished\"") {
		return "", fmt.Errorf("not a step finished event")
	}

	type EventLog struct {
		Timestamp   string                 `json:"timestamp"`
		MessageType string                 `json:"type"`
		EventType   string                 `json:"event_type"`
		Content     log.StepFinishedParams `json:"content"`
	}

	var eventLog EventLog
	err := json.Unmarshal(line, &eventLog)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	logger := log.NewLogger(log.LoggerOpts{LoggerType: log.ConsoleLogger, Writer: &buf})
	logger.PrintStepFinishedEvent(eventLog.Content)

	return buf.String(), nil
}

func convertMessageLog(line []byte) (string, error) {
	type MessageLog struct {
		Message string `json:"message"`
		Level   string `json:"level"`
	}

	var messageLog MessageLog
	err := json.Unmarshal(line, &messageLog)
	if err != nil {
		return "", err
	}

	if messageLog.Level == "" {
		return "", fmt.Errorf("invalid message log")
	}

	return createLogMsg(messageLog.Level, messageLog.Message), nil
}

var levelToANSIColorCode = map[level]ansiColorCode{
	errorLevel: redCode,
	warnLevel:  yellowCode,
	infoLevel:  blueCode,
	doneLevel:  greenCode,
	debugLevel: magentaCode,
}

func createLogMsg(lvl string, message string) string {
	color := levelToANSIColorCode[level(lvl)]
	if color != "" {
		return addColor(color, message)
	}
	return message
}

func addColor(color ansiColorCode, msg string) string {
	return string(color) + msg + string(resetCode)
}

type level string

const (
	errorLevel  level = "error"
	warnLevel   level = "warn"
	infoLevel   level = "info"
	doneLevel   level = "done"
	normalLevel level = "normal"
	debugLevel  level = "debug"
)

type ansiColorCode string

const (
	redCode     ansiColorCode = "\x1b[31;1m"
	yellowCode  ansiColorCode = "\x1b[33;1m"
	blueCode    ansiColorCode = "\x1b[34;1m"
	greenCode   ansiColorCode = "\x1b[32;1m"
	magentaCode ansiColorCode = "\x1b[35;1m"
	resetCode   ansiColorCode = "\x1b[0m"
)

func replaceVariableParts(line string) string {
	timeRegexp := regexp.MustCompile(`(\| time: .+\|)`)
	line = timeRegexp.ReplaceAllString(line, "[REPLACED]")

	runTimeRegexp := regexp.MustCompile(`(\| .+ sec \|)`)
	line = runTimeRegexp.ReplaceAllString(line, "[REPLACED]")

	totalRunTimeRegexp := regexp.MustCompile(`(\| Total runtime: .+ \|)`)
	line = totalRunTimeRegexp.ReplaceAllString(line, "[REPLACED]")

	return line
}

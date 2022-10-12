package integration

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsoleLogCanBeRestoredFromJSONLog(t *testing.T) {
	consoleLog := createConsoleLog(t)
	jsonLog := createJSONleLog(t)
	convertedConsoleLog := restoreConsoleLog(t, jsonLog)
	require.Equal(t, replaceVariableParts(consoleLog), replaceVariableParts(convertedConsoleLog))
}

func createConsoleLog(t *testing.T) string {
	execCmd := exec.Command(binPath(), "setup")
	outBytes, err := execCmd.CombinedOutput()
	require.NoError(t, err, string(outBytes))

	cmd := exec.Command(binPath(), "run", "fail_test", "--config", "log_format_test_bitrise.yml")
	out, err := cmd.CombinedOutput()
	require.EqualError(t, err, "exit status 1")
	return string(out)
}

func createJSONleLog(t *testing.T) []byte {
	execCmd := exec.Command(binPath(), "setup")
	outBytes, err := execCmd.CombinedOutput()
	require.NoError(t, err, string(outBytes))

	cmd := exec.Command(binPath(), "run", "fail_test", "--config", "log_format_test_bitrise.yml", "--output-format", "json")
	out, err := cmd.CombinedOutput()
	require.EqualError(t, err, "exit status 1")
	return out
}

func restoreConsoleLog(t *testing.T, log []byte) string {
	type Log struct {
		Message string `json:"message"`
		Level   string `json:"level"`
	}

	var consoleLog string
	lines := bytes.Split(log, []byte("\n"))
	for _, line := range lines {
		if string(line) == "" {
			continue
		}
		var log Log
		err := json.Unmarshal(line, &log)
		require.NoError(t, err, string(line))
		consoleLog += createLogMsg(log.Level, log.Message)
	}
	return consoleLog
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

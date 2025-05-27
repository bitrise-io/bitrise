package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	// Define flags
	var stepID string
	flag.StringVar(&stepID, "step-id", "script", "The ID of the step that is executing the command")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] -- <command> [args...]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	// Find the index of the "--" separator
	separatorIndex := -1
	for i, arg := range os.Args {
		if arg == "--" {
			separatorIndex = i
			break
		}
	}

	// Check if separator was found
	if separatorIndex == -1 || separatorIndex == len(os.Args)-1 {
		flag.Usage()
		os.Exit(1)
	}

	// Parse flags from arguments before the "--" separator
	flag.CommandLine.Parse(os.Args[1:separatorIndex])

	// Get the command and its arguments after the "--" separator
	cmdName := os.Args[separatorIndex+1]
	cmdArgs := []string{}
	if len(os.Args) > separatorIndex+2 {
		cmdArgs = os.Args[separatorIndex+2:]
	}

	// Generate command string for the trace log
	cmdString := cmdName
	if len(cmdArgs) > 0 {
		cmdString += " " + strings.Join(cmdArgs, " ")
	}

	// Current timestamp in microseconds for starting the command
	startTime := time.Now()
	startTimeUnixMicro := startTime.UnixMicro()

	// Get process ID
	pid := os.Getpid()

	// Print command start trace log
	fmt.Printf("BITRISE_TRACE:{\"ts\":%d,\"type\":\"command_start\",\"step_id\":\"%s\",\"command\":\"%s\",\"pid\":%d,\"tid\":1}\n",
		startTimeUnixMicro, stepID, cmdString, pid)

	// Prepare the command to run
	cmd := exec.Command(cmdName, cmdArgs...)

	// Set the command to use the current process's stdin, stdout, and stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()

	// Current timestamp in microseconds for ending the command
	endTime := time.Now()
	endTimeUnixMicro := endTime.UnixMicro()

	// Calculate duration in microseconds
	durationMicro := endTimeUnixMicro - startTimeUnixMicro

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	// Print command end trace log
	fmt.Printf("BITRISE_TRACE:{\"ts\":%d,\"type\":\"command_end\",\"step_id\":\"%s\",\"command\":\"%s\",\"exit_code\":%d,\"duration_us\":%d,\"pid\":%d,\"tid\":1}\n",
		endTimeUnixMicro, stepID, cmdString, exitCode, durationMicro, pid)

	// Exit with the same code as the command
	if err != nil {
		os.Exit(exitCode)
	}
}

package tools

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/tools/asynccmd"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// UnameGOOS ...
func UnameGOOS() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "Darwin", nil
	case "linux":
		return "Linux", nil
	}
	return "", fmt.Errorf("Unsupported platform (%s)", runtime.GOOS)
}

// UnameGOARCH ...
func UnameGOARCH() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64", nil
	}
	return "", fmt.Errorf("Unsupported architecture (%s)", runtime.GOARCH)
}

// InstallToolFromGitHub ...
func InstallToolFromGitHub(toolname, githubUser, toolVersion string) error {
	unameGOOS, err := UnameGOOS()
	if err != nil {
		return fmt.Errorf("Failed to determine OS: %s", err)
	}
	unameGOARCH, err := UnameGOARCH()
	if err != nil {
		return fmt.Errorf("Failed to determine ARCH: %s", err)
	}
	downloadURL := "https://github.com/" + githubUser + "/" + toolname + "/releases/download/" + toolVersion + "/" + toolname + "-" + unameGOOS + "-" + unameGOARCH

	return InstallFromURL(toolname, downloadURL)
}

// DownloadFile ...
func DownloadFile(downloadURL, targetDirPath string) error {
	outFile, err := os.Create(targetDirPath)
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Warnf("Failed to close (%s)", targetDirPath)
		}
	}()
	if err != nil {
		return fmt.Errorf("failed to create (%s), error: %s", targetDirPath, err)
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download from (%s), error: %s", downloadURL, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf("failed to close (%s) body", downloadURL)
		}
	}()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to download from (%s), error: %s", downloadURL, err)
	}

	return nil
}

// InstallFromURL ...
func InstallFromURL(toolBinName, downloadURL string) error {
	if len(toolBinName) < 1 {
		return fmt.Errorf("no Tool (bin) Name provided! URL was: %s", downloadURL)
	}

	tmpDir, err := pathutil.NormalizedOSTempDirPath("__tmp_download_dest__")
	if err != nil {
		return fmt.Errorf("failed to create tmp dir for download destination")
	}
	tmpDestinationPth := filepath.Join(tmpDir, toolBinName)

	if err := DownloadFile(downloadURL, tmpDestinationPth); err != nil {
		return fmt.Errorf("failed to download, error: %s", err)
	}

	bitriseToolsDirPath := configs.GetBitriseToolsDirPath()
	destinationPth := filepath.Join(bitriseToolsDirPath, toolBinName)

	if exist, err := pathutil.IsPathExists(destinationPth); err != nil {
		return fmt.Errorf("failed to check if file exist (%s), error: %s", destinationPth, err)
	} else if exist {
		if err := os.Remove(destinationPth); err != nil {
			return fmt.Errorf("failed to remove file (%s), error: %s", destinationPth, err)
		}
	}

	if err := MoveFile(tmpDestinationPth, destinationPth); err != nil {
		return fmt.Errorf("failed to copy (%s) to (%s), error: %s", tmpDestinationPth, destinationPth, err)
	}

	if err := os.Chmod(destinationPth, 0755); err != nil {
		return fmt.Errorf("failed to make file (%s) executable, error: %s", destinationPth, err)
	}

	return nil
}

// ------------------
// --- Stepman

// StepmanSetup ...
func StepmanSetup(collection string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "setup", "--collection", collection}
	return command.RunCommand("stepman", args...)
}

// StepmanActivate ...
func StepmanActivate(collection, stepID, stepVersion, dir, ymlPth string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "activate", "--collection", collection,
		"--id", stepID, "--version", stepVersion, "--path", dir, "--copyyml", ymlPth}
	return command.RunCommand("stepman", args...)
}

// StepmanUpdate ...
func StepmanUpdate(collection string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "update", "--collection", collection}
	return command.RunCommand("stepman", args...)
}

// StepmanRawStepLibStepInfo ...
func StepmanRawStepLibStepInfo(collection, stepID, stepVersion string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "step-info", "--collection", collection,
		"--id", stepID, "--version", stepVersion, "--format", "raw"}
	return command.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
}

// StepmanRawLocalStepInfo ...
func StepmanRawLocalStepInfo(pth string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "step-info", "--step-yml", pth, "--format", "raw"}
	return command.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
}

// StepmanJSONStepLibStepInfo ...
func StepmanJSONStepLibStepInfo(collection, stepID, stepVersion string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "step-info", "--collection", collection,
		"--id", stepID, "--version", stepVersion, "--format", "json"}

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	if err := command.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), "stepman", args...); err != nil {
		return outBuffer.String(), fmt.Errorf("Error: %s, details: %s", err, errBuffer.String())
	}

	return outBuffer.String(), nil
}

// StepmanJSONLocalStepInfo ...
func StepmanJSONLocalStepInfo(pth string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "step-info", "--step-yml", pth, "--format", "json"}

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	if err := command.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), "stepman", args...); err != nil {
		return outBuffer.String(), fmt.Errorf("Error: %s, details: %s", err, errBuffer.String())
	}

	return outBuffer.String(), nil
}

// StepmanRawStepList ...
func StepmanRawStepList(collection string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "step-list", "--collection", collection, "--format", "raw"}
	return command.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
}

// StepmanJSONStepList ...
func StepmanJSONStepList(collection string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "step-list", "--collection", collection, "--format", "json"}

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	if err := command.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), "stepman", args...); err != nil {
		return outBuffer.String(), fmt.Errorf("Error: %s, details: %s", err, errBuffer.String())
	}

	return outBuffer.String(), nil
}

//
// Share

// StepmanShare ...
func StepmanShare() error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "--toolmode"}
	return command.RunCommand("stepman", args...)
}

// StepmanShareAudit ...
func StepmanShareAudit() error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "audit", "--toolmode"}
	return command.RunCommand("stepman", args...)
}

// StepmanShareCreate ...
func StepmanShareCreate(tag, git, stepID string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "create", "--tag", tag, "--git", git, "--stepid", stepID, "--toolmode"}
	return command.RunCommand("stepman", args...)
}

// StepmanShareFinish ...
func StepmanShareFinish() error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "finish", "--toolmode"}
	return command.RunCommand("stepman", args...)
}

// StepmanShareStart ...
func StepmanShareStart(collection string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "start", "--collection", collection, "--toolmode"}
	return command.RunCommand("stepman", args...)
}

// ------------------
// --- Envman

// EnvmanInit ...
func EnvmanInit() error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "init"}
	return command.RunCommand("envman", args...)
}

// EnvmanInitAtPath ...
func EnvmanInitAtPath(envstorePth string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "init", "--clear"}
	return command.RunCommand("envman", args...)
}

// EnvmanAdd ...
func EnvmanAdd(envstorePth, key, value string, expand, skipIfEmpty bool) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "add", "--key", key, "--append"}
	if !expand {
		args = append(args, "--no-expand")
	}
	if skipIfEmpty {
		args = append(args, "--skip-if-empty")
	}

	envman := exec.Command("envman", args...)
	envman.Stdin = strings.NewReader(value)
	envman.Stdout = os.Stdout
	envman.Stderr = os.Stderr
	return envman.Run()
}

// ExportEnvironmentsList ...
func ExportEnvironmentsList(envstorePth string, envsList []envmanModels.EnvironmentItemModel) error {
	for _, env := range envsList {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return err
		}

		opts, err := env.GetOptions()
		if err != nil {
			return err
		}

		isExpand := envmanModels.DefaultIsExpand
		if opts.IsExpand != nil {
			isExpand = *opts.IsExpand
		}

		skipIfEmpty := envmanModels.DefaultSkipIfEmpty
		if opts.SkipIfEmpty != nil {
			skipIfEmpty = *opts.SkipIfEmpty
		}

		if err := EnvmanAdd(envstorePth, key, value, isExpand, skipIfEmpty); err != nil {
			return err
		}
	}
	return nil
}

// EnvmanClear ...
func EnvmanClear(envstorePth string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "clear"}
	out, err := command.New("envman", args...).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		errorMsg := err.Error()
		if errorutil.IsExitStatusError(err) && out != "" {
			errorMsg = out
		}
		return fmt.Errorf("failed to clear envstore (%s), error: %s", envstorePth, errorMsg)
	}
	return nil
}

// EnvmanRun runs a command through envman.
func EnvmanRun(envstorePth, workDirPth string, cmdArgs []string, timeout time.Duration, secrets []envmanModels.EnvironmentItemModel) (int, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "run"}
	args = append(args, cmdArgs...)

	if !configs.IsSecretFiltering {
		command := command.NewWithStandardOuts("envman", args...).SetStdin(os.Stdin).SetDir(workDirPth)

		if timeout <= 0 {
			exitCode, err := command.RunAndReturnExitCode()
			return exitCode, errors.WithStack(err)
		}

		// create a new process group for our process and its child processes
		cmd := command.GetCmd()
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		if err := cmd.Start(); err != nil {
			return -1, errors.WithStack(err)
		}

		// Setpgid: true creates a new process group for cmd and its subprocesses
		// this way cmd will not belong to its parent process group,
		// cmd will not be killed when you hit ^C in your terminal
		// to fix this, we listen and handle Interrupt signal manually
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
				log.Warnf("Failed to kill process, error: %+v", errors.WithStack(err))
				os.Exit(130)
			}
		}()
		defer func() {
			signal.Stop(c)
		}()
		//

		timer := time.AfterFunc(timeout, func() {
			if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
				log.Warnf("Failed to kill process, error: %+v", errors.WithStack(err))
				os.Exit(130)
			}
		})

		err := cmd.Wait()

		timer.Stop()

		exitCode := 0
		if err != nil {
			exitCode = 1
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					exitCode = status.ExitStatus()
				}
			}
			if err.Error() == "signal: killed" {
				err = errors.New("timeout")
			}
		}

		return exitCode, errors.WithStack(err)
	}

	log.Warnf("Secret filtering enabled")

	cmd := asynccmd.New("envman", args...)
	cmd.SetDir(workDirPth)
	cmd.SetTimeout(timeout)

	var secretValues []string
	for _, secret := range secrets {
		_, value, err := secret.GetKeyValuePair()
		if err != nil || len(value) < 1 { // secret to redact needs to be at least 1 char legth
			continue
		}
		secretValues = append(secretValues, value)
	}
	cmd.SetSecrets(secretValues)

	statusChan, logChan := cmd.Start()
	for line := range logChan {
		fmt.Println(line)
	}
	status := <-statusChan
	return status.Code, status.Err
}

// EnvmanJSONPrint ...
func EnvmanJSONPrint(envstorePth string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "print", "--format", "json", "--expand"}

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	if err := command.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), "envman", args...); err != nil {
		return outBuffer.String(), fmt.Errorf("Error: %s, details: %s", err, errBuffer.String())
	}

	return outBuffer.String(), nil
}

// MoveFile ...
func MoveFile(oldpath, newpath string) error {
	err := os.Rename(oldpath, newpath)
	if err == nil {
		return nil
	}

	if linkErr, ok := err.(*os.LinkError); ok {
		if linkErr.Err == unix.EXDEV {
			info, err := os.Stat(oldpath)
			if err != nil {
				return err
			}

			data, err := ioutil.ReadFile(oldpath)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(newpath, data, info.Mode())
			if err != nil {
				return err
			}

			return os.Remove(oldpath)
		}
	}

	return err
}

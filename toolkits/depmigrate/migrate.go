package depmigrate

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
)

type GoModMigrator struct {
	projectDir string
}

func NewGoModMigrator(projectDir string) (*GoModMigrator, error) {
	absPath, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project path: %v", err)
	}
	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("not a directory (%s)", absPath)
	}

	return &GoModMigrator{projectDir: projectDir}, nil
}

func (m GoModMigrator) IsGoPathModeStep() bool {
	goModPath := filepath.Join(m.projectDir, "go.mod")
	_, err := os.Stat(goModPath)

	return err != nil
}

func (m GoModMigrator) Migrate(goBinaryPath, goRoot, packageName string) error {
	if packageName == "" {
		return fmt.Errorf("package name not specified")
	}

	cmds := []*command.Model{
		command.New(goBinaryPath, "mod", "init", packageName),
		command.New(goBinaryPath, "mod", "tidy"),
		command.New(goBinaryPath, "mod", "vendor"),
	}

	if err := os.RemoveAll(filepath.Join(m.projectDir, "vendor")); err != nil {
		return fmt.Errorf("failed to remove vendor directory: %v", err)
	}

	for _, cmd := range cmds {
		cmd.SetDir(m.projectDir)
		cmd.AppendEnvs("GOROOT=" + goRoot)

		goModPath := filepath.Join(m.projectDir, "go.mod")
		if _, err := os.Stat(goModPath); err != nil {
			log.Debugf("go.mod does not exists: %s", err)
		} else {
			contents, err := ioutil.ReadFile(goModPath)
			if err != nil {
				return err
			}
			log.Debugf("go.mod exist at %s, contents: %s", goModPath, contents)
		}

		log.Debugf("$ %s", cmd.PrintableCommandArgs())
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			if errorutil.IsExitStatusError(err) {
				return fmt.Errorf("command `%s` failed, output: %s", cmd.PrintableCommandArgs(), out)
			}

			return fmt.Errorf("failed to run command `%s`: %v", cmd.PrintableCommandArgs(), err)
		}
	}

	return nil
}

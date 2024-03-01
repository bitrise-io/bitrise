package toolkits

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
)

func getGoEnv(cmdRunner commandRunner, goBinaryPath string, envKey string) (string, error) {
	envCmd := command.New(goBinaryPath, "env", "-json", envKey)

	outputData, err := cmdRunner.runForOutput(envCmd)
	if err != nil {
		return "", err
	}

	goEnvs := make(map[string]string)
	if err := json.Unmarshal([]byte(outputData), &goEnvs); err != nil {
		return "", fmt.Errorf("failed to unmarshall go env output (%s): %v", outputData, err)
	}

	if _, ok := goEnvs[envKey]; !ok {
		return "", nil
	}

	return goEnvs[envKey], nil
}

func isGoPathModeStep(projectDir string) bool {
	goModPath := filepath.Join(projectDir, "go.mod")
	_, err := os.Stat(goModPath)

	return err != nil
}

func migrateToGoModules(stepAbsDirPath, packageName string) error {
	if packageName == "" {
		return fmt.Errorf("package name not specified")
	}

	// The go directive (https://golang.org/ref/mod#go-mod-file-go) sets go CLI and language features
	// Setting it to a fixed version so a future backward incompatible change does code
	goModTemplate := `module %s
go 1.16`
	goModContents := fmt.Sprintf(goModTemplate, packageName)
	if err := ioutil.WriteFile(filepath.Join(stepAbsDirPath, "go.mod"), []byte(goModContents), 0600); err != nil {
		return fmt.Errorf("failed to write go.mod file: %v", err)
	}

	return nil
}
